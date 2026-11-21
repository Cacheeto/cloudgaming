package main

import (
	"io"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"syscall"
	"time"
	"unsafe"
)

// En Windows todo lo del host va en %APPDATA%\CloudGaming.
func lugaresDelSistema() lugares {
	dir, err := os.UserConfigDir()
	if err != nil {
		dir = os.TempDir()
	}
	dir = filepath.Join(dir, "CloudGaming")
	return lugares{config: dir, estado: dir}
}

// prepararRegistro: el host corre sin ventana, así que los mensajes van a host.log
// (se empieza de nuevo cuando pasa de 1 MB).
func prepararRegistro() {
	ruta := filepath.Join(lugar.estado, "host.log")
	modo := os.O_CREATE | os.O_WRONLY | os.O_APPEND
	if info, err := os.Stat(ruta); err == nil && info.Size() > 1<<20 {
		modo |= os.O_TRUNC
	}
	archivo, err := os.OpenFile(ruta, modo, 0o600)
	if err != nil {
		return
	}
	// El archivo primero: abierto sin consola (por la tarea programada) la salida de errores
	// falla, y MultiWriter deja de escribir en el primero que falla.
	log.SetOutput(io.MultiWriter(archivo, os.Stderr))
	log.SetFlags(log.Ldate | log.Ltime)
}

const sinVentana = 0x08000000 // CREATE_NO_WINDOW

// ocultarVentana: que los programas de consola (tailscale, nvidia-smi, tasklist) no abran
// una ventana negra cada vez que el host los usa.
func ocultarVentana(cmd *exec.Cmd) {
	cmd.SysProcAttr = &syscall.SysProcAttr{HideWindow: true, CreationFlags: sinVentana}
}

func programaTailscale() string {
	if ruta, err := exec.LookPath("tailscale"); err == nil {
		return ruta
	}
	return filepath.Join(carpetaProgramas(), "Tailscale", "tailscale.exe")
}

func carpetaProgramas() string {
	if dir := os.Getenv("ProgramFiles"); dir != "" {
		return dir
	}
	return `C:\Program Files`
}

func dirSunshine() string {
	return filepath.Join(carpetaProgramas(), "Sunshine", "config")
}

// ----------------------------------------------------------------------- PC

var (
	kernel32           = syscall.NewLazyDLL("kernel32.dll")
	procGetSystemTimes = kernel32.NewProc("GetSystemTimes")
	procMemoryStatusEx = kernel32.NewProc("GlobalMemoryStatusEx")
	procGetTickCount64 = kernel32.NewProc("GetTickCount64")
)

var (
	user32                    = syscall.NewLazyDLL("user32.dll")
	procGetKeyboardLayoutName = user32.NewProc("GetKeyboardLayoutNameW")
)

// distribucionTeclado: la distribución de teclado de Windows con el nombre que le da Linux
// («latam», «es», «us»), para que la app traduzca lo que se escribe con el teclado del celu.
// Con otra: «windows-<código>», y la app manda el texto tal cual (Windows lo escribe bien con
// cualquier distribución).
func distribucionTeclado() string {
	var nombre [9]uint16 // KL_NAMELENGTH
	if r, _, _ := procGetKeyboardLayoutName.Call(uintptr(unsafe.Pointer(&nombre[0]))); r == 0 {
		return ""
	}
	codigo := strings.ToUpper(syscall.UTF16ToString(nombre[:]))
	switch valor, err := strconv.ParseUint(codigo, 16, 32); {
	case err != nil:
	case codigo == "00000409":
		return "us"
	case codigo == "0000040A" || codigo == "00000C0A": // España
		return "es"
	case valor>>16 == 0 && valor&0x3FF == 0x0A: // las demás de español: la latinoamericana
		return "latam"
	}
	return "windows-" + codigo
}

func tiemposCPU() (inactivo, total uint64, ok bool) {
	var ocioso, nucleo, usuario syscall.Filetime
	r, _, _ := procGetSystemTimes.Call(uintptr(unsafe.Pointer(&ocioso)),
		uintptr(unsafe.Pointer(&nucleo)), uintptr(unsafe.Pointer(&usuario)))
	if r == 0 {
		return 0, 0, false
	}
	valor := func(f syscall.Filetime) uint64 { return uint64(f.HighDateTime)<<32 | uint64(f.LowDateTime) }
	// El tiempo del núcleo ya incluye el ocioso.
	return valor(ocioso), valor(nucleo) + valor(usuario), true
}

