// Company logos are served by a logo proxy named in PUBLIC_LOGO_ORIGIN (the public deployment
// runs one at https://logo.freehire.me, fronting a free favicon service that resolves the real
// brand mark for most company names). On a miss the proxy returns a clean 404 so each consumer
// falls back to its own placeholder (the SVG monogram in the SPA, a monogram tile in the OG
// card). Resolves by company name, since that is all most call sites carry.
//
// Unset, there are no logo URLs at all: every consumer shows its placeholder, and no company
// name a person looks at leaves their browser. That is the default for a self-hosted copy,
// which must not tell somebody else's server what its user is browsing.
//
// The origin is handed in (setLogoOrigin) rather than read from $app/env here, so this module
// stays importable outside the SvelteKit runtime — the docs and OG generators load it too.
let logoOrigin = '';

/** Points every logo URL at the proxy at `origin`; empty or absent turns logos off. Called
 *  once by each runtime's hook (hooks.server.ts, hooks.client.ts) with PUBLIC_LOGO_ORIGIN. */
export function setLogoOrigin(origin: string | undefined): void {
  logoOrigin = (origin ?? '').trim().replace(/\/+$/, '');
}

function logoBase(): string {
  return logoOrigin;
}

/** The proxy logo URL for a company name, or null when there is no name. */
export function companyLogoUrl(name: string): string | null {
  const base = logoBase();
  if (!name || !base) return null;
  return `${base}/${encodeURIComponent(name)}`;
}

/** The proxy logo URL for a SOURCE, resolved from its DISPLAY NAME — "Greenhouse",
 *  "BambooHR", "Telegram" — exactly as `companyLogoUrl` resolves a company.
 *
 *  Not from a host, which is what the first version of this did. The proxy resolves a
 *  brand from a name; a host is either a 404 (a per-tenant subdomain such as
 *  `jobs.smartrecruiters.com` or `2020companies.wd1.myworkdayjobs.com`) or, worse, the
 *  right image for the WRONG company, because an ATS posting's URL is often on the
 *  employer's own domain — production served Bankrate's mark for Greenhouse and ZEREN
 *  GROUP's for SuccessFactors. Measured against the live proxy on 2026-09-16: 14 of 15
 *  source names resolve; hosts were about half 404 and half wrong brand.
 *
 *  A miss still 404s, so every caller needs its own fallback — the proxy cannot tell the
 *  difference between "no such brand" and "not today". */
export function sourceLogoUrl(displayName: string, domain?: string): string | null {
  const base = logoBase();
  if (!displayName || !base) return null;
  const url = `${base}/${encodeURIComponent(displayName)}`;
  return domain ? `${url}?domain=${encodeURIComponent(domain)}` : url;
}

/** A source's own domain, for the sources whose logo the proxy cannot find — or could find
 *  wrongly — from the name alone.
 *
 *  An OVERRIDE, not a registry: a source missing from here resolves by name exactly as it
 *  does today, so the map going stale costs a logo rather than breaking one. That is the
 *  deliberate difference from the posting-host sampling this replaced, which was wrong for
 *  every source it covered.
 *
 *  Measured against the live proxy 2026-09-17: of the 145 sources carrying 500+ jobs, these
 *  14 had no logo by name, and all 14 resolve with the domain. `successfactors` settles why
 *  a nicer name was never the whole fix — its label was already right and the proxy still
 *  had nothing. The domain also removes a doubt a 200 cannot: a name resolved WRONGLY
 *  answers 200 with somebody else's mark (see cmd/publish-logo-domains and the `g2i` case).
 *
 *  Add an entry when a source's card shows the wrong mark or none; verify it resolves before
 *  committing, the same way these were. */
export const SOURCE_LOGO_DOMAINS: Record<string, string> = {
  successfactors: 'successfactors.com',
  zohorecruit: 'zoho.com',
  jazzhr: 'jazzhr.com',
  adpmyjobs: 'adp.com',
  applicantpro: 'applicantpro.com',
  '4dayweek': '4dayweek.io',
  isolvedhire: 'isolvedhcm.com',
  hrmdirect: 'hrmdirect.com',
  freshteam: 'freshworks.com',
  catsone: 'catsone.com',
  sber: 'sber.ru',
  alfabank: 'alfabank.ru',
  aijobs: 'aijobs.net',
  remotedotcom: 'remote.com',
  // By name the proxy answers 200 with another company's mark (checked 2026-10-30).
  arc: 'arc.dev',
};
