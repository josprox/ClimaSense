#!/bin/sh
set -eu

root="$(CDPATH= cd -- "$(dirname -- "$0")/../.." && pwd)"
[ "$(id -u)" = "0" ] || { echo "Ejecute con sudo" >&2; exit 1; }
[ "$(uname -m)" = "aarch64" ] || { echo "Se requiere Raspberry Pi OS Lite de 64 bits" >&2; exit 1; }
[ -x "$root/dist/joss-linux-arm64" ] || { echo "Falta dist/joss-linux-arm64; ejecute build-runtime.sh" >&2; exit 1; }
[ -f "$root/dist/plugins/climasense_hardware.jp" ] || { echo "Falta el paquete hardware; ejecute build-plugins.sh" >&2; exit 1; }
[ -f "$root/dist/plugins/climasense_transport.jp" ] || { echo "Falta el paquete transport; ejecute build-plugins.sh" >&2; exit 1; }

export DEBIAN_FRONTEND=noninteractive
echo "Esperando disponibilidad del administrador de paquetes..."
apt-get -o DPkg::Lock::Timeout=300 update
apt-get -o DPkg::Lock::Timeout=300 install -y --no-install-recommends hostapd dnsmasq wpasupplicant \
  wireless-tools iw gpiod i2c-tools isc-dhcp-client ca-certificates

id climasense >/dev/null 2>&1 || useradd --system --home-dir /data/climasense \
  --shell /usr/sbin/nologin --groups i2c climasense
usermod -a -G i2c climasense
install -d -o climasense -g climasense -m 0750 /data/climasense
install -d -m 0755 /opt/climasense/edge /opt/climasense/plugins /usr/local/lib/climasense
install -m 0755 "$root/dist/joss-linux-arm64" /usr/local/bin/joss

edge_source="$root/edge"
if [ -d "$edge_source" ]; then
  tar -C "$edge_source" --exclude='.env' --exclude='plugins' --exclude='storage' \
    --exclude='tests' --exclude='log.txt' --exclude='*.sqlite*' -cf - . | tar -C /opt/climasense/edge -xf -
else
  edge_source="$root"
  tar -C "$edge_source" -cf - activate.joss env.example joss.yaml main.joss onboarding.joss src | \
    tar -C /opt/climasense/edge -xf -
fi
if [ ! -f /data/climasense/.env ]; then
  install -o climasense -g climasense -m 0600 "$edge_source/env.example" /data/climasense/.env
else
  chown climasense:climasense /data/climasense/.env
  chmod 0600 /data/climasense/.env
fi
ln -sfn /data/climasense/.env /opt/climasense/edge/.env
install -d -m 0755 /opt/climasense/edge/plugins/climasense_hardware/0.1.1
install -d -m 0755 /opt/climasense/edge/plugins/climasense_transport/0.2.0
install -m 0644 "$root/dist/plugins/climasense_hardware.jp" /opt/climasense/edge/plugins/climasense_hardware/0.1.1/climasense_hardware.jp
install -m 0644 "$root/dist/plugins/climasense_transport.jp" /opt/climasense/edge/plugins/climasense_transport/0.2.0/climasense_transport.jp
find /opt/climasense/edge -type d -exec chmod 0755 {} +
find /opt/climasense/edge -type f -exec chmod 0644 {} +
chown -R climasense:climasense /opt/climasense/edge /data/climasense
find /data/climasense -type f -exec chmod go-rwx {} +

boot_config=
for candidate in /boot/firmware/config.txt /boot/config.txt; do
  [ -f "$candidate" ] && { boot_config="$candidate"; break; }
done
[ -n "$boot_config" ] || { echo "No se encontro config.txt de Raspberry Pi" >&2; exit 1; }
need_i2c=0
need_gpio=0
grep -Eq '^[[:space:]]*dtparam=i2c_arm=on([[:space:]]|$)' "$boot_config" || need_i2c=1
grep -Eq '^[[:space:]]*gpio=17=ip,pu([[:space:]]|$)' "$boot_config" || need_gpio=1
if [ "$need_i2c" -eq 1 ] || [ "$need_gpio" -eq 1 ]; then
  printf '\n[all]\n' >>"$boot_config"
  [ "$need_i2c" -eq 0 ] || printf '%s\n' 'dtparam=i2c_arm=on' >>"$boot_config"
  [ "$need_gpio" -eq 0 ] || printf '%s\n' 'gpio=17=ip,pu' >>"$boot_config"
fi

install -m 0644 "$root/os/raspios/hostapd.conf" /etc/climasense-hostapd.conf
install -m 0644 "$root/os/raspios/dnsmasq.conf" /etc/climasense-dnsmasq.conf
install -m 0755 "$root/os/raspios/libexec/climasense-onboarding" /usr/local/lib/climasense/onboarding
install -m 0755 "$root/os/raspios/libexec/climasense-network" /usr/local/lib/climasense/network
install -m 0755 "$root/os/raspios/libexec/climasense-activation" /usr/local/lib/climasense/activation
install -m 0755 "$root/os/raspios/libexec/climasense-activation-ready" /usr/local/lib/climasense/activation-ready
install -m 0755 "$root/os/raspios/libexec/climasense-wifi-interface" /usr/local/lib/climasense/wifi-interface
install -m 0755 "$root/os/raspios/libexec/climasense-maintenance-button" /usr/local/lib/climasense/maintenance-button

selected_iface="$(/usr/local/lib/climasense/wifi-interface ap)"
echo "ClimaSense usara la interfaz Wi-Fi: $selected_iface"
rm -f /data/climasense/onboarding.failed
install -d -m 0755 /etc/NetworkManager/conf.d
printf '%s\n' '[keyfile]' "unmanaged-devices=interface-name:$selected_iface" > /etc/NetworkManager/conf.d/99-climasense-wifi.conf
for unit in hostapd dnsmasq; do systemctl disable --now "$unit" 2>/dev/null || true; done
systemctl disable getty@tty1.service 2>/dev/null || true
systemctl mask getty@tty1.service
install -m 0644 "$root/os/raspios/systemd/"*.service /etc/systemd/system/
systemctl daemon-reload
systemctl enable climasense-onboarding.service climasense-network.service climasense-activation.service climasense-edge.service climasense-maintenance-button.service
echo "ClimaSense instalado sin interrumpir la conexion actual."
echo "Al reiniciar, $selected_iface cambiara temporalmente a hotspot de configuracion."
