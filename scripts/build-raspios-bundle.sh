#!/bin/sh
set -eu

root="$(CDPATH= cd -- "$(dirname -- "$0")/.." && pwd)"
bundle="$root/dist/climasense-raspios-installer.tar.gz"

sh "$root/scripts/test-raspios.sh"

[ -x "$root/dist/joss-linux-arm64" ] || {
  echo "Falta dist/joss-linux-arm64; ejecute scripts/build-runtime.sh" >&2
  exit 1
}
[ -f "$root/dist/plugins/climasense_hardware.jp" ] || {
  echo "Falta climasense_hardware.jp; ejecute scripts/build-plugins.sh" >&2
  exit 1
}
[ -f "$root/dist/plugins/climasense_transport.jp" ] || {
  echo "Falta climasense_transport.jp; ejecute scripts/build-plugins.sh" >&2
  exit 1
}

tar -C "$root" -czf "$bundle" \
  --exclude='edge/.env' \
  --exclude='edge/storage' \
  --exclude='edge/plugins' \
  --exclude='edge/tests' \
  --exclude='edge/log.txt' \
  --exclude='*.sqlite*' \
  os/raspios edge \
  dist/joss-linux-arm64 \
  dist/plugins/climasense_hardware.jp \
  dist/plugins/climasense_transport.jp

(cd "$root/dist" && sha256sum "$(basename "$bundle")" > "$(basename "$bundle").sha256")
echo "Paquete creado: $bundle"
