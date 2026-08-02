# ClimaSense Server

Esta rama contiene únicamente el servicio SaaS de ClimaSense. Usa siempre la última release estable de Joss; `>=3.6.3` permanece como contrato mínimo para los canales WebSocket. Expone paneles administrativos y de cliente, activación de dispositivos, telemetría, análisis ambiental y actualizaciones autenticadas en tiempo real.

Los paneles mantienen una conexión WebSocket persistente con `/ws/dashboard`. El navegador obtiene primero un token efímero en `/api/v1/auth/ws-token`; los eventos se filtran por rol, usuario y organización. Las mutaciones producen una actualización inmediata y, además, el cliente reconcilia el estado cada 20 segundos para detectar dispositivos que dejaron de enviar datos aunque el WebSocket del navegador siga conectado.

Cada lote de telemetría aceptado actualiza `devices.last_seen_at` y publica una invalidación inmediata al panel. El estado «en línea» significa que el servidor recibió datos durante los últimos tres minutos; «activo» sólo describe que la credencial del dispositivo no ha sido revocada.

El análisis contextual se ejecuta dentro del proceso del servidor al iniciar y después cada 600 segundos, configurable mediante `ANALYSIS_INTERVAL_SECONDS` con un mínimo de 60. Para cada dispositivo activo con ubicación asignada, combina las muestras nuevas, el tipo y rango térmico del espacio, sus coordenadas y el clima exterior obtenido de Open-Meteo. Si no hay telemetría posterior al último análisis, no genera duplicados. Cada resultado se publica por WebSocket al panel correspondiente.

## Desarrollo y pruebas

Configura `env.joss` a partir de `env.example` para desarrollo. Para ejecutar el conjunto completo en Linux o WSL:

```sh
./scripts/bootstrap.sh
./scripts/test.sh
```

`bootstrap.sh` descarga la última release oficial, valida el SHA-256 y la conserva por versión en `cache/joss-release`. Las pruebas no compilan Joss-language ni dependen de una instalación global.

La prueba levanta el servidor en `127.0.0.1:18080`, valida rutas públicas y protegidas, abre WebSockets reales de dispositivo y panel, envía un lote firmado, comprueba `last_seen_at` y confirma que el panel recibe la invalidación inmediatamente. También valida el análisis contextual según ubicación, evita análisis duplicados y comprueba que dos intentos concurrentes no puedan consumir el mismo código de activación.

## Despliegue

El `Dockerfile` usa `JOSS_VERSION=latest` con el instalador oficial no interactivo y está preparado para Dokploy. Una versión concreta sólo se pasa como argumento cuando se necesite reproducir un despliegue anterior. Define secretos distintos y robustos para `JWT_SECRET`, `APP_KEY` y `SAAS_BOOTSTRAP_KEY`; en producción conserva `SESSION_COOKIE_SECURE=true` y sirve el sitio mediante HTTPS/WSS.

Durante el build, Docker instala el paquete versionado `climasense_transport.jp` en la ubicación de autoload y ejecuta una llamada criptográfica real. El despliegue falla de inmediato si el payload nativo no puede registrarse; no se publica un contenedor que vaya a responder 500 al generar o consumir códigos de activación.

No subas `.env`, `env.joss`, sesiones, bases SQLite ni bitácoras. La única dependencia binaria rastreada en esta rama es `plugins/climasense_transport/0.2.0/climasense_transport.jp`, necesaria para que una compilación autónoma pueda iniciar sin depender del monorepo.
