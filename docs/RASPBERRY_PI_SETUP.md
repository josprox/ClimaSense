# Configuracion de Raspberry Pi

1. Grabe la imagen siguiendo `FLASHING.md`.
2. Encienda la Raspberry y conectese desde un telefono o computadora a `Configuracion ClimaSense`.
3. Abra `http://192.168.4.1:8080`, seleccione la red Wi-Fi e ingrese su contrasena.
4. Ingrese el numero de activacion y cree un nombre y contrasena de 8 a 63 caracteres para el hotspot privado de mantenimiento.
5. La Raspberry apaga el punto de acceso, se conecta a Internet y consume el codigo una sola vez, ligado al serial de la placa. No se requiere pantalla ni teclado.

Para volver a configurar la red despues de la activacion, conecte un pulsador normalmente abierto entre GPIO17 (pin fisico 11) y GND (pin fisico 9). Mantengalo presionado durante 5 segundos. El Edge pausa temporalmente la telemetria y publica el hotspot privado configurado durante el alta. Conectese a esa red, abra `http://192.168.4.1:8080` y cambie el Wi-Fi o las credenciales del hotspot. Al guardar, el hotspot se apaga y el Edge retoma la conexion al servidor.

No conecte GPIO17 directamente a 3.3 V ni a 5 V. `config.txt` configura una resistencia pull-up interna, por lo que el pulsador solo debe cerrar el circuito entre GPIO17 y GND.
5. El sistema guarda la configuración en `.env` y `device.token`, inicia el servicio y comienza a enviar por WebSocket.
6. En el portal de empresa, asigne el dispositivo a una ubicacion con coordenadas y rango termico.
7. Compruebe `/dev/i2c-1`, `/etc/init.d/S95climasense status` y los logs de consola.

No conecte ni retire el sensor con la placa energizada.
