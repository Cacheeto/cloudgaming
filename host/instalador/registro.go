//go:build windows

package main

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"os/user"
	"path/filepath"
	"regexp"
	"strings"
	"time"
	"unicode/utf16"
)

// El registro: una copia de lo que muestra el instalador (sin colores y con la hora), más el
// detalle de los programas que corre. Si algo falla, quien instaló puede mandar ese archivo en
// vez de una foto de la ventana. Las contraseñas nunca se anotan.

var (
	registro      *os.File
	rutaRegistro  string
	inicioDeLinea = true
	reColor       = regexp.MustCompile(`\x1b\[[0-9;]*m`)
)

// abrirRegistro: cada vez que se abre el instalador se agrega al final del mismo archivo (con
// un encabezado), y se empieza de nuevo cuando pasa de 1 MB. Sin registro, el instalador
// anda igual.
func abrirRegistro() {
	rutaRegistro = filepath.Join(tmp, "instalador.log")
	modo := os.O_CREATE | os.O_WRONLY | os.O_APPEND
	if info, err := os.Stat(rutaRegistro); err == nil && info.Size() > 1<<20 {
		modo |= os.O_TRUNC
	}
	archivo, err := os.OpenFile(rutaRegistro, modo, 0o600)
	if err != nil {
		return
	}
	registro = archivo
	cuenta := "?"
	if yo, err := user.Current(); err == nil {
		cuenta = yo.Username
	}
	if esAdministrador() {
		cuenta += " (administrador)"
	}
	ver, _ := exec.Command("cmd", "/c", "ver").Output()
	anotar(fmt.Sprintf("\n===== %s · %s · %s · %s\n", time.Now().Format("2006-01-02"),
		strings.TrimSpace(desdeOEM(ver)), cuenta, strings.Join(os.Args, " ")))
}

// anotar escribe solo en el registro.
func anotar(texto string) {
	if registro == nil {
		return
	}
	var b strings.Builder
	for _, r := range reColor.ReplaceAllString(texto, "") {
		if inicioDeLinea && r != '\n' {
			b.WriteString(time.Now().Format("15:04:05 "))
			inicioDeLinea = false
		}
		b.WriteRune(r)
		if r == '\n' {
			inicioDeLinea = true
		}
	}
	registro.WriteString(b.String())
}

// mostrar escribe en la ventana y en el registro.
func mostrar(texto string) {
	fmt.Print(texto)
	anotar(texto)
}

// anotarComando: qué programa corrió y cómo le fue; si falló, también lo que escribió.
func anotarComando(programa string, args []string, codigo int, salida string) {
	anotar(fmt.Sprintf("      · %s %s → %d\n", filepath.Base(programa), strings.Join(args, " "), codigo))
	if codigo != 0 {
		anotarLineas(ultimasLineas(salida, 15))
	}
}

// anotarFinal agrega las últimas líneas de otro registro (el de un MSI, el del host, el de
// Sunshine), para que todo lo necesario quede en un solo archivo.
func anotarFinal(ruta string, lineas int) {
	crudo, err := os.ReadFile(ruta)
	if err != nil {
		anotar("      (no pude leer " + ruta + ": " + err.Error() + ")\n")
		return
	}
	anotar("      --- lo último de " + ruta + ":\n")
	anotarLineas(ultimasLineas(decodificar(crudo), lineas))
	anotar("      ---\n")
}

func anotarLineas(lineas []string) {
	for _, linea := range lineas {
		anotar("        " + linea + "\n")
	}
}

func ultimasLineas(texto string, n int) []string {
	lineas := strings.Split(strings.TrimSpace(strings.ReplaceAll(texto, "\r\n", "\n")), "\n")
	if len(lineas) == 1 && lineas[0] == "" {
		return nil
	}
	if len(lineas) > n {
		lineas = lineas[len(lineas)-n:]
	}
	return lineas
}

// decodificar: los registros de Windows Installer pueden venir en UTF-16 (empiezan con FF FE).
func decodificar(crudo []byte) string {
	if !bytes.HasPrefix(crudo, []byte{0xFF, 0xFE}) {
		return string(crudo)
	}
	unidades := make([]uint16, (len(crudo)-2)/2)
	for i := range unidades {
		unidades[i] = uint16(crudo[2+2*i]) | uint16(crudo[3+2*i])<<8
	}
	return string(utf16.Decode(unidades))
}

// avisarRegistro: después de un error, dice dónde quedó el registro y, si la ventana se abrió
// con doble clic, abre la carpeta con el archivo marcado (para mandarlo por WhatsApp, por
// ejemplo).
func avisarRegistro() {
	if registro == nil {
		return
	}
	fmt.Println()
	fmt.Println("    Si no sabés cómo seguir, mandale este archivo a quien te pasó el instalador:")
	fmt.Println("    " + rutaRegistro)
	if ventanaPropia() {
		exec.Command("explorer.exe", "/select,", rutaRegistro).Start()
	}
}
