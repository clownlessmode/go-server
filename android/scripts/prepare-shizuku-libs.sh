#!/usr/bin/env bash
set -euo pipefail

# Copies Shizuku native libs from an official manager APK into notepad-shizuku.
# Usage: ./scripts/prepare-shizuku-libs.sh /path/to/shizuku-manager.apk

if [[ $# -ne 1 ]]; then
  echo "Usage: $0 /path/to/shizuku-manager.apk" >&2
  exit 1
fi

APK="$1"
ROOT="$(cd "$(dirname "$0")/.." && pwd)"
DEST="$ROOT/notepad-shizuku/src/main/jniLibs"
TMP="$(mktemp -d)"

trap 'rm -rf "$TMP"' EXIT

unzip -q "$APK" "lib/*" -d "$TMP"
rm -rf "$DEST"
mkdir -p "$DEST"
cp -R "$TMP/lib/"* "$DEST"

echo "Copied Shizuku libs to $DEST"
