<script lang="ts">
  import { goto } from '$app/navigation';
  import { resolve } from '$app/paths';
  import { ArrowRight, Bot, Check, FileCheck2, FileUp, Hand, LockKeyhole, Puzzle, UserPlus } from '@lucide/svelte';
  import type { Component } from 'svelte';
  import FilterModal from './filters/FilterModal.svelte';
  import HeaderSearch from './HeaderSearch.svelte';
  import RecentJobsFeed from './RecentJobsFeed.svelte';
  import ChatExample from './ChatExample.svelte';
  import { ROLE_PLACEHOLDER } from '#lib/placeholderRoles.js';
  import { api } from '#lib/api.js';
  import { browseQuery, planForSuggestion } from '#lib/browseTarget.js';
  import type { ApplyPlan } from '#lib/apiSuggestions.js';
  import { countryLabel } from '#lib/facets.js';
  import { emptyFilters, filtersToParams } from '#lib/facetModel.js';
  import type { StagedFilters } from '#lib/stagedFilters.svelte.js';
  import { starterSuggestions } from '#lib/suggestions.js';
  import { Button, SectionLabel } from '#lib/ui/index.js';
  import type { CatalogScale, FacetCounts } from '#lib/types.js';
  import { formatCount } from '#lib/utils.js';

  // The local edition's homepage. Keenu runs on the person's own computer and is worked
  // by their own AI agent, so the page says that first, puts the search box under it,
  // and then answers the one question a fresh install raises — "what do I do now?" — as
  // four steps. Below them: how big this install's catalogue is, what the agent can be
  // asked, and the rules the server holds it to whatever it is told.
  //
  // The box is the header's own search component at hero size. On a page with no list
  // registered, that component already drives a target that turns every pick into a
  // link to the feed (see browseTarget) — so the dropdown, its keyboard walk and the
  // hotkeys all arrive here unchanged. The header renders no box on this route (see
  // TopBar): one screen, one field.
  let {
    counts,
    scale,
    signedIn = false,
  }: {
    /** The unfiltered category and country distributions, or null when the call
     *  failed. Feeds both rows of chips and the box's empty dropdown. */
    counts: FacetCounts | null;
    /** How big this install's catalogue is, or null before rollup-stats has run. */
    scale: CatalogScale | null;
    /** Whether the visitor has an account here: the first setup step is then done. */
    signedIn?: boolean;
  } = $props();

  /** How many shortcuts each row draws: what fits on one line at 1280px. */
  const CHIP_LIMIT = 6;
  const COUNTRY_CHIP_LIMIT = 3;

  type Chip = { href: string; label: string; count?: number };

  /** Where a chip goes, serialized by the same `browseQuery` a pick in the dropdown
   *  navigates through, so the link and the control cannot filter differently. Real
   *  `<a href>`s rather than click handlers. */
  function planHref(plan: ApplyPlan): string {
    return `${resolve('/jobs')}?${browseQuery(plan)}`;
  }

  function feedHref(param: string, value: string): string {
    return planHref({ facets: [[param, value]] });
  }

  /** What, rather than where — built by the same function the dropdown's empty state
   *  uses, biggest first. */
  const crafts = $derived<Chip[]>(
    starterSuggestions(counts, CHIP_LIMIT)
      .map((s) => ({ href: planHref(planForSuggestion(s)), label: s.label, count: s.count }))
      .sort((a, b) => (b.count ?? 0) - (a.count ?? 0)),
  );

  /** Where, rather than what: Remote, then the busiest countries this catalogue holds. */
  const places = $derived<Chip[]>([
    { href: feedHref('work_mode', 'remote'), label: 'Remote' },
    ...Object.entries(counts?.facets?.countries ?? {})
      .sort(([, a], [, b]) => b - a)
      .slice(0, COUNTRY_CHIP_LIMIT)
      .map(([code, count]) => ({ href: feedHref('countries', code), label: countryLabel(code), count })),
  ]);

  type Figure = { value: string; label: string };

  /** One catalogue figure, or nothing: a cold stats snapshot zeroes what it could not
   *  measure, and a page saying "0 companies" is worse than one saying nothing. */
  function figure(value: number | undefined, label: string): Figure | null {
    if (!value) return null;
    return { value: value.toLocaleString('en-US'), label };
  }

  /** The hero's number, floored to two significant figures (6,358 → "6,300+", never
   *  rounded up): a "+" figure must not overstate the catalogue. Before rollup-stats has
   *  published a figure the line says so in words. */
  function floorFigure(n: number): number {
    const step = 10 ** Math.max(0, Math.floor(Math.log10(n)) - 1);
    return Math.floor(n / step) * step;
  }
  const jobsFigure = $derived(
    scale?.unique_open_jobs ? `${formatCount(floorFigure(scale.unique_open_jobs))}+` : 'Thousands of',
  );

  const figures = $derived(
    [
      // The de-duplicated count, the one /jobs shows.
      figure(scale?.unique_open_jobs, 'open jobs'),
      figure(scale?.companies, 'companies hiring'),
      figure(scale?.sources, 'job boards read'),
    ].filter((f) => f !== null),
  );

  type Step = { title: string; body: string; href: string; action: string; icon: Component; done?: boolean };

  /** Getting a fresh install working, in the order it has to happen. The agent and the
   *  extension are both set up from the one integrations page, which builds its commands
   *  from the origin it is opened on. */
  const steps = $derived<Step[]>([
    {
      title: 'Create your account',
      body: 'An email and a password. Nothing is sent anywhere, and there is nothing to verify.',
      href: resolve('/signin'),
      action: 'Create account',
      icon: UserPlus,
      done: signedIn,
    },
    {
      title: 'Upload your CV',
      body: 'A PDF on your profile page. It is stored on this computer, not uploaded anywhere.',
      href: resolve('/my/profile'),
      action: 'Open profile',
      icon: FileUp,
    },
    {
      title: 'Connect your agent',
      body: 'Paste one command into Claude Code or Codex, then approve it here. No API key.',
      href: resolve('/my/integrations/agent'),
      action: 'Get the command',
      icon: Bot,
    },
    {
      title: 'Add the browser extension',
      body: 'Build it once. It fills application forms beside the posting; you press Submit.',
      href: resolve('/my/integrations/agent'),
      action: 'How to install',
      icon: Puzzle,
    },
  ]);

  /** What the server enforces whatever the agent is told — the reason a person can let
   *  an AI near their CV. Worded from LOCAL_EDITION.md's "Rules the server enforces". */
  const rules: { title: string; body: string; icon: Component }[] = [
    {
      title: 'Only what your CV says',
      body: 'An import keeps only what your uploaded file contains; anything else is dropped and reported to you.',
      icon: FileCheck2,
    },
    {
      title: 'Every claim cites your evidence',
      body: 'A tailored CV can state only achievements you made. One the agent infers waits for your Confirm.',
      icon: Check,
    },
    {
      title: 'You press Submit',
      body: 'The extension fills the form. Nothing is ever sent to an employer from the browser for you.',
      icon: Hand,
    },
    {
      title: 'Your data stays here',
      body: 'Your account, CVs and applications live in this computer’s Docker volumes, and nowhere else.',
      icon: LockKeyhole,
    },
  ];

  /** The All-filters modal, hosted here rather than by a list — there is no list. Here
   *  it COMPOSES a search: seeded empty, and its footer hands back the staged selection,
   *  which becomes the query the feed opens with. */
  let filtersOpen = $state(false);

  /** The distribution behind the modal's controls: the dynamic facets (Skills, City,
   *  Job language, Source) build their whole option list from it. `disjunctive` so each
   *  facet's numbers ignore its own selection. */
  const stagedCounts = (params: URLSearchParams) => api.facetCounts(params, { disjunctive: true });

  /** How many postings the staged selection would open, for the footer's Apply. */
  const previewCount = (params: URLSearchParams) => api.facetCounts(params).then((c) => c.total);

  function searchWithFilters(staged: StagedFilters) {
    filtersOpen = false;
    const query = filtersToParams(staged.value).toString();
    // eslint-disable-next-line svelte/no-navigation-without-resolve -- query appended to a resolved path
    void goto(query ? `${resolve('/jobs')}?${query}` : resolve('/jobs'));
  }
