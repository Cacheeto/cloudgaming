# Instalador de CloudGaming para Windows

`instalar-cloudgaming.exe` deja una PC con Windows lista para jugar desde el celu con la app
CloudGaming. Es un solo archivo: trae adentro el panel de la PC (CloudGaming Host) y los APK de
la app, y baja de internet lo demás. Es el equivalente de `host/linux/instalar.sh`, con los
mismos pasos y las mismas opciones.

---

## 1. Cómo funciona

### Lo que ve quien lo instala

1. **SmartScreen**: como el programa no está firmado, Windows muestra «Windows protegió su PC».
   Hay que tocar *Más información* → *Ejecutar de todas formas*.
2. **Permiso de administrador** (el cartel de «¿Quieres permitir que esta aplicación haga
   cambios…?»): una sola vez, para todo.
3. Una **ventana de consola** que pregunta «¿Seguimos?» y después muestra cada paso con una
   tilde verde (✓, ya está) o un punto amarillo (•, faltaba y lo hace).
4. Se abre el **navegador** para iniciar sesión en Tailscale con la cuenta de la persona
   (Google, Microsoft, GitHub…). Es la misma cuenta con la que después entra en el celu.
5. Al final, la **dirección para bajar la app** desde el celu (por el Wi-Fi de la casa) y
   algunos consejos. La ventana espera un Enter antes de cerrarse.

