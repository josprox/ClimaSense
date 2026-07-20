# Servidor

Configure `server/.env` fuera del repositorio a partir de `env.example`, ejecute `joss migrate` y despues `joss server start`. La configuracion entregada inicia HTTP sin certificado en `http://localhost:8080` y usa PostgreSQL sin SSL.

No defina `TLS_CERT_FILE` ni `TLS_KEY_FILE`: si cualquiera aparece en `.env`, Joss intentara habilitar HTTPS. Este modo es deliberadamente inseguro y solo debe exponerse en una red confiable.

Los dispositivos se aprovisionan con `X-Provisioning-Key`; el token se devuelve una sola vez y solo su SHA-256 queda almacenado. Telemetria y configuracion usan HMAC. Historial y ultima medicion requieren `X-Operator-Token`.

El panel grafico queda disponible en `/`, su resumen JSON en `/api/v1/dashboard/summary` y la disponibilidad de base de datos en `/ready`. El esquema se verifico tanto con SQLite como con PostgreSQL real; la carga concurrente continua pendiente.

## Primer administrador y clientes

1. Defina `SAAS_BOOTSTRAP_KEY` en `.env`; si se omite, la primera instalacion acepta `APP_KEY` como fallback.
2. Abra `/setup` y cree el unico administrador inicial.
3. Inicie sesion en `/login`, cree una empresa y entregue al cliente el correo y contrasena inicial.
4. Genere un codigo de activacion para esa empresa y entreguelo con el producto.
5. El cliente entra en `/app`, crea una o mas ubicaciones y asigna cada Edge activado.

`Cron::schedule("*/10 * * * *")` ejecuta el analisis ambiental. Tambien puede dispararse manualmente desde `POST /api/v1/admin/analysis/run` para diagnostico. Open-Meteo se consulta con latitud/longitud; si falla, el analisis reutiliza el ultimo snapshot disponible y, si tampoco existe, continua sin contexto exterior.
