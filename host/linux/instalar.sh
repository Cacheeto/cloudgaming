#!/usr/bin/env bash
# CloudGaming Host para Linux: deja esta PC lista para jugar desde el celu con la app
# CloudGaming. Instala (o completa) lo que falte:
#   1. Tailscale: la red privada entre el celu y la PC, con tu cuenta.
#   2. Sunshine: el que manda la imagen al celu y recibe los controles.
#   3. CloudGaming Host: el panel que usa la app; arranca solo al entrar a la sesión.
#
# Se corre con tu usuario, no con sudo: la clave de administrador la pide solo para instalar.
#   ./instalar.sh                instalar o completar todo
#   ./instalar.sh --revisar      mostrar qué está y qué falta, sin tocar nada
#   ./instalar.sh --desinstalar  quitar CloudGaming Host (Tailscale y Sunshine quedan)
#   ./instalar.sh --si           instalar sin preguntar antes de empezar
set -euo pipefail

AQUI=$(cd "$(dirname "$0")" && pwd)
HOST_NUEVO=$AQUI/cloudgaming-host
DESTINO=$HOME/.local/share/cloudgaming-host
UNIDAD=cloudgaming-host.service
UNIDADES=${XDG_CONFIG_HOME:-$HOME/.config}/systemd/user
DATOS=${XDG_CONFIG_HOME:-$HOME/.config}/cloudgaming-panel # los del host (mismo nombre que el panel viejo)
PUERTO=47777
SUNSHINE_REPO=SunshineStream/Sunshine
SUNSHINE_FLATPAK=dev.lizardbyte.app.Sunshine
SUNSHINE_UNIDADES="app-dev.lizardbyte.app.Sunshine.service sunshine.service homebrew.sunshine.service"
# Solo para probar el instalador en una máquina sin nadie delante: no espera el inicio de
# sesión de Tailscale.
SIN_LOGIN=${CLOUDGAMING_SIN_LOGIN:-}

MODO=instalar
PREGUNTAR=true
for arg in "$@"; do
  case "$arg" in
    --revisar) MODO=revisar ;;
    --desinstalar) MODO=desinstalar ;;
    --si | -y) PREGUNTAR=false ;;
    -h | --help) sed -n '2,12p' "$0" | cut -c3-; exit 0 ;;
    *) echo "No conozco la opción «$arg» (probá --help)" >&2; exit 1 ;;
  esac
done

if [ -t 1 ]; then
  N=$'\e[1m' V=$'\e[32m' A=$'\e[33m' R=$'\e[31m' F=$'\e[0m'
else
  N='' V='' A='' R='' F=''
fi
titulo() { printf '\n%s%s%s\n' "$N" "$1" "$F"; }
ok() { printf '  %s✓%s %s\n' "$V" "$F" "$1"; }
falta() { printf '  %s•%s %s\n' "$A" "$F" "$1"; FALTAN=$((FALTAN + 1)); }
dice() { printf '    %s\n' "$1"; }
error() { printf '\n%sNo pude seguir:%s %s\n' "$R" "$F" "$1" >&2; exit 1; }
revisando() { [ "$MODO" = revisar ]; }
FALTAN=0

como_admin() {
  if [ "$(id -u)" -eq 0 ]; then
    "$@"
  elif command -v sudo >/dev/null; then
    sudo "$@"
  else
    error "hace falta sudo para instalar (o correr como administrador: $*)"
  fi
}

# bajar URL ARCHIVO («-» = a la salida)
bajar() {
  if command -v curl >/dev/null; then
    curl -fsSL --retry 2 -o "$2" "$1"
  elif command -v wget >/dev/null; then
    wget -q -O "$2" "$1"
  else
    error "hace falta curl o wget para bajar los programas"
  fi
}

# ------------------------------------------------------------------ el sistema

[ "$(id -u)" -ne 0 ] || error "corrélo con tu usuario, sin sudo: la clave de administrador te la pide cuando hace falta"
[ "$(uname -m)" = x86_64 ] || error "por ahora CloudGaming Host es solo para PC de 64 bits (x86_64)"
systemctl --user show-environment >/dev/null 2>&1 || error "no encuentro systemd de tu usuario (systemctl --user)"
[ -r /etc/os-release ] || error "no sé qué Linux es este (falta /etc/os-release)"
# shellcheck disable=SC1091
. /etc/os-release
ATOMICO=false # Fedora Silverblue, Kinoite, Bazzite...: el sistema es de solo lectura
[ -e /run/ostree-booted ] && ATOMICO=true
es() { case " ${ID:-} ${ID_LIKE:-} " in *" $1 "*) return 0 ;; esac; return 1; }