</script>

<!-- One row of shortcuts into the catalogue: the crafts and the places share it. -->
{#snippet chipRow(label: string, chips: Chip[])}
  <div class="flex flex-col items-center gap-2">
    <p class="text-xs uppercase tracking-wide text-muted-foreground/80" aria-hidden="true">{label}</p>
    <nav aria-label={label} class="flex flex-wrap justify-center gap-2">
      {#each chips as chip (chip.href)}
        <!-- eslint-disable svelte/no-navigation-without-resolve -- feedHref resolves /jobs and appends the serialized filter -->
        <a
          href={chip.href}
          class="rounded-full border border-border bg-card px-3 py-1.5 text-sm text-muted-foreground transition-colors hover:border-brand hover:text-foreground"
        >
          {chip.label}
          {#if chip.count !== undefined}
            <span class="ml-1 text-xs tabular-nums text-muted-foreground/80">{chip.count.toLocaleString('en-US')}</span>
          {/if}
        </a><!-- eslint-enable svelte/no-navigation-without-resolve -->
      {/each}
    </nav>
  </div>
{/snippet}

<!-- Hero. `svh` rather than `vh`: on mobile Safari `100vh` counts browser chrome that is
     not there. `dot-grid dot-grid-centred` is the shared landing backdrop from app.css. -->
<section
  class="dot-grid dot-grid-centred -mx-4 flex min-h-[calc(100svh-14rem)] flex-col items-center justify-center px-4 pb-12 pt-12"
>
  <div class="flex w-full max-w-2xl flex-col items-center gap-6">
    <div class="flex flex-col items-center gap-5 text-center">
      <p
        class="inline-flex items-center gap-2 rounded-full border border-border bg-card px-3 py-1 text-xs font-medium text-muted-foreground"
      >
        <span class="relative flex size-2" aria-hidden="true">
          <span
            class="absolute inline-flex size-full animate-ping rounded-full bg-success opacity-60 motion-reduce:hidden"
          ></span>
          <span class="relative inline-flex size-2 rounded-full bg-success"></span>
        </span>
        Running privately on this computer
      </p>
      <h1 class="text-balance text-4xl font-semibold tracking-tight sm:text-5xl">
        Your job search, run by
        <span class="text-brand-strong">your own AI agent.</span>
      </h1>
      <p class="max-w-xl text-balance text-base leading-relaxed text-muted-foreground sm:text-lg">
        {jobsFigure} live postings, read straight from company career pages. Search them here, or let Claude Code or Codex
        find the right ones, tailor your CV and fill the application.
      </p>
    </div>

    <!-- `relative z-30`: the layer every panel the box opens is drawn in — above the chips
         that follow it, below the sticky header (`z-40`). -->
    <div class="relative z-30 flex w-full">
      <HeaderSearch
        placeholder={ROLE_PLACEHOLDER}
        label="Search jobs, companies and skills"
        size="hero"
        autofocus
        {counts}
        onOpenFilters={() => (filtersOpen = true)}
      />
    </div>

    {#if crafts.length > 0}
      {@render chipRow('Biggest specializations', crafts)}
    {/if}
    {#if places.length > 1}
      <!-- Only with real countries behind it: Remote alone is a constant, not a measurement. -->
      {@render chipRow('Locations', places)}
    {/if}
  </div>
</section>

<!-- Setup. Numbered cards in the order the steps have to happen; the first one ticks
     itself off once the visitor has an account. -->
<section class="border-t border-border py-12 sm:py-16" aria-labelledby="setup-heading">
  <div class="flex flex-col gap-2">
    <SectionLabel text="get started" />
    <h2 id="setup-heading" class="text-2xl font-semibold tracking-tight sm:text-3xl">
      Four steps, about five minutes.
    </h2>
  </div>
  <ol class="mt-8 grid grid-cols-1 gap-4 sm:grid-cols-2 lg:grid-cols-4">
    {#each steps as step, i (step.title)}
      {@const Icon = step.icon}
      <li
        class={[
          'flex flex-col gap-3 rounded-xl border bg-card p-5 transition-colors',
          step.done ? 'border-success-strong/40' : 'border-border hover:border-brand/50',
        ]}
      >
        <div class="flex items-center justify-between">
          <span class="flex size-9 items-center justify-center rounded-lg bg-brand-muted text-brand-strong">
            <Icon class="size-4.5" aria-hidden="true" />
          </span>
          {#if step.done}
            <span class="inline-flex items-center gap-1 text-xs font-medium text-success-strong">
              <Check class="size-3.5" aria-hidden="true" /> Done
            </span>
          {:else}
            <span class="text-xs font-medium tabular-nums text-muted-foreground" aria-hidden="true">0{i + 1}</span>
          {/if}
        </div>
        <h3 class="font-semibold tracking-tight">{step.title}</h3>
        <p class="text-sm leading-relaxed text-muted-foreground">{step.body}</p>
        {#if !step.done}
          <!-- eslint-disable svelte/no-navigation-without-resolve -- every step's href is a resolve() result -->
          <a
            href={step.href}
            class="group mt-auto inline-flex items-center gap-1.5 pt-1 text-sm font-medium text-brand-strong hover:underline hover:underline-offset-4"
          >
            {step.action}
            <ArrowRight class="size-4 transition-transform group-hover:translate-x-0.5" aria-hidden="true" />
          </a><!-- eslint-enable svelte/no-navigation-without-resolve -->
        {/if}
      </li>
    {/each}
  </ol>
</section>

{#if figures.length > 0}
  <!-- How big this install's catalogue is, from the one snapshot rollup-stats publishes;
       a figure it could not measure is dropped rather than shown as a zero. -->
  <section class="border-t border-border py-10">
    <SectionLabel as="h2" text="this install, today" />
    <dl class="mt-6 grid grid-cols-2 gap-x-6 gap-y-6 sm:grid-cols-3">
      {#each figures as f (f.label)}
        <!-- `flex-col-reverse` so the number reads first while the DOM keeps the term
             before its description. -->
        <div class="flex flex-col-reverse gap-1">
          <dt class="text-sm text-muted-foreground">{f.label}</dt>
          <dd class="text-3xl font-semibold tabular-nums tracking-tight sm:text-4xl">{f.value}</dd>
        </div>
      {/each}
    </dl>
  </section>
{/if}

<!-- What the agent is for, beside what it may not do. `grid-cols-1` is load-bearing: with
     no column template the single track below `lg` sizes to its widest content. -->
<section class="border-t border-border py-12 sm:py-16">
  <div class="grid grid-cols-1 gap-6 lg:grid-cols-2">
    <div class="flex flex-col gap-5 rounded-2xl border border-border bg-card p-6 sm:p-8">
      <div class="flex flex-col gap-2">
        <h2 class="text-2xl font-semibold tracking-tight">Ask in plain words.</h2>
        <p class="text-sm leading-relaxed text-muted-foreground">
          Your agent searches this catalogue, reads your CV, and keeps your board up to date — on the subscription you
          already pay for.
        </p>
      </div>
      <div class="[zoom:0.8]"><ChatExample /></div>
      <div class="mt-auto flex flex-wrap items-center gap-3">
        <Button href={resolve('/my/integrations/agent')} variant="cta" size="md" class="group">
          Connect your agent
          <ArrowRight class="size-4 transition-transform group-hover:translate-x-1" />
        </Button>
        <Button href={resolve('/how-it-works')} variant="ghost" size="md">How it works</Button>
      </div>
    </div>

    <div class="flex flex-col gap-6 rounded-2xl border border-border bg-card p-6 sm:p-8">
      <div class="flex flex-col gap-2">
        <h2 class="text-2xl font-semibold tracking-tight">Rules your agent cannot break.</h2>
        <p class="text-sm leading-relaxed text-muted-foreground">
          The server enforces these, whatever the agent is told. That is what makes it safe to hand it your CV.
        </p>
      </div>
      <ul class="flex flex-col gap-5">
        {#each rules as rule (rule.title)}
          {@const Icon = rule.icon}
          <li class="flex gap-4">
            <span class="flex size-9 shrink-0 items-center justify-center rounded-lg bg-brand-muted text-brand-strong">
              <Icon class="size-4.5" aria-hidden="true" />
            </span>
            <div class="flex flex-col gap-1">
              <h3 class="text-sm font-semibold">{rule.title}</h3>
              <p class="text-sm leading-relaxed text-muted-foreground">{rule.body}</p>
            </div>
          </li>
        {/each}
      </ul>
    </div>
  </div>
</section>

<RecentJobsFeed />

<!-- Seeded empty and applied by navigating: see `searchWithFilters`. -->
<FilterModal
  seed={emptyFilters()}
  {counts}
  {stagedCounts}
  {previewCount}
  open={filtersOpen}
  onClose={() => (filtersOpen = false)}
  applyLabel="Search jobs"
  onApply={searchWithFilters}
/>
