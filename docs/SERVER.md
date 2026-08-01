# Servidor

Configure `server/.env` fuera del repositorio a partir de `env.example`, ejecute `joss migrate` y despues `joss server start`. Se requiere Joss 3.6.3 o posterior; 3.6.1 no implementa los canales que mantienen los paneles en vivo. El Dockerfile fija 3.6.3 y el script de pruebas rechaza versiones anteriores. El proceso escucha HTTP en el puerto interno 8080; el dominio publico debe terminar HTTPS/WSS en el proxy. PostgreSQL sin SSL solo es aceptable en una red privada equivalente.

Si el proxy termina TLS, no defina `TLS_CERT_FILE` ni `TLS_KEY_FILE` en Joss. El puerto 8080 no debe publicarse directamente a Internet.

Mantenga `SESSION_COOKIE_SECURE=true` en produccion. Solo debe cambiarse a `false` para una prueba local servida directamente por `http://127.0.0.1`; de lo contrario el navegador no enviara la sesion sobre HTTP, como corresponde.

Los dispositivos se aprovisionan con `X-Provisioning-Key`; el token se devuelve una sola vez y solo su SHA-256 queda almacenado. Telemetria y configuracion usan HMAC. Historial y ultima medicion requieren `X-Operator-Token`.

El panel grafico queda disponible en `/`, su resumen JSON en `/api/v1/dashboard/summary` y la disponibilidad de base de datos en `/ready`. El esquema se verifico tanto con SQLite como con PostgreSQL real; la carga concurrente continua pendiente.

`/admin`, `/app` y el dashboard mantienen `WS /ws/dashboard`. El primer mensaje contiene el JWT obtenido por `GET /api/v1/auth/ws-token`; el servidor valida el token manualmente y responde `ready`. Activaciones, telemetria, heartbeats, ubicaciones, codigos y analisis publican `refresh`, y el navegador consulta el resumen inmediatamente. La reconexion usa backoff con jitter y una consulta de respaldo cada 20 segundos solo mientras el socket no esta listo.

Los canales de Joss son locales al proceso. Mantenga una sola replica del servidor para entrega en vivo. Antes de escalar horizontalmente se necesita un backplane pub/sub y afinidad de WebSocket; sin ello la base de datos seguira correcta, pero un navegador conectado a otra replica dependera del respaldo periodico. El proxy inverso debe conservar HTTP/1.1 y reenviar `Upgrade` y `Connection`; HTTPS externo convierte automaticamente el cliente a `wss://`.

## Primer administrador y clientes

1. Defina `SAAS_BOOTSTRAP_KEY` en `.env`; si se omite, la primera instalacion acepta `APP_KEY` como fallback.
2. Abra `/setup` y cree el unico administrador inicial.
3. Inicie sesion en `/login`, cree una empresa y entregue al cliente el correo y contrasena inicial.
4. Genere un codigo de activacion para esa empresa y entreguelo con el producto.
5. El cliente entra en `/app`, crea una o mas ubicaciones y asigna cada Edge activado.

`Cron::schedule("*/10 * * * *")` ejecuta el analisis ambiental. Tambien puede dispararse manualmente desde `POST /api/v1/admin/analysis/run` para diagnostico. Open-Meteo se consulta con latitud/longitud; si falla, el analisis reutiliza el ultimo snapshot disponible y, si tampoco existe, continua sin contexto exterior.
