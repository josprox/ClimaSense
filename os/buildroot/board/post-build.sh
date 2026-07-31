#!/bin/sh
set -eu

target_dir="$1"
board_dir="$(CDPATH= cd -- "$(dirname -- "$0")" && pwd)"
artifacts="$board_dir/artifacts"

# Defensa adicional para builds incrementales: la cuenta root nunca conserva
# una contrasena vacia aunque la configuracion base de Raspberry Pi la active.
sed -i 's|^root:[^:]*:|root:*:|' "$target_dir/etc/shadow"
rm -f "$target_dir/etc/init.d/S40network" "$target_dir/etc/init.d/S80dnsmasq"
# Equipo cerrado: no se instala ningun getty ni prompt de login local.
sed -i '/getty/d' "$target_dir/etc/inittab"

install -d -m 0755 "$target_dir/opt/climasense" "$target_dir/opt/climasense/edge" "$target_dir/opt/climasense/plugins"
install -m 0755 "$artifacts/joss-linux-arm64" "$target_dir/usr/bin/joss"
cp -R "$artifacts/edge/." "$target_dir/opt/climasense/edge/"
cp -R "$artifacts/plugins/." "$target_dir/opt/climasense/edge/plugins/"

install -d -o 1000 -g 1000 -m 0750 "$target_dir/data/climasense"
install -o 1000 -g 1000 -m 0600 "$artifacts/env.example" "$target_dir/data/climasense/.env"
ln -sfn /data/climasense/.env "$target_dir/opt/climasense/edge/.env"

find "$target_dir/opt/climasense/edge" -type f -name '*.joss' -exec chmod 0644 {} \;

# La Zero 2 W revision 1 usa el nombre 43436s para el mismo CLM blob.
# Buildroot instala el blob 43436; el enlace evita operar con canales limitados.
if [ -f "$target_dir/lib/firmware/brcm/brcmfmac43436-sdio.clm_blob" ]; then
  ln -sfn brcmfmac43436-sdio.clm_blob \
    "$target_dir/lib/firmware/brcm/brcmfmac43436s-sdio.clm_blob"
fi

grep -q 'getty' "$target_dir/etc/inittab" && {
  echo "ERROR: la imagen todavia contiene un prompt de login" >&2
  exit 1
}
[ ! -e "$target_dir/etc/init.d/S80dnsmasq" ] || {
  echo "ERROR: dnsmasq debe ser administrado exclusivamente por S44" >&2
  exit 1
}
for required_binary in usr/bin/gpioget usr/bin/joss usr/sbin/dnsmasq usr/sbin/hostapd; do
  [ -x "$target_dir/$required_binary" ] || {
    echo "ERROR: falta /$required_binary en la imagen" >&2
    exit 1
  }
done
[ -x "$target_dir/usr/sbin/iwlist" ] || [ -x "$target_dir/sbin/iwlist" ] || {
  echo "ERROR: falta iwlist en la imagen" >&2
  exit 1
}
[ -f "$target_dir/lib/firmware/regulatory.db" ] || {
  echo "ERROR: falta regulatory.db en la imagen" >&2
  exit 1
}
