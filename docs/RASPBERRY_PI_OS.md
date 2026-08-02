# Migracion a Raspberry Pi OS

ClimaSense migra a Raspberry Pi OS Lite de 64 bits para usar el kernel y el firmware oficiales de la Raspberry Pi Zero 2 W. Buildroot se conserva temporalmente como referencia, pero no es la ruta de despliegue nueva.

1. Desde Raspberry Pi Imager grabe Raspberry Pi OS Lite (64-bit), configure usuario, SSH, Raspberry Pi Connect y la red del adaptador USB.
2. Arranque la Raspberry y confirme acceso por SSH o Connect.
3. En el equipo de desarrollo ejecute `sh scripts/bootstrap.sh`; descargara la ultima release oficial y extraera de ella el runtime ARM64. Compile los plugins y ejecute `sh scripts/build-raspios-bundle.sh`.
4. Copie `dist/climasense-raspios-installer.tar.gz` a la Raspberry y extraigalo.
5. Ejecute `sudo sh os/raspios/install.sh`. La conexion actual se conserva hasta reiniciar.
6. Reinicie. El adaptador seleccionado publicara `Configuracion ClimaSense` en `http://192.168.4.1:8080`.

El equipo no espera datos en la terminal. El instalador deshabilita el prompt de login del monitor local; SSH y Raspberry Pi Connect permanecen habilitados y los servicios ClimaSense arrancan mediante systemd.

En el primer arranque Raspberry Pi OS puede mantener `dpkg` ocupado mediante PackageKit. El instalador espera hasta cinco minutos a que termine; nunca elimine manualmente `/var/lib/dpkg/lock-frontend`.

El instalador comprueba que el adaptador USB soporte modo AP antes de cambiar la configuracion. Si no lo soporta, termina con error y conserva la conexion administrada por Raspberry Pi OS.

Si el hotspot, DHCP o el portal web no pueden arrancar, el modo de rescate elimina la exclusion de NetworkManager y recupera la conexion que Raspberry Pi Imager dejo guardada. El detalle queda en `/data/climasense/onboarding.log` y la causa resumida en `/data/climasense/onboarding.failed`; volver a ejecutar el instalador limpia la marca y permite otro intento.

El instalador deja solamente la interfaz Wi-Fi seleccionada fuera de NetworkManager y controla ese radio con `hostapd`, `dnsmasq` y `wpa_supplicant`. Detecta interfaces integradas y USB, prefiere una USB compatible con modo AP y guarda la seleccion en `/data/climasense/wifi.interface`. Puede fijarse manualmente con `CLIMASENSE_WIFI_INTERFACE`. Los datos persistentes se conservan en `/data/climasense` y la aplicacion en `/opt/climasense/edge`.

Con un solo adaptador Wi-Fi, SSH y Raspberry Pi Connect se desconectaran temporalmente mientras el radio funciona como hotspot. Regresaran cuando el portal guarde la red cliente y el equipo vuelva a conectarse a Internet.

Si el adaptador no logra asociarse o conseguir una direccion IP en la red guardada, systemd inicia automaticamente el hotspot de recuperacion para permitir corregir el SSID o la contrasena.

Despues de la activacion, mantener GPIO17 presionado durante cinco segundos detiene temporalmente Edge y vuelve a publicar el hotspot privado configurado durante el alta. Al guardar los cambios, el adaptador regresa a modo cliente y se restablecen Edge, SSH y Raspberry Pi Connect.

Las contrasenas WPA2 del Wi-Fi cliente y del hotspot deben tener entre 8 y 63 caracteres. Una clave de exactamente 8 caracteres es valida. Las redes protegidas con claves de 4 o 6 caracteres no cumplen WPA2-PSK y no se aceptan.

El instalador habilita I2C y configura GPIO17 con resistencia pull-up en `config.txt`; ambos cambios se aplican en el primer reinicio.

Antes de distribuir una imagen final, valide en hardware: AP inicial, alta Wi-Fi, activacion, telemetria y recuperacion tras reinicio.

La validacion fisica actual se realizo con un adaptador Wi-Fi USB porque el radio integrado de la placa probada falla. El sensor detectado fue BME280 ID `0x60` en `0x76`; despues de instalar el plugin 0.1.1, Edge registro `Medicion almacenada` y `Sincronizacion: {"ok":true,"sent":1}`.

Para actualizar sin borrar la configuracion, extraiga un bundle nuevo en un directorio temporal y vuelva a ejecutar `sudo sh os/raspios/install.sh`. `/data/climasense` se conserva. Reinicie solamente cuando cambien parametros de arranque, I2C o seleccion de red.
