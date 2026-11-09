#!/bin/sh
# Start the local edition: every service in Docker, then open the site.
#
#   ./scripts/local-up.sh
#
# Builds the images on first run (several minutes) and on any later run whose sources
# changed. Data lives in Docker volumes and survives `docker compose down`; only
# `docker compose down -v` erases it.
set -eu

cd "$(dirname "$0")/.."

origin="${WEB_ORIGIN:-http://localhost:8090}"

# Build the Go image and the web image one after the other: together they need more memory
# than an 8 GB laptop has spare, and an out-of-memory build fails without saying why.
export COMPOSE_PARALLEL_LIMIT="${COMPOSE_PARALLEL_LIMIT:-1}"
docker compose up -d --build

# Search answers 500 until the Meilisearch index exists, and nothing else creates it:
# routine changes reach the index through the drain, which assumes it is there.
# MSYS_NO_PATHCONV: Git Bash on Windows would otherwise rewrite the in-container path
# /app/reindex into a Windows one; everywhere else the variable means nothing.
MSYS_NO_PATHCONV=1 docker compose run --rm --entrypoint /app/reindex app

printf 'Waiting for %s ' "$origin"
i=0
until curl -fsS -o /dev/null "$origin/"; do
  i=$((i + 1))
  if [ "$i" -ge 60 ]; then
    printf '\nThe site did not answer after two minutes; run: docker compose logs app web\n' >&2
    exit 1
  fi
  printf '.'
  sleep 2
done

printf '\nfreehire is running at %s\n' "$origin"
