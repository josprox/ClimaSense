# Grabacion

Verifique primero el dispositivo de bloque; un destino incorrecto destruye datos.

```sh
./scripts/flash-image.sh /dev/sdX
```

El script exige escribir `FLASH`, usa bloques de 4 MiB, `fsync` y `sync`. Alternativamente use Raspberry Pi Imager con `dist/climasense-os-rpi-zero-2-w.img` y verifique el SHA-256 publicado.

El flujo normal ya no requiere escribir identidad ni token en la tarjeta: conecte pantalla y teclado y el asistente solicitara Wi-Fi y activacion. Para un despliegue headless puede preconfigurar solo estos dos archivos:

```sh
sudo mount /dev/sdX2 /mnt/climasense
sudo install -o 1000 -g 1000 -m 0600 os/buildroot/board/rootfs-overlay/etc/wpa_supplicant.conf.example /mnt/climasense/data/climasense/wpa_supplicant.conf
printf '%s\n' 'CODIGO-DE-ACTIVACION' | sudo tee /mnt/climasense/data/climasense/activation.code >/dev/null
sudo chown 1000:1000 /mnt/climasense/data/climasense/activation.code
sudo chmod 0600 /mnt/climasense/data/climasense/activation.code
sudo umount /mnt/climasense
```

Edite la red antes de desmontar. La URL SaaS se fija en `config.example.json` durante el build; el dispositivo obtiene identidad y token al consumir el codigo. Sustituya `/dev/sdX2` solo despues de verificar el dispositivo correcto.
