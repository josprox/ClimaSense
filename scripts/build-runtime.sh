#!/bin/sh
set -eu

root="$(CDPATH= cd -- "$(dirname -- "$0")/.." && pwd)"
sh "$root/scripts/prepare-joss-release.sh"
echo "Runtime ARM64 actualizado desde la ultima release: $root/dist/joss-linux-arm64"
