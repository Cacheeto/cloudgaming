#!/usr/bin/env bash
# CloudGaming: compila el motor de red (libtailscale, el de la app de Tailscale para Android)
# y lo deja en app/libs/libtailscale.aar, de donde lo toma build-cloudgaming.sh.
# Solo hace falta correrlo la primera vez y cuando se cambia la versión fijada acá abajo.
# Correr dentro del contenedor:  distrobox enter fedora -- ./red-cloudgaming.sh
# Baja prioridad y 6 trabajos: la PC se usa para jugar mientras compila.
set -euo pipefail

AQUI=$(cd "$(dirname "$0")" && pwd)
CACHE=$HOME/.cache/cloudgaming-android
export ANDROID_HOME=${ANDROID_HOME:-$CACHE/sdk-gradle}
export ANDROID_NDK_HOME=$ANDROID_HOME/ndk/29.0.14206865
SALIDA=$AQUI/app/libs/libtailscale.aar
# arm64 para el celu y arm (32 bits) para la tele; build-cloudgaming.sh elige cuál va en cada APK.
# Para probar en el emulador:  DESTINOS=android/arm64,android/arm,android/amd64 ./red-cloudgaming.sh
DESTINOS=${DESTINOS:-android/arm64,android/arm}

# Versión fijada de tailscale-android (BSD-3): la estable 1.104.7, etiqueta
# 1.104.7-tfb3b9b296-g226f029a3. Para actualizar: cambiar el commit por el de otra etiqueta
# estable (número del medio par), y si su go.toolchain.rev cambió, también GO_REV y
# GO_SHA256 (sha256sum del .tar.gz bajado). Si cambió libtailscale/interfaces.go, hay que
# acompañarlo en EntornoRed, ServicioRed y VigiaRed.
TS_REPO=https://github.com/tailscale/tailscale-android.git
TS_COMMIT=226f029a3ef5d0263dfb6cf4c79a243eec99ed86
GO_REV=1635316a95598e9fe1bff7267604daa94aa1c8ec
GO_SHA256=366968d5a7a240585533149572b8becc9bac8d9e6ff2712a4865a21c16b9f814

FUENTES=$CACHE/tailscale-android
export TOOLCHAINDIR=$CACHE/tailscale-go-$GO_REV

# gomobile busca platforms/android-<número> y no entiende "android-37.0".
if ! ls -d "$ANDROID_HOME"/platforms/android-[0-9][0-9] >/dev/null 2>&1; then
  echo "Falta la plataforma android-36: correr preparar-sdk-cloudgaming.sh" >&2
  exit 1
fi

echo "· fuentes de tailscale-android ($TS_COMMIT)"
if [[ ! -d $FUENTES/.git ]]; then
  git init -q "$FUENTES"
  git -C "$FUENTES" remote add origin "$TS_REPO"
fi
if [[ $(git -C "$FUENTES" rev-parse -q --verify HEAD 2>/dev/null) != "$TS_COMMIT" ]]; then
  git -C "$FUENTES" fetch -q --depth 1 origin "$TS_COMMIT"
  git -C "$FUENTES" checkout -q --detach FETCH_HEAD
fi
if [[ $(<"$FUENTES/go.toolchain.rev") != "$GO_REV" ]]; then
  echo "go.toolchain.rev cambió: actualizar GO_REV y GO_SHA256 en este script" >&2
  exit 1
fi

if [[ ! -x $TOOLCHAINDIR/bin/go ]]; then
  echo "· Go de Tailscale ($GO_REV)"
  mkdir -p "$CACHE/descargas"
  tgz=$CACHE/descargas/tailscale-go-$GO_REV.tar.gz
  [[ -f $tgz ]] || curl -sSfL -o "$tgz" \
    "https://github.com/tailscale/go/releases/download/build-$GO_REV/linux-amd64.tar.gz"
  echo "$GO_SHA256  $tgz" | sha256sum -c -
  mkdir -p "$TOOLCHAINDIR"
  tar -C "$TOOLCHAINDIR" --strip-components=1 -xf "$tgz"
fi

# Lo mismo que hace el Makefile de tailscale-android, sin Docker ni make.
cd "$FUENTES"
export GOTOOLCHAIN=local TS_USE_TOOLCHAIN=1
export GOPATH=$CACHE/go GOCACHE=$CACHE/go-build GOBIN=$FUENTES/android/build/go/bin
export GOMAXPROCS=6
export PATH=$FUENTES/tool:$GOBIN:$PATH
unset GOROOT

echo "· gomobile"
nice -n 19 ./tool/go install golang.org/x/mobile/cmd/gobind golang.org/x/mobile/cmd/gomobile
nice -n 19 ./tool/go run tailscale.com/cmd/mkversion > tailscale.version

echo "· libtailscale ($DESTINOS); la primera vez tarda varios minutos"
mkdir -p "$(dirname "$SALIDA")"
nice -n 19 gomobile bind -target "$DESTINOS" -androidapi 26 \
  -tags "$(./build-tags.sh) ts_omit_cachenetmap" \
  -ldflags "-s -w -linkmode=external -extldflags=-Wl,-z,max-page-size=16384 $(./version-ldflags.sh)" \
  -o "$SALIDA" ./libtailscale
echo "Listo: $SALIDA ($(du -h "$SALIDA" | cut -f1))"
