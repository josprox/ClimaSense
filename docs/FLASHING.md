# Grabacion

Verifique primero el dispositivo de bloque; un destino incorrecto destruye datos.

```sh
./scripts/flash-image.sh /dev/sdX
```

El script exige escribir `FLASH`, usa bloques de 4 MiB, `fsync` y `sync`. Alternativamente use Raspberry Pi Imager con `dist/climasense-os-rpi-zero-2-w.img` y verifique el SHA-256 publicado.

El flujo normal ya no requiere escribir identidad ni token en la tarjeta, ni usar pantalla o teclado: conectese al punto de acceso `Configuracion ClimaSense` y complete Wi-Fi, activacion y credenciales del hotspot de mantenimiento en `http://192.168.4.1:8080`.

La URL SaaS se fija en `edge/env.example` o `.env` durante el build; el dispositivo obtiene identidad y token al consumir el codigo. La preconfiguracion headless requiere preparar conjuntamente `wpa_supplicant.conf`, `activation.code` y un `hostapd.conf` WPA2 dentro de `/data/climasense`; si falta cualquiera, el dispositivo conserva el flujo seguro mediante portal.
