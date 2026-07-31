# Solucion de problemas

- Sin `/dev/i2c-1`: confirme `dtparam=i2c_arm=on`, kernel `I2C_CHARDEV` y device tree.
- `permission denied`: confirme usuario `climasense`, grupo `i2c` y modo `0660`.
- Chip ID `0x60`: el modulo es BME280; este driver lo rechaza intencionalmente.
- Cola crece: revise hora UTC, DNS, URL HTTP, credencial y respuesta del servidor.
- Firma invalida: compare JSON, secuencia, nonce y reloj; no reutilice solicitudes.
- Servicio reinicia: ejecute el diagnostico del plugin y revise espacio disponible en `/data`.
- Una respuesta de error aparece como HTTP 200: use `ApiResponse::error` en controladores ClimaSense. Es la capa de compatibilidad para el dispatcher JSON de Joss 3.6.1; no sustituya esa llamada por `Response::error` hasta actualizar a un runtime que preserve `status_code`.
- `empty($map["key"])` rechaza un valor que si existe: en Joss 3.6.1 asigne primero `$value = $map["key"]` y despues evalue `empty($value)`. La misma precaucion aplica a llamadas dentro de `empty()`.
# Diagnostico de arranque

El equipo no requiere ni ofrece login local. Si aparece `climasense login:`, la tarjeta contiene una imagen anterior y debe regenerarse y grabarse nuevamente.

Durante un primer arranque correcto deben aparecer mensajes `ClimaSense:` indicando que se espera `wlan0`, se levanta el hotspot y el portal queda listo. El detalle persistente se guarda en `/data/climasense/onboarding.log`.

`export_store: invalid GPIO 17` identifica una imagen anterior que usaba la interfaz GPIO sysfs. La implementacion actual utiliza `gpioget gpiochip0 17`.
