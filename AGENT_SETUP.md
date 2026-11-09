# Setting up Keenu Job Agent — instructions for an AI agent

You are an AI coding agent (Claude Code, Codex, or similar) and a person has asked you to set
up or use **Keenu Job Agent** from <https://github.com/groot-code24/keenu-job-agent>. Follow
these steps in order. Run the commands yourself where you can; tell the person plainly when a
step needs them.

Keenu runs entirely on the person's computer, in Docker. It is free and needs no API key: the
AI is **you**, connected to the person's own server over MCP. The server stores their data and
enforces the rules; it makes no AI calls of its own.

## 1. Check the prerequisites

```sh
git --version
docker info --format '{{.ServerVersion}}'   # must print a version, not an error
docker compose version                      # must be v2
```

- **Docker missing or not running:** ask the person to install and start
  [Docker Desktop](https://www.docker.com/products/docker-desktop/), then continue. Do not try to
  install it silently: it needs their consent and, on Windows, may need a restart.
- **Memory:** the published images just need pulling. Building them locally (the fallback) needs
  about 4 GB of free memory.

## 2. Get the code

Clone into the person's home folder unless they name another place:

```sh
git clone https://github.com/groot-code24/keenu-job-agent.git ~/keenu-job-agent
cd ~/keenu-job-agent
```

If it is already cloned there, `git pull` instead.

## 3. Start it

| OS | Command (from the repository folder) |
|---|---|
| macOS / Linux | `./scripts/local-up.sh` |
| Windows | `powershell -ExecutionPolicy Bypass -File scripts\local-up.ps1` |

The script generates this install's own secrets, pulls the images, starts everything and waits
until <http://localhost:8090> answers. The first start takes a few minutes; if it has to build
the images instead, 20–50 minutes. Run it in the background if your tool would otherwise time
out, then wait for it to print `Keenu Job Agent is running`. You can also poll
`http://localhost:8090/health` until it answers 200.

## 4. Connect yourself

Register the person's local server as an MCP server named `keenu`:

```sh
# Claude Code
claude mcp add --transport http --scope user keenu http://localhost:8090/api/v1/mcp

# Codex
codex mcp add keenu --url http://localhost:8090/api/v1/mcp
```

## 5. What the person must do

Tell them, in this order:

1. **Open <http://localhost:8090> and create an account** (email and password). No email is sent
   and no verification is needed.
2. **Upload their CV as a PDF** on the profile page (<http://localhost:8090/my/profile>).
3. **Approve your connection.** In Claude Code: start a new session (or run `/mcp`), choose
   `keenu`, choose *Authenticate*. In Codex: run `codex mcp login keenu`. A browser tab opens on
   their Keenu; they approve it there.

When the `keenu` tools are available to you (for example `read_my_cv_file`), setup is done.

## 6. Using it

- **Import the CV:** `read_my_cv_file`, then `import_my_cv` with the CV restated in the given
  shape. Copy values exactly; the server keeps only what the file's text contains and lists what
  it dropped, which you should tell the person.
- **Find jobs:** `search_jobs` and `get_job`. A fresh install collects jobs from a starter list
  of companies within its first minutes; more can be added (see `deploy/local/README.md`).
- **Tailor a CV:** `tailor_cv_for_job` → `tailoring_context` → `experience_search` → `edit_cv`
  (each claim cites an `evidence_id` whose `can_write_cv` is true) → `cv_job_match` and
  `cv_page_count` to check. Tell the person they can review, undo and download it at
  <http://localhost:8090/my/cvs>.
- **Apply:** the person builds and loads the browser extension (see `LOCAL_EDITION.md`), opens
  the application page with the side panel open, and you use `read_application_form` and
  `fill_application_form`. You never submit. The person reviews and presses Submit.

## Rules you cannot work around

The server enforces these, so do not try. Explain them to the person when they come up.

- A CV states only what the person did. An `edit_cv` op without a valid `evidence_id` is refused.
- What you add with `experience_add` is your reading. It cannot be printed until the person
  presses **Confirm** on their Experience page. Ask them about anything their record does not
  show, rather than inventing it.
- Nothing is submitted from the browser for them.

## Troubleshooting

| Symptom | Fix |
|---|---|
| `Docker is not running` | Start Docker Desktop and rerun the start script. |
| Port 8090, 8080, 5432, 7700 or 9000 already in use | Stop what uses it, or set `WEB_HOST_PORT` / `HIRE_HOST_PORT` / `DB_HOST_PORT` / `MEILI_HOST_PORT` / `S3_HOST_PORT` in `.env`. A different web port means a different `WEB_ORIGIN`, which the script handles by building the web image locally. |
| Build killed or very slow | Not enough memory: ask the person to close heavy apps, then rerun. |
| "Requests from other sites are refused" when signing in | Open exactly <http://localhost:8090>, not `127.0.0.1` or a LAN address. |
| `keenu` tools missing after `mcp add` | Start a new session, then authenticate (step 5.3). |
| `no CV file uploaded yet` | The person uploads the PDF on the profile page. |
| `no browser is attached` | The extension's side panel must be open on the page. |
| Search shows nothing | Jobs are still being collected; wait a few minutes. `docker compose logs crawler` shows progress. |

Updating later: `git pull`, then run the start script again.
