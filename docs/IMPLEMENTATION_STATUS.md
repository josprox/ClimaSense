# Estado de implementacion

Actualizado tras la migracion y prueba fisica en Raspberry Pi OS.

| Componente | Estado | Evidencia |
|---|---|---|
| Runtime Joss | IMPLEMENTADO Y PROBADO | Servidor fijado a Joss `>=3.6.3`; suite y canales WebSocket ejecutados con 3.6.3 compilado desde el repositorio indicado |
| Plugin I2C Linux | IMPLEMENTADO Y PROBADO EN HARDWARE | `/dev/i2c-1` operativo y dispositivo detectado en `0x76` |
| Driver BMP280/BME280 | IMPLEMENTADO Y PROBADO | Vectores, errores y chip IDs `0x58`/`0x60` cubiertos; BME280 fisico leido correctamente |
| Cola Edge SQLite | IMPLEMENTADO Y PROBADO | WAL, orden, idempotencia, retencion y eliminacion posterior al ACK |
| Onboarding Edge | IMPLEMENTADO | Portal sin consola, Wi-Fi, activacion, hotspot WPA2 y recuperacion systemd incluidos |
| Interfaz Wi-Fi USB | IMPLEMENTADO Y PROBADO EN HARDWARE | Seleccion dinamica de interfaz con modo AP; el equipo probado opera con USB por fallo del radio integrado |
| Edge completo | OPERATIVO | Activacion completada, medicion almacenada y lote sincronizado (`sent:1`) en Raspberry Pi OS |
| Seguridad de dispositivo | IMPLEMENTADO Y PROBADO | HMAC, ventana temporal, nonce, comparacion constante, token individual y ACK por dispositivo |
| SaaS multiempresa | IMPLEMENTADO Y PROBADO | Roles, organizaciones, ubicaciones, activacion y aislamiento por membresia |
| Paneles en vivo | IMPLEMENTADO Y PROBADO E2E | JWT en primer mensaje, canales admin/usuario/empresa, evento HTTP→WS, reconexion y fallback |
| Inteligencia ambiental | IMPLEMENTADO Y PROBADO | Cron, muestras recientes, Open-Meteo/fallback y recomendacion contextual explicable |
| Raspberry Pi OS | RUTA SOPORTADA | Instalador idempotente, servicios systemd, I2C/GPIO, rescate de red y bundle con checksum |
| Buildroot | LEGACY | Conservado para referencia; no es la ruta recomendada ni resuelve el radio fisicamente defectuoso |
| OTA | NO IMPLEMENTADO | El instalador puede reaplicarse conservando datos, pero aun no existen firma, health check ni rollback OTA |

## Contratos relevantes de Joss

- Los paneles requieren Joss 3.6.3 porque `subscribe`, `publish`, `subscriberCount`, `onClose` y la publicacion desde handlers HTTP no estan disponibles en el binario 3.6.1 instalado previamente.
- El upgrade WebSocket ocurre antes del middleware HTTP. `/ws/dashboard` recibe el JWT como primer mensaje y llama `Auth::validateToken`; `/ws/edge` conserva su autenticacion HMAC propia.
- Los canales WebSocket son locales al proceso. El despliegue en vivo debe usar una replica hasta incorporar un backplane pub/sub.
- `ApiResponse` sigue usando `Response::raw` para conservar codigos HTTP reales en los runtimes compatibles usados por el proyecto.
- Las variables numericas del entorno se convierten mediante `JSON::parse`; se evita depender de coerciones implicitas entre cadenas y enteros.

## Validacion ejecutada

`scripts/test.sh` pasa con Joss 3.6.3 e incluye Go test/vet, RPC de plugins, configuracion y cola Edge, migraciones, indices, rutas, codigos HTTP y una prueba real que inicia sesion, autentica `/ws/dashboard`, realiza una mutacion HTTP y recibe el evento `refresh`.

En la Raspberry Pi se confirmo BME280 en `0x76`, activacion persistente, servicio Edge activo, `Medicion almacenada: secuencia=1` y `Sincronizacion: {"ok":true,"sent":1}`. Quedan como validaciones prolongadas el corte de energia, direccion `0x77`, recuperacion repetida de AP y mantenimiento por GPIO17 bajo fallos consecutivos.
