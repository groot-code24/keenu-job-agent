import { describe, expect, it } from 'vitest';
import template from '#lib/server/agentSetup.md?raw';
import { GET } from './+server';

async function fetchFrom(origin: string): Promise<string> {
  const res = await GET({ url: new URL(`${origin}/agent-setup.md`) } as Parameters<typeof GET>[0]);
  expect(res.headers.get('content-type')).toContain('text/markdown');
  return res.text();
}

describe('GET /agent-setup.md', () => {
  // A self-hosted copy must send the agent to ITSELF: the static file this replaced named the
  // public deployment, so every agent set up from a local install connected elsewhere.
  it('names the server it was fetched from', async () => {
    const text = await fetchFrom('http://localhost:8090');
    expect(text).toContain('http://localhost:8090/api/v1/mcp');
    expect(text).not.toContain('https://freehire.me');
  });

  it('serves the public deployment its text unchanged', async () => {
    expect(await fetchFrom('https://freehire.me')).toBe(template);
  });
});
