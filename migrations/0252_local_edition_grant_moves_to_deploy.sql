-- The local edition's "every account holds Ultra" moves out of the schema and into the
-- deployment that wants it (deploy/local/local-edition.sql, applied by docker-compose.yml's
-- local-edition step after every migrate).
--
-- 0251 set the grant as the column's default, which made it true of EVERY database built from
-- these migrations — the integration tests' too, whose paywall and tier assertions need a new
-- account to start with nothing. A grant is a property of a deployment, not of the schema.
-- Rows 0251 already granted keep their grant; only the default goes.
ALTER TABLE public.users
    ALTER COLUMN ultra_until_granted DROP DEFAULT;
