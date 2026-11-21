//go:build windows

// Instalador de CloudGaming Host para Windows: un solo programa que deja esta PC lista
// para jugar desde el celu con la app CloudGaming. Instala (o completa) lo que falte:
//  1. Tailscale: la red privada entre el celu y la PC, con tu cuenta.
//  2. Sunshine: el que manda la imagen al celu y recibe los controles, con ViGEmBus,
//     el mando virtual (sin él, el joystick no llega a los juegos).
//  3. CloudGaming Host: el panel que usa la app; arranca solo al entrar a Windows.
//
// Es la versión para Windows de linux/instalar.sh, con los mismos pasos y opciones. Trae
// adentro el host y los APK de la app (los pone ahí empaquetar.sh, en adentro/).
//
//	instalar-cloudgaming.exe                instalar o completar todo
//	instalar-cloudgaming.exe --revisar      mostrar qué está y qué falta, sin tocar nada
//	instalar-cloudgaming.exe --desinstalar  quitar CloudGaming Host (Tailscale y Sunshine quedan)
//	instalar-cloudgaming.exe --si           instalar sin preguntar antes de empezar
//
// Se abre con doble clic: pide permiso de administrador una vez (se vuelve a abrir como
// administrador en otra ventana) y al terminar espera un Enter antes de cerrarse.
package main

import (
	"bufio"
	"embed"
	"flag"
	"fmt"
	"os"
	"os/user"
	"path/filepath"
	"strings"
)

//go:embed adentro/cloudgaming-host.exe
var hostAdentro []byte

//go:embed adentro/descargas/*.apk
var apksAdentro embed.FS

const (
	puerto = 47777
	tarea  = "CloudGaming Host" // la tarea programada que abre el host al iniciar sesión
	regla  = "CloudGaming Host" // la regla del cortafuegos
)

var (
	revisar     = flag.Bool("revisar", false, "mostrar qué está y qué falta, sin tocar nada")
	desinstalar = flag.Bool("desinstalar", false, "quitar CloudGaming Host (Tailscale y Sunshine quedan)")
	si          = flag.Bool("si", false, "instalar sin preguntar antes de empezar")
	elevadoPor  = flag.String("elevado-por", "", "(uso interno) la cuenta que abrió el instalador antes de pedir permiso")
)

// Solo para probar el instalador en una máquina sin nadie delante: no espera el inicio de
// sesión de Tailscale.
var sinLogin = os.Getenv("CLOUDGAMING_SIN_LOGIN") != ""

var (
	destino = filepath.Join(carpetaProgramas(), "CloudGaming") // el host y los APK
	datos   string                                             // los del host (%APPDATA%\CloudGaming)
	tmp     string                                             // descargas, el registro y la copia del host para probar el acceso a Sunshine
	entrada = bufio.NewReader(os.Stdin)
	faltan  int
)

func main() {
	flag.Usage = func() {
		fmt.Fprintln(os.Stderr, "Uso: instalar-cloudgaming.exe [--revisar | --desinstalar] [--si]")
		flag.PrintDefaults()
	}
	flag.Parse()
	prepararConsola()
	tmp = filepath.Join(os.TempDir(), "cloudgaming-instalador")
	if err := os.MkdirAll(tmp, 0o700); err != nil {
		errorFatal("no puedo crear " + tmp + ": " + err.Error())
	}
	abrirRegistro()

	if !*revisar && !esAdministrador() {
		if *elevadoPor != "" {
			errorFatal("Windows no le dio permiso de administrador al instalador")
		}
		dice("Windows te va a pedir permiso de administrador: decile que sí.")
		yo, err := user.Current()
		if err != nil {
			errorFatal("no sé con qué cuenta de Windows estás: " + err.Error())
		}
		args := append(os.Args[1:], "--elevado-por="+yo.Uid)
		if err := abrirComoAdministrador(args); err != nil {
			errorFatal("sin permiso de administrador no puedo instalar nada")
		}
		os.Exit(0) // sigue en la ventana nueva
	}
	if *elevadoPor != "" {
		// Si la cuenta no es administradora, Windows pide la clave de otra que sí lo es, y
		// el instalador terminaría preparando esa otra cuenta.
		if yo, err := user.Current(); err == nil && yo.Uid != *elevadoPor {
			errorFatal(fmt.Sprintf("el permiso de administrador vino de otra cuenta (%s). Instalalo desde la cuenta de Windows con la que vas a jugar, y que esa cuenta sea administradora.", yo.Username))
		}
	}

	config, err := os.UserConfigDir()
	if err != nil {
		errorFatal("no encuentro la carpeta de datos de tu cuenta (%APPDATA%)")
	}
	datos = filepath.Join(config, "CloudGaming")

	switch {
	case *desinstalar:
		quitar()
		salir(0)
	case *revisar:
		ponerTitulo("CloudGaming: revisar esta PC")
	default:
		ponerTitulo("CloudGaming: preparar esta PC")
		titulo("CloudGaming: preparar esta PC para jugar desde el celu")
		dice("Instalo lo que falte de Tailscale, Sunshine y CloudGaming Host.")
		dice("Los programas se bajan de internet: tarda unos minutos.")
		if !*si && !preguntarSiNo("¿Seguimos?") {
			salir(0)
		}
	}

	pasoTailscale()
	pasoSunshine()
	pasoMando()
	pasoCortafuegos()
	pasoHost()
	os.RemoveAll(filepath.Join(tmp, "descargas"))

	if *revisar {
		mostrar("\n")
		switch faltan {
		case 0:
			ok("Está todo listo.")
		case 1:
			dice("Falta 1 cosa: abrí el instalador para completarla.")
		default:
			dice(fmt.Sprintf("Faltan %d cosas: abrí el instalador para completarlas.", faltan))
		}
		salir(0)
	}
	instrucciones()
	mostrar("\n")
	dice("Lo que se hizo quedó anotado en " + rutaRegistro)
	salir(0)
}

