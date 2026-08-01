# Configuracion de Raspberry Pi

1. Instale Raspberry Pi OS Lite de 64 bits y el paquete siguiendo `RASPBERRY_PI_OS.md`.
2. Reinicie y conectese desde un telefono o computadora a `Configuracion ClimaSense`.
3. Abra `http://192.168.4.1:8080`, seleccione la red Wi-Fi e ingrese su contrasena WPA2 de 8 a 63 caracteres.
4. Ingrese el codigo de activacion y cree nombre y contrasena para el hotspot privado de mantenimiento.
5. La Raspberry cambia a modo cliente, obtiene Internet, consume el codigo ligado al serial e inicia telemetria. No requiere teclado, pantalla ni login local.
6. En `/app`, asigne el dispositivo a una ubicacion con coordenadas y rango termico.

Para reconfigurar despues del alta, conecte un pulsador normalmente abierto entre GPIO17 (pin fisico 11) y GND (pin fisico 9) y mantengalo presionado cinco segundos. El Edge se pausa, aparece el hotspot privado y el portal permite cambiar Wi-Fi o sus propias credenciales. Al guardar, vuelve el modo cliente y se restablecen Edge, SSH y Raspberry Pi Connect.

No conecte GPIO17 a 3.3 V ni a 5 V: la resistencia pull-up es interna. No conecte ni retire el sensor con la placa energizada. Valide con `i2cdetect -y 1` y `systemctl status climasense-activation.service climasense-edge.service`; los servicios anteriores `/etc/init.d/S*` solo pertenecen a Buildroot.
