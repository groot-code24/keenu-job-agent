import { readFileSync } from 'node:fs';
import { join } from 'node:path';
import { describe, expect, it } from 'vitest';

// A SOURCE-TEXT AUDIT, not a mounted-component test — web/ has no component-test
// infrastructure at all (see jobActionStrip.test.ts's own comment: no Svelte
// plugin, no DOM in vitest.config.ts).
const SOURCE = readFileSync(join(import.meta.dirname, 'Footer.svelte'), 'utf8');

// The offset of the `{/if}` that closes the `{#if !compact}` block: before it is the
// link columns the compact footer drops, after it the bottom bar every footer keeps.
const GUARD_OPEN = SOURCE.indexOf('{#if !compact}');
const GUARD_END = SOURCE.indexOf('{/if}', GUARD_OPEN);

describe('Footer compact mode', () => {
  it('declares a compact prop defaulting to false', () => {
    expect(SOURCE).toMatch(/let\s*\{\s*compact\s*=\s*false\s*\}/);
  });

  it('renders the link groups only inside the {#if !compact} guard', () => {
    const groupsAt = SOURCE.indexOf('{#each groups as group');
    expect(GUARD_OPEN, 'expected an {#if !compact} guard').toBeGreaterThan(-1);
    expect(groupsAt).toBeGreaterThan(GUARD_OPEN);
    expect(groupsAt).toBeLessThan(GUARD_END);
  });

  it('still renders the bottom bar (copyright, cookie settings, source link) unconditionally', () => {
    const bottomBar = SOURCE.slice(GUARD_END);
    expect(bottomBar).toContain('©');
    expect(bottomBar).toContain('Cookie settings');
    expect(bottomBar).toContain('REPOSITORY_URL');
  });

  it("only borders the bottom bar's own top when compact (no divider from a section that isn't rendered)", () => {
    const commentAt = SOURCE.indexOf('Bottom bar:');
    const innerDivAt = SOURCE.indexOf('mx-auto flex max-w-6xl flex-col');
    expect(SOURCE.slice(commentAt, innerDivAt)).toContain('compact ?');
  });
});

describe('Footer on a local install', () => {
  it('loads nothing from another server: no hosted badge or image', () => {
    // The Product Hunt badge was an <img> from api.producthunt.com on every page view.
    expect(SOURCE).not.toMatch(/<img\b/);
    expect(SOURCE).not.toContain('producthunt');
  });

  it('offers no store listing: those extensions and apps talk to the public deployment, not this install', () => {
    for (const store of ['chromewebstore.google.com', 'addons.mozilla.org', 'apps.apple.com', 'EXTENSION_STORE_URL']) {
      expect(SOURCE).not.toContain(store);
    }
  });

  it('links its own repository and credits the project it is built from', () => {
    expect(SOURCE).toContain("'https://github.com/groot-code24/keenu-job-agent'");
    expect(SOURCE).toContain("'https://github.com/strelov1/freehire'");
  });

  it('opens external links in a new tab with a safe rel', () => {
    for (const match of SOURCE.matchAll(/href=\{(?:REPOSITORY_URL|UPSTREAM_URL)\}([\s\S]*?)>/g)) {
      expect(match[1]).toContain('target="_blank"');
      expect(match[1]).toContain('rel="noopener noreferrer"');
    }
  });
});
