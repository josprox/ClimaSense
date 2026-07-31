# Compilacion incremental de la imagen

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
