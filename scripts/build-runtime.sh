#!/bin/sh
set -eu

root="$(CDPATH= cd -- "$(dirname -- "$0")/.." && pwd)"
: "${JOSS_SOURCE:?Defina JOSS_SOURCE con la ruta del repositorio Joss de solo lectura}"
[ -f "$JOSS_SOURCE/go.mod" ] || { echo "JOSS_SOURCE no contiene go.mod" >&2; exit 1; }
minimum_joss="3.6.3"
source_version="$(sed -n 's/^const Version = "\([0-9][0-9.]*\)"/\1/p' "$JOSS_SOURCE/pkg/version/version.go")"
[ -n "$source_version" ] || { echo "No se pudo determinar la version de Joss en JOSS_SOURCE" >&2; exit 1; }
[ "$(printf '%s\n%s\n' "$minimum_joss" "$source_version" | sort -V | head -n 1)" = "$minimum_joss" ] || {
  echo "ClimaSense requiere Joss >= $minimum_joss; JOSS_SOURCE contiene $source_version" >&2
  exit 1
}
mkdir -p "$root/dist"

(cd "$JOSS_SOURCE" && CGO_ENABLED=0 GOOS=linux GOARCH=arm64 go build -trimpath -ldflags '-s -w' -o "$root/dist/joss-linux-arm64" ./cmd/joss)
(cd "$root/dist" && sha256sum joss-linux-arm64 > joss-linux-arm64.sha256)
echo "Runtime ARM64 creado sin modificar Joss: $root/dist/joss-linux-arm64"
