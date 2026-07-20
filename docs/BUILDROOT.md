# Buildroot

La version fijada es 2025.02.16 LTS y su SHA-256 esta en `os/buildroot/buildroot.sha256`. El script parte de `raspberrypizero2w_64_defconfig`, aplica `climasense.fragment` y un fragmento de kernel.

La configuracion habilita ARM64, Wi-Fi Broadcom, I2C chardev, ext4, CA TLS, rng-tools, i2c-tools y watchdog. No instala escritorio, X11, Wayland, audio ni NetworkManager. Se usa `wpa_supplicant` por su menor huella.

La cuenta root se bloquea y no se instala una cuenta interactiva predeterminada. `climasense` tiene UID/GID 1000 fijo, contrasena bloqueada, shell `/bin/false` y pertenencia al grupo `i2c`; la configuracion inicial se prepara montando la particion ext4 de la imagen como explica la guia de flashing.

La configuracion incluye el firmware/NVRAM `brcmfmac-sdio-firmware-rpi` para el Wi-Fi del Zero 2 W, ademas de `brcmfmac`, `i2c-dev`, los buses BCM2835 y el soporte BMP280 del kernel.

La compilacion completa requiere Linux/WSL, toolchain de Buildroot, red y varios gigabytes libres. En WSL, el arbol se extrae en `/tmp/climasense-buildroot` para conservar enlaces y permisos que NTFS no representa fielmente; `CLIMASENSE_BUILD_DIR` permite cambiarlo. No se debe marcar la imagen como funcional hasta arrancarla en una Pi Zero 2 W real.
