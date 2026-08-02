#!/bin/sh
set -eu

root="$(CDPATH= cd -- "$(dirname -- "$0")/.." && pwd)"
release_base="${JOSS_RELEASE_BASE:-https://github.com/josprox/Joss-language/releases/latest/download}"
cache_root="$root/cache/joss-release"
manifest="$cache_root/release-manifest.json"
mkdir -p "$cache_root" "$root/dist/tools" "$root/dist"

manifest_tmp="$manifest.partial"
curl -fsSL "$release_base/release-manifest.json" -o "$manifest_tmp"
version="$(sed -n 's/^[[:space:]]*"version"[[:space:]]*:[[:space:]]*"\([^"]*\)".*/\1/p' "$manifest_tmp" | head -n 1)"
[ -n "$version" ] || { echo "No se pudo determinar la ultima version publicada de Joss" >&2; exit 1; }
mv -f "$manifest_tmp" "$manifest"

artifact_sha() {
  artifact="$1"
  awk -v artifact="$artifact" '
    index($0, "\"name\": \"" artifact "\"") { found=1; next }
    found && /"sha256"/ {
      line=$0
      sub(/^.*"sha256"[[:space:]]*:[[:space:]]*"/, "", line)
      sub(/".*$/, "", line)
      print line
      exit
    }
  ' "$manifest"
}

download_artifact() {
  artifact="$1"
  expected="$(artifact_sha "$artifact")"
  [ -n "$expected" ] || { echo "El manifiesto no contiene $artifact" >&2; exit 1; }
  version_dir="$cache_root/$version"
  archive="$version_dir/$artifact"
  mkdir -p "$version_dir"
  if [ ! -f "$archive" ] || ! printf '%s  %s\n' "$expected" "$archive" | sha256sum -c - >/dev/null 2>&1; then
    echo "Descargando Joss $version: $artifact..." >&2
    partial="$archive.partial"
    curl -fsSL "$release_base/$artifact" -o "$partial"
    printf '%s  %s\n' "$expected" "$partial" | sha256sum -c - >/dev/null
    mv -f "$partial" "$archive"
  fi
  printf '%s\n' "$archive"
}

os="$(uname -s)"
arch="$(uname -m)"
case "$os:$arch" in
  Linux:x86_64) host_archive=jossecurity-linux.zip; host_binary=joss-linux-amd64 ;;
  Linux:aarch64) host_archive=jossecurity-linux.zip; host_binary=joss-linux-arm64 ;;
  Linux:armv7l) host_archive=jossecurity-linux.zip; host_binary=joss-linux-armv7 ;;
  Darwin:x86_64) host_archive=jossecurity-macos.zip; host_binary=joss-macos-amd64 ;;
  Darwin:arm64) host_archive=jossecurity-macos.zip; host_binary=joss-macos-arm64 ;;
  *) echo "Plataforma no soportada para Joss: $os $arch" >&2; exit 1 ;;
esac

host_zip="$(download_artifact "$host_archive")"
host_extract="$cache_root/$version/${host_archive%.zip}"
host_digest="$(sha256sum "$host_zip" | awk '{print $1}')"
if [ "$(cat "$host_extract/.complete" 2>/dev/null || true)" != "$host_digest" ]; then
  rm -rf "$host_extract"
  mkdir -p "$host_extract"
  unzip -q "$host_zip" -d "$host_extract"
  printf '%s\n' "$host_digest" >"$host_extract/.complete"
fi
host_source="$(find "$host_extract" -type f -name "$host_binary" | head -n 1)"
[ -n "$host_source" ] || { echo "No se encontro $host_binary en $host_archive" >&2; exit 1; }
install -m 0755 "$host_source" "$root/dist/tools/joss"

linux_zip="$(download_artifact jossecurity-linux.zip)"
linux_extract="$cache_root/$version/jossecurity-linux"
linux_digest="$(sha256sum "$linux_zip" | awk '{print $1}')"
if [ "$(cat "$linux_extract/.complete" 2>/dev/null || true)" != "$linux_digest" ]; then
  rm -rf "$linux_extract"
  mkdir -p "$linux_extract"
  unzip -q "$linux_zip" -d "$linux_extract"
  printf '%s\n' "$linux_digest" >"$linux_extract/.complete"
fi
arm_source="$(find "$linux_extract" -type f -name joss-linux-arm64 | head -n 1)"
[ -n "$arm_source" ] || { echo "La release $version no contiene joss-linux-arm64" >&2; exit 1; }
install -m 0755 "$arm_source" "$root/dist/joss-linux-arm64"

printf '%s\n' "$version" >"$root/dist/joss-linux-arm64.version"
(cd "$root/dist" && sha256sum joss-linux-arm64 >joss-linux-arm64.sha256)
echo "Joss $version preparado desde la release oficial y verificado con SHA-256."
