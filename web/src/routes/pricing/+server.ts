import { redirect } from '@sveltejs/kit';
import type { RequestHandler } from './$types';

// The local edition sells nothing: every account holds Ultra (migration 0251), so there is no
// plan to choose between. The route stays so the links and bookmarks that name it still land
// somewhere useful — the caller's own plan page, which shows what they have.
export const GET: RequestHandler = () => redirect(308, '/my/plan');
