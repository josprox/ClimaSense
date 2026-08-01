#!/bin/sh
set -eu

root="$(CDPATH= cd -- "$(dirname -- "$0")/.." && pwd)"
mkdir -p "$root/dist"
cp "$root/main.joss" "$root/dist/climasense-edge.joss"
tar -C "$root" --exclude='*.sqlite*' --exclude='plugins' --exclude='dist' \
  --exclude='docs' --exclude='os' --exclude='scripts' --exclude='tests' \
  --exclude='.git' --exclude='.env' --exclude='log.txt' \
  -czf "$root/dist/climasense-edge.tar.gz" .
(cd "$root/dist" && sha256sum climasense-edge.joss climasense-edge.tar.gz > climasense-edge.sha256)
