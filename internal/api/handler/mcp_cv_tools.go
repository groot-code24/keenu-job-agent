package handler

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"maps"
	"slices"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/strelov1/freehire/internal/ai/assistant"
	"github.com/strelov1/freehire/internal/candidate/cv"
	"github.com/strelov1/freehire/internal/candidate/fitanalysis"
	"github.com/strelov1/freehire/internal/candidate/matchanalysis"
	"github.com/strelov1/freehire/internal/candidate/resume"
	"github.com/strelov1/freehire/internal/candidate/resumeextract"
)

// The CV-tailoring tools a candidate's OWN agent uses over MCP — Claude Code or Codex,
// signed in with the candidate's own subscription, on their own machine.
//
// In the website the tailoring agent is a model this process calls, inside a session it
// owns. Here the agent is outside: the server makes no model call at all, and its job is
// to hold the data and keep the rules. The rules are the same ones, enforced in the same
// service paths, because most of these tools ARE the in-app agent's, served with the CV
// named by argument instead of by session:
//
//   - a CV edit that states something about the candidate must cite banked evidence the
//     candidate asserted (cvedit's evidence gate), so the agent cannot write a claim nobody
//     made;
//   - what the agent banks on its own reading is agent_inferred, which no CV may print until
//     the candidate confirms it on the website — there is no transcript here to check the
//     candidate's words against, so nothing the agent banks can raise its own standing;
//   - a parse of the CV file the agent submits is kept only where the file's own text
//     supports it (resumeextract.KeepSupported), which is what entitles it to cv_import.

// mcpCVParseModel stamps a structured CV the candidate's own agent parsed, where the
// in-process extractor would stamp its model id.
const mcpCVParseModel = "candidate-agent"

// mcpCVTools is the tailoring tool set served on the signed-in MCP server.
func (h *assistantHandlers) mcpCVTools() []assistant.Tool {
	return []assistant.Tool{
		h.mcpReadMyCVFileTool(),
		h.mcpImportMyCVTool(),
		h.mcpListMyCVsTool(),
		h.mcpTailorCVForJobTool(),
		h.mcpTailoringContextTool(),
		h.mcpGetCVTool(),
		h.mcpEditCVTool(),
		h.mcpCVJobMatchTool(),
		h.mcpCVPageCountTool(),
		mcpServed(h.cvEvidenceFidelityTool(), "Check evidence fidelity", assistant.EffectReads),
		mcpServed(h.experienceSearchTool(), "Search experience bank", assistant.EffectReads),
		mcpServed(h.experienceGetTool(), "Get achievements", assistant.EffectReads),
		mcpServed(h.experienceEmploymentsTool(), "List roles and projects", assistant.EffectReads),
		h.mcpExperienceAddTool(),
	}
}

// mcpServed gives an in-app tool what the MCP server announces it with.
func mcpServed(t assistant.Tool, title string, effect assistant.Effect) assistant.Tool {
	t.Title, t.Effect = title, effect
	return t
}

// uploadedCVStore is resumeStore for a tool that cannot proceed without the uploaded file,
// answering a sentence the agent can relay instead of degrading.
func (h *assistantHandlers) uploadedCVStore() (*resume.Store, error) {
	store := h.resumeStore()
	if store == nil || !store.Enabled() {
		return nil, fiber.NewError(fiber.StatusServiceUnavailable, "CV storage is not configured on this server")
	}
	return store, nil
}

func (h *assistantHandlers) cvSurface() error {
	if h.cv == nil || h.cv.cvStore == nil {
		return fiber.NewError(fiber.StatusServiceUnavailable, "CVs are not available on this server")
	}
	return nil
}

// noCVFile is the answer when the candidate has not uploaded a CV: the one step the agent
// cannot take for them, since the upload is a file on their machine and a browser form.
var noCVFile = fiber.NewError(fiber.StatusConflict,
	"no CV file uploaded yet — the candidate uploads it on the website's profile page (PDF)")

