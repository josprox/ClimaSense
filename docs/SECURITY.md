# Seguridad

El proceso Joss puede escuchar HTTP dentro de la red privada, pero el endpoint publico configurado por Edge es `https://climasense.joss.red` y debe terminar TLS en un proxy confiable. La firma de aplicacion se conserva: HMAC-SHA256 sobre timestamp, secuencia, nonce y SHA-256 del JSON canonico. La clave HMAC es SHA-256 del token; el servidor no guarda el token original. Las comparaciones de HMAC y hashes de acceso se realizan en tiempo constante dentro del plugin.

HTTP no protege tokens, firmas ni telemetria frente a observacion o alteracion de red. Nunca exponga directamente el puerto interno; el proxy publico debe ofrecer HTTPS/WSS con un certificado valido.

El servidor limita desfase a 300 segundos, registra nonces aceptados y aplica el rate limiting nativo de Joss. No registre tokens, Wi-Fi, firmas completas ni cuerpos sensibles. Rote una credencial marcando la anterior como revocada.

Las contrasenas de usuarios SaaS se procesan con bcrypt mediante `Auth::create`; la sesion usa JWT en una cookie HTTP-only, `SameSite=Lax` y `Secure` por defecto. Solo las pruebas HTTP locales definen `SESSION_COOKIE_SECURE=false`. El acceso administrativo se valida por rol y el acceso de cliente por membresia de empresa. Los codigos de activacion contienen 128 bits aleatorios, se guardan como SHA-256 y se consumen una sola vez. La imagen Edge publica no incluye token, empresa ni identidad final.

`server/storage` y `server/log.txt` son estado local ignorado y no se distribuyen. Una revision encontro que ambos habian sido versionados anteriormente; por ello, en cualquier instalacion que haya usado esa revision se debe rotar `JWT_SECRET` y cerrar sesiones existentes. Eliminarlos del ultimo commit no borra copias presentes en el historial remoto.

El hotspot inicial solo existe durante el alta local. Antes de activar el equipo, el usuario debe crear las credenciales del hotspot de mantenimiento WPA2. Hostapd necesita conservar su PSK localmente; `/data/climasense/hostapd.conf` y `maintenance-ap.pass` se almacenan con modo `0600` dentro de la particion de datos. La clave nunca se devuelve por la API ni se muestra de nuevo. El hotspot de mantenimiento permanece apagado durante la operacion normal y solo se habilita mediante una pulsacion fisica prolongada en GPIO17.

Las dos rutas WebSocket autentican dentro del primer mensaje porque Joss realiza el upgrade antes del middleware HTTP. `/ws/edge` mantiene HMAC, timestamp, secuencia y nonce; su ACK identifica al dispositivo. `/ws/dashboard` valida el JWT de sesion y suscribe canales por rol, usuario y empresa. La cookie sigue siendo HTTP-only: un endpoint autenticado y no cacheable entrega el token al JavaScript solo para memoria; nunca se guarda en URL, `localStorage` ni logs. Los eventos solo indican que debe recargarse una vista y no incluyen telemetria ni datos de otra empresa.

En Raspberry Pi OS, el instalador no cambia la cuenta creada mediante Imager ni deshabilita SSH/Connect. Oculta el getty del monitor local para que el dispositivo no espere un login fisico y ejecuta la aplicacion con la cuenta de sistema no interactiva `climasense`, contrasena bloqueada y grupos de hardware estrictamente necesarios.

Pendiente: politica automatica de retencion de nonces, recuperacion de contrasena con correo, mTLS opcional, backplane WebSocket para multiples replicas, auditoria externa y pruebas de penetracion. El proxy HTTPS debe reenviar `Upgrade` y `Connection`.
