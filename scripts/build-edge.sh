#!/bin/sh
set -eu

root="$(CDPATH= cd -- "$(dirname -- "$0")/.." && pwd)"
mkdir -p "$root/dist"
cp "$root/edge/service.joss" "$root/dist/climasense-edge.joss"
tar -C "$root/edge" --exclude='.env' --exclude='*.sqlite*' --exclude='log.txt' --exclude='plugins' --exclude='storage' -czf "$root/dist/climasense-edge.tar.gz" .
(cd "$root/dist" && sha256sum climasense-edge.joss climasense-edge.tar.gz > climasense-edge.sha256)
