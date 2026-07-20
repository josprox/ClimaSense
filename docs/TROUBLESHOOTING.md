# Solucion de problemas

- Sin `/dev/i2c-1`: confirme `dtparam=i2c_arm=on`, kernel `I2C_CHARDEV` y device tree.
- `permission denied`: confirme usuario `climasense`, grupo `i2c` y modo `0660`.
- Chip ID `0x60`: el modulo es BME280; este driver lo rechaza intencionalmente.
- Cola crece: revise hora UTC, DNS, URL HTTP, credencial y respuesta del servidor.
- Firma invalida: compare JSON, secuencia, nonce y reloj; no reutilice solicitudes.
- Servicio reinicia: ejecute el diagnostico del plugin y revise espacio disponible en `/data`.
- Una respuesta de error aparece como HTTP 200: use `ApiResponse::error` en controladores ClimaSense. Es la capa de compatibilidad para el dispatcher JSON de Joss 3.6.1; no sustituya esa llamada por `Response::error` hasta actualizar a un runtime que preserve `status_code`.
- `empty($map["key"])` rechaza un valor que si existe: en Joss 3.6.1 asigne primero `$value = $map["key"]` y despues evalue `empty($value)`. La misma precaucion aplica a llamadas dentro de `empty()`.
