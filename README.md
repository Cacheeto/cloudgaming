# CloudGaming

Jugá en el celu o en la tele a los juegos de tu PC, en casa o afuera, con un menú hecho para
manejarse con el joystick.

> **Hecho con IA.** El código, los scripts, los instaladores y los textos de este proyecto,
> incluido este README, los escribió **Claude**, la IA de [Anthropic](https://www.anthropic.com),
> trabajando con [Claude Code](https://claude.com/claude-code). Mi parte fue decir qué quería,
> probarlo en mis equipos y elegir el camino en cada paso. Más detalles en
> [Cómo se hizo](#cómo-se-hizo).

Son dos piezas que se hablan por una red privada de [Tailscale](https://tailscale.com):

- **La app** (`cliente/`): una versión de [Moonlight para Android](https://github.com/moonlight-stream/moonlight-android)
  con un menú propio. Trae Tailscale adentro (no hace falta instalar la app de Tailscale), encuentra
  sola las PC de tu cuenta, se empareja sin escribir el PIN y se actualiza sola desde la PC. Hay
  una versión para el celu (64 bits) y otra para Android TV (32 bits), que inicia sesión con un
  código QR porque la tele no tiene navegador.
- **El host** (`host/`): el panel de la PC, en Go. Es lo que muestra el menú de la app: arranca
  los juegos, maneja [Sunshine](https://github.com/LizardByte/Sunshine) (el que manda la imagen y
  recibe los controles), muestra los datos que gasta el stream y sirve los APK para instalar la
  app. Anda en Linux y en Windows.

```
  Celu o tele (app CloudGaming)                     PC
 ┌──────────────────────────┐   red de Tailscale   ┌───────────────────────────────────┐
 │ menú del panel  ─────────┼──────────────────────┼─▶ CloudGaming Host (puerto 47777) │
 │ (juegos, PIN, datos)     │                      │     habla con Sunshine y Tailscale│
 │ stream (Moonlight) ◀─────┼──────────────────────┼── Sunshine: imagen y sonido       │
 │ joystick ────────────────┼──────────────────────┼─▶ Sunshine                        │
 └──────────────────────────┘                      └───────────────────────────────────┘
        ▲  por el Wi-Fi de la casa: solo la descarga de los APK (/descargas/*.apk)
```

## Instalar en una PC

Los instaladores dejan todo listo: Tailscale, Sunshine y el panel, que arranca solo.

- **Linux** (Fedora, Bazzite y otras atómicas, Ubuntu, Debian, Arch, openSUSE): ver
  [`host/linux/LEEME.txt`](host/linux/LEEME.txt).
- **Windows**: ver [`host/instalador/LEEME.md`](host/instalador/LEEME.md).

Al terminar te dicen cómo bajar la app al celu desde la misma PC.

## Compilar

Bajá el repo con el submódulo de Moonlight:

```
git clone --recursive https://github.com/Cacheeto/cloudgaming.git
```

**La app** se compila en un contenedor con Java (por ejemplo, `distrobox enter fedora`), desde `cliente/`:

```
./preparar-sdk-cloudgaming.sh   # el SDK de Android, una sola vez
./red-cloudgaming.sh            # el motor de Tailscale (libtailscale); la primera vez tarda
./build-cloudgaming.sh          # deja cloudgaming.apk y cloudgaming-tv.apk en host/web/descargas
```

Para que Android los instale, los APK salen firmados con la llave de
`~/.local/share/cloudgaming-app` (fuera del repo). Se crea una sola vez y hay que guardarla:
Android solo deja actualizar la app si la versión nueva viene firmada con la misma llave.

```
mkdir -p ~/.local/share/cloudgaming-app && cd ~/.local/share/cloudgaming-app
head -c 24 /dev/urandom | base64 > clave && chmod 600 clave
keytool -genkeypair -keystore firma.p12 -storetype PKCS12 -alias cloudgaming \
  -keyalg RSA -keysize 3072 -validity 10000 -dname "CN=CloudGaming" \
  -storepass:file clave -keypass:file clave
```

**El host**, desde `host/`:

```
./compilar.sh     # programas para Linux y Windows, en host/dist
./empaquetar.sh   # los instaladores (tar.gz para Linux, .exe para Windows), con los APK adentro
```

Necesita Go; si no está, usa el que baja `red-cloudgaming.sh`.

## El resto de las carpetas

`config/` y `scripts/` son un ejemplo de cómo se arma una PC con Bazzite para jugar a WoW con
Battle.net (por [Faugus](https://github.com/Faugus/faugus-launcher) y Proton): la PC entra sola
al prenderla y queda con la sesión bloqueada, se desbloquea durante el stream con una pantalla
virtual del tamaño del celu, y Battle.net arranca de cero para tomar ese tamaño. Cada archivo
explica dónde va.

## Seguridad

- El panel solo atiende a la propia PC y a la red de Tailscale. Desde la red de la casa solo
  deja bajar los APK.
- Los equipos de tu cuenta de Tailscale entran sin contraseña; los demás necesitan el usuario y
  la contraseña de Sunshine.
- La contraseña de Sunshine y la llave de firma de la app quedan en la PC, nunca en el repo.

## Cómo se hizo

Todo el proyecto se hizo conversando con Claude en la terminal, con
[Claude Code](https://claude.com/claude-code) y el modelo **Claude Opus 5.5**. Yo contaba qué
quería («que la app encuentre sola la PC», «que el teclado del celu escriba como el de la PC»),
lo probaba en el celu, la tele y la PC, y decidía cuando había que elegir entre opciones.
Claude escribía el código y la documentación, lo compilaba y lo probaba en la PC
(compilaciones, una máquina virtual con Windows para el instalador) y guardaba cada cambio en
git, con Claude como coautor.

Programas que se usaron:

| Para qué | Programas |
|----------|-----------|
| Escribirlo | Claude Code con Claude Opus 5.5 (Anthropic), git, [GitHub CLI](https://cli.github.com) |
| La PC | [Bazzite](https://bazzite.gg) (Fedora Atomic) con KDE Plasma, [distrobox](https://distrobox.it) (un contenedor de Fedora para compilar), [Homebrew](https://brew.sh) |
| El host | Go; el panel en HTML, CSS y JavaScript (la música de fondo se genera en el momento, con Web Audio) |
| La app | Java, el SDK y el NDK de Android, Gradle |
| Los scripts y los instaladores | Bash, Go (el instalador de Windows) y Python (el primer panel, ya retirado, y herramientas de prueba) |
| El stream y la red | [Moonlight](https://moonlight-stream.org) (la base de la app), [Sunshine](https://github.com/LizardByte/Sunshine), [Tailscale](https://tailscale.com) |
| Jugar a WoW en Linux | Battle.net con [Faugus Launcher](https://github.com/Faugus/faugus-launcher), Proton y umu |
| Probar | QEMU (Windows 11 en una máquina virtual), un celu Android y una tele con Android TV |

## Créditos y licencia

La app parte de [Moonlight para Android](https://github.com/moonlight-stream/moonlight-android),
de Cameron Gutman, Diego Waxemberg y colaboradores. No es la app oficial de Moonlight y usa su
propio identificador (`ar.cloudgaming.app`). También usa
[tailscale-android](https://github.com/tailscale/tailscale-android) (BSD-3) y
[qrcodegen](https://github.com/nayuki/QR-Code-generator) (MIT). En Windows, el instalador suma
[ViGEmBus](https://github.com/nefarius/ViGEmBus) para el mando virtual.

Todo se distribuye bajo la licencia [GPLv3](LICENSE), la misma de Moonlight.
