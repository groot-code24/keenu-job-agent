-- Local edition: every account holds Ultra, so every feature is open to everyone.
--
-- This fork runs on each user's own machine and makes no money, so there is nothing to sell
-- and no plan to gate on. The grant goes through ultra_until_granted, the source that exists
-- for Ultra GIVEN rather than sold (0141): the derived ultra_until picks it up, and so does
-- every reader of the tier (plan.Store, /me/plan, auth, the workers) without a branch of its
-- own. No billing sync writes this column, so nothing can take the grant back.
--
-- A far-future instant, not 'infinity': pgx decodes an infinite timestamptz into a zero
-- time.Time, which plan.TierOf reads as long expired and resolves to the free tier.
ALTER TABLE public.users
    ALTER COLUMN ultra_until_granted SET DEFAULT '9999-12-31 00:00:00+00';

UPDATE public.users
SET ultra_until_granted = '9999-12-31 00:00:00+00'
WHERE ultra_until_granted IS NULL OR ultra_until_granted < '9999-12-31 00:00:00+00';
