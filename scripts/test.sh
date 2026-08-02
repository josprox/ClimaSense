#!/bin/sh
set -eu

root="$(CDPATH= cd -- "$(dirname -- "$0")/.." && pwd)"
joss="$root/scripts/joss-current.sh"
echo "Probando con $(sh "$joss" version) desde la release oficial"
sh "$root/scripts/install-server-plugin.sh"

(cd "$root/plugins/climasense_transport" && go test ./... && go vet ./...)

env_backup=""
if [ -f "$root/env.joss" ]; then
  env_backup="$(mktemp)"
  cp "$root/env.joss" "$env_backup"
fi
cp "$root/tests/env.test.joss" "$root/env.joss"
cleanup_tests() {
  rm -f "$root/tests/server-test.sqlite"*
  if [ -n "$env_backup" ]; then
    cp "$env_backup" "$root/env.joss"
    rm -f "$env_backup"
  else
    rm -f "$root/env.joss"
  fi
}
trap cleanup_tests EXIT INT TERM
(cd "$root" && sh "$joss" migrate && sh "$joss" run tests/schema.joss && sh "$joss" run tests/saas-schema.joss && sh "$joss" run tests/database-ready.joss && sh "$joss" run tests/client-flow.joss && sh "$joss" run tests/live-updates.joss && sh "$joss" run tests/transport-token.joss && sh "$joss" run tests/syntax.joss)

if command -v node >/dev/null 2>&1; then
  node --check "$root/public/js/live.js"
  node --check "$root/public/js/dashboard.js"
  node --check "$root/public/js/admin.js"
  node --check "$root/public/js/tenant.js"
fi

(
  cd "$root"
  sh "$joss" server start > tests/server-http.log 2>&1 &
  server_pid=$!
  cleanup_http() {
    kill "$server_pid" 2>/dev/null || true
    wait "$server_pid" 2>/dev/null || true
    rm -f tests/server-http.log tests/server-cookie.txt tests/register.json tests/organization.json tests/activation-code.json
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
  register_status="$(curl -s -o /dev/null -w '%{http_code}' http://127.0.0.1:18080/register)"
  admin_status="$(curl -s -o /dev/null -w '%{http_code}' http://127.0.0.1:18080/api/v1/admin/summary)"
  tenant_status="$(curl -s -o /dev/null -w '%{http_code}' http://127.0.0.1:18080/api/v1/tenant/summary)"
  ws_token_status="$(curl -s -o /dev/null -w '%{http_code}' http://127.0.0.1:18080/api/v1/auth/ws-token)"
  [ "$operator_status" = "401" ] || { echo "operator status=$operator_status, want 401" >&2; exit 1; }
  [ "$provision_status" = "401" ] || { echo "provision status=$provision_status, want 401" >&2; exit 1; }
  [ "$ready_status" = "200" ] || { echo "ready status=$ready_status, want 200" >&2; exit 1; }
  [ "$login_status" = "200" ] || { echo "login status=$login_status, want 200" >&2; exit 1; }
  [ "$register_status" = "200" ] || { echo "register status=$register_status, want 200" >&2; exit 1; }
  [ "$admin_status" = "401" ] || { echo "admin status=$admin_status, want 401" >&2; exit 1; }
  [ "$tenant_status" = "401" ] || { echo "tenant status=$tenant_status, want 401" >&2; exit 1; }
  [ "$ws_token_status" = "401" ] || { echo "ws token status=$ws_token_status, want 401" >&2; exit 1; }

  register_status="$(curl -sS -o tests/register.json -w '%{http_code}' \
    -c tests/server-cookie.txt -H 'Content-Type: application/json' \
    --data '{"email":"edge-test@example.invalid","password":"Test-only-password-2026!","name":"Edge Test"}' \
    http://127.0.0.1:18080/api/v1/auth/register)"
  [ "$register_status" = "201" ] || { cat tests/register.json; cat tests/server-http.log; exit 1; }

  organization_status="$(curl -sS -o tests/organization.json -w '%{http_code}' \
    -b tests/server-cookie.txt -H 'Content-Type: application/json' \
    --data '{"name":"ClimaSense Test","slug":"climasense-test","contact_email":"edge-test@example.invalid"}' \
    http://127.0.0.1:18080/api/v1/tenant/organizations)"
  [ "$organization_status" = "201" ] || { cat tests/organization.json; cat tests/server-http.log; exit 1; }

  activation_status="$(curl -sS -o tests/activation-code.json -w '%{http_code}' \
    -b tests/server-cookie.txt -H 'Content-Type: application/json' \
    --data '{"label":"Raspberry HTTP Test"}' \
    http://127.0.0.1:18080/api/v1/tenant/activation-codes)"
  [ "$activation_status" = "201" ] || { cat tests/activation-code.json; cat tests/server-http.log; exit 1; }
  grep -Eq '"code":"CS-[A-Za-z0-9_-]+"' tests/activation-code.json || { cat tests/activation-code.json; exit 1; }

  if ! (cd "$root/plugins/climasense_transport" && go run "$root/tests/dashboard-live.go"); then
    cat tests/server-http.log
    exit 1
  fi
  echo "server-http-ok"
)
