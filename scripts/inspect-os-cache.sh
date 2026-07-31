#!/bin/sh
set -eu

cache_root="${CLIMASENSE_CACHE_ROOT:-$HOME/.cache/climasense-os}"
build_dir="${CLIMASENSE_BUILD_DIR:-$cache_root/build}"
download_dir="${BR2_DL_DIR:-$cache_root/downloads}"

echo "Arbol incremental: $build_dir"
if [ -d "$build_dir" ]; then
  du -sh "$build_dir"
else
  echo "  aun no existe"
fi

echo "Cache de descargas: $download_dir"
if [ -d "$download_dir" ]; then
  du -sh "$download_dir"
  find "$download_dir" -type f | wc -l | awk '{print "  archivos: " $1}'
else
  echo "  aun no existe"
fi
