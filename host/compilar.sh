#!/usr/bin/env bash
# CloudGaming Host: compila el programa de la PC para Linux y para Windows (64 bits),
# con la página del panel adentro. Deja los ejecutables en host/dist/.
# Usa el Go que ya está para la app (el de Tailscale, en ~/.cache/cloudgaming-android)
# si no hay otro. Baja prioridad: la PC se usa para jugar mientras compila.
set -euo pipefail

AQUI=$(cd "$(dirname "$0")" && pwd)
CACHE=$HOME/.cache/cloudgaming-android
GO=$(command -v go || ls -d "$CACHE"/tailscale-go-*/bin/go 2>/dev/null | head -1)
if [ -z "$GO" ]; then
  echo "No encuentro Go: correr primero cliente/red-cloudgaming.sh (lo baja)" >&2
  exit 1
fi
export GOTOOLCHAIN=local GOPATH=$CACHE/go CGO_ENABLED=0 GOFLAGS=-trimpath

cd "$AQUI"
mkdir -p dist
for destino in linux/amd64 windows/amd64; do
  so=${destino%/*} arq=${destino#*/}
  salida=dist/cloudgaming-host-$so-$arq
  banderas="-s -w"
  if [ "$so" = windows ]; then
    salida=$salida.exe
    banderas="$banderas -H=windowsgui" # sin ventana de consola
  fi
  echo "· $so/$arq"
  GOOS=$so GOARCH=$arq nice -n 19 "$GO" vet .
  GOOS=$so GOARCH=$arq nice -n 19 "$GO" build -p 6 -ldflags "$banderas" -o "$salida" .
  echo "Listo: $salida ($(du -h "$salida" | cut -f1))"
done
