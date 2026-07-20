# Pruebas

Ejecute `./scripts/test.sh` en Linux/WSL con Go, Joss y curl disponibles. Incluye tests Go de compensacion BMP280, chip ID, fallo de bus, cierre, HMAC, manipulacion, replay temporal, cliente WebSocket, activacion y rechazo de CA no confiable; carga RPC real de ambos JP desde Joss; parser y configuracion Edge; cola SQLite; migraciones, esquema SaaS e indices. La prueba HTTP local adicional comprueba `/health`, `/ready`, el resumen del dashboard y que los rechazos sin sesion devuelvan codigos reales.

La integracion SaaS fue recorrida con SQLite: bootstrap de administrador, login JWT, alta de Tecnito Tech, codigo de activacion, dispositivo ligado por serial, ubicacion, asignacion, telemetria WS con ACK, consulta Open-Meteo y analisis contextual. La migracion y consultas de las 14 tablas funcionales tambien se ejecutaron contra PostgreSQL real.

Las formulas se verifican con el vector del datasheet: temperatura aproximada `25.08 C` y presion `100653.27 Pa`.

La imagen construida se valida con SHA-256, tabla MBR, particiones FAT32/ext4, `e2fsck` y una auditoria interna de runtime, plugins, firmware Wi-Fi, usuarios, permisos y scripts de inicio.

No estan cubiertos sin hardware: bus real, ambas direcciones fisicas, desconexion electrica, asociacion Wi-Fi, watchdog, corte de energia y arranque de la imagen.