TMP=$(mktemp -d)
trap 'rm -rf "$TMP"' EXIT

# ------------------------------------------------------------------ desinstalar

if [ "$MODO" = desinstalar ]; then
  titulo "Quitar CloudGaming Host"
  systemctl --user disable --now "$UNIDAD" 2>/dev/null || true
  rm -f "$UNIDADES/$UNIDAD"
  systemctl --user daemon-reload
  rm -rf "$DESTINO"
  ok "CloudGaming Host quitado. Tailscale y Sunshine quedan instalados."
  dice "Tus datos (consumo, sesiones, acceso a Sunshine) siguen en $DATOS: borralos a mano si querés."
  exit 0
fi

# ------------------------------------------------------------------ 1. Tailscale

tailscale_estado() {
  tailscale status --json --peers=false 2>/dev/null | sed -n 's/.*"BackendState": *"\([A-Za-z]*\)".*/\1/p' | head -1
}

paso_tailscale() {
  titulo "1. Tailscale (la red privada entre el celu y la PC)"
  if ! command -v tailscale >/dev/null; then
    falta "Tailscale no está instalado"
    revisando && return
    $ATOMICO && error "en un sistema de solo lectura instalá Tailscale con rpm-ostree (https://tailscale.com/download/linux), reiniciá y volvé a correr esto"
    dice "Lo instalo con el instalador oficial de Tailscale…"
    bajar https://tailscale.com/install.sh "$TMP/tailscale.sh"
    sh "$TMP/tailscale.sh"
  fi
  ok "Tailscale instalado"

  if ! systemctl is-active --quiet tailscaled; then
    falta "el servicio de Tailscale está apagado"
    revisando && return
    como_admin systemctl enable --now tailscaled
    sleep 2
  fi
  ok "servicio de Tailscale andando"

  # «Operador»: así el host le puede consultar a Tailscale sin ser administrador.
  local operador estado
  operador=$(tailscale debug prefs 2>/dev/null | sed -n 's/.*"OperatorUser": *"\([^"]*\)".*/\1/p' | head -1)
  estado=$(tailscale_estado)
  if [ "$estado" != Running ]; then
    falta "esta PC no inició sesión en Tailscale"
    revisando && return
    if [ -n "$SIN_LOGIN" ]; then
      dice "(prueba: no espero el inicio de sesión)"
      como_admin tailscale set --operator="$USER"
      return
    fi
    dice "Abrí el enlace que aparece abajo e iniciá sesión con tu cuenta (Google, Microsoft,"
    dice "GitHub…). En el celu vas a entrar con ${N}esa misma cuenta${F}."
    como_admin tailscale up --operator="$USER"
    estado=$(tailscale_estado)
    [ "$estado" = Running ] || error "Tailscale no quedó conectado (estado: ${estado:-desconocido})"
  elif [ "$operador" != "$USER" ]; then
    falta "Tailscale no deja que tu usuario lo consulte"
    revisando && return
    como_admin tailscale set --operator="$USER"
  fi
  local cuenta
  cuenta=$(tailscale status --json --peers=false 2>/dev/null | sed -n 's/.*"LoginName": *"\([^"]*\)".*/\1/p' | head -1)
  ok "conectada a Tailscale${cuenta:+ con la cuenta $cuenta}"
}

# ------------------------------------------------------------------ 2. Sunshine

SUN_TIPO='' # nativo, brew o flatpak
SUN_CONFIG=$HOME/.config/sunshine

sunshine_detectar() {
  SUN_TIPO=''
  SUN_CONFIG=$HOME/.config/sunshine
  if command -v sunshine >/dev/null; then
    SUN_TIPO=nativo
  elif [ -x /home/linuxbrew/.linuxbrew/bin/sunshine ]; then
    SUN_TIPO=brew
  elif command -v flatpak >/dev/null && flatpak info "$SUNSHINE_FLATPAK" >/dev/null 2>&1; then
    SUN_TIPO=flatpak
    SUN_CONFIG=$HOME/.var/app/$SUNSHINE_FLATPAK/config/sunshine
  fi
}

sunshine_correr() {
  case "$SUN_TIPO" in
    nativo) sunshine "$@" ;;
    brew) /home/linuxbrew/.linuxbrew/bin/sunshine "$@" ;;
    flatpak) flatpak run --command=sunshine "$SUNSHINE_FLATPAK" "$@" ;;
  esac
}

