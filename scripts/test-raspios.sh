#!/bin/sh
set -eu

root="$(CDPATH= cd -- "$(dirname -- "$0")/.." && pwd)"
fail() { echo "Raspberry Pi OS audit: $1" >&2; exit 1; }
contains() { grep -Fq "$2" "$1" || fail "$1 no contiene: $2"; }

for script in "$root/os/raspios/install.sh" "$root/os/raspios/libexec/"* "$root/scripts/build-raspios-bundle.sh"; do
  sh -n "$script" || fail "sintaxis shell invalida en $script"
done

contains "$root/os/raspios/systemd/climasense-onboarding.service" "Type=oneshot"
contains "$root/os/raspios/systemd/climasense-onboarding.service" "Before=climasense-network.service"
contains "$root/os/raspios/systemd/climasense-network.service" "OnFailure=climasense-network-recovery.service"
contains "$root/os/raspios/systemd/climasense-activation.service" "Restart=on-failure"
contains "$root/os/raspios/systemd/climasense-edge.service" "ExecCondition=/usr/local/lib/climasense/activation-ready"
contains "$root/os/raspios/install.sh" "climasense-activation-ready"
contains "$root/os/raspios/install.sh" "dtparam=i2c_arm=on"
contains "$root/os/raspios/libexec/climasense-onboarding" "onboarding.failed"

if grep -R -n 'wlan0' "$root/os/raspios" "$root/edge/activate.joss" "$root/edge/app/OnboardingController.joss"; then
  fail "se encontro una interfaz wlan0 fija en la ruta Raspberry Pi OS"
fi

echo "raspios-audit-ok"
