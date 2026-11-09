# Keenu Job Agent

Keenu Job Agent is a local edition of [freehire](https://github.com/strelov1/freehire), the
open-source job aggregator, used under its MIT license. It runs entirely on your own computer.
Every feature is open to every account, and nothing is sold. The AI work runs in **your own**
Claude Code or Codex, on the subscription you already have, so this server needs no API key and
makes no AI calls itself.

The easiest way to set it up is to let your agent do it: give Claude Code or Codex the
repository URL and ask it to follow [AGENT_SETUP.md](AGENT_SETUP.md).

## What you need

- [Docker Desktop](https://www.docker.com/products/docker-desktop/).
- [Claude Code](https://docs.anthropic.com/en/docs/claude-code) or Codex, signed in with your own
  subscription.
- [Node.js](https://nodejs.org/) 24, only to build the browser extension.

## Start it

One command fetches the code into `~/keenu-job-agent` and starts everything. It needs only
Docker Desktop, which it starts for you if it is installed but not running:

```sh
# macOS / Linux
curl -fsSL https://raw.githubusercontent.com/groot-code24/keenu-job-agent/main/scripts/install.sh | sh
```

```powershell
# Windows (PowerShell)
irm https://raw.githubusercontent.com/groot-code24/keenu-job-agent/main/scripts/install.ps1 | iex
```

Running it again updates the code and restarts, keeping your data. From a copy you already
have, the start scripts do the same without fetching anything:

```sh
./scripts/local-up.sh                                        # macOS / Linux
powershell -ExecutionPolicy Bypass -File scripts\local-up.ps1  # Windows
```

The script pulls ready-made images (published by this repository's `images` workflow), so the
first start takes a few minutes. If it cannot pull them, or you set `KEENU_BUILD=1`, it builds
them on your machine instead: 20–50 minutes and about 4 GB of free memory. Then open
**<http://localhost:8090>** and sign up.

On first start the script writes this install's own secrets (`JWT_SECRET`, `MEILI_MASTER_KEY`)
into `.env`. On Windows it also adds `deploy/local/compose.loopback6.yml` there, so the site
answers on the IPv6 loopback too: Windows tries that first for `localhost` and otherwise waits
about 200 ms on every new connection. Every port listens on `127.0.0.1` only, so nobody else on your network can reach
it. Set `BIND_ADDRESS=0.0.0.0` in `.env` only if you mean to share it.

`docker compose down` stops everything. Your data stays in Docker volumes; only
`docker compose down -v` erases it.

## Connect your AI agent

In the app, open **Integrations → Your AI agent**. That page shows the exact commands for your
install. In short:

```sh
claude mcp add --transport http keenu http://localhost:8090/api/v1/mcp   # then /mcp → Authenticate
codex mcp add keenu --url http://localhost:8090/api/v1/mcp && codex mcp login keenu
```

Approving the connection in the browser tab that opens is all the sign-in there is. No email
verification and no API key are needed.

Then work in your own words:

1. Upload your CV (PDF) on your profile page, and ask your agent to *import my CV*.
2. Ask it to *find jobs that fit me and tailor a CV for the best one*.
3. Build and load the extension (below), open an application, and ask it to *fill this application;
   don't submit*.

## Rules the server enforces

These hold whatever your agent is told.

- **Only what your CV says.** An import keeps only values your uploaded file actually contains.
  Anything else is dropped and reported back to you.
- **Every claim needs your evidence.** A CV edit that states something about you must cite an
  achievement you made. One the agent adds by itself cannot be printed until you press
  **Confirm** on your Experience page.
- **Nothing is submitted for you from the browser.** The extension fills fields; you review and
  press Submit. (Auto-apply, which does submit, stays off until you turn on *agents may apply* in
  your settings.)

## The browser extension

```sh
cd extension && npm install && npm run build:local
```

In Chrome, open `chrome://extensions`, turn on **Developer mode**, choose **Load unpacked** and
pick `extension/.output/chrome-mv3-localedition`. This build talks only to
`http://localhost:8090` and sends no analytics anywhere.

## Jobs

On every start, the stack checks a starter list of company job boards (Greenhouse, Lever, Ashby)
against the live boards and adds the ones that answer. A crawler then keeps them fresh. To add
companies, see [deploy/local/README.md](deploy/local/README.md).

## Privacy

Your account, CVs and tracking stay on your computer. The only things that leave it are the
crawler fetching public job boards, and whatever your agent reads, such as your CV text and job
descriptions. That goes to Anthropic or OpenAI as part of your own conversation with your agent.

## Limits

- No email is sent, so there are no verification codes, password-reset mails or job-alert emails.
- Without an LLM key on the server, its built-in AI features (the in-app assistant, dictation,
  the server-side fit analysis) stay off. Your agent does that work instead.
- Building the images locally needs about 4 GB of free memory. If a build dies, close other apps
  and run the start script again.
- Company logos are off by default, so no company name you look at leaves your computer. Set
  `PUBLIC_LOGO_ORIGIN` to a logo proxy to turn them on. Analytics never load unless you set
  `PUBLIC_GA_MEASUREMENT_ID` yourself.
