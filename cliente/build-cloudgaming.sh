#!/usr/bin/env bash
# CloudGaming: compila el cliente (Moonlight + menú propio) y lo deja donde el panel lo sirve:
# cloudgaming.apk para el celu (64 bits) y cloudgaming-tv.apk para la tele (32 bits).
# Correr dentro del contenedor:  distrobox enter fedora -- ./build-cloudgaming.sh
# Baja prioridad y 6 trabajos: la PC se usa para jugar mientras compila.
set -euo pipefail

AQUI=$(cd "$(dirname "$0")" && pwd)
export ANDROID_HOME=${ANDROID_HOME:-$HOME/.cache/cloudgaming-android/sdk-gradle}
DESCARGAS=$AQUI/../host/web/descargas

cd "$AQUI"
if ! unzip -l app/libs/libtailscale.aar 2>/dev/null | grep -q armeabi-v7a; then
  echo "Falta el motor de red (de 64 y 32 bits): correr primero red-cloudgaming.sh" >&2
  exit 1
fi
APK=app/build/outputs/apk/nonRoot/release/app-nonRoot-release.apk
for destino in arm64-v8a:cloudgaming.apk armeabi-v7a:cloudgaming-tv.apk; do
  abi=${destino%%:*} salida=$DESCARGAS/${destino#*:}
  echo "· $abi"
  nice -n 19 ./gradlew --no-daemon --max-workers=6 assembleNonRootRelease -PcloudgamingAbis="$abi" "$@"
  "$ANDROID_HOME/build-tools/37.0.0/apksigner" verify --print-certs "$APK" 2>/dev/null | head -1
  cp "$APK" "$salida"
  echo "Listo: $salida ($(du -h "$salida" | cut -f1))"
done
