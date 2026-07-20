# Estado de implementacion

| Componente | Estado | Evidencia |
|---|---|---|
| Inspeccion Joss 3.6.1 | IMPLEMENTADO Y PROBADO | Los 20 documentos de `Joss-language/docs` y los contratos ejecutables relevantes fueron revisados; baseline registrado |
| Plugin I2C Linux | IMPLEMENTADO, PENDIENTE DE PRUEBA EN HARDWARE | Compila linux/arm64 sin CGO; ioctl, timeout, reintentos y cierre |
| Driver BMP280 | IMPLEMENTADO, PENDIENTE DE PRUEBA EN HARDWARE | Vector de compensacion y errores probados |
| Cola Edge SQLite | IMPLEMENTADO Y PROBADO | Persistencia, orden, idempotencia y ACK WebSocket por dispositivo probados en Windows |
| Onboarding Edge | IMPLEMENTADO, PENDIENTE DE PRUEBA EN HARDWARE | Primer arranque solicita Wi-Fi y activacion; admite preseed headless; serial ligado al codigo |
| Edge completo | PARCIAL | Ciclo Joss y WebSocket funcionales; workers concurrentes, heartbeat/config remota pendientes |
| Seguridad de aplicacion | IMPLEMENTADO Y PROBADO | HMAC, replay temporal y comparacion constante probados; transporte configurado deliberadamente como HTTP sin TLS |
| SaaS multiempresa | IMPLEMENTADO Y PROBADO | Admin, clientes, organizaciones, ubicaciones, JWT/roles, activacion, aislamiento por empresa y paneles verificados |
| Telemetria WebSocket | IMPLEMENTADO Y PROBADO | Envelope HMAC, replay, persistencia, duplicados y ACK `accepted_through` recorridos E2E |
| Inteligencia ambiental | IMPLEMENTADO Y PROBADO | Cron de 10 minutos, rango por ubicacion, 20 muestras, Open-Meteo, fallback y recomendacion explicable |
| Servidor Joss | OPERATIVO | 28 rutas HTTP/WS, SQLite y PostgreSQL real probados; panel admin y tenant revisados visualmente |
| OTA | NO INICIADO | Solo modelo y endpoint de consulta |
| Buildroot | IMPLEMENTADO Y PROBADO EN BUILD | Buildroot 2025.02.16 completo en WSL; kernel 6.6.28-v8, DTB Zero 2 W, Wi-Fi Broadcom, I2C y BMP280 compilados |
| Imagen `.img` | IMPLEMENTADA, PENDIENTE DE PRUEBA EN HARDWARE | SHA-256, MBR, FAT32, ext4 y `e2fsck` verificados; Joss, Edge, JP, firmware Wi-Fi, permisos y servicios auditados dentro de la imagen |

## Linea base del runtime

El repositorio Joss original no fue modificado. `go test ./...` en la copia de consulta ejecuto correctamente todos los paquetes salvo `cmd/joss`, que en esa copia clonada carecia de `cmd/joss/runner_windows.exe`. El repositorio indicado por el usuario si contiene ese artefacto y permitio compilar el runtime ARM64 externamente.

Joss 3.6.1 crea `Response::json` y `Response::error` con el campo interno `status_code`, pero su dispatcher HTTP JSON consulta `status`. ClimaSense no modifica el runtime: `ApiResponse` serializa JSON mediante `Response::raw`, cuyo dispatcher si conserva `status_code`. Se comprobaron respuestas reales `401 Unauthorized` para acceso operador y provisionamiento sin credencial.

En el mismo runtime, `empty()` comprueba correctamente identificadores, pero considera inexistente una expresion de indice de mapa o una llamada. ClimaSense asigna primero esos resultados a variables y aplica `empty()` sobre la variable. La carga del JSON Edge se ejecuta en pruebas para impedir regresiones de este contrato.

Los callbacks anonimos de `WebSocket::onMessage` no conservan `$ws` en Joss 3.6.1. ClimaSense registra un metodo ligado (`$this->handleMessage`) y guarda la conexion en la instancia del controlador; asi el ACK se envia por el mismo socket sin modificar el runtime.

Joss 3.6.1 traduce `double()` como `DOUBLE`, tipo no aceptado por PostgreSQL, y puede registrar la migracion aunque falle la creacion. ClimaSense usa `DECIMAL(14,4)` para las magnitudes del sensor e incluye migraciones compensatorias que completan instalaciones parciales sin borrar datos.

No se ha medido RAM en hardware. La imagen previa mide 288 MiB; debe regenerarse con el plugin 0.2.0 antes de grabar hardware. Falta validar el asistente, Wi-Fi, I2C/BMP280, watchdog y consumo real en una Pi Zero 2 W fisica.
