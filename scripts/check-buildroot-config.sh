#!/bin/sh
set -eu

root="$(CDPATH= cd -- "$(dirname -- "$0")/.." && pwd)"
version="$(cat "$root/os/buildroot/buildroot.version")"
archive="buildroot-$version.tar.xz"
build_dir="${CLIMASENSE_BUILD_DIR:-${TMPDIR:-/tmp}/climasense-buildroot}"
source="$build_dir/buildroot-$version"

(cd "$root/output/downloads" && sha256sum -c "$root/os/buildroot/buildroot.sha256")
if [ ! -d "$source" ]; then
  mkdir -p "$build_dir"
  tar -C "$build_dir" -xf "$root/output/downloads/$archive"
fi
mkdir -p "$source/board/climasense"
cp -R "$root/os/buildroot/board/." "$source/board/climasense/"
make -C "$source" raspberrypizero2w_64_defconfig
cat "$root/os/buildroot/climasense.fragment" >> "$source/.config"
make -C "$source" olddefconfig
grep -E 'BR2_(aarch64|PACKAGE_WPA_SUPPLICANT|PACKAGE_I2C_TOOLS|PACKAGE_BRCMFMAC_SDIO_FIRMWARE_RPI(_WIFI)?|LINUX_KERNEL_CONFIG_FRAGMENT_FILES|ROOTFS_OVERLAY|PACKAGE_RPI_FIRMWARE_CONFIG_FILE)=' "$source/.config"
