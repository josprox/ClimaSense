#!/bin/sh
set -eu

root="$(CDPATH= cd -- "$(dirname -- "$0")/.." && pwd)"
version="$(cat "$root/os/buildroot/buildroot.version")"
archive="buildroot-$version.tar.xz"
archive_cache="$root/output/downloads"
cache_root="${CLIMASENSE_CACHE_ROOT:-$HOME/.cache/climasense-os}"
build_dir="${CLIMASENSE_BUILD_DIR:-$cache_root/build}"
download_dir="${BR2_DL_DIR:-$cache_root/downloads}"
source="$build_dir/buildroot-$version"

[ "$(uname -s)" = "Linux" ] || {
  echo "Este script debe ejecutarse dentro de Linux o WSL." >&2
  exit 1
}

mkdir -p "$archive_cache" "$build_dir" "$download_dir"

if [ ! -f "$archive_cache/$archive" ]; then
  curl --fail --location --output "$archive_cache/$archive" \
    "https://buildroot.org/downloads/$archive"
fi
(cd "$archive_cache" && sha256sum -c "$root/os/buildroot/buildroot.sha256")

if [ ! -d "$source" ]; then
  tar -C "$build_dir" -xf "$archive_cache/$archive"
fi

make -C "$source" raspberrypizero2w_64_defconfig
cat "$root/os/buildroot/climasense.fragment" >> "$source/.config"
make -C "$source" olddefconfig

echo "Descargando anticipadamente las fuentes seleccionadas..."
make -C "$source" BR2_DL_DIR="$download_dir" source

echo "Cache preparada en: $download_dir"
echo "Ahora use: ./scripts/build-os-cached.sh"
