#!/bin/sh
set -eu

root="$(CDPATH= cd -- "$(dirname -- "$0")/.." && pwd)"
source_package="$root/plugins/climasense_transport/0.2.0/climasense_transport.jp"
installed_package="$root/plugins/climasense_transport/climasense_transport.jp"

[ -s "$source_package" ] || {
  echo "Falta el paquete versionado: $source_package" >&2
  exit 1
}

install -m 0644 "$source_package" "$installed_package"
echo "Plugin servidor instalado: $installed_package"
