package handler

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/gofiber/fiber/v2"

	"github.com/strelov1/freehire/internal/ai/assistant"
	"github.com/strelov1/freehire/internal/ai/autofillagent"
)

// The application tools a candidate's OWN agent uses over MCP to fill a form in the
// candidate's own browser, through the extension's side panel — the same browser-tool wire
// the in-app autofill drives (internal/ai/browsertools), with the candidate's agent in the
// planner's seat.
//
// Two things hold whatever the agent does:
//   - nothing is ever SUBMITTED. The extension writes values into fields and stops; the
//     candidate reads the form and presses Submit themselves (extension/AGENTS.md);
//   - a fill is addressed by frame AND form as read_form reported them, so a value meant for
//     the application cannot land in a job-alert signup on the same page.
//
// The unattended path — auto-apply, which does submit — stays where it was: OAuth-only,
// behind users.agents_may_apply (assistant_auto_apply_tools.go).

// browserCall runs one browser tool on the caller's own channel, with the same bound and the
// same two failure sentences as read_current_page.
func (h *assistantHandlers) browserCall(ctx context.Context, userID int64, tool string, args any) (json.RawMessage, error) {
	if h.browserTools == nil {
		return nil, errors.New(noBrowserAttached)
	}
	caller := h.browserTools.NewCaller(userID)
	defer caller.Close()
	callCtx, cancel := context.WithTimeout(ctx, readCurrentPageTimeout)
	defer cancel()
	raw, err := caller.Call(callCtx, tool, args)
	if err != nil {
		if callCtx.Err() != nil {
			return nil, fmt.Errorf("the browser did not answer in %s — the page may still be loading, or the panel lost its connection; ask the candidate to reload the page and try again", readCurrentPageTimeout)
		}
		return nil, fmt.Errorf("%w — %s", err, noBrowserAttached)
	}
	return raw, nil
}

func (h *assistantHandlers) mcpApplyTools() []assistant.Tool {
	return []assistant.Tool{
		mcpServed(h.readCurrentPageTool(), "Read my open page", assistant.EffectReads),
		h.mcpReadApplicationFormTool(),
		h.mcpFillApplicationFormTool(),
		h.mcpAutofillProfileTool(),
	}
}

func (h *assistantHandlers) mcpReadApplicationFormTool() assistant.Tool {
	return assistant.Tool{
		Name:  "read_application_form",
		Title: "Read the application form",
		Description: "The form fields on the page open in the candidate's browser (the freehire side panel must be " +
			"open on it): each field's label, type, current value, whether it is required, its options, and the " +
			"frame and form it sits in; plus the CV upload controls, which mark which form is the application.",
		Effect: assistant.EffectReads,
		Schema: map[string]any{"type": "object", "properties": map[string]any{}, "additionalProperties": false},
		Run: func(ctx context.Context, userID int64, raw json.RawMessage) (any, error) {
			if err := assistant.DecodeArgs(raw, &struct{}{}); err != nil {
				return nil, err
			}
			return h.browserCall(ctx, userID, "read_form", nil)
		},
	}
}

func (h *assistantHandlers) mcpFillApplicationFormTool() assistant.Tool {
	return assistant.Tool{
		Name:  "fill_application_form",
		Title: "Fill the application form",
		Description: "Write values into fields of the form open in the candidate's browser. Each fill names the " +
			"field's label and the frame and form read_application_form reported for it. Nothing is submitted: the " +
			"candidate reviews the form and presses Submit. The result says, per label, whether it was filled, not " +
			"found, or ambiguous (two matching fields — nothing is written then).",
		Effect: assistant.EffectOverwrites, // replaces what the fields held
		Schema: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"fills": map[string]any{
					"type": "array",
					"items": map[string]any{
						"type": "object",
						"properties": map[string]any{
							"label": map[string]any{"type": "string"},
							"value": map[string]any{"type": "string"},
							"frame": map[string]any{"type": "integer"},
							"form":  map[string]any{"type": "integer"},
						},
						"required":             []string{"label", "value", "frame", "form"},
						"additionalProperties": false,
					},
				},
			},
			"required":             []string{"fills"},
			"additionalProperties": false,
		},
		Run: func(ctx context.Context, userID int64, raw json.RawMessage) (any, error) {
			// autofillagent.Fill is the wire shape the extension's fill_simple reads, pinned
			// by the fixture extension/lib/tools/testdata/fill-simple-call.json.
			var in struct {
				Fills []autofillagent.Fill `json:"fills"`
			}
			if err := assistant.DecodeArgs(raw, &in); err != nil {
				return nil, err
			}
			if len(in.Fills) == 0 {
				return nil, fiber.NewError(fiber.StatusBadRequest, "fills is empty")
			}
			return h.browserCall(ctx, userID, "fill_simple", map[string]any{"fills": in.Fills})
		},
	}
}

func (h *assistantHandlers) mcpAutofillProfileTool() assistant.Tool {
	return assistant.Tool{
		Name:  "get_autofill_profile",
		Title: "Get my application details",
		Description: "The candidate's own details for application forms — name, email, phone, location, links — " +
			"as they keep them on the website. These are the values contact fields are filled with.",
		Effect: assistant.EffectReads,
		Schema: map[string]any{"type": "object", "properties": map[string]any{}, "additionalProperties": false},
		Run: func(ctx context.Context, userID int64, raw json.RawMessage) (any, error) {
			if err := assistant.DecodeArgs(raw, &struct{}{}); err != nil {
				return nil, err
			}
			if h.autofillProfiles == nil {
				return nil, fiber.NewError(fiber.StatusServiceUnavailable, "application details are not available on this server")
			}
			return h.autofillProfiles.Assemble(ctx, userID)
		},
	}
}