sunshine_unidad() {
  local u
  for u in $SUNSHINE_UNIDADES; do
    if systemctl --user cat "$u" >/dev/null 2>&1; then
      echo "$u"
      return 0
    fi
  done
  return 1
}

# El paquete de la última versión de Sunshine que le va a este Linux (una expresión para
# elegirlo entre los archivos publicados). Sin paquete propio: se usa el de Flathub.
sunshine_patron() {
  $ATOMICO && return 1
  local v
  case "${ID:-}" in
    fedora) echo "\.fc${VERSION_ID}\.x86_64\.rpm$" ;;
    ubuntu) echo "(\+|%2B)ubuntu${VERSION_ID}_amd64\.deb$" ;;
    debian) echo "(\+|%2B)debian${VERSION_CODENAME:-}_amd64\.deb$" ;;
    opensuse-tumbleweed) echo "\.suse\.tw\.x86_64\.rpm$" ;;
    opensuse-leap) echo "\.suse\.lp${VERSION_ID//./}\.x86_64\.rpm$" ;;
    *)
      if es arch; then
        echo "-x86_64\.pkg\.tar\.zst$"
      elif es fedora; then # Nobara, Ultramarine…: misma versión que su Fedora
        echo "\.fc${VERSION_ID%%.*}\.x86_64\.rpm$"
      elif es ubuntu; then # Mint, Pop!_OS, Zorin…: el paquete del Ubuntu en que se basan
        case "${UBUNTU_CODENAME:-}" in
          jammy) v=22.04 ;;
          noble) v=24.04 ;;
          resolute) v=26.04 ;;
          *) return 1 ;;
        esac
        echo "(\+|%2B)ubuntu${v}_amd64\.deb$"
      else
        return 1
      fi
      ;;
  esac
}

sunshine_instalar() {
  local patron url archivo
  if patron=$(sunshine_patron); then
    url=$(bajar "https://api.github.com/repos/$SUNSHINE_REPO/releases/latest" - |
      grep -o '"browser_download_url": *"[^"]*"' | cut -d'"' -f4 | grep -E -e "$patron" | head -1 || true)
  fi
  if [ -n "${url:-}" ]; then
    archivo=$TMP/${url##*/}
    archivo=${archivo//%2B/+} # en la dirección, el «+» del nombre va como %2B
    dice "Bajo ${archivo##*/}…"
    bajar "$url" "$archivo"
    case "$archivo" in
      *.deb) como_admin apt-get install -y "$archivo" ;;
      *.pkg.tar.zst) como_admin pacman -U --noconfirm "$archivo" ;;
      *.rpm)
        if command -v zypper >/dev/null; then
          como_admin zypper --non-interactive install --allow-unsigned-rpm "$archivo"
        else
          como_admin dnf install -y "$archivo"
        fi
        ;;
    esac
  elif command -v flatpak >/dev/null; then
    dice "No hay paquete de Sunshine para este Linux: lo instalo desde Flathub…"
    flatpak remote-add --user --if-not-exists flathub https://dl.flathub.org/repo/flathub.flatpakrepo
    flatpak install --user -y flathub "$SUNSHINE_FLATPAK"
    # Reglas para el teclado, el mouse y los joysticks virtuales, y el servicio de usuario.
    flatpak run --command=additional-install.sh "$SUNSHINE_FLATPAK"
  else
    error "no encontré cómo instalar Sunshine en ${PRETTY_NAME:-este Linux}: instalalo a mano (https://github.com/$SUNSHINE_REPO) y volvé a correr esto"
  fi
}

# Sunshine guarda su usuario en sunshine_state.json (solo el usuario y la clave cifrada).
sunshine_tiene_usuario() {
  grep -q '"username"' "$SUN_CONFIG/sunshine_state.json" 2>/dev/null
}

# guardar_acceso USUARIO CLAVE: el host los prueba contra Sunshine y, si andan, los guarda.
guardar_acceso() {
  printf '%s\n%s\n' "$1" "$2" | "$HOST_NUEVO" -guardar-sunshine
}

CLAVE_NUEVA=''