func (h *assistantHandlers) mcpReadMyCVFileTool() assistant.Tool {
	return assistant.Tool{
		Name:  "read_my_cv_file",
		Title: "Read my CV file",
		Description: "The text of the CV file the candidate uploaded, as the server reads it, and when it was " +
			"uploaded. This is the text a submitted parse is checked against: a value kept by " +
			"import_my_cv must appear in it word for word (case, punctuation and line breaks aside).",
		Effect: assistant.EffectReads,
		Schema: map[string]any{"type": "object", "properties": map[string]any{}, "additionalProperties": false},
		Run: func(ctx context.Context, userID int64, raw json.RawMessage) (any, error) {
			if err := assistant.DecodeArgs(raw, &struct{}{}); err != nil {
				return nil, err
			}
			store, err := h.uploadedCVStore()
			if err != nil {
				return nil, err
			}
			text, uploadedAt, err := store.TextAndUploadedAt(ctx, userID)
			if errors.Is(err, resume.ErrNotStored) {
				return nil, noCVFile
			}
			if err != nil {
				return nil, err
			}
			return map[string]any{"uploaded_at": uploadedAt.UTC().Format(time.RFC3339), "text": text}, nil
		},
	}
}

// periodDateSchema is perioddate.PeriodDate on the wire.
var periodDateSchema = map[string]any{
	"type":        "object",
	"description": "A date as the CV prints it: the year, and the month (1-12) only when the CV gives one.",
	"properties": map[string]any{
		"year":  map[string]any{"type": "integer"},
		"month": map[string]any{"type": "integer"},
	},
	"required":             []string{"year"},
	"additionalProperties": false,
}

func stringList(description string) map[string]any {
	return map[string]any{"type": "array", "items": map[string]any{"type": "string"}, "description": description}
}

// structuredCVSchema is resumeextract.Structured on the wire.
var structuredCVSchema = map[string]any{
	"type":        "object",
	"description": "The CV restated in a fixed shape. Copy every value from the CV text as written; leave out what it does not say.",
	"properties": map[string]any{
		"full_name":      map[string]any{"type": "string"},
		"headline":       map[string]any{"type": "string", "description": "The line under the name, e.g. \"Senior Backend Engineer\"."},
		"location":       map[string]any{"type": "string"},
		"email":          map[string]any{"type": "string"},
		"phone":          map[string]any{"type": "string"},
		"summary":        map[string]any{"type": "string"},
		"total_years":    map[string]any{"type": "integer", "description": "Only when the CV itself states it (\"8 years of experience\")."},
		"links":          stringList("Profile and portfolio URLs as printed."),
		"languages":      stringList("Spoken languages."),
		"skills":         stringList("Skills as the CV lists them."),
		"certifications": stringList("Certifications as printed."),
		"experience": map[string]any{
			"type": "array",
			"items": map[string]any{
				"type": "object",
				"properties": map[string]any{
					"title":      map[string]any{"type": "string"},
					"company":    map[string]any{"type": "string"},
					"location":   map[string]any{"type": "string"},
					"start":      periodDateSchema,
					"end":        periodDateSchema,
					"current":    map[string]any{"type": "boolean", "description": "True only when the CV says the role continues (Present)."},
					"summary":    map[string]any{"type": "string"},
					"highlights": stringList("Each bullet under the role, copied whole."),
					"stack":      stringList("Technologies the CV names for this role."),
				},
				"additionalProperties": false,
			},
		},
		"education": map[string]any{
			"type": "array",
			"items": map[string]any{
				"type": "object",
				"properties": map[string]any{
					"degree":      map[string]any{"type": "string"},
					"institution": map[string]any{"type": "string"},
					"year":        periodDateSchema,
				},
				"additionalProperties": false,
			},
		},
		"projects": map[string]any{
			"type": "array",
			"items": map[string]any{
				"type": "object",
				"properties": map[string]any{
					"name":       map[string]any{"type": "string"},
					"link":       map[string]any{"type": "string"},
					"highlights": stringList("Each bullet under the project, copied whole."),
				},
				"additionalProperties": false,
			},
		},
	},
	"additionalProperties": false,
}

