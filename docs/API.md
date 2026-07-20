# API

Endpoints implementados:

- `GET /`
- `GET /login`, `GET /setup`, `GET /admin`, `GET /app`
- `POST /api/v1/auth/login` y `POST /api/v1/auth/logout`
- `POST /api/v1/setup/admin`
- `GET /api/v1/admin/summary`
- `POST /api/v1/admin/organizations`
- `POST /api/v1/admin/activation-codes`
- `POST /api/v1/admin/analysis/run`
- `GET /api/v1/tenant/summary`
- `POST /api/v1/tenant/locations`
- `POST /api/v1/tenant/devices/{device_id}/location`
- `POST /api/v1/edge/activate`
- `WS /ws/edge`
- `POST /api/v1/devices/provision`
- `POST /api/v1/devices/heartbeat`
- `POST /api/v1/telemetry`
- `POST /api/v1/telemetry/batch`
- `GET /api/v1/devices/{device_id}/config`
- `GET /api/v1/devices/{device_id}/updates`
- `GET /api/v1/devices/{device_id}/latest`
- `GET /api/v1/devices/{device_id}/measurements?from=&to=&limit=`
- `GET /health` y `GET /ready`

Las solicitudes de dispositivo llevan `X-Device-ID`, `X-Timestamp`, `X-Sequence`, `X-Nonce` y `X-Signature`. El cuerpo maximo de lote implementado es 100 mediciones. `received_at` lo asigna el servidor.

En WebSocket esos mismos campos viajan dentro de un envelope `telemetry_batch` junto con `payload`. El ACK incluye `device_id`, `accepted`, `duplicates` y `accepted_through`; el Edge no elimina su cola hasta validar la confirmacion correspondiente.

Las APIs `admin` requieren una sesion Joss con rol `admin`; las APIs `tenant` resuelven la empresa desde `organization_users` y nunca aceptan un `organization_id` proporcionado por el navegador. El codigo de activacion se devuelve una sola vez, se almacena como hash y cambia de `available` a `claimed` atomically durante el alta del Edge.

Los errores usan codigos HTTP reales mediante `ApiResponse`: `400` para telemetria invalida, `401` para autenticacion rechazada, `404` para ausencia de datos, `409` para activacion o correo duplicado y `503` cuando el servicio no esta listo. Altas y activaciones exitosas devuelven `201`.
