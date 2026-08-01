# Aplicacion Edge

`edge/service.joss` carga configuracion validada, inicializa SQLite con WAL y `synchronous=FULL`, mide BMP280 o BME280 con reintentos acotados, asigna una secuencia monotona, persiste y envia lotes por WebSocket. Solo elimina filas despues de recibir un ACK que contiene su propio `device_id` y `accepted_through`. La clave primaria evita duplicados y la retencion elimina primero registros antiguos al superar el limite.

`edge/activate.joss` intercambia el codigo de activacion de un solo uso y el serial de la Raspberry por un `device_id` y token individual. El token se escribe con modo `0600`; el servidor conserva solo SHA-256. La misma imagen puede distribuirse a todos los clientes porque no contiene identidad ni credenciales finales.

La version actual ejecuta muestreo y sincronizacion en el mismo ciclo; el timeout WebSocket esta acotado, pero una llamada lenta puede retrasar el siguiente muestreo. Separar ambos workers mediante canales Joss permanece pendiente.

En Raspberry Pi OS, `climasense-activation.service` completa identidad y token antes de que `climasense-edge.service` opere normalmente. Un `DEVICE_ID` vacio se representa internamente como `replace-after-provisioning` para permitir validaciones de instalacion, pero no sustituye la activacion ni crea una credencial valida.
