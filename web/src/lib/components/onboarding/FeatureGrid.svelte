<script lang="ts">
  // What a feature does, in four short lines — the body of the tour's tailoring and extension steps.
  import type { Component } from 'svelte';

  interface Feature {
    icon: Component<{ class?: string }>;
    title: string;
    body: string;
  }

  let { items }: { items: Feature[] } = $props();
</script>

<ul class="mt-6 grid gap-x-8 gap-y-4 sm:grid-cols-2">
  {#each items as f (f.title)}
    <li class="flex gap-2.5">
      <f.icon class="mt-0.5 size-4 shrink-0 text-brand-strong" />
      <span class="min-w-0">
        <span class="block text-sm font-medium">{f.title}</span>
        <span class="block text-xs leading-snug text-muted-foreground">{f.body}</span>
      </span>
    </li>
  {/each}
</ul>
