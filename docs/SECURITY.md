# Seguridad

El despliegue actual usa HTTP sin certificado por decision explicita del proyecto. La firma de aplicacion se conserva: HMAC-SHA256 sobre timestamp, secuencia, nonce y SHA-256 del JSON canonico. La clave HMAC es SHA-256 del token; el servidor no guarda el token original. Las comparaciones de HMAC y hashes de acceso se realizan en tiempo constante dentro del plugin.

HTTP no protege tokens, firmas ni telemetria frente a observacion o alteracion de red. Use esta configuracion solamente en una red local confiable; para exponer el servicio a Internet vuelva a habilitar TLS o coloque un proxy HTTPS delante.

El servidor limita desfase a 300 segundos, registra nonces aceptados y aplica el rate limiting nativo de Joss. No registre tokens, Wi-Fi, firmas completas ni cuerpos sensibles. Rote una credencial marcando la anterior como revocada.

Las contrasenas de usuarios SaaS se procesan con bcrypt mediante `Auth::create`; la sesion usa JWT HTTP-only. El acceso administrativo se valida por rol y el acceso de cliente por membresia de empresa. Los codigos de activacion contienen 128 bits aleatorios, se guardan como SHA-256 y se consumen una sola vez. La imagen Edge publica no incluye token, empresa ni identidad final.

El hotspot inicial solo existe durante el alta local. Antes de activar el equipo, el usuario debe crear las credenciales del hotspot de mantenimiento WPA2. Hostapd necesita conservar su PSK localmente; `/data/climasense/hostapd.conf` y `maintenance-ap.pass` se almacenan con modo `0600` dentro de la particion de datos. La clave nunca se devuelve por la API ni se muestra de nuevo. El hotspot de mantenimiento permanece apagado durante la operacion normal y solo se habilita mediante una pulsacion fisica prolongada en GPIO17.

La ruta WebSocket autentica dentro del mensaje porque Joss realiza el upgrade antes del middleware HTTP. Cada envelope mantiene HMAC, timestamp, secuencia y nonce; el ACK identifica al dispositivo para impedir que una conexion acepte la confirmacion de otra.

En ClimaSense OS, la cuenta `root` queda bloqueada y el post-build elimina defensivamente cualquier contrasena vacia heredada del defconfig. El servicio usa la cuenta no interactiva `climasense`, con UID/GID 1000 estable, contrasena bloqueada y solamente el grupo suplementario `i2c`.

Pendiente: politica automatica de retencion de nonces, recuperacion de contrasena con correo, mTLS opcional, auditoria externa y pruebas de penetracion. Para un SaaS expuesto a Internet, el HTTP actual debe quedar detras de un proxy HTTPS que reenvie `Upgrade` y `Connection`.
