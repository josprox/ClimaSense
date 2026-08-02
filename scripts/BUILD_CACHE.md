# Compilacion incremental de la imagen

> Flujo historico de Buildroot. El despliegue actual con Raspberry Pi OS genera un instalador mediante `scripts/build-raspios-bundle.sh` y no recompila una imagen completa.

Joss tampoco se recompila. `scripts/bootstrap.sh` descarga la ultima release oficial, verifica el SHA-256 y la guarda por version en `cache/joss-release`. El cache solo vuelve a descargar el ZIP cuando cambia la release o falla su checksum.

Estos scripts son alternativos y no modifican `build-os.sh`.

La primera preparacion descarga las fuentes seleccionadas:

```sh
./scripts/prefetch-os-sources.sh
```

Para construir conservando el arbol compilado y las descargas entre sesiones
de WSL:

```sh
./scripts/build-os-cached.sh
./scripts/manifest.sh
```

Las siguientes ejecuciones reutilizan por defecto:

```text
~/.cache/climasense-os/build
~/.cache/climasense-os/downloads
```

Para consultar el espacio utilizado:

```sh
./scripts/inspect-os-cache.sh
```

Puede cambiar la ubicacion sin editar scripts:

```sh
export CLIMASENSE_CACHE_ROOT="$HOME/otra-ruta"
```
