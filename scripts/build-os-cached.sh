#!/bin/sh
set -eu

root="$(CDPATH= cd -- "$(dirname -- "$0")/.." && pwd)"
cache_root="${CLIMASENSE_CACHE_ROOT:-$HOME/.cache/climasense-os}"

# El arbol de compilacion queda en el filesystem Linux de WSL: conserva
# permisos/enlaces, sobrevive entre sesiones y permite compilacion incremental.
export CLIMASENSE_BUILD_DIR="${CLIMASENSE_BUILD_DIR:-$cache_root/build}"

# Buildroot guarda aqui todas las fuentes descargadas de kernel, firmware,
# toolchain y paquetes. Al estar fuera de /tmp no se vuelven a descargar.
export BR2_DL_DIR="${BR2_DL_DIR:-$cache_root/downloads}"

mkdir -p "$CLIMASENSE_BUILD_DIR" "$BR2_DL_DIR"

echo "Arbol incremental: $CLIMASENSE_BUILD_DIR"
echo "Cache de descargas: $BR2_DL_DIR"
echo "Los paquetes ya presentes no se descargaran nuevamente."

version="$(cat "$root/os/buildroot/buildroot.version")"
source="$CLIMASENSE_BUILD_DIR/buildroot-$version"
hardware_stamp="$CLIMASENSE_BUILD_DIR/climasense-hardware.sha256"
hardware_hash="$(
  sha256sum \
    "$root/os/buildroot/board/config.txt" \
    "$root/os/buildroot/board/cmdline.txt" \
    "$root/os/buildroot/board/kernel.fragment" |
    sha256sum |
    awk '{print $1}'
)"
previous_hash="$(cat "$hardware_stamp" 2>/dev/null || true)"

if [ -f "$source/.config" ] && [ "$hardware_hash" != "$previous_hash" ]; then
  echo "Cambio de hardware detectado: reconstruyendo solo kernel y firmware."
  make -C "$source" linux-dirclean rpi-firmware-dirclean
fi

"$root/scripts/build-os.sh" "$@"
printf '%s\n' "$hardware_hash" > "$hardware_stamp"