// mcpImportMyCVTool stores a parse of the uploaded CV made by the candidate's own agent: the
// in-process extractor's job, done outside, and checked on the way in. What the file's text
// supports is stored exactly as the extractor's answer would be — the structured CV the
// profile and every new CV are seeded from, and the work history banked as cv_import
// evidence. What it does not support is dropped and named back, so the agent can tell the
// candidate rather than quietly lose it.
func (h *assistantHandlers) mcpImportMyCVTool() assistant.Tool {
	return assistant.Tool{
		Name:  "import_my_cv",
		Title: "Import my CV",
		Description: "Store a parse of the candidate's uploaded CV file: their profile, and every role, project and " +
			"bullet in their experience bank, where the bullets become evidence a tailored CV may cite. Each value " +
			"is kept only if the uploaded file's text contains it word for word; the result lists every value " +
			"that was dropped and why. Importing again adds what is new and never removes banked experience.",
		Effect: assistant.EffectOverwrites, // replaces the stored structured CV for this upload
		Schema: map[string]any{
			"type":                 "object",
			"properties":           map[string]any{"cv": structuredCVSchema},
			"required":             []string{"cv"},
			"additionalProperties": false,
		},
		Run: func(ctx context.Context, userID int64, raw json.RawMessage) (any, error) {
			var in struct {
				CV resumeextract.Structured `json:"cv"`
			}
			if err := assistant.DecodeArgs(raw, &in); err != nil {
				return nil, err
			}
			store, err := h.uploadedCVStore()
			if err != nil {
				return nil, err
			}
			text, uploadedAt, err := store.TextAndUploadedAt(ctx, userID)
			if errors.Is(err, resume.ErrNotStored) {
				return nil, noCVFile
			}
			if err != nil {
				return nil, err
			}
			st := in.CV
			dropped := resumeextract.KeepSupported(&st, text)
			st.Sanitize()
			if err := store.SetStructured(ctx, userID, st, mcpCVParseModel, uploadedAt); err != nil {
				return nil, err
			}
			h.resume.importExperience(ctx, userID, st, uploadedAt.UTC().Format(time.RFC3339))

			bullets := 0
			for _, e := range st.Experience {
				bullets += len(e.Highlights)
			}
			for _, p := range st.Projects {
				bullets += len(p.Highlights)
			}
			if dropped == nil {
				dropped = []resumeextract.Unsupported{}
			}
			return map[string]any{
				"stored": map[string]any{
					"roles": len(st.Experience), "projects": len(st.Projects), "bullets": bullets,
					"education": len(st.Education), "skills": len(st.Skills),
				},
				"dropped": dropped,
				"dropped_reason": "the uploaded CV's text does not contain these values word for word; " +
					"nothing was stored for them",
			}, nil
		},
	}
}

func (h *assistantHandlers) mcpListMyCVsTool() assistant.Tool {
	return assistant.Tool{
		Name:  "list_my_cvs",
		Title: "List my CVs",
		Description: "The candidate's CVs: base CVs (is_default marks the one tailoring copies from when none is " +
			"named) and tailored copies, each with the job it was tailored for.",
		Effect: assistant.EffectReads,
		Schema: map[string]any{"type": "object", "properties": map[string]any{}, "additionalProperties": false},
		Run: func(ctx context.Context, userID int64, raw json.RawMessage) (any, error) {
			if err := assistant.DecodeArgs(raw, &struct{}{}); err != nil {
				return nil, err
			}
			if err := h.cvSurface(); err != nil {
				return nil, err
			}
			bases, err := h.cv.cvStore.Bases(ctx, userID)
			if err != nil {
				return nil, err
			}
			tailored, err := h.cv.cvStore.ListTailored(ctx, userID)
			if err != nil {
				return nil, err
			}
			outBases := make([]map[string]any, 0, len(bases))
			for _, b := range bases {
				outBases = append(outBases, map[string]any{
					"cv_id": b.ID.String(), "title": b.Title, "is_default": b.IsDefault,
					"updated_at": b.UpdatedAt.UTC().Format(time.RFC3339),
				})
			}
			outTailored := make([]map[string]any, 0, len(tailored))
			for _, t := range tailored {
				outTailored = append(outTailored, map[string]any{
					"cv_id": t.ID.String(), "title": t.Title, "job_slug": t.JobSlug,
					"job_title": t.JobTitle, "company": t.JobCompany,
					"updated_at": t.UpdatedAt.UTC().Format(time.RFC3339),
				})
			}
			return map[string]any{"base_cvs": outBases, "tailored_cvs": outTailored}, nil
		},
	}
}

