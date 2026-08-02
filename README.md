# ClimaSense AI

ClimaSense AI es una plataforma SaaS multiempresa para Raspberry Pi Zero 2 W y sensores BMP280 o BME280, escrita principalmente en Joss. Del BME280 se utilizan temperatura y presion; la humedad queda reservada para una ampliacion posterior. Incluye panel administrativo, portal por empresa, multiples ubicaciones, activacion de dispositivos, telemetria WebSocket, paneles actualizados en vivo, clima exterior y recomendaciones ambientales cada diez minutos. I2C, sensores, cliente WebSocket y criptografia viven en plugins JP v2 escritos en Go. El proyecto obtiene siempre la ultima release estable de Joss; `>=3.6.3` permanece solamente como contrato minimo de compatibilidad.

El flujo operativo es: el administrador crea la empresa y un codigo de compra; el Edge solicita Wi-Fi y ese codigo durante el primer arranque; el servidor lo vincula a la empresa; el cliente asigna una ubicacion y rango termico; finalmente el motor `climasense-context-v1` compara las muestras interiores con Open-Meteo y genera una recomendacion explicable.

## Construccion

En Linux o WSL:

```sh
./scripts/bootstrap.sh
./scripts/build-runtime.sh
./scripts/build-plugins.sh
./scripts/build-edge.sh
./scripts/build-server.sh
./scripts/test.sh
./scripts/build-raspios-bundle.sh
./scripts/manifest.sh
```

`bootstrap.sh` consulta `releases/latest`, descarga el ZIP oficial para Linux o macOS, verifica su SHA-256 con `release-manifest.json` y lo conserva en `cache/joss-release`. Del ZIP de Linux extrae el runtime ARM64 para la Raspberry. No compila Joss-language y no depende del `joss` global instalado. Cuando se publica una version nueva, la siguiente ejecucion descarga esa version; las compilaciones posteriores reutilizan el cache.

La ruta de despliegue actual usa Raspberry Pi OS Lite de 64 bits instalado con Raspberry Pi Imager. `build-raspios-bundle.sh` produce `dist/climasense-raspios-installer.tar.gz`; el procedimiento completo esta en `docs/RASPBERRY_PI_OS.md`. La integracion anterior con Buildroot se conserva unicamente como referencia.

Los paneles `/admin` y `/app` reciben invalidaciones por `WS /ws/dashboard` y consultan de inmediato su resumen autorizado. Si el socket se interrumpe, el navegador reconecta con backoff y usa una consulta periodica de respaldo hasta recuperar el canal.

Nunca copie `env.example` sin sustituir credenciales mediante un canal seguro. Los tokens no deben entrar al repositorio. El primer administrador se crea en `/setup` con `SAAS_BOOTSTRAP_KEY` (o `APP_KEY` como fallback de primera instalacion).
