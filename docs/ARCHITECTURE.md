# Arquitectura

El Edge ejecuta `joss run service.joss`, mide mediante `climasense_hardware.jp`, persiste primero en SQLite/WAL y despues envia lotes por `/ws/edge` con `climasense_transport.jp`. El servidor Joss verifica HMAC, ventana temporal, nonce y secuencia, confirma cada lote por WebSocket, evita duplicados mediante una restriccion unica y persiste en PostgreSQL.

Los sidecars JP v2 son procesos autocontenidos sin CGO. Cada operacion I2C abre, configura, usa y cierra `/dev/i2c-N`; no mantiene descriptores entre llamadas. Joss conserva la logica de configuracion, adquisicion, cola, endpoints y consultas.

La capa SaaS aisla `organizations`, `organization_users`, `locations`, codigos de activacion, dispositivos, clima y analisis. Joss Auth administra contrasenas bcrypt, JWT y roles `admin`/`client`. El administrador ve todo el sistema; el cliente solo consulta su `organization_id`.

Cada diez minutos, Cron ejecuta `AnalysisService`: toma hasta 20 muestras recientes, consulta la condicion exterior de Open-Meteo, aplica el rango configurado para la ubicacion y guarda un veredicto, severidad y recomendacion. La version `climasense-context-v1` es un motor inteligente contextual y explicable; no se presenta como modelo predictivo entrenado.

Buildroot instala el runtime, ambos JP, el codigo Edge, BusyBox init, wpa_supplicant, i2c-tools y watchdog. `/data/climasense` contiene configuracion, token y SQLite. `S44` mantiene el punto de acceso y el portal web durante el onboarding; cuando el navegador entrega Wi-Fi, activacion y credenciales del hotspot privado, `S45` conecta la red y `S55` consume el codigo sin solicitar datos por consola. `S97` supervisa GPIO17 y, tras una pulsacion de 5 segundos, pausa el Edge y cambia temporalmente el unico radio Wi-Fi a modo AP. Al guardar, vuelve a modo cliente y reinicia el servicio Edge.
