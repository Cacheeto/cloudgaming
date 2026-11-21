#!/usr/bin/env bash
# Arma un SDK de Android "estándar" (con sdkmanager) para compilar el cliente con Gradle.
# Correr dentro del contenedor:  distrobox enter fedora -- ./preparar-sdk.sh
set -euo pipefail

SDK=${ANDROID_HOME:-$HOME/.cache/cloudgaming-android/sdk-gradle}
CMDLINE_URL=https://dl.google.com/android/repository/commandlinetools-linux-16111833_latest.zip
CMDLINE_SHA1=e025545c62a8e64c7559119566a569fb1dec5f60

mkdir -p "$SDK"
if [[ ! -x $SDK/cmdline-tools/latest/bin/sdkmanager ]]; then
  echo "· cmdline-tools"
  tmp=$(mktemp -d)
  trap 'rm -rf "$tmp"' EXIT
  curl -sSfL "$CMDLINE_URL" -o "$tmp/c.zip"
  echo "$CMDLINE_SHA1  $tmp/c.zip" | sha1sum -c -
  unzip -q "$tmp/c.zip" -d "$tmp"
  mkdir -p "$SDK/cmdline-tools"
  mv "$tmp/cmdline-tools" "$SDK/cmdline-tools/latest"
fi

# android-36 es para gomobile (red-cloudgaming.sh), que no entiende "android-37.0".
echo "· paquetes (sdkmanager verifica las sumas solo)"
# "yes" acepta las licencias; sin pipefail porque "yes" termina con SIGPIPE.
set +o pipefail
yes | "$SDK/cmdline-tools/latest/bin/sdkmanager" --sdk_root="$SDK" \
  "platform-tools" "platforms;android-37.0" "platforms;android-36" "build-tools;37.0.0" "ndk;29.0.14206865" \
  | grep -v '^\[' || true
set -o pipefail
test -d "$SDK/ndk/29.0.14206865"
echo "Listo: $SDK"
