#!/bin/sh
# ============================================================
# ClimaSense Server — Entrypoint Docker para Dokploy
# ============================================================

set -e

ENV_FILE="/app/env.joss"

echo "[entrypoint] Generando env.joss desde variables de entorno..."

python3 -c '
import os
ignore = {"PATH", "HOME", "HOSTNAME", "TERM", "SHLVL", "PWD", "_", "OLDPWD", "DEBIAN_FRONTEND", "LANG"}
with open("/app/env.joss", "w") as f:
    for k, v in os.environ.items():
        if k in ignore or k.startswith(("PYTHON", "PIP", "GPG")):
            continue
        val = v.replace("\"", "\\\"")
        f.write(f"{k}=\"{val}\"\n")
'

echo "[entrypoint] ✓ env.joss generado con éxito."

echo "[entrypoint] Ejecutando migraciones de base de datos..."
joss migrate

echo "[entrypoint] Iniciando ClimaSense Server..."
exec joss server start
