# ClimaSense Server

Esta rama contiene únicamente el servicio SaaS de ClimaSense. Requiere Joss 3.6.3 o posterior y expone paneles administrativos y de cliente, activación de dispositivos, telemetría, análisis ambiental y actualizaciones autenticadas en tiempo real.

Los paneles mantienen una conexión WebSocket persistente con `/ws/dashboard`. El navegador obtiene primero un token efímero en `/api/v1/auth/ws-token`; los eventos se filtran por rol, usuario y organización. Si el WebSocket se interrumpe, el cliente reconecta con espera progresiva y usa una consulta de respaldo cada 20 segundos.

## Desarrollo y pruebas

Configura `env.joss` a partir de `env.example` para desarrollo. Para ejecutar el conjunto completo en Linux o WSL:

```sh
./scripts/test.sh
```

La prueba levanta el servidor en `127.0.0.1:18080`, valida rutas públicas y protegidas, abre un WebSocket real, crea telemetría por HTTP y confirma que el panel recibe la invalidación inmediatamente. También comprueba que dos intentos concurrentes no puedan consumir el mismo código de activación.

## Despliegue

El `Dockerfile` instala exactamente Joss 3.6.3 y está preparado para Dokploy. Define secretos distintos y robustos para `JWT_SECRET`, `APP_KEY` y `SAAS_BOOTSTRAP_KEY`; en producción conserva `SESSION_COOKIE_SECURE=true` y sirve el sitio mediante HTTPS/WSS.

No subas `.env`, `env.joss`, sesiones, bases SQLite ni bitácoras. La única dependencia binaria rastreada en esta rama es `plugins/climasense_transport/0.2.0/climasense_transport.jp`, necesaria para que una compilación autónoma pueda iniciar sin depender del monorepo.
