#!/usr/bin/env bash
# CloudGaming Host: arma lo que se le pasa a otra persona para preparar su PC.
#   dist/cloudgaming-host-linux.tar.gz: el programa, el instalador y los APK de la app.
#   dist/instalar-cloudgaming.exe: el instalador de Windows, con el programa y los APK adentro.
# Compila primero (compilar.sh). Los APK salen de web/descargas (los deja ahí
# cliente/build-cloudgaming.sh).
set -euo pipefail

AQUI=$(cd "$(dirname "$0")" && pwd)
"$AQUI/compilar.sh"

cd "$AQUI/dist"
for apk in cloudgaming.apk cloudgaming-tv.apk; do
  [ -f "$AQUI/web/descargas/$apk" ] || { echo "Falta web/descargas/$apk: compilar la app primero" >&2; exit 1; }
done

CARPETA=cloudgaming-host-linux
rm -rf "$CARPETA"
mkdir -p "$CARPETA/descargas"
cp cloudgaming-host-linux-amd64 "$CARPETA/cloudgaming-host"
install -m 755 "$AQUI/linux/instalar.sh" "$CARPETA/instalar.sh"
cp "$AQUI/linux/LEEME.txt" "$CARPETA/"
cp "$AQUI"/web/descargas/cloudgaming.apk "$AQUI"/web/descargas/cloudgaming-tv.apk "$CARPETA/descargas/"
tar -czf "$CARPETA.tar.gz" "$CARPETA"
rm -rf "$CARPETA"
echo "Listo: dist/$CARPETA.tar.gz ($(du -h "$CARPETA.tar.gz" | cut -f1))"

# Windows: el instalador lleva adentro (go:embed) lo que se copia en instalador/adentro.
CACHE=$HOME/.cache/cloudgaming-android
GO=$(command -v go || ls -d "$CACHE"/tailscale-go-*/bin/go 2>/dev/null | head -1)
export GOTOOLCHAIN=local GOPATH=$CACHE/go CGO_ENABLED=0 GOFLAGS=-trimpath GOOS=windows GOARCH=amd64
ADENTRO=$AQUI/instalador/adentro
rm -rf "$ADENTRO"
mkdir -p "$ADENTRO/descargas"
cp cloudgaming-host-windows-amd64.exe "$ADENTRO/cloudgaming-host.exe"
cp "$AQUI"/web/descargas/cloudgaming.apk "$AQUI"/web/descargas/cloudgaming-tv.apk "$ADENTRO/descargas/"
(cd "$AQUI" && nice -n 19 "$GO" vet ./instalador && nice -n 19 "$GO" build -p 6 -ldflags "-s -w" -o dist/instalar-cloudgaming.exe ./instalador)
rm -rf "$ADENTRO"
echo "Listo: dist/instalar-cloudgaming.exe ($(du -h instalar-cloudgaming.exe | cut -f1))"
