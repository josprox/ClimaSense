#!/bin/sh
set -eu

root="$(CDPATH= cd -- "$(dirname -- "$0")/.." && pwd)"
joss="$root/dist/tools/joss"
work="$(mktemp -d)"
server_pid=""

cleanup() {
  if [ -n "$server_pid" ]; then
    kill "$server_pid" 2>/dev/null || true
    wait "$server_pid" 2>/dev/null || true
  fi
  rm -rf "$work"
}
trap cleanup EXIT HUP INT TERM

mkdir -p "$work/app/views" \
  "$work/plugins/climasense_hardware/0.1.1" \
  "$work/plugins/climasense_transport/0.2.0"
cp "$root/main.joss" "$root/onboarding.joss" "$root/routes.joss" "$root/joss.yaml" "$work/"
cp -R "$root/src" "$work/"
cp "$root/src/views/onboarding.joss.html" "$work/app/views/onboarding.joss.html"
cp "$root/dist/plugins/climasense_hardware.jp" \
  "$work/plugins/climasense_hardware/0.1.1/climasense_hardware.jp"
cp "$root/dist/plugins/climasense_transport.jp" \
  "$work/plugins/climasense_transport/0.2.0/climasense_transport.jp"
sed 's/^PORT=.*/PORT="18080"/' "$root/env.example" >"$work/.env"
printf '\nCLIMASENSE_DATA_DIR="%s/data"\nDEVICE_TOKEN_FILE="%s/data/device.token"\n' \
  "$work" "$work" >>"$work/.env"
mkdir -p "$work/data"
printf 'network={}\n' >"$work/data/wpa_supplicant.conf"

(
  cd "$work"
  exec "$joss" server start
) >"$work/portal.log" 2>&1 &
server_pid=$!

attempts=0
while [ "$attempts" -lt 15 ]; do
  if curl -fs http://127.0.0.1:18080/ >"$work/index.html"; then
    break
  fi
  kill -0 "$server_pid" 2>/dev/null || { cat "$work/portal.log"; exit 1; }
  sleep 1
  attempts=$((attempts + 1))
done

grep -q 'ClimaSense Edge' "$work/index.html"
curl -fsS http://127.0.0.1:18080/api/setup/status >"$work/status.json"
grep -q '"ok":true' "$work/status.json"
curl -fsS http://127.0.0.1:18080/api/wifi/scan >"$work/scan.json"
grep -q '"ok":true' "$work/scan.json"
grep -q '"networks":' "$work/scan.json"

status="$(curl -sS -o "$work/connect.json" -w '%{http_code}' \
  -H 'Content-Type: application/json' -d '{}' \
  http://127.0.0.1:18080/api/wifi/connect)"
[ "$status" = 422 ] || { cat "$work/connect.json"; exit 1; }
grep -q '"error":' "$work/connect.json"

status="$(curl -sS -o "$work/complete.json" -w '%{http_code}' \
  -H 'Content-Type: application/json' \
  -d '{"code":"TEST-CODE","ap_ssid":"ClimaSense Test","ap_pass":"testpass123"}' \
  http://127.0.0.1:18080/api/setup/complete)"
[ "$status" = 202 ] || { cat "$work/complete.json"; exit 1; }
python3 - "$work/complete.json" <<'PY'
import json
import sys

with open(sys.argv[1], encoding="utf-8") as response:
    payload = json.load(response)
assert payload["ok"] is True
assert payload["maintenance"] is False
PY
grep -q '^TEST-CODE$' "$work/data/activation.code"
grep -q '^ssid=ClimaSense Test$' "$work/data/hostapd.conf"

echo "portal-http-api-ok"
