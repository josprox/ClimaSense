#!/bin/sh
set -eu

root="$(CDPATH= cd -- "$(dirname -- "$0")/.." && pwd)"
mkdir -p "$root/dist"
tar -C "$root/server" \
  --exclude='.env' \
  --exclude='env.joss' \
  --exclude='*.sqlite*' \
  --exclude='plugins' \
  --exclude='storage' \
  --exclude='log.txt' \
  -czf "$root/dist/climasense-server.tar.gz" .
cp "$root/dist/climasense-server.tar.gz" "$root/dist/climasense-server"
(cd "$root/dist" && sha256sum climasense-server.tar.gz climasense-server > climasense-server.tar.gz.sha256)
(cd "$root/dist" && sha256sum climasense-server > climasense-server.sha256)
