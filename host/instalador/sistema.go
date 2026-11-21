//go:build windows

package main

import (
	"bufio"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/netip"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"syscall"
	"time"
	"unicode/utf16"
	"unicode/utf8"
	"unsafe"
)

var (
	kernel32               = syscall.NewLazyDLL("kernel32.dll")
	procSetConsoleMode     = kernel32.NewProc("SetConsoleMode")
	procSetConsoleTitle    = kernel32.NewProc("SetConsoleTitleW")
	procGetConsoleProcList = kernel32.NewProc("GetConsoleProcessList")
	procMultiByteToWide    = kernel32.NewProc("MultiByteToWideChar")
	shell32                = syscall.NewLazyDLL("shell32.dll")
	procShellExecute       = shell32.NewProc("ShellExecuteW")
)

// Lo instala MSI con un reinicio pendiente (no hace falta para seguir, pero conviene).
var reinicioPendiente bool

func carpetaProgramas() string {
	if dir := os.Getenv("ProgramFiles"); dir != "" {
		return dir
	}
	return `C:\Program Files`
}

func existe(ruta string) bool {
	_, err := os.Stat(ruta)
	return err == nil
}

// correr ejecuta un programa y devuelve su código de salida y lo que escribió
// (-1 si no se pudo correr). Queda anotado en el registro.
func correr(programa string, args ...string) (int, string) {
	codigo, texto := 0, ""
	salida, err := exec.Command(programa, args...).CombinedOutput()
	switch fallo, ok := err.(*exec.ExitError); {
	case err == nil:
		texto = desdeOEM(salida)
	case ok:
		codigo, texto = fallo.ExitCode(), desdeOEM(salida)
	default:
		codigo, texto = -1, err.Error()
	}
	anotarComando(programa, args, codigo, texto)
	return codigo, texto
}

// desdeOEM: los programas de consola de Windows (sc, netsh, cmd) escriben en la página de
// códigos de la consola (en español, la 850), no en UTF-8.
func desdeOEM(crudo []byte) string {
	if utf8.Valid(crudo) {
		return string(crudo)
	}
	const paginaOEM = 1 // CP_OEMCP
	origen, largo := uintptr(unsafe.Pointer(&crudo[0])), uintptr(len(crudo))
	n, _, _ := procMultiByteToWide.Call(paginaOEM, 0, origen, largo, 0, 0)
	if n == 0 {
		return string(crudo)
	}
	unidades := make([]uint16, n)
	procMultiByteToWide.Call(paginaOEM, 0, origen, largo, uintptr(unsafe.Pointer(&unidades[0])), n)
	return string(utf16.Decode(unidades))
}

func leerLineas(r io.Reader) <-chan string {
	lineas := make(chan string)
	go func() {
		escaner := bufio.NewScanner(r)
		for escaner.Scan() {
			lineas <- escaner.Text()
		}
		close(lineas)
	}()
	return lineas
}

// ------------------------------------------------------------------ bajar e instalar

// bajar deja el archivo en la carpeta temporal del instalador, mostrando cuánto falta.
func bajar(url, nombre string) string {
	dir := filepath.Join(tmp, "descargas")
	os.MkdirAll(dir, 0o700)
	ruta := filepath.Join(dir, nombre)
	cliente := &http.Client{Transport: &http.Transport{
		Proxy:                 http.ProxyFromEnvironment,
		ResponseHeaderTimeout: 30 * time.Second,
	}}
	respuesta, err := cliente.Get(url)
	if err != nil {
		errorFatal("no pude bajar " + nombre + " (¿hay internet?): " + err.Error())
	}
	defer respuesta.Body.Close()
	if respuesta.StatusCode != 200 {
		errorFatal("no pude bajar " + nombre + ": el servidor contestó " + respuesta.Status)
	}
	archivo, err := os.Create(ruta)
	if err != nil {
		errorFatal("no puedo escribir " + ruta + ": " + err.Error())
	}
	defer archivo.Close()
	avance := &avance{nombre: nombre, total: respuesta.ContentLength}
	if _, err := io.Copy(archivo, io.TeeReader(respuesta.Body, avance)); err != nil {
		fmt.Println()
		errorFatal("se cortó la descarga de " + nombre + ": " + err.Error())
	}
	avance.terminar()
	return ruta
}

