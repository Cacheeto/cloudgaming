// CloudGaming Host: el programa de la PC. Sirve el menú (la página de web/) y una API
// chica que habla con:
//   - Sunshine, el que manda la imagen, por su API en https://127.0.0.1:47990;
//   - Tailscale, la red: para reconocer los equipos de la misma cuenta y para medir
//     cuántos datos usa el stream cuando el celu está fuera de casa;
//   - WoW y Battle.net, si están instalados.
//
// Solo atiende conexiones desde la propia PC y desde la red de Tailscale (y, desde la
// red de casa, la descarga de la app). Anda en Linux y en Windows: lo que cambia entre
// los dos está en los archivos *_linux.go y *_windows.go.
//
// Reemplaza al panel en Python (el viejo servidor.py), con los mismos archivos de datos.
package main

import (
	"bufio"
	"embed"
	"errors"
	"flag"
	"fmt"
	"io/fs"
	"log"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"
	"time"
)

const version = "1.1"

//go:embed web/*.*
var webAdentro embed.FS

var (
	puerto       = flag.Int("puerto", 47777, "puerto del panel")
	dirWeb       = flag.String("web", "", "servir la página desde esta carpeta (para desarrollo) en vez de la que trae adentro")
	dirDescargas = flag.String("descargas", "", "carpeta con los APK de la app (por defecto, «descargas» dentro de -web o junto al programa)")
	verVersion   = flag.Bool("version", false, "mostrar la versión y salir")
	guardarSun   = flag.Bool("guardar-sunshine", false, "leer usuario y contraseña de Sunshine (una línea cada uno), probarlos y guardarlos; para el instalador")
)

// Lo que el resto del programa comparte, como en el panel en Python.
var (
	lugar     lugares
	sunshine  *Sunshine
	contador  *Contador
	sesiones  *Sesiones
	cuenta    *Cuenta
	recientes *Recientes
	medidor   *MedidorPC
)

func main() {
	flag.Parse()
	if *verVersion {
		fmt.Println("CloudGaming Host", version)
		return
	}
	lugar = lugaresDelSistema()
	for _, dir := range []string{lugar.config, lugar.estado} {
		if err := os.MkdirAll(dir, 0o700); err != nil {
			log.Fatalf("no puedo crear %s: %v", dir, err)
		}
	}
	if *guardarSun {
		os.Exit(guardarAccesoSunshine())
	}
	prepararRegistro()

	web, descargas := carpetasWeb()
	sunshine = nuevoSunshine()
	contador = nuevoContador()
	sesiones = nuevasSesiones()
	cuenta = &Cuenta{cache: map[string]respuestaCuenta{}}
	recientes = nuevosRecientes()
	medidor = &MedidorPC{}

	señales := make(chan os.Signal, 1)
	signal.Notify(señales, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-señales
		contador.guardar()
		os.Exit(0)
	}()
	go contador.bucle()
	go recientes.bucle()

	servidor := &http.Server{
		Addr:              fmt.Sprintf(":%d", *puerto),
		Handler:           &Panel{web: web, descargas: descargas},
		ReadHeaderTimeout: 10 * time.Second,
	}
	log.Printf("CloudGaming Host %s en el puerto %d", version, *puerto)
	err := servidor.ListenAndServe()
	contador.guardar()
	log.Fatal(err)
}

// guardarAccesoSunshine: el instalador pasa por la entrada el usuario y la contraseña de
// Sunshine (que no queden a la vista en la lista de procesos). Si Sunshine los acepta,
// quedan guardados para el panel; sin usuario, se prueba el acceso que ya estaba guardado.
// Sunshine puede estar arrancando: se le da medio minuto.
// Sale con 0 si el acceso anda, 2 si Sunshine lo rechaza, 3 si no responde y 4 si no hay
// nada guardado.
func guardarAccesoSunshine() int {
	lector := bufio.NewReader(os.Stdin)
	usuario, _ := lector.ReadString('\n')
	clave, _ := lector.ReadString('\n')
	cred := &credSunshine{Usuario: strings.TrimRight(usuario, "\r\n"), Clave: strings.TrimRight(clave, "\r\n")}
	sunshine = nuevoSunshine()
	nuevo := cred.Usuario != ""
	if !nuevo {
		if sunshine.cred == nil {
			return 4
		}
		cred = sunshine.cred
	}
	for intento := 0; ; intento++ {
		err := sunshine.probar(cred)
		var errSun *ErrorSunshine
		switch {
		case err == nil:
			if nuevo {
				sunshine.guardarCred(cred)
				fmt.Fprintln(os.Stderr, "Acceso a Sunshine guardado.")
			}
			return 0
		case errors.As(err, &errSun) && errSun.mensaje == "credenciales_sunshine":
			fmt.Fprintln(os.Stderr, "Sunshine rechazó el usuario o la contraseña.")
			return 2
		case intento >= 15:
			fmt.Fprintf(os.Stderr, "Sunshine no responde: %v\n", err)
			return 3
		}
		time.Sleep(2 * time.Second)
	}
}

// carpetasWeb: de dónde sale la página y dónde están los APK de la app.
func carpetasWeb() (fs.FS, string) {
	var web fs.FS
	descargas := *dirDescargas
	if *dirWeb != "" {
		web = os.DirFS(*dirWeb)
		if descargas == "" {
			descargas = filepath.Join(*dirWeb, "descargas")
		}
	} else {
		web, _ = fs.Sub(webAdentro, "web")
		if descargas == "" {
			if programa, err := os.Executable(); err == nil {
				descargas = filepath.Join(filepath.Dir(programa), "descargas")
			}
		}
	}
	return web, descargas
}
