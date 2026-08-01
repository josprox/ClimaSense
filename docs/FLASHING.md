# Grabacion e instalacion

La ruta soportada no graba una imagen ClimaSense propia. Use Raspberry Pi Imager para instalar Raspberry Pi OS Lite de 64 bits, configure usuario, SSH, Raspberry Pi Connect y una red temporal funcional —preferentemente mediante el adaptador USB si el radio integrado falla— y arranque la placa.

Genere `dist/climasense-raspios-installer.tar.gz` con `scripts/build-raspios-bundle.sh`, copielo a la Raspberry y ejecute `sudo sh os/raspios/install.sh`. El instalador conserva la conexion actual hasta el reinicio. El procedimiento detallado y los comandos de actualizacion estan en `RASPBERRY_PI_OS.md`.

Tras reiniciar aparece `Configuracion ClimaSense`; en `http://192.168.4.1:8080` se capturan Wi-Fi, codigo de activacion y credenciales del hotspot privado. No se escribe identidad ni token en la tarjeta y no se solicita ningun dato por consola. `scripts/flash-image.sh` y la imagen Buildroot se mantienen exclusivamente para reproducir el camino historico.
