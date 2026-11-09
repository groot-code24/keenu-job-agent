<script lang="ts">
  import { resolve } from '$app/paths';
  import { reopen } from '#lib/consent.svelte.js';
  import { ProviderIcon } from '#lib/ui/index.js';

  // The local edition's footer. It links what this install actually serves and nothing
  // else: no store listings (those extensions talk to the public deployment, not to this
  // computer — the extension that works here is built from the repository), no
  // community or social profiles of the project this one is built from, and no
  // third-party image (a hosted badge is a request to somebody else's server on every
  // page view, which this edition promises not to make).
  //
  // Internal links go through resolve() (base-path safe), mirroring the header.
  const groups = [
    {
      title: 'Find work',
      links: [
        { label: 'Jobs', href: resolve('/jobs') },
        { label: 'Companies', href: resolve('/companies') },
        { label: 'Collections', href: resolve('/collections') },
        { label: 'Jobs by role', href: resolve('/roles') },
        { label: 'Skills glossary', href: resolve('/skills') },
      ],
    },
    {
      title: 'Your agent',
      links: [
        { label: 'Connect your agent', href: resolve('/my/integrations/agent') },
        { label: 'Your CV', href: resolve('/my/profile') },
        { label: 'Tailored CVs', href: resolve('/my/cvs') },
        { label: 'Tracked jobs', href: resolve('/my/tracking') },
      ],
    },
    {
      title: 'This install',
      links: [
        { label: 'How it works', href: resolve('/how-it-works') },
        { label: 'Where jobs come from', href: resolve('/sources') },
        { label: 'Crawler status', href: resolve('/status') },
        { label: 'API docs', href: resolve('/docs/api') },
      ],
    },
  ];

  /** This edition's own repository: where its code, issues and updates live. */
  const REPOSITORY_URL = 'https://github.com/groot-code24/keenu-job-agent';
  /** The project it is built from, credited under its MIT license. */
  const UPSTREAM_URL = 'https://github.com/strelov1/freehire';

  const year = new Date().getFullYear();

  // Compact: the account shell (/my/*) is an app-like surface with its own sidebar nav,
  // so the link columns below are chrome that doesn't belong there — only the bottom bar
  // still applies. Full-bleed account routes (/my/assistant/*, /tailor/*) render no footer
  // at all, compact or otherwise — see routes/+layout.svelte's hideFooter.
  let { compact = false }: { compact?: boolean } = $props();
</script>

<footer class="border-t border-border">
  {#if !compact}
    <div class="mx-auto grid max-w-6xl gap-8 px-4 py-10 sm:grid-cols-5 sm:py-12">
      <div class="flex max-w-xs flex-col gap-3 sm:col-span-2">
        <p class="text-sm font-semibold tracking-tight">Keenu Job Agent</p>
        <p class="text-sm leading-relaxed text-muted-foreground">
          A private job search that runs on this computer. Your CV, your applications and your account never leave it.
        </p>
      </div>
      <!-- Navigation groups. Each is a named landmark (aria-label) so screen readers get a
           title without adding headings to the page outline. -->
      {#each groups as group (group.title)}
        <nav class="flex flex-col gap-3" aria-label={group.title}>
          <p class="text-xs font-medium uppercase tracking-wider text-muted-foreground">
            {group.title}
          </p>
          <ul class="flex flex-col gap-2">
            {#each group.links as link (link.href)}
              <li>
                <!-- eslint-disable-next-line svelte/no-navigation-without-resolve -- internal route already passed through resolve() when building `groups`; the linter can't trace it via the variable -->
                <a href={link.href} class="text-sm text-muted-foreground transition-colors hover:text-foreground">
                  {link.label}
                </a>
              </li>
            {/each}
          </ul>
        </nav>
      {/each}
    </div>
  {/if}

  <!-- Bottom bar: copyright and cookie settings on the left, where the code lives on the
       right. Split off by its own top border only when it follows the link groups above —
       compact, the footer's own top border already does that job. -->
  <div class={compact ? undefined : 'border-t border-border'}>
    <div
      class="mx-auto flex max-w-6xl flex-col gap-3 px-4 py-4 text-xs text-muted-foreground sm:flex-row sm:items-center sm:justify-between sm:gap-1"
    >
      <div class="flex items-center gap-4">
        <p>© {year} Keenu Job Agent</p>
        <!-- Re-opens the consent banner so a prior cookie choice can be changed —
             withdrawal as easy as granting (GDPR). -->
        <button type="button" onclick={reopen} class="text-muted-foreground transition-colors hover:text-foreground">
          Cookie settings
        </button>
      </div>
      <p class="flex flex-wrap items-center gap-x-1">
        Built on
        <!-- eslint-disable svelte/no-navigation-without-resolve -- external repository URLs opened in a new tab; not internal routes -->
        <a
          href={UPSTREAM_URL}
          target="_blank"
          rel="noopener noreferrer"
          class="underline underline-offset-2 hover:text-foreground"
        >
          freehire</a
        >, MIT licensed.
        <a
          href={REPOSITORY_URL}
          target="_blank"
          rel="noopener noreferrer"
          class="ml-1 inline-flex items-center gap-1 font-medium text-foreground transition-colors hover:text-muted-foreground"
        >
          <ProviderIcon provider="github" class="size-3.5" />
          View source
        </a><!-- eslint-enable svelte/no-navigation-without-resolve -->
      </p>
    </div>
  </div>
</footer>
