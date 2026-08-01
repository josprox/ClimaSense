# Pruebas

Ejecute `./scripts/test.sh` en Linux/WSL con Go, Joss 3.6.3 o posterior y curl disponibles. Incluye tests Go de compensacion BMP280/BME280, chip ID, fallo de bus, cierre, HMAC, manipulacion, replay temporal, cliente WebSocket, activacion y rechazo de CA no confiable; carga RPC real de ambos JP desde Joss; parser y configuracion Edge; cola SQLite; migraciones, esquema SaaS e indices. La prueba HTTP local comprueba salud y rechazos sin sesion. Una prueba E2E adicional inicia sesion, autentica `/ws/dashboard`, crea una empresa por HTTP y exige recibir el evento en vivo correspondiente.

La integracion SaaS fue recorrida con SQLite: bootstrap de administrador, login JWT, alta de Tecnito Tech, codigo de activacion, dispositivo ligado por serial, ubicacion, asignacion, telemetria WS con ACK, consulta Open-Meteo y analisis contextual. La migracion y consultas de las 14 tablas funcionales tambien se ejecutaron contra PostgreSQL real.

Las formulas se verifican con el vector del datasheet: temperatura aproximada `25.08 C` y presion `100653.27 Pa`.

El paquete Raspberry Pi OS se valida con `scripts/test-raspios.sh`: estructura, servicios systemd, permisos, instalador, seleccion de interfaz y scripts de recuperacion. `build-raspios-bundle.sh` genera su SHA-256.

En hardware se valido Raspberry Pi OS, adaptador Wi-Fi USB, I2C en `0x76`, identificacion BME280 `0x60`, almacenamiento de una medicion y sincronizacion con ACK. Aun requieren pruebas repetidas: desconexion electrica del sensor, direccion `0x77`, corte de energia durante escritura, pulsacion prolongada de GPIO17 y recuperacion tras multiples fallos consecutivos de asociacion.
