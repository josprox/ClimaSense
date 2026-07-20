#!/bin/sh
set -eu

root="$(CDPATH= cd -- "$(dirname -- "$0")/.." && pwd)"
mkdir -p "$root/dist"
cp "$root/edge/main.joss" "$root/dist/climasense-edge.joss"
tar -C "$root/edge" --exclude='*.sqlite*' --exclude='plugins' -czf "$root/dist/climasense-edge.tar.gz" .
(cd "$root/dist" && sha256sum climasense-edge.joss climasense-edge.tar.gz > climasense-edge.sha256)
