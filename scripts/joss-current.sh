#!/bin/sh
set -eu

root="$(CDPATH= cd -- "$(dirname -- "$0")/.." && pwd)"

if [ -n "${JOSS_BIN:-}" ]; then
  [ -x "$JOSS_BIN" ] || { echo "JOSS_BIN no es ejecutable: $JOSS_BIN" >&2; exit 1; }
  exec "$JOSS_BIN" "$@"
fi

binary="$root/dist/tools/joss"
if [ ! -x "$binary" ]; then
  sh "$root/scripts/prepare-joss-release.sh" >&2
fi
exec "$binary" "$@"
