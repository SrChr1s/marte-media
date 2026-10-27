#!/usr/bin/env bash
# Downloads the binaries that ship inside the installers into build/deps/<os>-<arch>/:
#   yt-dlp  (Unlicense)  from the official release, checked against SHA2-256SUMS
#   ffmpeg  (LGPL build) from BtbN/FFmpeg-Builds, checked against its checksums.sha256
#   deno    (MIT)        from denoland/deno, checked against its .sha256sum. yt-dlp needs a JavaScript
#                        runtime to extract from YouTube, and deno is the only one it enables by default.
# Usage: scripts/fetch-deps.sh [linux-amd64] [windows-amd64]   (default: both)
#        DEPS="deno" scripts/fetch-deps.sh                      (only some of: ytdlp ffmpeg deno)
set -euo pipefail

cd "$(dirname "$0")/.."
YTDLP=https://github.com/yt-dlp/yt-dlp/releases/latest/download
FFMPEG=https://github.com/BtbN/FFmpeg-Builds/releases/download/latest
DENO=https://github.com/denoland/deno/releases/latest/download

targets=("$@")
[ ${#targets[@]} -eq 0 ] && targets=(linux-amd64 windows-amd64)
wanted() { [[ " ${DEPS:-ytdlp ffmpeg deno} " == *" $1 "* ]]; }

tmp=$(mktemp -d)
trap 'rm -rf "$tmp"' EXIT

fetch() { curl -fsSL --retry 3 -o "$2" "$1"; }

# verify <sums-file> <file> <name-in-sums>
verify() {
  local want
  want=$(awk -v n="$3" '{f=$2; sub(/^\*/,"",f); sub(/^\.\//,"",f)} f==n {print $1; exit}' "$1")
  [ -n "$want" ] || { echo "no checksum listed for $3" >&2; exit 1; }
  echo "$want  $2" | sha256sum -c --status - || { echo "checksum mismatch: $3" >&2; exit 1; }
}

wanted ytdlp  && fetch "$YTDLP/SHA2-256SUMS" "$tmp/ytdlp.sums"
wanted ffmpeg && fetch "$FFMPEG/checksums.sha256" "$tmp/ffmpeg.sums"

for t in "${targets[@]}"; do
  out="build/deps/$t"
  mkdir -p "$out"
  case "$t" in
    linux-amd64)
      ytdlp_asset=yt-dlp_linux;  ytdlp_dst=yt-dlp
      ff_asset=ffmpeg-master-latest-linux64-lgpl.tar.xz
      deno_asset=deno-x86_64-unknown-linux-gnu.zip; deno_dst=deno ;;
    windows-amd64)
      ytdlp_asset=yt-dlp.exe;    ytdlp_dst=yt-dlp.exe
      ff_asset=ffmpeg-master-latest-win64-lgpl.zip
      deno_asset=deno-x86_64-pc-windows-msvc.zip; deno_dst=deno.exe ;;
    *) echo "unknown target: $t" >&2; exit 1 ;;
  esac

  if wanted ytdlp; then
    echo "==> $t: yt-dlp"
    fetch "$YTDLP/$ytdlp_asset" "$tmp/$ytdlp_asset"
    verify "$tmp/ytdlp.sums" "$tmp/$ytdlp_asset" "$ytdlp_asset"
    install -m 0755 "$tmp/$ytdlp_asset" "$out/$ytdlp_dst"
  fi

  if wanted ffmpeg; then
    echo "==> $t: ffmpeg"
    fetch "$FFMPEG/$ff_asset" "$tmp/$ff_asset"
    verify "$tmp/ffmpeg.sums" "$tmp/$ff_asset" "$ff_asset"
    x="$tmp/x-$t"; mkdir -p "$x"
    case "$ff_asset" in
      *.tar.xz) tar -xJf "$tmp/$ff_asset" -C "$x" --wildcards '*/bin/ffmpeg' '*/LICENSE.txt' ;;
      *.zip)    unzip -q -j "$tmp/$ff_asset" '*/bin/ffmpeg.exe' '*/LICENSE.txt' -d "$x/flat" ;;
    esac
    ffbin=$(find "$x" -name 'ffmpeg' -o -name 'ffmpeg.exe' | head -n1)
    lic=$(find "$x" -name 'LICENSE.txt' | head -n1)
    install -m 0755 "$ffbin" "$out/$(basename "$ffbin")"
    cp "$lic" "$out/FFMPEG-LICENSE.txt"
  fi

  if wanted deno; then
    echo "==> $t: deno"
    fetch "$DENO/$deno_asset" "$tmp/$deno_asset"
    fetch "$DENO/$deno_asset.sha256sum" "$tmp/$deno_asset.sha256sum"
    # Linux ships "<hash>  <name>"; the Windows asset is PowerShell's Get-FileHash output ("Hash : <hash>", CRLF).
    want=$(tr -d '\r' < "$tmp/$deno_asset.sha256sum" | awk '/^Hash/ {print tolower($3); exit} $1 ~ /^[0-9a-fA-F]{64}$/ {print tolower($1); exit}')
    [ -n "$want" ] || { echo "no checksum for $deno_asset" >&2; exit 1; }
    echo "$want  $tmp/$deno_asset" | sha256sum -c --status - || { echo "checksum mismatch: $deno_asset" >&2; exit 1; }
    unzip -q -o -j "$tmp/$deno_asset" "$deno_dst" -d "$tmp/x-deno-$t"
    install -m 0755 "$tmp/x-deno-$t/$deno_dst" "$out/$deno_dst"
  fi
done

echo "done:"; ls -lh build/deps/*
