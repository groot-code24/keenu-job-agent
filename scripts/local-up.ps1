# Start Keenu Job Agent (the local edition) on Windows: every service in Docker, then the site.
#
#   powershell -ExecutionPolicy Bypass -File scripts\local-up.ps1
#
# The Windows twin of scripts/local-up.sh; keep the two doing the same steps. It pulls the
# images the repository publishes and builds them locally only when it cannot pull them, when
# KEENU_BUILD=1, or when WEB_ORIGIN is not the default (the web image bakes its origin in).
# Data lives in Docker volumes and survives `docker compose down`; only `-v` erases it.
$ErrorActionPreference = 'Stop'

Set-Location (Join-Path $PSScriptRoot '..')

$defaultOrigin = 'http://localhost:8090'
$origin = if ($env:WEB_ORIGIN) { $env:WEB_ORIGIN } else { $defaultOrigin }

docker info *> $null
if ($LASTEXITCODE -ne 0) { throw 'Docker is not running. Start Docker Desktop and run this again.' }

# Secrets of this install's own: docker-compose.yml falls back to well-known development
# values, which would let anyone who knows them forge a session on this machine.
function New-Secret {
  $bytes = New-Object byte[] 32
  [System.Security.Cryptography.RandomNumberGenerator]::Create().GetBytes($bytes)
  -join ($bytes | ForEach-Object { $_.ToString('x2') })
}
if (-not (Test-Path .env)) { New-Item -ItemType File .env | Out-Null }
$envText = Get-Content .env -Raw
if ($null -eq $envText) { $envText = '' }
foreach ($name in 'JWT_SECRET', 'MEILI_MASTER_KEY') {
  if ($envText -notmatch "(?m)^$name=") { Add-Content .env "$name=$(New-Secret)" -Encoding ascii }
}

$build = $false
if ($env:KEENU_BUILD -eq '1' -or $origin -ne $defaultOrigin) {
  $build = $true
} else {
  docker compose pull app web
  if ($LASTEXITCODE -ne 0) {
    Write-Warning 'Could not pull the published images; building them here (20-50 minutes the first time).'
    $build = $true
  }
}
if ($build) {
  # One image at a time: together the Go and web builds need more memory than an 8 GB
  # laptop has spare, and an out-of-memory build fails without saying why.
  docker compose build app
  if ($LASTEXITCODE -ne 0) { throw 'building the app image failed' }
  docker compose build web
  if ($LASTEXITCODE -ne 0) { throw 'building the web image failed' }
}
docker compose up -d --no-build
if ($LASTEXITCODE -ne 0) { throw 'docker compose up failed' }

# Search answers 500 until the Meilisearch index exists; it is created once, not rebuilt on
# every start (routine changes reach it through the drain).
$meiliKey = ((Get-Content .env | Where-Object { $_ -match '^MEILI_MASTER_KEY=' }) | Select-Object -Last 1) -replace '^MEILI_MASTER_KEY=', ''
$meiliPort = if ($env:MEILI_HOST_PORT) { $env:MEILI_HOST_PORT } else { '7700' }
$status = 0
for ($i = 0; $i -lt 30; $i++) {
  try {
    Invoke-WebRequest "http://127.0.0.1:$meiliPort/indexes/jobs" -Headers @{ Authorization = "Bearer $meiliKey" } -UseBasicParsing -TimeoutSec 5 | Out-Null
    $status = 200; break
  } catch {
    if ($_.Exception.Response) { $status = [int]$_.Exception.Response.StatusCode; break }
    Start-Sleep -Seconds 2
  }
}
if ($status -eq 404) {
  docker compose run --rm --entrypoint /app/reindex app
  if ($LASTEXITCODE -ne 0) { throw 'creating the search index failed' }
}

Write-Host -NoNewline "Waiting for $origin "
$ready = $false
for ($i = 0; $i -lt 90; $i++) {
  try {
    Invoke-WebRequest "$origin/health" -UseBasicParsing -TimeoutSec 5 | Out-Null
    $ready = $true; break
  } catch {
    Write-Host -NoNewline '.'
    Start-Sleep -Seconds 2
  }
}
if (-not $ready) { throw 'The site did not answer after three minutes; run: docker compose logs app web' }

Write-Host "`nKeenu Job Agent is running at $origin"
Write-Host "Connect your AI agent: $origin/my/integrations/agent"
if (-not $env:KEENU_NO_BROWSER) { Start-Process $origin }
