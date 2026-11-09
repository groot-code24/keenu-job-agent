import template from '#lib/server/agentSetup.md?raw';
import type { RequestHandler } from './$types';

// The setup instructions a terminal agent fetches (AgentSetupButton) name the MCP server to
// connect to. They used to be a static file naming the public deployment outright, so a
// self-hosted copy — the local edition on localhost — sent every agent set up from its own
// pages to somebody else's server. The text is unchanged; the public origin in it is replaced
// by the origin this request reached, which on the public deployment is the same string.
const PUBLIC_ORIGIN = 'https://freehire.me';

export const GET: RequestHandler = ({ url }) =>
  new Response(template.replaceAll(PUBLIC_ORIGIN, url.origin), {
    headers: { 'content-type': 'text/markdown; charset=utf-8', 'cache-control': 'public, max-age=300' },
  });
