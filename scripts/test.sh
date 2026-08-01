#!/bin/sh
set -eu

root="$(CDPATH= cd -- "$(dirname -- "$0")/.." && pwd)"
minimum_joss="3.6.3"
installed_joss="$(joss version | sed -n 's/^Joss v\([0-9][0-9.]*\).*/\1/p')"
[ -n "$installed_joss" ] || { echo "No se pudo determinar la version de Joss" >&2; exit 1; }
[ "$(printf '%s\n%s\n' "$minimum_joss" "$installed_joss" | sort -V | head -n 1)" = "$minimum_joss" ] || {
  echo "ClimaSense requiere Joss >= $minimum_joss; instalado: $installed_joss" >&2
  exit 1
}

for plugin in climasense_hardware climasense_transport; do
  (cd "$root/plugins/$plugin" && go test ./... && go vet ./...)
done
(cd "$root" && CLIMASENSE_TEST_INTEGER=17 joss run tests/env-number.joss && joss run tests/syntax.joss && joss run tests/plugins.joss && rm -f tests/queue-test.sqlite* && joss run tests/queue.joss && rm -f tests/queue-test.sqlite*)
sh "$root/scripts/test-raspios.sh"
