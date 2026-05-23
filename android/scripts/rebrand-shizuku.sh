#!/usr/bin/env bash
# Rebrand official Shizuku manager: change app name and launcher icon only.
# Requires: apktool, zipalign, apksigner (Android build-tools)
set -euo pipefail

ROOT="$(cd "$(dirname "$0")/.." && pwd)"
OUT_DIR="$ROOT/dist"
TMP_DIR="$(mktemp -d)"
SHIZUKU_VERSION="${SHIZUKU_VERSION:-v13.6.0}"
SHIZUKU_APK_NAME="shizuku-${SHIZUKU_VERSION}.r1086.2650830c-release.apk"
SHIZUKU_URL="https://github.com/RikkaApps/Shizuku/releases/download/${SHIZUKU_VERSION}/${SHIZUKU_APK_NAME}"
APP_LABEL="${APP_LABEL:-Блокнот}"
ICON_SRC="${ICON_SRC:-$ROOT/branding/notepad-icon.png}"

trap 'rm -rf "$TMP_DIR"' EXIT

command -v apktool >/dev/null || { echo "Install apktool: brew install apktool" >&2; exit 1; }
command -v zipalign >/dev/null || { echo "Install Android build-tools (zipalign)" >&2; exit 1; }
command -v apksigner >/dev/null || { echo "Install Android build-tools (apksigner)" >&2; exit 1; }

mkdir -p "$OUT_DIR"
SRC_APK="$TMP_DIR/shizuku-original.apk"
curl -L -o "$SRC_APK" "$SHIZUKU_URL"

DECODED="$TMP_DIR/decoded"
apktool d -f -o "$DECODED" "$SRC_APK"

# Rename app label in all values* strings.xml
find "$DECODED/res" -name 'strings.xml' -print0 | while IFS= read -r -d '' file; do
  if grep -q 'name="app_name"' "$file"; then
    perl -i -pe "s/(<string name=\"app_name\">).*?(<\/string>)/\${1}${APP_LABEL}\${2}/" "$file"
  fi
done

# Replace launcher icons if branding icon provided
if [[ -f "$ICON_SRC" ]]; then
  for dir in "$DECODED/res"/mipmap-*; do
    [[ -d "$dir" ]] || continue
    for icon in ic_launcher.webp ic_launcher_round.webp ic_launcher.png ic_launcher_round.png; do
      if [[ -f "$dir/$icon" ]]; then
        rm -f "$dir/$icon"
      fi
    done
    # apktool prefers png for replaced icons
    cp "$ICON_SRC" "$dir/ic_launcher.png"
    cp "$ICON_SRC" "$dir/ic_launcher_round.png"
  done
fi

BUILT="$TMP_DIR/built.apk"
apktool b -o "$BUILT" "$DECODED"

ALIGNED="$TMP_DIR/aligned.apk"
zipalign -f 4 "$BUILT" "$ALIGNED"

KEYSTORE="$ROOT/branding/debug.keystore"
if [[ ! -f "$KEYSTORE" ]]; then
  keytool -genkeypair -v \
    -keystore "$KEYSTORE" \
    -storepass android -keypass android \
    -alias androiddebugkey -keyalg RSA -keysize 2048 -validity 10000 \
    -dname "CN=Rebellion Debug"
fi

FINAL="$OUT_DIR/shizuku-notepad.apk"
apksigner sign \
  --ks "$KEYSTORE" \
  --ks-pass pass:android \
  --key-pass pass:android \
  --out "$FINAL" \
  "$ALIGNED"

echo "Built: $FINAL"
echo "Label: $APP_LABEL"
