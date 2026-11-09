# Start the local edition on Windows: every service in Docker, then open the site.
#
#   powershell -ExecutionPolicy Bypass -File scripts\local-up.ps1
#
# The Windows twin of scripts/local-up.sh; keep the two doing the same steps. Builds the
# images on first run (several minutes) and on any later run whose sources changed. Data
# lives in Docker volumes and survives `docker compose down`; only `docker compose down -v`
# erases it.
$ErrorActionPreference = 'Stop'

Set-Location (Join-Path $PSScriptRoot '..')

$origin = if ($env:WEB_ORIGIN) { $env:WEB_ORIGIN } else { 'http://localhost:8090' }

# Build the Go image and the web image one after the other: together they need more memory
# than an 8 GB laptop has spare, and an out-of-memory build fails without saying why.
if (-not $env:COMPOSE_PARALLEL_LIMIT) { $env:COMPOSE_PARALLEL_LIMIT = '1' }
docker compose up -d --build
if ($LASTEXITCODE -ne 0) { throw 'docker compose up failed' }

# Search answers 500 until the Meilisearch index exists, and nothing else creates it:
# routine changes reach the index through the drain, which assumes it is there.
docker compose run --rm --entrypoint /app/reindex app
if ($LASTEXITCODE -ne 0) { throw 'creating the search index failed' }

Write-Host -NoNewline "Waiting for $origin "
$ready = $false
for ($i = 0; $i -lt 60; $i++) {
  try {
    Invoke-WebRequest "$origin/" -UseBasicParsing -TimeoutSec 5 | Out-Null
    $ready = $true
    break
  } catch {
    Write-Host -NoNewline '.'
    Start-Sleep -Seconds 2
  }
}
if (-not $ready) { throw 'The site did not answer after two minutes; run: docker compose logs app web' }

Write-Host "`nfreehire is running at $origin"
Start-Process $origin
