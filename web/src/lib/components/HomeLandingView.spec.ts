import { render, screen } from '@testing-library/svelte';
import { describe, expect, it, vi } from 'vitest';
import type { CatalogScale } from '#lib/types.js';

vi.mock('$app/paths', () => ({ resolve: (p: string) => p }));
vi.mock('$app/navigation', () => ({
  goto: vi.fn(),
  afterNavigate: vi.fn(),
  beforeNavigate: vi.fn(),
  onNavigate: vi.fn(),
}));
vi.mock('#lib/api.js', () => ({
  api: {
    facetCounts: vi.fn(() => new Promise(() => {})),
    recentJobs: vi.fn(() => new Promise(() => {})),
    searchJobs: vi.fn(() => new Promise(() => {})),
  },
}));
// The live feed opens an SSE stream; it is not what these tests are about.
vi.mock('./RecentJobsFeed.svelte', () => ({ default: () => {} }));
vi.stubGlobal(
  'IntersectionObserver',
  class {
    observe() {}
    unobserve() {}
    disconnect() {}
  },
);

import HomeLandingView from './HomeLandingView.svelte';

const scale = { unique_open_jobs: 5_212_340, companies: 700_009, sources: 437, ats_platforms: 178 } as CatalogScale;

describe('HomeLandingView hero', () => {
  it('quotes the live catalogue size, compacted', () => {
    render(HomeLandingView, { props: { counts: null, scale } });
    expect(screen.getByText(/5\.2M\+ live postings/)).toBeTruthy();
  });

  // A self-hosted catalogue holds thousands, not millions: flooring to a fixed tenth of a
  // million read "0+ live postings" there.
  it('floors a small catalogue to two significant figures, never to zero', () => {
    render(HomeLandingView, { props: { counts: null, scale: { ...scale, unique_open_jobs: 6_358 } } });
    expect(screen.getByText(/6(\.3K|,300)\+ live postings/)).toBeTruthy();
    expect(screen.queryByText(/\b0\+ live postings/)).toBeNull();
  });

  it('floors rather than rounds, so the "+" never overstates', () => {
    render(HomeLandingView, { props: { counts: null, scale: { ...scale, unique_open_jobs: 4_968_000 } } });
    expect(screen.getByText(/4\.9M\+ live postings/)).toBeTruthy();
  });

  // A cold snapshot must not make the hero say nothing, or overstate: the fallback is a floor.
  // Before rollup-stats has published a figure the line says so in words: quoting another
  // deployment's catalogue size would overstate a self-hosted one by orders of magnitude.
  it('says it in words, with no figure, when the snapshot is missing', () => {
    render(HomeLandingView, { props: { counts: null, scale: null } });
    expect(screen.getByText(/Thousands of live postings/)).toBeTruthy();
    expect(screen.queryByText(/4\.9M/)).toBeNull();
  });

  it('links the agent and API doors from the hero sentence', () => {
    render(HomeLandingView, { props: { counts: null, scale: null } });
    expect(screen.getByRole('link', { name: 'your AI agent' }).getAttribute('href')).toBe('/agents');
    expect(screen.getByRole('link', { name: 'the API' }).getAttribute('href')).toBe('/jobs-api');
  });
});
