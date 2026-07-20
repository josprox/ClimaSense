#!/bin/sh
set -eu

[ "$#" -eq 1 ] || { echo "Uso: $0 /dev/sdX" >&2; exit 2; }
device="$1"
case "$device" in /dev/sd? | /dev/mmcblk? ) ;; *) echo "Dispositivo no permitido: $device" >&2; exit 2;; esac
root="$(CDPATH= cd -- "$(dirname -- "$0")/.." && pwd)"
image="$root/dist/climasense-os-rpi-zero-2-w.img"
[ -f "$image" ] || { echo "No existe $image" >&2; exit 1; }
echo "Se sobrescribira completamente $device con $image"
printf "Escriba FLASH para continuar: "
read answer
[ "$answer" = "FLASH" ] || { echo "Cancelado"; exit 1; }
sudo dd if="$image" of="$device" bs=4M conv=fsync status=progress
sync