// jobBySlugForAgent resolves a job the caller named, answering a sentence the agent can act
// on for one that does not exist.
func (h *assistantHandlers) jobBySlugForAgent(ctx context.Context, slug string) (int64, error) {
	slug = strings.TrimSpace(slug)
	if slug == "" {
		return 0, fiber.NewError(fiber.StatusBadRequest, "job_slug is required")
	}
	if h.queries == nil {
		return 0, fiber.NewError(fiber.StatusServiceUnavailable, "jobs are not available on this server")
	}
	id, err := h.queries.GetJobIDBySlug(ctx, slug)
	if errors.Is(err, pgx.ErrNoRows) {
		return 0, fiber.NewError(fiber.StatusNotFound, "no job with slug "+slug)
	}
	return id, err
}

func (h *assistantHandlers) mcpTailorCVForJobTool() assistant.Tool {
	return assistant.Tool{
		Name:  "tailor_cv_for_job",
		Title: "Tailor a CV for a job",
		Description: "Create the candidate's tailored copy of a base CV for one job, or return the copy that already " +
			"exists for that job. The base CV is left untouched; the job is placed on the candidate's tracking " +
			"board as preparing. Needs a CV upload or an existing base CV to copy from.",
		Effect: assistant.EffectAdds,
		Schema: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"job_slug":   map[string]any{"type": "string", "description": "The job's slug, as job search returns it."},
				"base_cv_id": map[string]any{"type": "string", "description": "The base CV to copy. Omit for the default base."},
			},
			"required":             []string{"job_slug"},
			"additionalProperties": false,
		},
		Run: func(ctx context.Context, userID int64, raw json.RawMessage) (any, error) {
			var in struct {
				JobSlug  string `json:"job_slug"`
				BaseCVID string `json:"base_cv_id"`
			}
			if err := assistant.DecodeArgs(raw, &in); err != nil {
				return nil, err
			}
			if err := h.cvSurface(); err != nil {
				return nil, err
			}
			jobID, err := h.jobBySlugForAgent(ctx, in.JobSlug)
			if err != nil {
				return nil, err
			}
			job, err := h.queries.GetJob(ctx, jobID)
			if err != nil {
				return nil, err
			}
			baseID, err := h.cv.requestedTailorBase(ctx, userID, in.BaseCVID)
			if err != nil {
				return nil, err
			}
			base, tailored, created, err := h.cv.tailorCopy(ctx, userID, job, baseID)
			if err != nil {
				return nil, err
			}
			if h.cv.jobs != nil {
				if err := h.cv.jobs.EnsureOnBoard(ctx, userID, job.ID); err != nil {
					log.Printf("mcp: placing vacancy on board after tailor user=%d job=%d: %v", userID, job.ID, err)
				}
			}
			return map[string]any{
				"cv_id": tailored.ID.String(), "base_cv_id": base.ID.String(), "created": created,
				"job": map[string]any{"slug": job.PublicSlug, "title": job.Title, "company": job.Company},
			}, nil
		},
	}
}

