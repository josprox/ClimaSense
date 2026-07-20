# Configuracion de Raspberry Pi

1. Grabe la imagen siguiendo `FLASHING.md`.
2. Conecte pantalla y teclado para el primer arranque, o preinstale `wpa_supplicant.conf` y `activation.code` en `/data/climasense` para una instalacion sin pantalla.
3. Encienda la Raspberry. `S44climasense-onboarding` solicita SSID y contrasena sin mostrarla; almacena solo el PSK derivado.
4. Cuando exista red, `S55climasense-activation` solicita el numero entregado por la empresa. El codigo se consume una sola vez y se liga al serial de la placa.
5. El sistema guarda la configuración en `.env` y `device.token`, inicia el servicio y comienza a enviar por WebSocket.
6. En el portal de empresa, asigne el dispositivo a una ubicacion con coordenadas y rango termico.
7. Compruebe `/dev/i2c-1`, `/etc/init.d/S95climasense status` y los logs de consola.

No conecte ni retire el sensor con la placa energizada.
