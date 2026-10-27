#!/usr/bin/env bash
# Builds installers into build/bin. Needs build/deps/ (run scripts/fetch-deps.sh first).
#   scripts/package.sh linux     -> marte-media_<ver>_amd64.deb + marte-media-<ver>.x86_64.rpm  (needs nfpm)
#   scripts/package.sh windows   -> marte-media-amd64-installer.exe                            (needs makensis)
set -euo pipefail

cd "$(dirname "$0")/.."
VERSION=$(python3 -c "import json;print(json.load(open('wails.json'))['info']['productVersion'])")
export VERSION

case "${1:-}" in
  linux)
    [ -x build/deps/linux-amd64/ffmpeg ] || { echo "run scripts/fetch-deps.sh first" >&2; exit 1; }
    wails build -tags webkit2_41
    nfpm package -f packaging/nfpm.yaml -p deb -t build/bin/
    nfpm package -f packaging/nfpm.yaml -p rpm -t build/bin/
    ;;
  windows)
    [ -f build/deps/windows-amd64/ffmpeg.exe ] || { echo "run scripts/fetch-deps.sh first" >&2; exit 1; }
    wails build -platform windows/amd64 -nsis
    ;;
  *) echo "usage: $0 linux|windows" >&2; exit 1 ;;
esac

ls -lh build/bin