paso_sunshine() {
  titulo "2. Sunshine (el que manda la imagen al celu)"
  sunshine_detectar
  if [ -z "$SUN_TIPO" ]; then
    falta "Sunshine no está instalado"
    revisando && return
    sunshine_instalar
    sunshine_detectar
    [ -n "$SUN_TIPO" ] || error "Sunshine no quedó instalado"
  fi
  ok "Sunshine instalado ($SUN_TIPO)"

  local unidad
  if ! unidad=$(sunshine_unidad); then
    falta "Sunshine no tiene servicio para arrancar solo"
    revisando && return
    error "no encuentro el servicio de Sunshine (probé $SUNSHINE_UNIDADES)"
  fi
  if ! systemctl --user is-enabled --quiet "$unidad" 2>/dev/null || ! systemctl --user is-active --quiet "$unidad"; then
    falta "Sunshine no arranca solo con la sesión"
    if ! revisando; then
      systemctl --user enable --now "$unidad"
      ok "Sunshine arranca solo con la sesión ($unidad)"
    fi
  else
    ok "Sunshine arranca solo con la sesión ($unidad)"
  fi

  # Acceso a la API de Sunshine para el panel: con él la app se empareja sola (el PIN).
  if [ -s "$DATOS/sunshine.json" ] && printf '\n\n' | "$HOST_NUEVO" -guardar-sunshine 2>/dev/null; then
    ok "el panel tiene acceso a Sunshine"
    return
  fi
  falta "el panel no tiene acceso a Sunshine"
  revisando && return

  local usuario clave
  if sunshine_tiene_usuario; then
    dice "Sunshine ya tiene usuario y contraseña (los de su página, https://localhost:47990)."
    for _ in 1 2 3; do
      read -r -p "    Usuario de Sunshine (vacío para crear uno nuevo): " usuario
      [ -n "$usuario" ] || break
      read -r -s -p "    Contraseña: " clave
      echo
      if guardar_acceso "$usuario" "$clave"; then
        ok "el panel tiene acceso a Sunshine"
        return
      fi
    done
    dice "Creo un usuario nuevo para Sunshine (reemplaza al anterior)."
  fi
  usuario=cloudgaming
  clave=$(od -An -N15 -tx1 /dev/urandom | tr -d ' \n')
  sunshine_correr --creds "$usuario" "$clave" >/dev/null 2>&1 || error "Sunshine no aceptó el usuario nuevo"
  systemctl --user restart "$unidad" # lee el usuario al arrancar
  guardar_acceso "$usuario" "$clave" || error "Sunshine no aceptó el usuario nuevo (¿arrancó bien? mirá: journalctl --user -u $unidad)"
  CLAVE_NUEVA=$clave
  ok "el panel tiene acceso a Sunshine (usuario «$usuario»)"
}

# ------------------------------------------------------------------ cortafuegos

paso_cortafuegos() {
  titulo "Cortafuegos"
  if systemctl is-active --quiet firewalld 2>/dev/null; then
    if [ "$(firewall-cmd --get-zone-of-interface=tailscale0 2>/dev/null || true)" = trusted ]; then
      ok "la red de Tailscale pasa (firewalld, zona trusted)"
      return
    fi
    falta "la red de Tailscale no está en la zona de confianza de firewalld"
    revisando && return
    como_admin firewall-cmd --permanent --zone=trusted --add-interface=tailscale0 >/dev/null
    como_admin firewall-cmd --reload >/dev/null
    ok "la red de Tailscale pasa (firewalld, zona trusted)"
  elif systemctl is-active --quiet ufw 2>/dev/null && grep -qi '^ENABLED=yes' /etc/ufw/ufw.conf 2>/dev/null; then
    if revisando; then
      dice "ufw está activo: al instalar se deja pasar la red de Tailscale"
      return
    fi
    como_admin ufw allow in on tailscale0 >/dev/null
    ok "la red de Tailscale pasa (ufw)"
  else
    ok "sin cortafuegos que cerrar el paso (ni firewalld ni ufw activos)"
  fi
}

# ------------------------------------------------------------------ 3. CloudGaming Host

# host_responde: el panel contesta en su puerto (con bash solo: curl puede no estar).
host_responde() {
  for _ in $(seq 1 20); do
    if (exec 3<>"/dev/tcp/127.0.0.1/$PUERTO" &&
      printf 'GET /api/hola HTTP/1.0\r\nHost: localhost\r\n\r\n' >&3 &&
      grep -q '"cloudgaming"' <&3) 2>/dev/null; then
      return 0
    fi
    sleep 0.5
  done
  return 1
}

