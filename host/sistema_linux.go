package main

import (
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"syscall"
	"time"
)

// Las mismas carpetas que el panel en Python: así se conservan las sesiones, el
// consumo de datos y los juegos recientes.
func lugaresDelSistema() lugares {
	casa, _ := os.UserHomeDir()
	config := os.Getenv("XDG_CONFIG_HOME")
	if config == "" {
		config = filepath.Join(casa, ".config")
	}
	estado := os.Getenv("XDG_STATE_HOME")
	if estado == "" {
		estado = filepath.Join(casa, ".local", "state")
	}
	return lugares{config: filepath.Join(config, "cloudgaming-panel"), estado: filepath.Join(estado, "cloudgaming-panel")}
}

// prepararRegistro: los mensajes van a la salida de error (systemd los guarda con la hora).
func prepararRegistro() {
	log.SetFlags(0)
}

func ocultarVentana(*exec.Cmd) {}

func programaTailscale() string {
	return "tailscale"
}

// dirSunshine: la carpeta de configuración de Sunshine (apps.json, sunshine.log): la
// común, o la del flatpak.
func dirSunshine() string {
	casa, _ := os.UserHomeDir()
	config := os.Getenv("XDG_CONFIG_HOME")
	if config == "" {
		config = filepath.Join(casa, ".config")
	}
	candidatas := []string{
		filepath.Join(config, "sunshine"),
		filepath.Join(casa, ".var", "app", "dev.lizardbyte.app.Sunshine", "config", "sunshine"),
	}
	for _, dir := range candidatas {
		if _, err := os.Stat(filepath.Join(dir, "sunshine.conf")); err == nil {
			return dir
		}
	}
	return candidatas[0]
}

// ----------------------------------------------------------------------- PC

func tiemposCPU() (inactivo, total uint64, ok bool) {
	crudo, err := os.ReadFile("/proc/stat")
	if err != nil {
		return 0, 0, false
	}
	campos := strings.Fields(strings.SplitN(string(crudo), "\n", 2)[0])
	if len(campos) < 6 {
		return 0, 0, false
	}
	for i, campo := range campos[1:] {
		v, err := strconv.ParseUint(campo, 10, 64)
		if err != nil {
			return 0, 0, false
		}
		total += v
		if i == 3 || i == 4 { // idle e iowait
			inactivo += v
		}
	}
	return inactivo, total, true
}

func memoria() (total, disponible int64) {
	crudo, err := os.ReadFile("/proc/meminfo")
	if err != nil {
		return 0, 0
	}
	disponible = -1
	for _, linea := range strings.Split(string(crudo), "\n") {
		clave, valor, ok := strings.Cut(linea, ":")
		campos := strings.Fields(valor)
		if !ok || len(campos) == 0 {
			continue
		}
		kb, _ := strconv.ParseInt(campos[0], 10, 64)
		switch clave {
		case "MemTotal":
			total = kb * 1024
		case "MemAvailable":
			disponible = kb * 1024
		}
	}
	if disponible < 0 {
		disponible = total
	}
	return total, disponible
}

func encendida() (float64, bool) {
	crudo, err := os.ReadFile("/proc/uptime")
	if err != nil {
		return 0, false
	}
	campos := strings.Fields(string(crudo))
	if len(campos) == 0 {
		return 0, false
	}
	seg, err := strconv.ParseFloat(campos[0], 64)
	return seg, err == nil
}

// notificar: sale en la pantalla de la PC, que es la que se ve en el stream.
func notificar(titulo, texto string, urgente bool) {
	args := []string{"-a", "Panel CloudGaming", "-i", "network-cellular"}
	if urgente {
		args = append(args, "-u", "critical")
	}
	cmd := exec.Command("notify-send", append(args, titulo, texto)...)
	if cmd.Start() == nil {
		go cmd.Wait()
	}
}

// Las unidades de usuario con las que suele andar Sunshine (la de Bazzite y Homebrew, y
// la de los paquetes de las distros).
var unidadesSunshine = []string{"app-dev.lizardbyte.app.Sunshine.service", "sunshine.service"}

func reiniciarSunshine() {
	for _, unidad := range unidadesSunshine {
		if codigo, _ := correr(5*time.Second, "systemctl", "--user", "cat", unidad); codigo == 0 {
			correr(30*time.Second, "systemctl", "--user", "restart", unidad)
			return
		}
	}
	log.Printf("no encuentro la unidad de Sunshine para reiniciarlo")
}

var reFuenteGnome = regexp.MustCompile(`\('xkb', '([^']+)'\)`)