func revisando() bool { return *revisar }

func instrucciones() {
	titulo("¡Listo! Ahora, en el celu:")
	casa := ipDeCasa()
	if casa == "" {
		casa = "IP-DE-ESTA-PC"
	}
	apk := fmt.Sprintf("http://%s:%d/descargas/cloudgaming.apk", casa, puerto)
	dice("1. Con el celu en el Wi-Fi de esta casa, abrí en el navegador:")
	dice("   " + negrita(apk))
	dice("   e instalá la app (Android te pide permiso para instalar desde el navegador).")
	dice("   Para una tele con Android TV: cloudgaming-tv.apk en la misma dirección.")
	dice("2. Abrí CloudGaming y entrá con la misma cuenta de Tailscale que usaste en esta PC.")
	dice("   La app encuentra esta PC sola.")
	mostrar("\n")
	dice("Consejos:")
	if reinicioPendiente {
		dice("• Reiniciá la PC cuando puedas: algo de lo que se instaló lo pide.")
	}
	dice("• Que Windows entre solo a tu cuenta al prender la PC: el panel arranca al iniciar")
	dice("  sesión (Sunshine y Tailscale arrancan con Windows).")
	dice("• En https://login.tailscale.com/admin/machines, en esta PC: «Disable key expiry»,")
	dice("  así no hay que volver a iniciar sesión cada unos meses.")
	if claveNueva != "" {
		guardada := filepath.Join(datos, "sunshine.json")
		dice("• Para entrar a la página de Sunshine (https://localhost:47990): usuario «cloudgaming»,")
		fmt.Printf("      contraseña «%s» (también queda en %s).\n", claveNueva, guardada)
		anotar("      contraseña: no se anota en el registro (está en " + guardada + ").\n")
	}
}

// ------------------------------------------------------------------ lo que se ve

var colores bool // la consola entiende los códigos de color (se activan en prepararConsola)

func color(codigo, texto string) string {
	if !colores {
		return texto
	}
	return "\x1b[" + codigo + "m" + texto + "\x1b[0m"
}

func negrita(texto string) string { return color("1", texto) }

func titulo(texto string) { mostrar("\n" + negrita(texto) + "\n") }
func ok(texto string)     { mostrar("  " + color("32", "✓") + " " + texto + "\n") }
func dice(texto string)   { mostrar("    " + texto + "\n") }

func falta(texto string) {
	mostrar("  " + color("33", "•") + " " + texto + "\n")
	faltan++
}

func errorFatal(texto string) {
	mensaje := "\n" + color("31", "No pude seguir:") + " " + texto + "\n"
	fmt.Fprint(os.Stderr, mensaje)
	anotar(mensaje)
	avisarRegistro()
	salir(1)
}

// salir: si la ventana es del instalador (se abrió con doble clic), espera un Enter para
// que se pueda leer lo que pasó antes de que se cierre.
func salir(codigo int) {
	if ventanaPropia() {
		fmt.Print("\nApretá Enter para cerrar esta ventana.")
		entrada.ReadString('\n')
	}
	os.Exit(codigo)
}

func preguntar(texto string) string {
	mostrar("    " + texto)
	linea, _ := entrada.ReadString('\n')
	anotar(linea)
	return strings.TrimSpace(linea)
}

func preguntarSiNo(texto string) bool {
	respuesta := strings.ToLower(preguntar(texto + " [S/n] "))
	return respuesta == "" || strings.HasPrefix(respuesta, "s")
}

// preguntarClave: como preguntar, sin mostrar lo que se escribe.
func preguntarClave(texto string) string {
	mostrar("    " + texto)
	restaurar := ocultarEco()
	linea, _ := entrada.ReadString('\n')
	restaurar()
	fmt.Println()
	anotar("(no se anota)\n")
	return strings.TrimRight(linea, "\r\n")
}
