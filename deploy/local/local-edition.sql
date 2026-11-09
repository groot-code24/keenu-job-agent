-- The local edition sells nothing, so every account holds Ultra — every feature, no limit.
-- Applied by docker-compose.yml's `local-edition` step after every migrate; idempotent.
--
-- It lives here rather than in a migration because it is a property of THIS deployment, not
-- of the schema: the integration tests build their databases from migrations/, and their
-- paywall assertions need a new account to start with nothing (see migration 0252).
--
-- The grant goes through ultra_until_granted, the source for Ultra given rather than sold:
-- every reader of the tier picks it up through the derived ultra_until, and no billing sync
-- writes this column. A far-future instant, not 'infinity', which pgx decodes as a zero time.
ALTER TABLE public.users
    ALTER COLUMN ultra_until_granted SET DEFAULT '9999-12-31 00:00:00+00';

UPDATE public.users
SET ultra_until_granted = '9999-12-31 00:00:00+00'
WHERE ultra_until_granted IS NULL OR ultra_until_granted < '9999-12-31 00:00:00+00';
