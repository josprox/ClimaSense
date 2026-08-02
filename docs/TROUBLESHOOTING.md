# Solucion de problemas

- Sin `/dev/i2c-1`: confirme `dtparam=i2c_arm=on`, kernel `I2C_CHARDEV` y device tree.
- `permission denied`: confirme usuario `climasense`, grupo `i2c` y modo `0660`.
- Chip ID `0x60`: el modulo es BME280 y esta soportado para temperatura y presion. Si un plugin anterior lo rechaza, reinstale el paquete que contiene `climasense_hardware` 0.1.1 o posterior.
- Cola crece: revise hora UTC, DNS, URL HTTP, credencial y respuesta del servidor.
- Firma invalida: compare JSON, secuencia, nonce y reloj; no reutilice solicitudes.
- Servicio reinicia: ejecute el diagnostico del plugin y revise espacio disponible en `/data`.
- Una respuesta de error aparece como HTTP 200: ejecute `scripts/bootstrap.sh` para preparar la ultima release de Joss y conserve `ApiResponse::error` en los controladores ClimaSense. La suite comprueba los codigos reales.
- `empty($map["key"])` rechaza un valor que si existe: asigne primero `$value = $map["key"]` y despues evalue `empty($value)`. La misma precaucion aplica a llamadas dentro de `empty()` en codigo que deba seguir siendo compatible.
- El panel solo cambia al recargar: compruebe que el despliegue usa la ultima release de Joss, `WS /ws/dashboard` y que el proxy reenvie `Upgrade`/`Connection`. La interfaz indica "respaldo por consulta periodica" cuando el socket esta caido. En varias replicas sin backplane, mantenga una sola replica para entrega inmediata.
- `GET /api/v1/auth/ws-token` devuelve `401`: la sesion expiro; vuelva a iniciar sesion. El token no debe copiarse a la URL ni almacenarse manualmente.
# Diagnostico de arranque

El equipo no requiere ni ofrece login local. Si aparece `climasense login:`, la tarjeta contiene una imagen anterior y debe regenerarse y grabarse nuevamente.

Durante un primer arranque correcto se levanta el hotspot y el portal queda listo. En Raspberry Pi OS la interfaz Wi-Fi se detecta dinamicamente —incluidos adaptadores USB— y se guarda en `/data/climasense/wifi.interface`; no se presupone que se llame `wlan0`. El detalle persistente se guarda en `/data/climasense/onboarding.log`.

Si el Wi-Fi integrado no aparece a nivel kernel pero el USB funciona, no fuerce `wlan0`: reinstale con el adaptador USB conectado o defina `CLIMASENSE_WIFI_INTERFACE` durante la instalacion. El instalador solo selecciona una interfaz que anuncie modo AP.

`export_store: invalid GPIO 17` identifica una imagen anterior que usaba la interfaz GPIO sysfs. En Raspberry Pi OS la implementacion detecta automaticamente si `gpioget` usa la interfaz 1.x o 2.x de libgpiod.
