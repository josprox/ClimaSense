#!/bin/bash
set -euo pipefail

echo "=================================================="
echo "    Preparando Entorno WSL para ClimaSense        "
echo "=================================================="

# 1. Instalar paquetes de sistema para Go y Buildroot
echo "[1/4] Instalando paquetes requeridos (Go, Buildroot, C toolchain)..."
sudo apt-get update
sudo apt-get install -y build-essential golang-go git curl wget bc bzip2 cpio unzip rsync file python3 libncurses-dev libssl-dev gawk

# 2. Verificar o ubicar JOSS_SOURCE
JOSS_PATH="/mnt/c/Users/joss/Documents/proyectos/Joss-language"
if [ ! -d "$JOSS_PATH" ]; then
    echo "ERROR: No se encontro el repositorio Joss en $JOSS_PATH" >&2
    exit 1
fi

export JOSS_SOURCE="$JOSS_PATH"
echo "[2/4] JOSS_SOURCE detectado en: $JOSS_SOURCE"

# 3. Compilar e instalar joss CLI en /usr/local/bin
echo "[3/4] Compilando binario 'joss' de Linux e instalando en /usr/local/bin..."
(cd "$JOSS_SOURCE" && go build -o /tmp/joss ./cmd/joss)
sudo mv /tmp/joss /usr/local/bin/joss
sudo chmod +x /usr/local/bin/joss

# 4. Validar ejecucion del bootstrap
echo "[4/4] Validando herramientas instaladas..."
echo "Go: $(go version)"
echo "Joss: $(joss version 2>/dev/null || echo 'joss instalado')"

echo "=================================================="
echo " ¡Entorno WSL configurado con exito!             "
echo " Puedes proceder a compilar la imagen Edge con:   "
echo "   export JOSS_SOURCE=\"$JOSS_PATH\"               "
echo "   ./scripts/bootstrap.sh                        "
echo "   ./scripts/build-runtime.sh                    "
echo "   ./scripts/build-plugins.sh                    "
echo "   ./scripts/build-edge.sh                       "
echo "   ./scripts/build-server.sh                     "
echo "   ./scripts/test.sh                             "
echo "   ./scripts/build-os.sh                         "
echo "=================================================="
