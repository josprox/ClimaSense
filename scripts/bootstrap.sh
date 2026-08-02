#!/bin/sh
set -eu

root="$(CDPATH= cd -- "$(dirname -- "$0")/.." && pwd)"

for command in go curl unzip sha256sum tar; do
  command -v "$command" >/dev/null 2>&1 || { echo "Falta dependencia: $command" >&2; exit 1; }
done

sh "$root/scripts/prepare-joss-release.sh"
sh "$root/scripts/install-server-plugin.sh"
echo "Go: $(go version)"
echo "Joss release: $(sh "$root/scripts/joss-current.sh" version)"
echo "Bootstrap de ClimaSense completado"
