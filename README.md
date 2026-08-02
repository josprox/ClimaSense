# ClimaSense Edge

Esta rama contiene únicamente el agente Edge de ClimaSense para Raspberry Pi OS Lite de 64 bits. Incluye el portal cautivo, activación web, hotspot protegido de mantenimiento, selección dinámica del adaptador Wi-Fi USB, cola local de telemetría y lectura BMP280/BME280 por I2C. Usa siempre la última release estable de Joss; `>=3.6.3` es únicamente el contrato mínimo de compatibilidad.

La instalación recomendada parte de Raspberry Pi Imager. Raspberry Pi Connect o una conexión existente pueden conservarse durante la instalación; después del reinicio, ClimaSense abre temporalmente el hotspot de configuración y no solicita credenciales en la consola.

## Construcción

En Linux o WSL:

```sh
./scripts/bootstrap.sh
./scripts/build-runtime.sh
./scripts/build-plugins.sh
./scripts/test.sh
./scripts/build-raspios-bundle.sh
./scripts/manifest.sh
```

`bootstrap.sh` descarga la última release oficial, valida el SHA-256 y conserva el ZIP por versión en `cache/joss-release`. También extrae `joss-linux-arm64` para el instalador de Raspberry; no compila ni requiere clonar Joss-language.

El paquete resultante es `dist/climasense-raspios-installer.tar.gz`. En la Raspberry Pi:

```sh
tar -xzf climasense-raspios-installer.tar.gz
sudo sh os/raspios/install.sh
sudo reboot
```

El instalador conserva `/data/climasense/.env` y el token del dispositivo durante actualizaciones. Para cableado, diagnóstico y recuperación consulta `docs/RASPBERRY_PI_OS.md`, `docs/I2C.md` y `docs/TROUBLESHOOTING.md`. Buildroot permanece sólo como ruta heredada; no es la opción recomendada para hardware con Wi-Fi USB.

No subas `.env`, tokens, bases SQLite, bitácoras ni paquetes `.jp` generados. Las redes WPA/WPA2 personales requieren contraseñas de 8 a 63 caracteres; una clave como `u3b6Fthb` sí es válida.
