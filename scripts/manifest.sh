#!/bin/sh
set -eu

root="$(CDPATH= cd -- "$(dirname -- "$0")/.." && pwd)"
mkdir -p "$root/dist"
generated="$(date -u +%Y-%m-%dT%H:%M:%SZ)"
image_status="not_built"
[ -f "$root/dist/climasense-os-rpi-zero-2-w.img" ] && image_status="built"
joss_version="$(cat "$root/dist/joss-linux-arm64.version" 2>/dev/null || printf 'unknown')"

{
cat <<EOF
{
  "schema_version": 1,
  "project": "ClimaSense AI",
  "generated_at": "$generated",
  "architecture": "linux-arm64",
  "buildroot": "$(cat "$root/os/buildroot/buildroot.version")",
  "joss": "$joss_version",
  "image_status": "$image_status",
  "hardware_validation": "pending",
  "artifacts": [
EOF
first=1
find "$root/dist" -type f ! -name '*.sha256' ! -name 'manifest.json' | sort | while IFS= read -r artifact; do
  relative="${artifact#"$root/dist/"}"
  bytes="$(wc -c < "$artifact" | tr -d ' ')"
  digest="$(sha256sum "$artifact" | awk '{print $1}')"
  [ "$first" = "1" ] || printf ',\n'
  first=0
  printf '    {"path":"%s","bytes":%s,"sha256":"%s"}' "$relative" "$bytes" "$digest"
done
printf '\n  ]\n}\n'
} > "$root/dist/manifest.json"
(cd "$root/dist" && sha256sum manifest.json > manifest.json.sha256)
