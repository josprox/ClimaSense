# Arquitectura

El Edge ejecuta `joss run main.joss`, mide mediante `climasense_hardware.jp`, persiste primero en SQLite/WAL y despues envia lotes por `/ws/edge` con `climasense_transport.jp`. El servidor Joss verifica HMAC, ventana temporal, nonce y secuencia, confirma cada lote por WebSocket, evita duplicados mediante una restriccion unica y persiste en PostgreSQL.

Los sidecars JP v2 son procesos autocontenidos sin CGO. Cada operacion I2C abre, configura, usa y cierra `/dev/i2c-N`; no mantiene descriptores entre llamadas. Joss conserva la logica de configuracion, adquisicion, cola, endpoints y consultas.

La capa SaaS aisla `organizations`, `organization_users`, `locations`, codigos de activacion, dispositivos, clima y analisis. Joss Auth administra contrasenas bcrypt, JWT y roles `admin`/`client`. El administrador ve todo el sistema; el cliente solo consulta su `organization_id`.

Cada diez minutos, Cron ejecuta `AnalysisService`: toma hasta 20 muestras recientes, consulta la condicion exterior de Open-Meteo, aplica el rango configurado para la ubicacion y guarda un veredicto, severidad y recomendacion. La version `climasense-context-v1` es un motor inteligente contextual y explicable; no se presenta como modelo predictivo entrenado.

Buildroot instala el runtime, ambos JP, el codigo Edge, BusyBox init, wpa_supplicant, i2c-tools y watchdog. `/data/climasense` contiene configuracion, token y SQLite. Los servicios `S44` y `S55` realizan el onboarding de Wi-Fi y activacion desde la consola del primer arranque.