type avance struct {
	nombre       string
	total, hecho int64
	ultimo       time.Time
}

func (a *avance) Write(p []byte) (int, error) {
	a.hecho += int64(len(p))
	if time.Since(a.ultimo) > 300*time.Millisecond {
		a.ultimo = time.Now()
		a.mostrar()
	}
	return len(p), nil
}

func (a *avance) mostrar() {
	if a.total > 0 {
		fmt.Printf("\r    bajando %s… %d %% de %.0f MB ", a.nombre, a.hecho*100/a.total, float64(a.total)/1e6)
	} else {
		fmt.Printf("\r    bajando %s… %.0f MB ", a.nombre, float64(a.hecho)/1e6)
	}
}

func (a *avance) terminar() {
	a.mostrar()
	fmt.Println()
	anotar(fmt.Sprintf("    bajado %s (%.1f MB)\n", a.nombre, float64(a.hecho)/1e6))
}

func sha256De(ruta string) string {
	archivo, err := os.Open(ruta)
	if err != nil {
		return ""
	}
	defer archivo.Close()
	suma := sha256.New()
	io.Copy(suma, archivo)
	return hex.EncodeToString(suma.Sum(nil))
}

// instalarMSI instala sin preguntas y sin reiniciar, con las propiedades del MSI que se le
// pasen. Si falla, el detalle queda en un registro al lado.
func instalarMSI(ruta, nombre string, propiedades ...string) {
	dice("Instalo " + nombre + "…")
	detalle := filepath.Join(tmp, strings.ToLower(nombre)+"-msi.log")
	args := append([]string{"/i", ruta, "/qn", "/norestart", "/l*", detalle}, propiedades...)
	codigo, _ := correr("msiexec", args...)
	switch codigo {
	case 0:
	case 3010, 1641:
		reinicioPendiente = true
	default:
		anotarFinal(detalle, 40)
		errorFatal(fmt.Sprintf("el instalador de %s falló (código %d; el detalle está en %s)", nombre, codigo, detalle))
	}
}

// ------------------------------------------------------------------ servicios

// sc.exe escribe el nombre del estado en inglés aunque Windows esté en otro idioma.
func servicioExiste(nombre string) bool {
	codigo, _ := correr("sc", "query", nombre)
	return codigo == 0
}

func servicioAndando(nombre string) bool {
	codigo, salida := correr("sc", "query", nombre)
	return codigo == 0 && strings.Contains(salida, "RUNNING")
}

func servicioAutomatico(nombre string) bool {
	codigo, salida := correr("sc", "qc", nombre)
	return codigo == 0 && strings.Contains(salida, "AUTO_START")
}

// arrancarServicio: que arranque con Windows, y ahora. «net start» espera a que arranque.
func arrancarServicio(nombre string) {
	correr("sc", "config", nombre, "start=", "auto")
	if !servicioAndando(nombre) {
		if codigo, salida := correr("net", "start", nombre); codigo != 0 && !servicioAndando(nombre) {
			errorFatal("no pude arrancar el servicio " + nombre + ": " + strings.TrimSpace(salida))
		}
	}
}

func reiniciarServicio(nombre string) {
	correr("net", "stop", nombre)
	arrancarServicio(nombre)
}

// ------------------------------------------------------------------ la red

// ipDeCasa: la dirección de esta PC en la red de casa (la de la salida a internet).
func ipDeCasa() string {
	casa := []netip.Prefix{
		netip.MustParsePrefix("192.168.0.0/16"),
		netip.MustParsePrefix("10.0.0.0/8"),
		netip.MustParsePrefix("172.16.0.0/12"),
	}
	privada := func(ip netip.Addr) bool {
		for _, red := range casa {
			if red.Contains(ip) {
				return true
			}
		}
		return false
	}
	if conexion, err := net.Dial("udp4", "1.1.1.1:80"); err == nil {
		local := conexion.LocalAddr().(*net.UDPAddr).AddrPort().Addr().Unmap()
		conexion.Close()
		if privada(local) {
			return local.String()
		}
	}
	direcciones, _ := net.InterfaceAddrs()
	for _, d := range direcciones {
		if red, ok := d.(*net.IPNet); ok {
			if ip, ok := netip.AddrFromSlice(red.IP); ok && privada(ip.Unmap()) {
				return ip.Unmap().String()
			}
		}
	}
	return ""
}