// mcpTailoringContextTool is cv_context without the analysis behind it. In the website the
// analysis is a model call the server makes and caches; here the candidate's agent IS the
// model, so it gets what an analysis is made from — the posting, its requirements, and the
// banked evidence the bank's own (model-free) retrieval finds for each — and judges the fit
// itself.
func (h *assistantHandlers) mcpTailoringContextTool() assistant.Tool {
	return assistant.Tool{
		Name:  "tailoring_context",
		Title: "Read a job against my experience",
		Description: "A job's posting as a tailoring reader needs it — title, company, the requirements the posting " +
			"lists, and its text — and, for each listed requirement, the candidate's banked achievements that " +
			"match it, each with its evidence id and whether a CV may print it (can_write_cv). An empty evidence " +
			"list means the bank holds nothing on that requirement. requirements is absent when the posting " +
			"could not be split into a list; the description still carries the whole posting.",
		Effect: assistant.EffectReads,
		Schema: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"job_slug": map[string]any{"type": "string", "description": "The job's slug."},
			},
			"required":             []string{"job_slug"},
			"additionalProperties": false,
		},
		Run: func(ctx context.Context, userID int64, raw json.RawMessage) (any, error) {
			var in struct {
				JobSlug string `json:"job_slug"`
			}
			if err := assistant.DecodeArgs(raw, &in); err != nil {
				return nil, err
			}
			jobID, err := h.jobBySlugForAgent(ctx, in.JobSlug)
			if err != nil {
				return nil, err
			}
			job, err := h.queries.GetJob(ctx, jobID)
			if err != nil {
				return nil, err
			}
			posting := fitanalysis.PostingJob(job, postingRequirements(job))
			asked := make([]matchanalysis.Requirement, 0, len(posting.Requirements))
			for _, r := range posting.Requirements {
				asked = append(asked, matchanalysis.Requirement{Text: r.Text})
			}
			return map[string]any{"job": posting, "requirements_evidence": h.evidenceFor(ctx, userID, asked)}, nil
		},
	}
}

// cvScoped serves a tool built for ONE session's CV to a caller who names the CV. The CV is
// loaded owner-scoped before the tool is built, so a cv_id belonging to someone else reads
// as missing; the inner tool then runs with the remaining arguments, unchanged.
//
// The schema is the inner tool's, minus the properties in omit (arguments that mean nothing
// outside the in-app session), plus cv_id. The inner tool is built once with a zero record
// to read that schema; building it touches no service, so this is safe on a handler that is
// only being listed.
func (h *assistantHandlers) cvScoped(
	name, title, description string, effect assistant.Effect,
	build func(rec cv.Record) (assistant.Tool, error),
	omit ...string,
) assistant.Tool {
	proto, _ := build(cv.Record{})
	props := map[string]any{"cv_id": map[string]any{"type": "string", "description": "The CV's id, from list_my_cvs or tailor_cv_for_job."}}
	if inner, ok := proto.Schema["properties"].(map[string]any); ok {
		for k, v := range inner {
			if !slices.Contains(omit, k) {
				props[k] = v
			}
		}
	}
	required := []string{"cv_id"}
	if inner, ok := proto.Schema["required"].([]string); ok {
		for _, r := range inner {
			if !slices.Contains(omit, r) {
				required = append(required, r)
			}
		}
	}
	return assistant.Tool{
		Name: name, Title: title, Description: description, Effect: effect,
		Schema: map[string]any{
			"type": "object", "properties": props, "required": required, "additionalProperties": false,
		},
		Run: func(ctx context.Context, userID int64, raw json.RawMessage) (any, error) {
			var args map[string]json.RawMessage
			if err := json.Unmarshal(raw, &args); err != nil {
				return nil, fmt.Errorf("%w: %w", assistant.ErrInvalidArguments, err)
			}
			var rawID string
			if err := json.Unmarshal(args["cv_id"], &rawID); err != nil {
				return nil, fiber.NewError(fiber.StatusBadRequest, "cv_id is required")
			}
			id, err := uuid.Parse(strings.TrimSpace(rawID))
			if err != nil {
				return nil, fiber.NewError(fiber.StatusBadRequest, "cv_id is not a CV id")
			}
			delete(args, "cv_id")
			for _, k := range omit {
				if _, sent := args[k]; sent {
					return nil, fmt.Errorf("%w: %s is not an argument of %s", assistant.ErrInvalidArguments, k, name)
				}
			}
			if err := h.cvSurface(); err != nil {
				return nil, err
			}
			rec, err := h.cv.cvStore.GetForModel(ctx, id, userID)
			if err != nil {
				return nil, cvToolError(err)
			}
			tool, err := build(rec)
			if err != nil {
				return nil, err
			}
			rest, err := json.Marshal(args)
			if err != nil {
				return nil, err
			}
			return tool.Run(ctx, userID, rest)
		},
	}
}

