# ClimaSense AI

ClimaSense AI es una plataforma SaaS multiempresa para Raspberry Pi Zero 2 W y BMP280, escrita principalmente en Joss. Incluye panel administrativo, portal por empresa, multiples ubicaciones, activacion de dispositivos, telemetria WebSocket, clima exterior y recomendaciones ambientales cada diez minutos. I2C, BMP280, cliente WebSocket y criptografia viven en plugins JP v2 escritos en Go porque esas capacidades no existen en Joss 3.6.1.

El flujo operativo es: el administrador crea la empresa y un codigo de compra; el Edge solicita Wi-Fi y ese codigo durante el primer arranque; el servidor lo vincula a la empresa; el cliente asigna una ubicacion y rango termico; finalmente el motor `climasense-context-v1` compara las muestras interiores con Open-Meteo y genera una recomendacion explicable.

## Construccion

En Linux o WSL:

```sh
export JOSS_SOURCE=/ruta/de/solo/lectura/Joss-language
./scripts/bootstrap.sh
./scripts/build-runtime.sh
./scripts/build-plugins.sh
./scripts/build-edge.sh
./scripts/build-server.sh
./scripts/test.sh
./scripts/build-os.sh
./scripts/manifest.sh
```

`build-os.sh` usa Buildroot 2025.02.16 LTS y produce `dist/climasense-os-rpi-zero-2-w.img`. La imagen fue construida y auditada estructuralmente; el arranque y los perifericos en hardware real todavia no han sido validados. Consulta `docs/IMPLEMENTATION_STATUS.md`.

Nunca copie `env.example` sin sustituir credenciales mediante un canal seguro. Los tokens no deben entrar al repositorio. El primer administrador se crea en `/setup` con `SAAS_BOOTSTRAP_KEY` (o `APP_KEY` como fallback de primera instalacion).