paso_host() {
  titulo "3. CloudGaming Host (el panel que usa la app)"
  if systemctl --user cat cloudgaming-panel.service >/dev/null 2>&1; then
    ok "esta PC ya tiene su panel CloudGaming (cloudgaming-panel.service): lo dejo como está"
    return
  fi
  local al_dia=false
  [ -x "$DESTINO/cloudgaming-host" ] && cmp -s "$HOST_NUEVO" "$DESTINO/cloudgaming-host" && al_dia=true
  if $al_dia && systemctl --user is-enabled --quiet "$UNIDAD" 2>/dev/null &&
    systemctl --user is-active --quiet "$UNIDAD"; then
    ok "CloudGaming Host instalado y andando"
    return
  fi
  if $al_dia; then falta "CloudGaming Host no está andando"; else falta "CloudGaming Host no está instalado (o es otra versión)"; fi
  revisando && return

  mkdir -p "$DESTINO/descargas" "$UNIDADES"
  install -m 755 "$HOST_NUEVO" "$DESTINO/cloudgaming-host"
  cp "$AQUI"/descargas/*.apk "$DESTINO/descargas/"
  cat >"$UNIDADES/$UNIDAD" <<EOF
[Unit]
Description=CloudGaming Host (el panel de la PC para la app CloudGaming)
After=tailscaled.service

[Service]
ExecStart=%h/.local/share/cloudgaming-host/cloudgaming-host
Restart=on-failure
RestartSec=5
Nice=10

[Install]
WantedBy=default.target
EOF
  systemctl --user daemon-reload
  systemctl --user enable "$UNIDAD" >/dev/null 2>&1
  systemctl --user restart "$UNIDAD"
  host_responde || error "CloudGaming Host no arrancó (mirá: journalctl --user -u $UNIDAD)"
  ok "CloudGaming Host instalado y andando (puerto $PUERTO)"
}

# ------------------------------------------------------------------ todo junto

if [ "$MODO" = instalar ]; then
  [ -x "$HOST_NUEVO" ] || error "falta el programa cloudgaming-host al lado de este instalador"
  titulo "CloudGaming: preparar esta PC para jugar desde el celu"
  dice "Instalo lo que falte de Tailscale, Sunshine y CloudGaming Host."
  dice "Para instalar te puede pedir tu clave de administrador."
  if $PREGUNTAR; then
    read -r -p "    ¿Seguimos? [S/n] " sigo
    case "${sigo:-s}" in [sS]*) ;; *) exit 0 ;; esac
  fi
fi

paso_tailscale
paso_sunshine
paso_cortafuegos
paso_host

if revisando; then
  echo
  case $FALTAN in
    0) ok "Está todo listo." ;;
    1) dice "Falta 1 cosa: corré ./instalar.sh para completarla." ;;
    *) dice "Faltan $FALTAN cosas: corré ./instalar.sh para completarlas." ;;
  esac
  exit 0
fi

titulo "¡Listo! Ahora, en el celu:"
ip_casa=$(ip -4 route get 1.1.1.1 2>/dev/null | sed -n 's/.* src \([0-9.]*\).*/\1/p' | head -1)
apk=http://${ip_casa:-IP-DE-ESTA-PC}:$PUERTO/descargas/cloudgaming.apk
dice "1. Con el celu en el Wi-Fi de esta casa, abrí en el navegador:"
dice "   ${N}$apk${F}"
if command -v qrencode >/dev/null; then
  qrencode -t ANSIUTF8 -m 2 "$apk" | sed 's/^/       /'
fi
dice "   e instalá la app (Android te pide permiso para instalar desde el navegador)."
dice "   Para una tele con Android TV: cloudgaming-tv.apk en la misma dirección."
dice "2. Abrí CloudGaming y entrá con la misma cuenta de Tailscale que usaste en esta PC."
dice "   La app encuentra esta PC sola."
echo
dice "Consejos:"
dice "• Que la sesión de esta PC arranque sola al prenderla (inicio de sesión automático):"
dice "  Sunshine y el panel arrancan con la sesión."
dice "• En https://login.tailscale.com/admin/machines, en esta PC: «Disable key expiry»,"
dice "  así no hay que volver a iniciar sesión cada unos meses."
if [ -n "$CLAVE_NUEVA" ]; then
  dice "• Para entrar a la página de Sunshine (https://localhost:47990): usuario «cloudgaming»,"
  dice "  contraseña «$CLAVE_NUEVA» (también queda en $DATOS/sunshine.json)."
fi
