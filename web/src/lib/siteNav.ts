// The site's top-level destinations, named once.
//
// Two navigations lead to them: the menu behind the header's burger (HeaderMenu),
// which lists all of them, and the homepage header's own row (TopBar), which on that
// one route stands in for the search box and shows a subset. They were two arrays, and
// the last time a destination moved — the feed, from `/` to `/jobs` — the menu's copy
// went on pointing at the old one until someone read both lists side by side. One list
// is how that stops being possible.
//
// A destination carries its own glyph, the way accountNavIcons pairs one to each
// account section, so the two navigations cannot draw the same page with different
// marks.

import {
  Bell,
  Bot,
  Briefcase,
  Building2,
  ChartColumn,
  Compass,
  Info,
  Layers,
  MessagesSquare,
  Radar,
  TrendingUp,
  Unlock,
  Wand,
} from '@lucide/svelte';
import type { LucideIcon } from '@lucide/svelte';
import type { Pathname } from '#lib/pathname.js';

export type SiteNavItem = {
  /** SvelteKit's own union of this app's route paths, not a string — so a destination
   *  that stops existing, or was never spelled right, is a compile error rather than a
   *  link that 404s. `satisfies` below checks each entry against it while keeping the
   *  literal `resolve()` needs. */
  href: Pathname;
  label: string;
  icon: LucideIcon;
};

export const NAV = {
  // The catalogue itself — what a visitor came to walk. Talent is the one that reads
  // the other way round: the anonymous candidate catalogue, for whoever is hiring.
  jobs: { href: '/jobs', label: 'Jobs', icon: Briefcase },
  companies: { href: '/companies', label: 'Companies', icon: Building2 },
  collections: { href: '/collections', label: 'Collections', icon: Layers },
  talent: { href: '/talent', label: 'Talent Network', icon: Radar },
  // The local edition's own front door: where the person connects Claude Code or Codex
  // to this install. Signed out, /my sends them to sign in first and back here.
  agent: { href: '/my/integrations/agent', label: 'Your agent', icon: Bot },

  // What this is and how it works — what a first-time visitor reads.
  howItWorks: { href: '/how-it-works', label: 'How it works', icon: Compass },
  about: { href: '/about', label: 'About', icon: Info },
  // The open-startup transparency page — already linked from the footer, but until
  // now unreachable from the header menu at all.
  open: { href: '/open', label: 'Open', icon: Unlock },

  // What the product does beyond listing jobs.
  cvTailoring: { href: '/features/tailor', label: 'CV tailoring', icon: Wand },
  jobNotifications: { href: '/features/notifications', label: 'Job notifications', icon: Bell },

  // What the catalogue says about the market.
  analytics: { href: '/analytics', label: 'Analytics', icon: ChartColumn },
  trends: { href: '/trends', label: 'Trends', icon: TrendingUp },
  discussions: { href: '/discussions', label: 'Discussions', icon: MessagesSquare },
} as const satisfies Record<string, SiteNavItem>;

/** What the homepage header shows where every other page shows the search box.
 *
 *  A shortcut to the menu's own top, not a second navigation with its own opinions
 *  about what matters. On a local install that is the catalogue and the agent that works
 *  it: Talent Network (a recruiter-facing directory of the public deployment) and About
 *  (who publishes that deployment) stay in the menu, not in the bar.
 *
 *  Five fit with room to spare at 1024px — `lg`, where the whole row first appears, and
 *  therefore the tightest case. Check at `lg` before adding one: the row has no overflow
 *  behaviour, so passing the bound crowds the controls on its right silently.
 *
 *  TopBar decides when the row is drawn — and, below `lg`, how much of it: it takes
 *  entries from the FRONT of this list, so the ORDER here is what a narrow screen gets.
 *  Jobs and the agent are what a phone shows. */
export const HEADER_LINKS = [NAV.jobs, NAV.agent, NAV.companies, NAV.collections, NAV.howItWorks] as const;
