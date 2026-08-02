#!/bin/bash
set -euo pipefail

echo "=================================================="
echo "    Preparando Entorno WSL para ClimaSense        "
echo "=================================================="

# 1. Instalar paquetes de sistema para Go y Buildroot
echo "[1/4] Instalando paquetes requeridos (Go, Buildroot, C toolchain)..."
sudo apt-get update
sudo apt-get install -y build-essential golang-go git curl wget bc bzip2 cpio unzip rsync file python3 libncurses-dev libssl-dev gawk

# 2. El runtime se obtiene de la ultima release oficial; no requiere clonar
# ni compilar el repositorio Joss-language.
echo "[2/4] Se usara la ultima release oficial de Joss."

# 3. Descargar el Joss anfitrion dentro de dist/tools y el runtime ARM64.
# No reemplaza /usr/local/bin/joss ni depende de una version global antigua.
echo "[3/4] Descargando y verificando la ultima release de Joss..."
sh "$(dirname "$0")/prepare-joss-release.sh"
JOSS_CURRENT="$(dirname "$0")/joss-current.sh"

# 4. Validar ejecucion del bootstrap
echo "[4/4] Validando herramientas instaladas..."
echo "Go: $(go version)"
echo "Joss actual: $(sh "$JOSS_CURRENT" version)"

echo "=================================================="
echo " ¡Entorno WSL configurado con exito!             "
echo " Puedes proceder a compilar la imagen Edge con:   "
echo "   ./scripts/bootstrap.sh                        "
echo "   ./scripts/build-runtime.sh                    "
echo "   ./scripts/build-plugins.sh                    "
echo "   ./scripts/build-edge.sh                       "
echo "   ./scripts/build-server.sh                     "
echo "   ./scripts/test.sh                             "
echo "   ./scripts/build-os.sh                         "
echo "=================================================="
