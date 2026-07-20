#!/bin/sh
set -eu

root="$(CDPATH= cd -- "$(dirname -- "$0")/.." && pwd)"
: "${JOSS_SOURCE:?Defina JOSS_SOURCE con la ruta del repositorio Joss de solo lectura}"
[ -f "$JOSS_SOURCE/go.mod" ] || { echo "JOSS_SOURCE no contiene go.mod" >&2; exit 1; }
mkdir -p "$root/dist"

(cd "$JOSS_SOURCE" && CGO_ENABLED=0 GOOS=linux GOARCH=arm64 go build -trimpath -ldflags '-s -w' -o "$root/dist/joss-linux-arm64" ./cmd/joss)
(cd "$root/dist" && sha256sum joss-linux-arm64 > joss-linux-arm64.sha256)
echo "Runtime ARM64 creado sin modificar Joss: $root/dist/joss-linux-arm64"
