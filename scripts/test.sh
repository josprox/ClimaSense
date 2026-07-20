#!/bin/sh
set -eu

root="$(CDPATH= cd -- "$(dirname -- "$0")/.." && pwd)"
for plugin in climasense_hardware climasense_transport; do
  (cd "$root/plugins/$plugin" && go test ./... && go vet ./...)
done
(cd "$root/edge" && joss run tests/syntax.joss && joss run tests/plugins.joss && rm -f tests/queue-test.sqlite* && joss run tests/queue.joss && rm -f tests/queue-test.sqlite*)

cp "$root/server/tests/env.test.joss" "$root/server/env.joss"
trap 'rm -f "$root/server/env.joss" "$root/server/tests/server-test.sqlite"*' EXIT INT TERM
(cd "$root/server" && joss migrate && joss run tests/schema.joss && joss run tests/saas-schema.joss && joss run tests/database-ready.joss && joss run tests/syntax.joss)

(
  cd "$root/server"
  joss server start > tests/server-http.log 2>&1 &
  server_pid=$!
  cleanup_http() {
    kill "$server_pid" 2>/dev/null || true
    wait "$server_pid" 2>/dev/null || true
    rm -f tests/server-http.log
  }
  trap cleanup_http EXIT INT TERM

  ready=0
  attempt=0
  while [ "$attempt" -lt 20 ]; do
    status="$(curl -s -o /dev/null -w '%{http_code}' http://127.0.0.1:18080/health || true)"
    [ "$status" = "200" ] && { ready=1; break; }
    attempt=$((attempt + 1))
    sleep 1
  done
  [ "$ready" = "1" ] || { cat tests/server-http.log; exit 1; }

  operator_status="$(curl -s -o /dev/null -w '%{http_code}' http://127.0.0.1:18080/api/v1/devices/test/latest)"
  provision_status="$(curl -s -o /dev/null -w '%{http_code}' -X POST -H 'Content-Type: application/json' --data '{}' http://127.0.0.1:18080/api/v1/devices/provision)"
  ready_status="$(curl -s -o /dev/null -w '%{http_code}' http://127.0.0.1:18080/ready)"
  login_status="$(curl -s -o /dev/null -w '%{http_code}' http://127.0.0.1:18080/login)"
  admin_status="$(curl -s -o /dev/null -w '%{http_code}' http://127.0.0.1:18080/api/v1/admin/summary)"
  tenant_status="$(curl -s -o /dev/null -w '%{http_code}' http://127.0.0.1:18080/api/v1/tenant/summary)"
  [ "$operator_status" = "401" ] || { echo "operator status=$operator_status, want 401" >&2; exit 1; }
  [ "$provision_status" = "401" ] || { echo "provision status=$provision_status, want 401" >&2; exit 1; }
  [ "$ready_status" = "200" ] || { echo "ready status=$ready_status, want 200" >&2; exit 1; }
  [ "$login_status" = "200" ] || { echo "login status=$login_status, want 200" >&2; exit 1; }
  [ "$admin_status" = "401" ] || { echo "admin status=$admin_status, want 401" >&2; exit 1; }
  [ "$tenant_status" = "401" ] || { echo "tenant status=$tenant_status, want 401" >&2; exit 1; }
  echo "server-http-ok"
)
