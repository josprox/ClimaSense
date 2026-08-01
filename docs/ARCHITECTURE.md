# Arquitectura

El Edge ejecuta `joss run service.joss`, mide mediante `climasense_hardware.jp`, persiste primero en SQLite/WAL y despues envia lotes por `/ws/edge` con `climasense_transport.jp`. El servidor Joss verifica HMAC, ventana temporal, nonce y secuencia, confirma cada lote por WebSocket, evita duplicados mediante una restriccion unica y persiste en PostgreSQL.

Los sidecars JP v2 son procesos autocontenidos sin CGO. Cada operacion I2C abre, configura, usa y cierra `/dev/i2c-N`; no mantiene descriptores entre llamadas. Joss conserva la logica de configuracion, adquisicion, cola, endpoints y consultas.

La capa SaaS aisla `organizations`, `organization_users`, `locations`, codigos de activacion, dispositivos, clima y analisis. Joss Auth administra contrasenas bcrypt, JWT y roles `admin`/`client`. El administrador ve todo el sistema; el cliente solo consulta su `organization_id`.

El transporte tiene dos rutas separadas. `/ws/edge` recibe un lote autenticado por HMAC, confirma y cierra; `/ws/dashboard` mantiene una conexion autenticada por JWT para cada navegador. Las mutaciones y los lotes publican una invalidacion en canales locales de administrador, usuario o empresa. El navegador agrupa eventos cercanos, vuelve a leer su resumen y reconecta con backoff; solo consulta cada 20 segundos cuando el socket no esta listo.

Cada diez minutos, Cron ejecuta `AnalysisService`: toma hasta 20 muestras recientes, consulta la condicion exterior de Open-Meteo, aplica el rango configurado para la ubicacion y guarda un veredicto, severidad y recomendacion. La version `climasense-context-v1` es un motor inteligente contextual y explicable; no se presenta como modelo predictivo entrenado.

La plataforma de dispositivo vigente es Raspberry Pi OS Lite de 64 bits. El instalador coloca la aplicacion en `/opt/climasense/edge`, conserva estado en `/data/climasense` y registra servicios systemd para red, portal, activacion, Edge y mantenimiento por GPIO17. Selecciona dinamicamente una interfaz integrada o USB compatible con AP. Buildroot permanece en el repositorio solo como implementacion historica.
