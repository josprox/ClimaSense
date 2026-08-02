#!/bin/sh
set -eu

root="$(CDPATH= cd -- "$(dirname -- "$0")/.." && pwd)"
joss="$root/scripts/joss-current.sh"
echo "Probando con $(sh "$joss" version) desde la release oficial"

for plugin in climasense_hardware climasense_transport; do
  (cd "$root/plugins/$plugin" && go test ./... && go vet ./...)
done
(cd "$root" && CLIMASENSE_TEST_INTEGER=17 sh "$joss" run tests/env-number.joss && sh "$joss" run tests/syntax.joss && sh "$joss" run tests/plugins.joss && rm -f tests/queue-test.sqlite* && sh "$joss" run tests/queue.joss && rm -f tests/queue-test.sqlite*)
sh "$root/scripts/test-raspios.sh"
