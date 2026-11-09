#!/bin/sh
# Install (or update) and start Keenu Job Agent on macOS or Linux with one command:
#
#   curl -fsSL https://raw.githubusercontent.com/groot-code24/keenu-job-agent/main/scripts/install.sh | sh
#
# It puts the code in ~/keenu-job-agent (KEENU_DIR to choose another folder), with git when
# it is installed and as a plain download when it is not, then runs scripts/local-up.sh,
# which needs only Docker. Running it again updates the code and restarts; this install's
# data and its .env stay as they are. The Windows twin is scripts/install.ps1.
set -eu

repo="https://github.com/groot-code24/keenu-job-agent"
dir="${KEENU_DIR:-$HOME/keenu-job-agent}"

if [ -d "$dir/.git" ]; then
  echo "Updating $dir"
  git -C "$dir" pull --ff-only
elif [ ! -e "$dir" ] && command -v git >/dev/null 2>&1; then
  echo "Downloading Keenu Job Agent into $dir"
  git clone --depth 1 "$repo.git" "$dir"
else
  # No git, or a folder an earlier git-less run made. The archive holds no .env, so
  # unpacking over the folder updates the code and keeps this install's secrets.
  echo "Downloading Keenu Job Agent into $dir"
  mkdir -p "$dir"
  curl -fsSL "$repo/archive/refs/heads/main.tar.gz" | tar -xz -C "$dir" --strip-components=1
fi

exec sh "$dir/scripts/local-up.sh"
