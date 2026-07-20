#!/bin/sh
set -eu

target_dir="$1"
board_dir="$(CDPATH= cd -- "$(dirname -- "$0")" && pwd)"
artifacts="$board_dir/artifacts"

# Defensa adicional para builds incrementales: la cuenta root nunca conserva
# una contrasena vacia aunque la configuracion base de Raspberry Pi la active.
sed -i 's|^root:[^:]*:|root:*:|' "$target_dir/etc/shadow"
rm -f "$target_dir/etc/init.d/S40network"

install -d -m 0755 "$target_dir/opt/climasense" "$target_dir/opt/climasense/edge" "$target_dir/opt/climasense/plugins"
install -m 0755 "$artifacts/joss-linux-arm64" "$target_dir/usr/bin/joss"
cp -R "$artifacts/edge/." "$target_dir/opt/climasense/edge/"
cp -R "$artifacts/plugins/." "$target_dir/opt/climasense/edge/plugins/"

install -d -o 1000 -g 1000 -m 0750 "$target_dir/data/climasense"
install -o 1000 -g 1000 -m 0600 "$artifacts/env.example" "$target_dir/data/climasense/.env"

find "$target_dir/opt/climasense/edge" -type f -name '*.joss' -exec chmod 0644 {} \;