// distribucionTeclado: la distribución de teclado de la PC con su nombre de XKB («latam»,
// «es», «us»…), para que la app traduzca lo que se escribe con el teclado del celu. Primero
// la del escritorio (KDE o GNOME) y si no, la del sistema; "" si no se sabe.
func distribucionTeclado() string {
	config := os.Getenv("XDG_CONFIG_HOME")
	if config == "" {
		casa, _ := os.UserHomeDir()
		config = filepath.Join(casa, ".config")
	}
	// KDE la guarda solo si se cambió en sus ajustes: Use=true y LayoutList=latam,us
	if datos, err := os.ReadFile(filepath.Join(config, "kxkbrc")); err == nil {
		valores := map[string]string{}
		for _, linea := range strings.Split(string(datos), "\n") {
			if clave, valor, ok := strings.Cut(strings.TrimSpace(linea), "="); ok {
				valores[clave] = valor
			}
		}
		if valores["Use"] == "true" && valores["LayoutList"] != "" {
			return primeraDistribucion(valores["LayoutList"])
		}
	}
	if strings.Contains(os.Getenv("XDG_CURRENT_DESKTOP"), "GNOME") {
		// [('xkb', 'latam'), ('xkb', 'us')]
		_, salida := correr(5*time.Second, "gsettings", "get", "org.gnome.desktop.input-sources", "sources")
		if m := reFuenteGnome.FindStringSubmatch(salida); m != nil {
			return primeraDistribucion(m[1])
		}
	}
	_, salida := correr(5*time.Second, "localectl", "status")
	for _, linea := range strings.Split(salida, "\n") {
		if valor, ok := strings.CutPrefix(strings.TrimSpace(linea), "X11 Layout:"); ok {
			return primeraDistribucion(valor)
		}
	}
	return ""
}

// primeraDistribucion: «latam+deadtilde,us» → «latam».
func primeraDistribucion(lista string) string {
	primera, _, _ := strings.Cut(strings.TrimSpace(lista), ",")
	primera, _, _ = strings.Cut(primera, "+")
	return strings.TrimSpace(primera)
}

// ---------------------------------------------------------------------- WoW
// Battle.net con Faugus (flatpak) y Proton: los procesos se reconocen por su línea de
// comando, que lleva el .exe de Windows.

var (
	patronWow       = regexp.MustCompile(`Wow\.exe`)
	patronBattlenet = regexp.MustCompile(`Battle\.net\.exe`)
	cmdBattlenet    = []string{"flatpak", "run", "--command=/app/bin/faugus-launcher",
		"io.github.Faugus.faugus-launcher", "--game", "battlenet"}
)

func wowInstalado() bool {
	casa, _ := os.UserHomeDir()
	faugus := false
	for _, dir := range []string{"/var/lib/flatpak/app", filepath.Join(casa, ".local/share/flatpak/app")} {
		if _, err := os.Stat(filepath.Join(dir, "io.github.Faugus.faugus-launcher")); err == nil {
			faugus = true
		}
	}
	_, err := os.Stat(filepath.Join(casa, "Faugus", "battlenet")) // el prefijo de Battle.net
	return faugus && err == nil
}

// procesos: los que tienen el patrón en su línea de comando (como pgrep -f).
func procesos(patron *regexp.Regexp) []int {
	entradas, _ := os.ReadDir("/proc")
	propio := os.Getpid()
	pids := []int{}
	for _, entrada := range entradas {
		pid, err := strconv.Atoi(entrada.Name())
		if err != nil || pid == propio {
			continue
		}
		crudo, err := os.ReadFile(filepath.Join("/proc", entrada.Name(), "cmdline"))
		if err == nil && patron.Match(crudo) {
			pids = append(pids, pid)
		}
	}
	return pids
}

func wowAbierto() bool       { return len(procesos(patronWow)) > 0 }
func battlenetAbierto() bool { return len(procesos(patronBattlenet)) > 0 }

func terminarWow(forzar bool) {
	senal := syscall.SIGTERM
	if forzar {
		senal = syscall.SIGKILL
	}
	for _, pid := range procesos(patronWow) {
		syscall.Kill(pid, senal)
	}
}

func abrirBattlenet() error {
	// Unidad propia: si el host se reinicia, Battle.net sigue vivo.
	args := append([]string{"--user", "--collect", "--quiet",
		"--unit", "cloudgaming-battlenet-" + strconv.FormatInt(time.Now().Unix(), 10), "--"}, cmdBattlenet...)
	cmd := exec.Command("systemd-run", args...)
	if err := cmd.Start(); err != nil {
		return err
	}
	go cmd.Wait()
	return nil
}