// abrirEnNavegador: por medio del Explorador, así el navegador no se abre como
// administrador (el instalador sí lo es).
func abrirEnNavegador(enlace string) {
	exec.Command("explorer.exe", enlace).Start()
}

// ------------------------------------------------------------------ Windows

const tokenElevation = 20

func esAdministrador() bool {
	token, err := syscall.OpenCurrentProcessToken()
	if err != nil {
		return false
	}
	defer token.Close()
	var elevado, largo uint32
	err = syscall.GetTokenInformation(token, tokenElevation, (*byte)(unsafe.Pointer(&elevado)), 4, &largo)
	return err == nil && elevado != 0
}

// abrirComoAdministrador vuelve a abrir el instalador pidiendo permiso (la ventana de
// «¿Quieres permitir que esta aplicación haga cambios…?»).
func abrirComoAdministrador(args []string) error {
	programa, err := os.Executable()
	if err != nil {
		return err
	}
	partes := make([]string, len(args))
	for i, arg := range args {
		partes[i] = syscall.EscapeArg(arg)
	}
	verbo, _ := syscall.UTF16PtrFromString("runas")
	ruta, _ := syscall.UTF16PtrFromString(programa)
	parametros, _ := syscall.UTF16PtrFromString(strings.Join(partes, " "))
	dir, _ := syscall.UTF16PtrFromString(filepath.Dir(programa))
	const mostrarNormal = 1
	r, _, _ := procShellExecute.Call(0, uintptr(unsafe.Pointer(verbo)), uintptr(unsafe.Pointer(ruta)),
		uintptr(unsafe.Pointer(parametros)), uintptr(unsafe.Pointer(dir)), mostrarNormal)
	if r <= 32 { // así avisa ShellExecute que falló (por ejemplo, si le dijeron que no)
		return fmt.Errorf("ShellExecute: %d", r)
	}
	return nil
}

const (
	ecoEntrada        = 0x0004 // ENABLE_ECHO_INPUT
	secuenciasVirtual = 0x0004 // ENABLE_VIRTUAL_TERMINAL_PROCESSING
)

// prepararConsola activa los colores, si la consola los entiende.
func prepararConsola() {
	var modo uint32
	if syscall.GetConsoleMode(syscall.Stdout, &modo) != nil {
		return
	}
	r, _, _ := procSetConsoleMode.Call(uintptr(syscall.Stdout), uintptr(modo|secuenciasVirtual))
	colores = r != 0
}

func ponerTitulo(texto string) {
	puntero, _ := syscall.UTF16PtrFromString(texto)
	procSetConsoleTitle.Call(uintptr(unsafe.Pointer(puntero)))
}

// ventanaPropia: la consola es solo del instalador (se abrió con doble clic, no desde una
// consola que ya estaba).
func ventanaPropia() bool {
	var procesos [4]uint32
	n, _, _ := procGetConsoleProcList.Call(uintptr(unsafe.Pointer(&procesos[0])), uintptr(len(procesos)))
	return n == 1
}

// ocultarEco: lo que se escriba no aparece en la pantalla, hasta llamar a la función que
// devuelve.
func ocultarEco() func() {
	var modo uint32
	if syscall.GetConsoleMode(syscall.Stdin, &modo) != nil {
		return func() {}
	}
	procSetConsoleMode.Call(uintptr(syscall.Stdin), uintptr(modo&^ecoEntrada))
	return func() { procSetConsoleMode.Call(uintptr(syscall.Stdin), uintptr(modo)) }
}

// utf16ConBOM: así espera schtasks el XML de la tarea.
func utf16ConBOM(texto string) []byte {
	unidades := utf16.Encode([]rune(texto))
	crudo := make([]byte, 2+2*len(unidades))
	crudo[0], crudo[1] = 0xFF, 0xFE
	for i, u := range unidades {
		crudo[2+2*i] = byte(u)
		crudo[3+2*i] = byte(u >> 8)
	}
	return crudo
}
