#!/bin/sh
# Start Keenu Job Agent (the local edition): every service in Docker, then wait for the site.
#
#   ./scripts/local-up.sh
#
# It pulls the images the repository publishes (.github/workflows/images.yml) and builds them
# locally only when it cannot pull them, when KEENU_BUILD=1, or when WEB_ORIGIN is not the
# default (the web image bakes its origin in at build time). Data lives in Docker volumes and
# survives `docker compose down`; only `docker compose down -v` erases it.
#
# The Windows twin is scripts/local-up.ps1; keep the two doing the same steps.
set -eu

cd "$(dirname "$0")/.."

default_origin="http://localhost:8090"
origin="${WEB_ORIGIN:-$default_origin}"

if ! command -v docker >/dev/null 2>&1; then
  echo "Docker is not installed. Install Docker Desktop (https://www.docker.com/products/docker-desktop/) and run this again." >&2
  exit 1
fi
# Installed but not running is the most common first-run failure. On macOS start it rather
# than send the person off to do it; a Linux daemon needs root, so that stays a message.
if ! docker info >/dev/null 2>&1 && [ "$(uname -s)" = Darwin ] && [ -d /Applications/Docker.app ]; then
  printf 'Starting Docker Desktop '
  open -a Docker
  i=0
  until docker info >/dev/null 2>&1; do
    i=$((i + 1))
    [ "$i" -ge 90 ] && break
    printf '.'
    sleep 2
  done
  echo
fi
if ! docker info >/dev/null 2>&1; then
  echo "Docker is not running. Start Docker Desktop (or the Docker daemon) and run this again." >&2
  exit 1
fi

# Secrets of this install's own. docker-compose.yml falls back to well-known development values,
# which would let anyone who knows them forge a session for an account on this machine.
random_hex() { od -An -N32 -tx1 /dev/urandom | tr -d ' \n'; }
touch .env
grep -q '^JWT_SECRET=' .env || printf 'JWT_SECRET=%s\n' "$(random_hex)" >>.env
grep -q '^MEILI_MASTER_KEY=' .env || printf 'MEILI_MASTER_KEY=%s\n' "$(random_hex)" >>.env

build=0
if [ "${KEENU_BUILD:-0}" = 1 ] || [ "$origin" != "$default_origin" ]; then
  build=1
elif ! docker compose pull app web; then
  echo "Could not pull the published images; building them here (20-50 minutes the first time)." >&2
  build=1
fi
if [ "$build" = 1 ]; then
  # One image at a time: together the Go and web builds need more memory than an 8 GB
  # laptop has spare, and an out-of-memory build fails without saying why.
  docker compose build app
  docker compose build web
fi
docker compose up -d --no-build

# Search answers 500 until the Meilisearch index exists. Routine changes reach it through the
# drain, which assumes it is there, so it is created once — not rebuilt on every start.
meili_key=$(sed -n 's/^MEILI_MASTER_KEY=//p' .env | tail -n 1)
i=0
until code=$(curl -s -o /dev/null -w '%{http_code}' -H "Authorization: Bearer $meili_key" "http://127.0.0.1:${MEILI_HOST_PORT:-7700}/indexes/jobs") && [ "$code" != 000 ]; do
  i=$((i + 1))
  [ "$i" -ge 30 ] && break
  sleep 2
done
if [ "${code:-000}" = 404 ]; then
  # MSYS_NO_PATHCONV: Git Bash on Windows would otherwise rewrite the in-container path.
  MSYS_NO_PATHCONV=1 docker compose run --rm --entrypoint /app/reindex app
fi

printf 'Waiting for %s ' "$origin"
i=0
until curl -fsS -o /dev/null "$origin/health"; do
  i=$((i + 1))
  if [ "$i" -ge 90 ]; then
    printf '\nThe site did not answer after three minutes; run: docker compose logs app web\n' >&2
    exit 1
  fi
  printf '.'
  sleep 2
done

printf '\nKeenu Job Agent is running at %s\n' "$origin"
printf 'Connect your AI agent: %s/my/integrations/agent\n' "$origin"
printf "Stop it with 'docker compose down' in %s (your data is kept).\n" "$(pwd)"
if [ -z "${KEENU_NO_BROWSER:-}" ]; then
  case "$(uname -s)" in
    Darwin) open "$origin" ;;
    *) if command -v xdg-open >/dev/null 2>&1; then xdg-open "$origin" >/dev/null 2>&1 || true; fi ;;
  esac
fi