Si algo falla, la ventana lo dice, muestra dónde quedó el registro y abre la carpeta con ese
archivo marcado (ver [Si algo falla](#si-algo-falla)).

### Los pasos

Cada paso **primero revisa** si ya está hecho y solo actúa si falta algo. Por eso se puede
volver a abrir las veces que haga falta: completa lo que quedó a medias y actualiza el panel si
el instalador trae una versión nueva.

| Paso | Qué revisa | Qué hace si falta |
|------|------------|-------------------|
| 1. Tailscale | Que esté instalado, que su servicio ande y que la PC haya iniciado sesión | Baja el MSI oficial (`pkgs.tailscale.com`), lo instala sin preguntas y corre `tailscale up --unattended`, abriendo el enlace de inicio de sesión en el navegador |
| 2. Sunshine | Que esté instalado, que su servicio arranque con Windows y que el panel tenga usuario y contraseña de Sunshine | Baja el MSI de la última versión publicada en GitHub (`SunshineStream/Sunshine`). Si Sunshine ya tenía usuario, los pide (hasta 3 intentos, la contraseña no se ve); si no, crea el usuario «cloudgaming» con una contraseña al azar |
| Mando virtual | Que exista el driver ViGEmBus | Baja ViGEmBus 1.22, comprueba su huella SHA-256 y lo instala; reinicia Sunshine para que lo vea |
| Cortafuegos | Que exista la regla «CloudGaming Host» | La crea: deja entrar al panel (TCP 47777) en todas las redes; el filtro fino lo hace el propio panel |
| 3. CloudGaming Host | Que el panel instalado sea igual al que trae el instalador, que esté la tarea programada y que el panel conteste | Cierra el panel viejo, copia el nuevo y los APK, crea la tarea programada, lo abre y espera a que conteste en `/api/hola` |

### Opciones

```
instalar-cloudgaming.exe                instalar o completar todo
instalar-cloudgaming.exe --revisar      mostrar qué está y qué falta, sin tocar nada
instalar-cloudgaming.exe --desinstalar  quitar CloudGaming Host (Tailscale, Sunshine y ViGEmBus quedan)
instalar-cloudgaming.exe --si           instalar sin preguntar antes de empezar
```

La variable de entorno `CLOUDGAMING_SIN_LOGIN=1` hace que no espere el inicio de sesión de
Tailscale: sirve solo para probar en una máquina sin nadie delante.

### Si algo falla

Todo lo que muestra la ventana queda anotado, con la hora de cada línea, en
`%TEMP%\cloudgaming-instalador\instalador.log`. También queda cada programa que corrió el
instalador (`sc`, `msiexec`, `netsh`, `schtasks`…) con su código de salida, y lo que escribió
si falló. Si falla la instalación de un MSI, el arranque del panel o el acceso a Sunshine, se
agrega lo último del registro correspondiente (el del MSI, `host.log` o `sunshine.log`), así
todo queda en un solo archivo. Cada vez que se abre el instalador se agrega al final, con un
encabezado (fecha, versión de Windows, cuenta y opciones). Las contraseñas no se anotan.

Cuando hay un error, la ventana muestra dónde está el archivo y abre la carpeta con el archivo
marcado: basta con mandarlo (por WhatsApp, por ejemplo) a quien pasó el instalador. Si todo
sale bien, la ruta aparece en la última línea.

### Cómo encaja con el resto

```
  Celu (app CloudGaming)                          PC con Windows
 ┌───────────────────────┐   red de Tailscale   ┌──────────────────────────────────────┐
 │ menú del panel  ──────┼──────────────────────┼─▶ CloudGaming Host (puerto 47777)    │
 │ (juegos, PIN, datos)  │                      │     habla con Sunshine y Tailscale   │
 │                       │                      │                                      │
 │ stream (Moonlight) ◀──┼──────────────────────┼── Sunshine: imagen y sonido          │
 │ joystick ─────────────┼──────────────────────┼─▶ Sunshine ─▶ ViGEmBus (mando Xbox)  │
 └───────────────────────┘                      └──────────────────────────────────────┘
        ▲  por el Wi-Fi de la casa: solo la descarga de los APK (/descargas/*.apk)
```

- **Tailscale** es la red privada entre el celu y la PC: funciona en la casa o afuera, sin
  abrir puertos en el router. La app reconoce las PC de la misma cuenta.
- **Sunshine** captura la pantalla, la manda comprimida al celu y recibe los controles.
- **ViGEmBus** es el driver que hace aparecer un mando de Xbox virtual en Windows: así el
  joystick del celu llega a los juegos. Sunshine para Windows no lo trae (su otra opción,
  Virtual HID, es paga).
- **CloudGaming Host** es el panel que usa el menú de la app: arranca juegos, empareja la app
  con Sunshine sola (el PIN), mide los datos que usa el stream, etc. Solo atiende a la propia
  PC y a la red de Tailscale; desde la red de casa solo deja bajar los APK.

### Qué queda instalado

| Qué | Dónde | Cómo arranca |
|-----|-------|--------------|
| Tailscale | `C:\Program Files\Tailscale` | Servicio «Tailscale», con Windows (modo *unattended*: sigue conectado aunque se cierre la sesión) |
| Sunshine | `C:\Program Files\Sunshine` (configuración y `sunshine.log` en `config\`) | Servicio «SunshineService», con Windows |
| ViGEmBus | Driver del sistema | Con Windows |
| CloudGaming Host | `C:\Program Files\CloudGaming` (programa y `descargas\*.apk`) | Tarea programada «CloudGaming Host», al iniciar sesión |
| Datos del panel | `%APPDATA%\CloudGaming` (`sunshine.json`, `host.log`, consumo, sesiones) | — |
| Regla del cortafuegos | «CloudGaming Host» | — |

La tarea programada corre el panel con la cuenta de quien instaló, **sin** permisos de
administrador, con prioridad normal (lo que abre, como Battle.net, hereda esa prioridad), sin
límite de tiempo, y si se cierra por un error Windows lo vuelve a abrir al minuto.

---

## 2. Cómo está desarrollado

### Archivos

| Archivo | Qué tiene |
|---------|-----------|
| `instalador.go` | `main`: opciones, pedido de permiso, el orden de los pasos, las instrucciones finales y lo que se ve en la consola (colores, preguntas, la contraseña oculta, la espera final) |
| `pasos.go` | Cada paso (Tailscale, Sunshine, mando, cortafuegos, host) y la desinstalación |
| `registro.go` | El registro de la instalación: lo que se muestra, los programas que corre y, si algo falla, lo último de otros registros |
| `sistema.go` | Lo de Windows: bajar archivos con avance, instalar MSI, servicios (`sc`, `net`), la IP de la casa, el navegador y las llamadas al sistema (permiso de administrador, modo de la consola) |
| `adentro/` | No está en git: `empaquetar.sh` copia ahí el host para Windows y los APK, compila y la borra |

Todos los archivos llevan `//go:build windows`: el paquete solo compila para Windows. Es parte
del módulo `cloudgaming/host`, sin dependencias externas: solo la biblioteca estándar de Go
(las funciones de Windows que faltan se llaman directo a `kernel32.dll` y `shell32.dll`).

### Cómo se compila

```
host/empaquetar.sh
```

Primero corre `compilar.sh` (el host para Linux y para Windows). Después arma el `.tar.gz` de
Linux y, para Windows, copia el host y los APK a `adentro/`. Los mete en el `.exe` con
`go:embed`, compila `dist/instalar-cloudgaming.exe` (unos 78 MB, casi todo los APK) y borra
`adentro/`. Los APK tienen que estar en `host/web/descargas/` (los deja
`cliente/build-cloudgaming.sh`).

El `.exe` es un programa de **consola** (a diferencia del host, que se compila sin ventana con
`-H=windowsgui`).

### Decisiones y por qué

- **Un solo `.exe` en Go, no un instalador MSI ni un script de PowerShell.** Se compila desde
  Linux sin herramientas de Windows, comparte lenguaje con el host y se abre con doble clic.
  Un `.ps1` choca con la política de ejecución de PowerShell, y un MSI propio necesita WiX y
  probar en Windows cada cambio.
- **Consola en vez de ventanas.** Es lo mismo que el instalador de Linux y alcanza para
  mostrar avance, preguntar y mostrar errores. Las tildes y los colores se ven bien en la
  consola de Windows 10/11 (se activan las secuencias de color con `SetConsoleMode`).
- **El permiso de administrador.** Windows ya lo pide solo al abrirlo, porque el nombre tiene
  «instalar» (Windows reconoce así a los instaladores). Si se le cambia el nombre, el programa
  se vuelve a abrir pidiéndolo (`ShellExecute` con `runas`) y le pasa `--elevado-por=<SID>`.
  Así comprueba que el permiso vino de la misma cuenta: si la cuenta no es administradora y
  otra persona pone su clave, el instalador prepararía la cuenta equivocada, y en ese caso
  avisa y se detiene. Ese control solo corre cuando el permiso lo pide el instalador: con el
  nombre original lo pide Windows antes de abrirlo y no hay con qué comparar (ver la
  recomendación 3).
- **El navegador se abre por medio del Explorador** (`explorer.exe <enlace>`): si lo abriera
  el instalador directamente, el navegador también correría como administrador.
- **El panel como tarea programada, no como servicio.** El panel tiene que correr dentro de la
  sesión de la persona: muestra avisos en pantalla, abre Battle.net y consulta a Tailscale con
  esa cuenta. Un servicio corre en otra sesión, sin escritorio. La tarea programada además
  permite volver a abrirlo si se cae.
- **La versión más nueva de Tailscale y Sunshine, pero ViGEmBus fijo.** Tailscale y Sunshine
  reciben arreglos de seguridad, así que conviene la última. ViGEmBus ya no se actualiza
  (repositorio archivado), así que se baja siempre el mismo archivo y se comprueba su huella:
  si alguien lo cambiara, el instalador no lo corre.
- **Tailscale sin su ventana de bienvenida** (`TS_NOLAUNCH=true`): esa ventana también invita
  a iniciar sesión y se superponía con la del instalador. El ícono de Tailscale aparece la
  próxima vez que se entra a Windows.
- **El acceso a Sunshine lo prueba el propio host** (`cloudgaming-host -guardar-sunshine`, el
  mismo que usa el instalador de Linux): el instalador lo escribe en una carpeta temporal, le
  pasa usuario y contraseña por la entrada (no quedan a la vista en la lista de procesos) y el
  host los guarda solo si Sunshine los acepta.

### Cómo se probó

En una máquina virtual con Windows 11 25H2 en español (QEMU con UEFI, Secure Boot y TPM),
manejada desde Linux con unos scripts que no están en el repo: crean y arrancan la VM, mandan
comandos de PowerShell por SSH, copian archivos y sacan capturas de pantalla.

Lo probado: instalación desde cero, segunda corrida sin nada que hacer (`--revisar` todo ✓),
actualización del panel, reinicio de Windows (el panel arranca solo), el recorrido con doble
clic (SmartScreen, permiso, consola, navegador), el usuario de Sunshine que ya existía (con
contraseña equivocada y correcta), la descarga del APK desde «la red de casa» (y que el resto
del panel la rechace), que Sunshine detecte ViGEmBus (`/api/virtual-input/status`) y el
registro (con un error forzado, ocupando el puerto del panel, y diciendo que no al permiso:
la carpeta se abre con el archivo marcado).

**Sin probar**: el inicio de sesión real en Tailscale y un stream de verdad (la VM no tiene
placa de video: Sunshine codifica por software). Tampoco con una cuenta de Windows que no sea
administradora, ni en Windows 10.

---

## 3. Recomendaciones para mejorarlo

Ordenadas por lo que más le cambia a quien lo usa. El esfuerzo es aproximado.

### Importantes

1. **Poder desinstalarlo desde «Aplicaciones instaladas»** (esfuerzo: medio). Hoy solo se
   quita con `--desinstalar` desde una consola, y nadie se va a acordar. Basta con registrar
   una entrada en `HKLM\Software\Microsoft\Windows\CurrentVersion\Uninstall` que apunte a un
   modo de desinstalación (por ejemplo, del propio host, para no dejar una copia de 78 MB).
2. **Firmar el `.exe`** (esfuerzo: medio; tiene costo). Evita el cartel de SmartScreen y, sobre
   todo, el **Control inteligente de aplicaciones** de Windows 11: si está activado (pasa en
   algunas instalaciones nuevas de Windows), bloquea los programas sin firma **sin opción de
   ejecutarlos igual**. Opciones: un certificado de firma de código (pago anual) o el servicio
   de firma de Microsoft (Azure Trusted Signing, pago mensual, pide verificar identidad;
   conviene confirmar si lo ofrecen a particulares desde Argentina).
   Mientras tanto, conviene avisarle a quien lo instale que puede pasar.
3. **Ícono y datos de versión en el `.exe`** (esfuerzo: poco). Con un recurso de Windows
   (`.syso`, se arma con `go-winres` o `rsrc`) el archivo muestra un ícono propio, nombre y
   versión en sus propiedades. Además el manifiesto decide cómo se pide el permiso, en vez de
   depender de que el nombre diga «instalar»: declarando que arranca sin permiso
   (`asInvoker`), siempre lo pide el instalador y siempre corre el control de «misma cuenta».
   Los antivirus también desconfían menos de los programas con datos de versión.

### Convenientes

4. **Comprobar la firma digital de lo que baja** (esfuerzo: poco). Tailscale y Sunshine se
   bajan por HTTPS de sus sitios oficiales, pero no se verifica quién firmó el MSI. Con
   `WinVerifyTrust` (o `Get-AuthenticodeSignature`) se puede exigir la firma de Tailscale Inc.
   y la de Sunshine, si la tiene.
5. **Que el inicio de sesión de Tailscale no espere para siempre** (esfuerzo: poco). Si la
   persona cierra el navegador sin entrar, la consola queda esperando sin decir nada. Un aviso
   a los pocos minutos («¿se abrió el navegador? copiá el enlace…») o un tiempo límite con la
   opción de reintentar lo haría más claro. La página de Tailscale también permite entrar
   escaneando un código QR con el celu, si ya tiene Tailscale ahí.
6. **Código QR con la dirección del APK** (esfuerzo: poco, suma una dependencia). Escanear la
   pantalla con la cámara del celu es más fácil que tipear `http://192.168.x.x:47777/…`. El de
   Linux lo hace con `qrencode` si está; en Go haría falta una biblioteca chica (por ejemplo,
   `rsc.io/qr`) y dibujarlo con caracteres de bloque.
7. **Tope de tiempo en las descargas trabadas** (esfuerzo: poco). Hoy hay tope para empezar
   a recibir (30 s), pero si una descarga se queda quieta a la mitad espera indefinidamente.
8. **Borrar lo temporal al terminar** (esfuerzo: mínimo). Se borran los instaladores bajados,
   pero queda la copia del host (8 MB) en la carpeta temporal (el registro sí queda, a
   propósito).
9. **Guardar las herramientas de prueba en el repo** (esfuerzo: poco). `windows.sh`, `win.sh`
   y `monitor.py` viven fuera del repo; en `scripts/vm/` quedarían junto al código y con historia.

### A futuro

10. **Fijar versiones conocidas de Tailscale y Sunshine** (con su huella), en vez de «la
    última». Protege de que una versión nueva cambie algo y rompa la instalación (por ejemplo,
    que Sunshine cambie el nombre del MSI o su API), a cambio de tener que actualizar el
    instalador a mano. Un punto medio: probar «la última» y, si falla, usar la fijada.
11. **El mando, a largo plazo.** ViGEmBus funciona, pero ya no recibe arreglos. Si un día
    Windows lo deja de aceptar, la salida es Virtual HID (el driver pago de Sunshine) o lo que
    Sunshine ofrezca entonces. Conviene revisar `sunshine.log` en cada actualización grande de
    Sunshine.
12. **Inicio de sesión automático.** El panel arranca al entrar a Windows, así que una PC que
    se prende y queda en la pantalla de inicio no responde a la app. El instalador solo lo
    aconseja; configurarlo solo implica guardar la contraseña de Windows (con la herramienta
    *Autologon* de Microsoft queda cifrada), y es una decisión que conviene dejarle a cada uno.
13. **Un solo instalador para Linux y Windows.** Hoy son dos programas con los mismos pasos (un
    script de bash y este en Go): cada cambio se hace dos veces. Pasar el de Linux a Go
    permitiría compartir la lógica, a cambio de perder la simpleza de un script que se lee y
    se corrige en cualquier editor.
14. **Servicios con la API de Windows** en vez de leer la salida de `sc.exe` (esfuerzo: medio,
    suma `golang.org/x/sys`). Hoy funciona porque `sc` escribe el estado en inglés en todos los
    idiomas, pero es más frágil que preguntarle al administrador de servicios directamente.
