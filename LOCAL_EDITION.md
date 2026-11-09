# Keenu Job Agent

Keenu Job Agent is a local edition of [freehire](https://github.com/strelov1/freehire), the
open-source job aggregator, used under its MIT license. It runs entirely on your own computer.
Every feature is open to every account, and nothing is sold. The AI work runs in **your own**
Claude Code or Codex, on the subscription you already have, so this server needs no API key and
makes no AI calls itself.

## What you need

- [Docker Desktop](https://www.docker.com/products/docker-desktop/), with about 4 GB of free memory
  the first time you start, while the images build.
- [Claude Code](https://docs.anthropic.com/en/docs/claude-code) or Codex, signed in with your own
  subscription.
- [Node.js](https://nodejs.org/) 24, only to build the browser extension.

## Start it

```sh
./scripts/local-up.sh                                        # macOS / Linux
powershell -ExecutionPolicy Bypass -File scripts\local-up.ps1  # Windows
```

The first start builds the images, which can take 20–30 minutes on a laptop. Later starts take
seconds. Then open **<http://localhost:8090>** and sign up.

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
- `docker compose` needs about 4 GB of free memory while building. If a build dies, close other
  apps and run the start script again.
