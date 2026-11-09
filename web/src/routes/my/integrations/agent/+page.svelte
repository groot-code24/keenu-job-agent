<script lang="ts">
  import { page } from '$app/state';
  import { resolve } from '$app/paths';
  import { locale } from '#lib/i18n/currentLocale.svelte.js';
  import { t } from '#lib/i18n/t.js';
  import { messages } from './messages';

  // Connecting the candidate's OWN agent (Claude Code, Codex) to this server's MCP endpoint,
  // the local edition's way of using AI: the agent runs on their subscription, and the
  // server only holds the data and keeps the rules (internal/api/mcpapp/AGENTS.md). Every
  // command is built from the origin this page was opened on, so it names this install —
  // never the public deployment.

  const s = $derived(t(messages, locale()));
  const mcpURL = $derived(`${page.url.origin}/api/v1/mcp`);
  const profileHref = resolve('/my/profile');

  const commands = $derived({
    claude: `claude mcp add --transport http keenu ${mcpURL}`,
    codex: `codex mcp add keenu --url ${mcpURL}\ncodex mcp login keenu`,
    extension: 'cd extension\nnpm install\nnpm run build:local',
    folder: 'extension/.output/chrome-mv3-localedition',
    importPrompt: 'Read my CV file on Keenu and import it.',
    tailorPrompt:
      'Find backend jobs on Keenu that fit my experience, pick the best one, and tailor a CV for it. ' +
      'Ask me before claiming anything my experience bank does not show.',
    applyPrompt:
      'Fill the application open in my browser with my details from Keenu. Do not submit; tell me what you could not fill.',
  });

  // Which block was copied last, so its button can say so. Clipboard access can be denied;
  // the text stays selectable either way.
  let copied = $state('');
  async function copy(id: string, text: string) {
    try {
      await navigator.clipboard.writeText(text);
      copied = id;
      setTimeout(() => (copied === id ? (copied = '') : null), 1500);
    } catch {
      copied = '';
    }
  }
</script>

{#snippet block(id: string, text: string)}
  <div class="relative">
    <pre class="overflow-x-auto rounded-lg border bg-muted/40 p-3 pr-20 text-xs leading-relaxed"><code>{text}</code
      ></pre>
    <button
      type="button"
      class="absolute top-2 right-2 rounded-md border bg-background px-2 py-1 text-xs hover:bg-muted"
      onclick={() => copy(id, text)}
    >
      {copied === id ? s.copied : s.copy}
    </button>
  </div>
{/snippet}

<div class="flex max-w-3xl flex-col gap-8">
  <p class="text-sm text-pretty text-muted-foreground">{s.intro}</p>

  <section class="flex flex-col gap-2">
    <h2 class="text-sm font-medium">{s.serverLabel}</h2>
    {@render block('url', mcpURL)}
  </section>

  <section class="flex flex-col gap-3">
    <h2 class="text-base font-semibold">{s.connectTitle}</h2>
    <h3 class="text-sm font-medium">{s.claudeLabel}</h3>
    {@render block('claude', commands.claude)}
    <p class="text-sm text-muted-foreground">{s.claudeAfter}</p>
    <h3 class="text-sm font-medium">{s.codexLabel}</h3>
    {@render block('codex', commands.codex)}
    <p class="text-sm text-muted-foreground">{s.codexAfter}</p>
  </section>

  <section class="flex flex-col gap-3">
    <h2 class="text-base font-semibold">{s.cvTitle}</h2>
    <p class="text-sm text-pretty">
      {s.cvBody}
      <a class="font-medium text-brand-strong hover:underline" href={profileHref}>{s.openProfile}</a>
    </p>
    <h3 class="text-sm font-medium">{s.promptLabel}</h3>
    {@render block('import', commands.importPrompt)}
  </section>

  <section class="flex flex-col gap-3">
    <h2 class="text-base font-semibold">{s.tailorTitle}</h2>
    <p class="text-sm text-pretty">{s.tailorBody}</p>
    <h3 class="text-sm font-medium">{s.promptLabel}</h3>
    {@render block('tailor', commands.tailorPrompt)}
  </section>

  <section class="flex flex-col gap-3">
    <h2 class="text-base font-semibold">{s.applyTitle}</h2>
    <p class="text-sm text-pretty">{s.applyBody}</p>
    {@render block('extension', commands.extension)}
    {@render block('folder', commands.folder)}
    <h3 class="text-sm font-medium">{s.promptLabel}</h3>
    {@render block('apply', commands.applyPrompt)}
  </section>
</div>
