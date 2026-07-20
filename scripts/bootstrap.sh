#!/bin/sh
set -eu

for command in go joss sha256sum tar; do
  command -v "$command" >/dev/null 2>&1 || { echo "Falta dependencia: $command" >&2; exit 1; }
done

echo "Go: $(go version)"
echo "Joss: $(joss version)"
echo "Bootstrap de ClimaSense completado"
