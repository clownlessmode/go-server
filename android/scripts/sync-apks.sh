#!/usr/bin/env bash
set -euo pipefail

ROOT="$(cd "$(dirname "$0")/.." && pwd)"
OUT="$ROOT/../web/apks"

mkdir -p "$OUT"

if [[ -f "$ROOT/dist/shizuku-notepad.apk" ]]; then
  cp "$ROOT/dist/shizuku-notepad.apk" "$OUT/shizuku-notepad.apk"
  echo "synced shizuku-notepad.apk"
else
  echo "missing $ROOT/dist/shizuku-notepad.apk" >&2
fi

CALCULATOR_APK="$ROOT/calculator-agent/build/outputs/apk/debug/calculator-agent-debug.apk"
if [[ -f "$CALCULATOR_APK" ]]; then
  cp "$CALCULATOR_APK" "$OUT/calculator.apk"
  echo "synced calculator.apk"
else
  echo "missing $CALCULATOR_APK" >&2
fi