func (h *assistantHandlers) mcpGetCVTool() assistant.Tool {
	return h.cvScoped("get_cv", "Read a CV",
		"A CV's current document — every section the candidate's CV prints, with the paths edit_cv addresses — "+
			"minus the contact block.",
		assistant.EffectReads,
		func(rec cv.Record) (assistant.Tool, error) { return h.cvGetTool(rec.ID), nil })
}

func (h *assistantHandlers) mcpEditCVTool() assistant.Tool {
	return h.cvScoped("edit_cv", "Edit a CV",
		"Apply a batch of edits to one CV, all or none; the batch is one entry in the CV's history on the "+
			"website, where the candidate can undo it. Each op names its path as get_cv shows it. An op that "+
			"states something about the candidate — a bullet, the summary, a technology, a skill, education, a "+
			"certification — carries evidence_id: an achievement from the experience bank whose can_write_cv is "+
			"true. One op without valid evidence refuses the whole batch. Each experience or project holds at "+
			fmt.Sprintf("most %d bullets. Contact details and the CV's title are not editable here.", cv.MaxBullets),
		assistant.EffectOverwrites,
		func(rec cv.Record) (assistant.Tool, error) {
			// One batch per call: each call is its own undoable entry, the grain the website
			// shows an outside agent's work at.
			return h.cvEditTool(rec.ID, uuid.New()), nil
		},
		// The run report belongs to the in-app autopilot, which writes its own plan first.
		"requirement", "requirement_status")
}

func (h *assistantHandlers) mcpCVJobMatchTool() assistant.Tool {
	return h.cvScoped("cv_job_match", "Score a CV against its job",
		"How well a tailored CV currently scores against the job it was tailored for — requirements coverage, "+
			"keyword match, job title match, seniority fit — recomputed from what the CV says now. No model call.",
		assistant.EffectReads,
		func(rec cv.Record) (assistant.Tool, error) {
			if rec.ID != uuid.Nil && rec.JobID == 0 {
				return assistant.Tool{}, fiber.NewError(fiber.StatusConflict,
					"this CV is not tailored to a job (or its job no longer exists)")
			}
			return h.cvJobMatchTool(rec.ID, rec.JobID), nil
		})
}

func (h *assistantHandlers) mcpCVPageCountTool() assistant.Tool {
	return h.cvScoped("cv_page_count", "Count a CV's pages",
		"Render a CV exactly as it downloads and report how many pages it fills. No model call.",
		assistant.EffectReads,
		func(rec cv.Record) (assistant.Tool, error) { return h.cvPageCountTool(rec.ID), nil })
}

// mcpExperienceAddTool is experience_add without `said`. In the website, `said` is checked
// against the session transcript and is the one way the agent can record the candidate's own
// words as theirs. Over MCP the conversation happens in the candidate's own client, which this
// server never sees, so the argument could only ever be ignored — and an argument that is
// silently ignored reads to the agent as one that worked.
func (h *assistantHandlers) mcpExperienceAddTool() assistant.Tool {
	t := h.experienceAddTool(uuid.Nil)
	t.Title, t.Effect = "Add an achievement", assistant.EffectAdds
	t.Description = "Record one achievement in the candidate's experience bank. It is recorded as the agent's " +
		"reading (can_write_cv false) until the candidate confirms it on the website's experience page; only " +
		"then may a CV print it."
	if props, ok := t.Schema["properties"].(map[string]any); ok {
		trimmed := maps.Clone(props)
		delete(trimmed, "said")
		schema := maps.Clone(t.Schema)
		schema["properties"] = trimmed
		t.Schema = schema
	}
	inner := t.Run
	t.Run = func(ctx context.Context, userID int64, raw json.RawMessage) (any, error) {
		var probe map[string]json.RawMessage
		if json.Unmarshal(raw, &probe) == nil {
			if _, sent := probe["said"]; sent {
				return nil, fmt.Errorf("%w: said is not an argument here — the candidate confirms an achievement "+
					"on the website instead", assistant.ErrInvalidArguments)
			}
		}
		return inner(ctx, userID, raw)
	}
	return t
}
