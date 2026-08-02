# ============================================================
# ClimaSense Server — Dockerfile para Dokploy
# ============================================================

FROM python:3.11-slim

# System deps: curl, ca-certificates, sqlite3
RUN apt-get update && apt-get install -y --no-install-recommends \
        curl \
        unzip \
        ca-certificates \
        sqlite3 \
    && rm -rf /var/lib/apt/lists/*

# Instala la ultima version publicada. Puede fijarse temporalmente con
# --build-arg JOSS_VERSION=x.y.z para reproducir un despliegue anterior.
ARG JOSS_VERSION=latest
ENV JOSS_VERSION=${JOSS_VERSION}

# Instalar binario joss desde el instalador oficial de Docker.
RUN curl -fsSL \
    https://raw.githubusercontent.com/josprox/Joss-language/main/install/docker-install.sh \
    | bash \
    && joss version

WORKDIR /app

# Copiar todo el código del servidor
COPY . .

# Instala el JP versionado donde el cargador lo prioriza y comprueba una
# llamada nativa real antes de aceptar la imagen.
RUN sh scripts/install-server-plugin.sh \
    && joss run tests/transport-token.joss \
    && chmod +x entrypoint.sh scripts/install-server-plugin.sh

# Directorio de almacenamiento para SQLite / archivos de sesión
RUN mkdir -p storage/logs

# Puerto por defecto del servidor (sobrescribible con PORT)
EXPOSE 8080

ENTRYPOINT ["/app/entrypoint.sh"]