type estadoMemoria struct {
	largo, carga                                   uint32
	total, disponible                              uint64
	paginacionTotal, paginacionLibre, virtualTotal uint64
	virtualLibre, virtualExtendidaLibre            uint64
}

func memoria() (total, disponible int64) {
	m := estadoMemoria{}
	m.largo = uint32(unsafe.Sizeof(m))
	if r, _, _ := procMemoryStatusEx.Call(uintptr(unsafe.Pointer(&m))); r == 0 {
		return 0, 0
	}
	return int64(m.total), int64(m.disponible)
}

func encendida() (float64, bool) {
	ms, _, _ := procGetTickCount64.Call()
	return float64(ms) / 1000, true
}

// notificar: un aviso de Windows (sale en la pantalla de la PC, que es la que se ve en el
// stream). El texto va por variables de entorno: nunca se mete adentro del script.
const scriptAviso = `
[Windows.UI.Notifications.ToastNotificationManager, Windows.UI.Notifications, ContentType = WindowsRuntime] | Out-Null
$xml = [Windows.UI.Notifications.ToastNotificationManager]::GetTemplateContent([Windows.UI.Notifications.ToastTemplateType]::ToastText02)
$t = $xml.GetElementsByTagName('text')
$t.Item(0).AppendChild($xml.CreateTextNode($env:CLOUDGAMING_TITULO)) | Out-Null
$t.Item(1).AppendChild($xml.CreateTextNode($env:CLOUDGAMING_TEXTO)) | Out-Null
$app = '{1AC14E77-02E7-4E5D-B744-2EB1AE5198B7}\WindowsPowerShell\v1.0\powershell.exe'
[Windows.UI.Notifications.ToastNotificationManager]::CreateToastNotifier($app).Show([Windows.UI.Notifications.ToastNotification]::new($xml))
`

func notificar(titulo, texto string, urgente bool) {
	cmd := exec.Command("powershell.exe", "-NoProfile", "-NonInteractive", "-Command", scriptAviso)
	cmd.Env = append(os.Environ(), "CLOUDGAMING_TITULO="+titulo, "CLOUDGAMING_TEXTO="+texto)
	ocultarVentana(cmd)
	if cmd.Start() == nil {
		go cmd.Wait()
	}
}

// reiniciarSunshine: en Windows Sunshine corre como servicio, que lo vuelve a abrir
// solo cuando se reinicia desde su API (como el botón de su página).
func reiniciarSunshine() {
	sunshine.pedir("POST", "/api/restart", nil)
}

// ---------------------------------------------------------------------- WoW

func lanzadorBattlenet() string {
	dir := os.Getenv("ProgramFiles(x86)")
	if dir == "" {
		dir = `C:\Program Files (x86)`
	}
	return filepath.Join(dir, "Battle.net", "Battle.net Launcher.exe")
}

func wowInstalado() bool {
	_, err := os.Stat(lanzadorBattlenet())
	return err == nil
}

func procesoAbierto(imagen string) bool {
	_, salida := correr(5*time.Second, "tasklist", "/FI", "IMAGENAME eq "+imagen, "/FO", "CSV", "/NH")
	return strings.Contains(strings.ToLower(salida), `"`+strings.ToLower(imagen)+`"`)
}

func wowAbierto() bool       { return procesoAbierto("Wow.exe") }
func battlenetAbierto() bool { return procesoAbierto("Battle.net.exe") }

func terminarWow(forzar bool) {
	args := []string{"/IM", "Wow.exe"}
	if forzar {
		args = append(args, "/F")
	}
	correr(10*time.Second, "taskkill", args...)
}

func abrirBattlenet() error {
	cmd := exec.Command(lanzadorBattlenet())
	if err := cmd.Start(); err != nil {
		return err
	}
	go cmd.Wait()
	return nil
}
