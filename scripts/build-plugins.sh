#!/bin/sh
set -eu

root="$(CDPATH= cd -- "$(dirname -- "$0")/.." && pwd)"
mkdir -p "$root/dist/plugins"

build_go_target() {
  plugin="$1" os="$2" arch="$3" output="$4"
  (cd "$root/plugins/$plugin" && CGO_ENABLED=0 GOOS="$os" GOARCH="$arch" go build -trimpath -ldflags '-s -w' -o "$output" ./cmd/plugin)
}

build_go_target climasense_hardware linux arm64 native/linux-arm64/climasense-hardware
build_go_target climasense_hardware linux amd64 native/linux-amd64/climasense-hardware
build_go_target climasense_transport linux arm64 native/linux-arm64/climasense-transport
build_go_target climasense_transport linux amd64 native/linux-amd64/climasense-transport

for plugin in climasense_hardware climasense_transport; do
  (cd "$root/plugins/$plugin" && go test ./... && joss run src/plugin.joss && joss build package . && joss package inspect "$plugin.jp")
  cp "$root/plugins/$plugin/$plugin.jp" "$root/dist/plugins/$plugin.jp"
  (cd "$root/dist/plugins" && sha256sum "$plugin.jp" > "$plugin.jp.sha256")
done
