# Install (or update) and start Keenu Job Agent on Windows with one command, in PowerShell:
#
#   irm https://raw.githubusercontent.com/groot-code24/keenu-job-agent/main/scripts/install.ps1 | iex
#
# It puts the code in ~\keenu-job-agent (KEENU_DIR to choose another folder), with git when
# it is installed and as a plain download when it is not, then runs scripts\local-up.ps1,
# which needs only Docker Desktop. Running it again updates the code and restarts; this
# install's data and its .env stay as they are. The macOS/Linux twin is scripts/install.sh.
$ErrorActionPreference = 'Stop'
# Windows PowerShell redraws its progress bar for every chunk a download reads, which slows
# the archive download to a crawl; and older Windows needs TLS 1.2 named to reach GitHub.
$ProgressPreference = 'SilentlyContinue'
[Net.ServicePointManager]::SecurityProtocol = [Net.ServicePointManager]::SecurityProtocol -bor [Net.SecurityProtocolType]::Tls12

$repo = 'https://github.com/groot-code24/keenu-job-agent'
$dir = if ($env:KEENU_DIR) { $env:KEENU_DIR } else { Join-Path $HOME 'keenu-job-agent' }
$hasGit = [bool](Get-Command git -ErrorAction SilentlyContinue)

if (Test-Path (Join-Path $dir '.git')) {
  Write-Host "Updating $dir"
  git -C $dir pull --ff-only
  if ($LASTEXITCODE -ne 0) { throw "git pull failed in $dir" }
} elseif ($hasGit -and -not (Test-Path $dir)) {
  Write-Host "Downloading Keenu Job Agent into $dir"
  git clone --depth 1 "$repo.git" $dir
  if ($LASTEXITCODE -ne 0) { throw 'git clone failed' }
} else {
  # No git, or a folder an earlier git-less run made. The archive holds no .env, so
  # unpacking over the folder updates the code and keeps this install's secrets.
  Write-Host "Downloading Keenu Job Agent into $dir"
  $tmp = Join-Path ([IO.Path]::GetTempPath()) ("keenu-" + [guid]::NewGuid())
  New-Item -ItemType Directory $tmp | Out-Null
  try {
    $zip = Join-Path $tmp 'keenu.zip'
    Invoke-WebRequest "$repo/archive/refs/heads/main.zip" -OutFile $zip -UseBasicParsing
    Expand-Archive $zip $tmp
    New-Item -ItemType Directory -Force $dir | Out-Null
    Copy-Item -Recurse -Force (Join-Path $tmp 'keenu-job-agent-main\*') $dir
  } finally {
    Remove-Item -Recurse -Force $tmp
  }
}

# A child process, so the execution policy of this machine does not stop the script; the
# same PowerShell that runs this one, since `powershell` is not on every PATH.
& (Get-Process -Id $PID).Path -NoProfile -ExecutionPolicy Bypass -File (Join-Path $dir 'scripts\local-up.ps1')
if ($LASTEXITCODE -ne 0) { throw 'Keenu Job Agent did not start; see the messages above.' }
