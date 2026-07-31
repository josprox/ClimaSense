#!/bin/sh
set -eu

root="$(CDPATH= cd -- "$(dirname -- "$0")/.." && pwd)"
version="$(cat "$root/os/buildroot/buildroot.version")"
archive="buildroot-$version.tar.xz"
cache="$root/output/downloads"
build_dir="${CLIMASENSE_BUILD_DIR:-${TMPDIR:-/tmp}/climasense-buildroot}"
source="$build_dir/buildroot-$version"
mkdir -p "$cache" "$build_dir" "$root/dist"

[ "$(uname -s)" = "Linux" ] || { echo "Buildroot requiere Linux o WSL" >&2; exit 1; }
PATH="/usr/local/sbin:/usr/local/bin:/usr/sbin:/usr/bin:/sbin:/bin"
export PATH
if [ ! -f "$cache/$archive" ]; then
  curl --fail --location --output "$cache/$archive" "https://buildroot.org/downloads/$archive"
fi
(cd "$cache" && sha256sum -c "$root/os/buildroot/buildroot.sha256")

if [ ! -d "$source" ]; then
  tar -C "$build_dir" -xf "$cache/$archive"
fi

rm -rf "$source/board/climasense"
mkdir -p "$source/board/climasense/artifacts/edge" "$source/board/climasense/artifacts/plugins/climasense_hardware/0.1.0" "$source/board/climasense/artifacts/plugins/climasense_transport/0.2.0"
cp -R "$root/os/buildroot/board/." "$source/board/climasense/"
chmod +x "$source/board/climasense/post-build.sh" "$source/board/climasense/rootfs-overlay/etc/init.d/"S*
cp "$root/dist/joss-linux-arm64" "$source/board/climasense/artifacts/joss-linux-arm64"
cp -R "$root/edge/app" "$root/edge/src" "$root/edge/routes.joss" "$root/edge/main.joss" "$root/edge/service.joss" "$root/edge/activate.joss" "$root/edge/joss.yaml" "$source/board/climasense/artifacts/edge/"
cp "$root/edge/env.example" "$source/board/climasense/artifacts/env.example"
cp "$root/dist/plugins/climasense_hardware.jp" "$source/board/climasense/artifacts/plugins/climasense_hardware/0.1.0/climasense_hardware.jp"
cp "$root/dist/plugins/climasense_transport.jp" "$source/board/climasense/artifacts/plugins/climasense_transport/0.2.0/climasense_transport.jp"

make -C "$source" raspberrypizero2w_64_defconfig
cat "$root/os/buildroot/climasense.fragment" >> "$source/.config"
make -C "$source" olddefconfig
make -C "$source"

cp "$source/output/images/sdcard.img" "$root/dist/climasense-os-rpi-zero-2-w.img"
(cd "$root/dist" && sha256sum climasense-os-rpi-zero-2-w.img > climasense-os-rpi-zero-2-w.img.sha256)
