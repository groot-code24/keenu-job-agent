# Contributing to Keenu Job Agent

Keenu Job Agent is a local edition of [freehire](https://github.com/strelov1/freehire), and
contributions are welcome. The repository is
<https://github.com/groot-code24/keenu-job-agent>.

## What belongs here, and what belongs upstream

**Here:** whatever makes the local edition work on someone's own computer. That covers running
it in Docker (`docker-compose.yml`, `scripts/local-up.*`, the crawler), the MCP tools a
candidate's own agent uses (`internal/api/handler/mcp_cv_tools.go`, `mcp_apply_tools.go`), the
starter job boards (`deploy/local/starter-boards`), the extension's local build, and
[LOCAL_EDITION.md](LOCAL_EDITION.md).

**Upstream:** a fix or feature for the shared core, such as a source adapter, the pipeline,
search or the CV editor. Send it to [freehire](https://github.com/strelov1/freehire) too, so both
projects get it and this fork can keep pulling updates.

## Getting started

```sh
git clone https://github.com/groot-code24/keenu-job-agent.git
cd keenu-job-agent
./scripts/local-up.sh        # or: powershell -ExecutionPolicy Bypass -File scripts\local-up.ps1
```

Then follow [LOCAL_EDITION.md](LOCAL_EDITION.md). For the toolchain behind the checks below
(Go, pnpm, sqlc, linters), `mise install` sets up the exact versions CI uses.

## Ground rules for this fork

- **The honesty rules are not negotiable.** A change must not let an agent write a claim the
  candidate did not make, or submit an application without the candidate's explicit opt-in.
  The import check, the CV evidence gate and "the extension never submits" exist for that, and
  a PR that weakens them will not be merged.
- **No AI costs on the server.** The local edition's AI runs in the user's own agent. A feature
  that needs the server to call a model must stay optional, and off unless a key is configured.
- **Nothing leaves the user's computer by default.** No analytics, telemetry or outbound calls
  beyond crawling public job boards.
- PR titles follow Conventional Commits (`feat: …`, `fix(ingest): …`); CI checks this.
- Before pushing Go: `gofmt -l .`, `go vet ./...`, `go vet -tags=integration ./...`,
  `go test ./...`. For the web app: `pnpm check` and `pnpm exec vitest run` in `web/`.

The rest of this guide is the original project's and still applies to the shared code.

---

# Contributing to freehire

Contributions are welcome. This guide exists to save both sides time.

## Philosophy

freehire's core is a small, opinionated pipeline: fetch → normalize → dedup →
upsert → enrich → serve. Everything a source adapter produces flows through one
schema and one wire shape.

**The extension point is the source, not the core.** Adding a company is one row
in the `boards` catalog — submit it through the "contribute a board" form on the
site, no PR needed. Adding an ATS platform is one new adapter in
`internal/ingest/sources` plus one line in `sources.All`. Adding a Telegram channel is
one row in `telegram_channels`. If your feature fits that shape, it is welcome.

Changes that bloat the core — new abstractions, config knobs, or error handling
that no current feature needs — are a harder sell. Build each feature correctly
and idiomatically, neither gold-plated nor a placeholder.

## The One Rule

**You must understand your code.** If you cannot explain what your change does
and how it interacts with the rest of the system, it will not be merged.

Using AI to write code is fine — this repository is built with it. Submitting
generated code you have not read is not. If you use an agent, run it from the
repository root so it picks up `AGENTS.md` automatically, and check that it
followed the rules there.

## Opening an issue

Use one of the [issue templates](.github/ISSUE_TEMPLATE). Bug reports, source
requests, and feature ideas are all fair game.

- Keep it concise. If it does not fit on one screen, it is probably two issues.
- Write in your own voice. A pasted LLM transcript is hard to act on.
- For a bug, include reproduction steps and the relevant logs.
- If you want to implement it yourself, say so — we will not race you.
- Check [.out-of-scope/](.out-of-scope/README.md) first: it lists ideas already
  declined, why, and what would change the answer.

Questions, half-formed ideas, and help requests belong in
[Discussions](https://github.com/strelov1/freehire/discussions) instead.

## Opening a pull request

Small PRs get reviewed fastest. For anything large or architectural, open an
issue first so we can agree on the shape before you spend the evening on it —
that is a suggestion, not a gate.

Run the same checks CI runs:

```bash
go build ./...
go vet ./...
gofmt -l .              # must print nothing
go test ./...           # unit tests
```

Note that `go test ./...` compiles no file behind the `integration` build tag.
If you changed a signature, `go vet -tags=integration ./...` is the cheap guard —
seconds, no Docker. If you touched anything Docker-dependent (queries, schema,
handlers), run the suite itself:

```bash
go test -tags=integration ./...   # needs Docker (testcontainers)
```

Regenerate committed artifacts when their source changed, and commit the result:

- Changed `internal/platform/db/queries/*.sql` or `migrations/` → `make sqlc`.
- Changed the Go contract types → `make gen-contracts`.

For the frontend (`web/`):

```bash
cd design-system && pnpm install && pnpm build   # design-system must build first
cd ../web && pnpm install && pnpm run check && pnpm run build
```

## Adding a source

This is the most welcome kind of contribution. See `AGENTS.md` for the source
adapter contract and the dedup/lifecycle conventions. In short:

- An ATS board adapter implements the `Source` interface in `internal/ingest/sources`
  and is registered in `sources.All`.
- A single outbound-link resolver implements `LinkSource` in
  `internal/ingest/linksource` and is registered in `linksource.All`.
- Always validate board slugs against the live provider before adding them — a
  board that yields nothing looks identical to a healthy one in the logs.

## Host dependencies

The `Dockerfile` installs what the server needs. If you run the backend outside
Docker (`make run`), put these tools on your `PATH`:

| Binary | Package | Needed for | If missing |
| :--- | :--- | :--- | :--- |
| `pdftotext` | `poppler-utils` | CV text extraction (résumé upload) | Uploads fail with a generic "invalid PDF" error, because a missing binary is reported the same way as a corrupt file |
| `typst` | `typst` (pinned to `0.15.0` in the `Dockerfile`) | CV PDF export | The PDF endpoint returns `501`; the rest of the CV builder still works. Set `TYPST_BIN` if it isn't on `PATH` |
| `google-chrome` / `chromium` | Chrome or Chromium | Optional: only when an auto-apply run (`cmd/auto-apply`) executes | That run fails to launch a browser; everything else works without it |

On Debian/Ubuntu, `sudo apt-get install -y poppler-utils` installs `pdftotext`.
`typst` isn't in apt: the `Dockerfile` downloads the static release tarball from
[GitHub releases](https://github.com/typst/typst/releases). Install Chrome or
Chromium from your usual source.

## Questions?

Ask in [GitHub Discussions](https://github.com/strelov1/freehire/discussions).
