//go:build windows

package main

import (
	"bytes"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io/fs"
	"net/http"
	"os"
	"os/exec"
	"os/user"
	"path/filepath"
	"regexp"
	"strings"
	"time"
)

const (
	tailscaleMSI = "https://pkgs.tailscale.com/stable/tailscale-setup-latest-amd64.msi"
	sunshineRepo = "SunshineStream/Sunshine"
	// ViGEmBus ya no se actualiza (su repositorio está archivado), así que se baja siempre
	// el mismo instalador y se comprueba que sea ese.
	vigemURL    = "https://github.com/nefarius/ViGEmBus/releases/download/v1.22.0/ViGEmBus_1.22.0_x64_x86_arm64.exe"
	vigemSHA256 = "89220a7865076b342892f98865f3499fb7c4cfd673159e89d352c360fd014c6a"
)

// ------------------------------------------------------------------ 1. Tailscale

func programaTailscale() string {
	if ruta, err := exec.LookPath("tailscale"); err == nil {
		return ruta
	}
	return filepath.Join(carpetaProgramas(), "Tailscale", "tailscale.exe")
}

type estadoTailscale struct {
	BackendState string
	Self         struct{ UserID int64 }
	User         map[string]struct{ LoginName string }
}

func (e estadoTailscale) cuenta() string {
	return e.User[fmt.Sprint(e.Self.UserID)].LoginName
}

// leerTailscale: el estado de Tailscale; recién instalado, su servicio tarda un poco en
// contestar.
func leerTailscale(ts string, espera time.Duration) estadoTailscale {
	var estado estadoTailscale
	limite := time.Now().Add(espera)
	for {
		codigo, salida := correr(ts, "status", "--json", "--peers=false")
		if codigo == 0 && json.Unmarshal([]byte(salida), &estado) == nil && estado.BackendState != "" {
			return estado
		}
		if time.Now().After(limite) {
			return estado
		}
		time.Sleep(time.Second)
	}
}

func pasoTailscale() {
	titulo("1. Tailscale (la red privada entre el celu y la PC)")
	ts := programaTailscale()
	if !existe(ts) {
		falta("Tailscale no está instalado")
		if revisando() {
			return
		}
		dice("Lo bajo y lo instalo (el instalador oficial de Tailscale)…")
		// Sin abrir su ventana de bienvenida, que también invita a iniciar sesión: acá eso lo
		// hace el instalador. Su ícono aparece la próxima vez que se entre a Windows.
		instalarMSI(bajar(tailscaleMSI, "tailscale.msi"), "Tailscale", "TS_NOLAUNCH=true")
		ts = programaTailscale()
		if !existe(ts) {
			errorFatal("Tailscale no quedó instalado")
		}
	}
	ok("Tailscale instalado")

	if !servicioAndando("Tailscale") {
		falta("el servicio de Tailscale está apagado")
		if revisando() {
			return
		}
		arrancarServicio("Tailscale")
	}
	estado := leerTailscale(ts, 30*time.Second)
	if estado.BackendState == "" {
		errorFatal("el servicio de Tailscale no contesta")
	}
	ok("servicio de Tailscale andando")

	if estado.BackendState != "Running" {
		falta("esta PC no inició sesión en Tailscale")
		if revisando() {
			return
		}
		if sinLogin {
			dice("(prueba: no espero el inicio de sesión)")
			return
		}
		dice("Se abre el navegador: iniciá sesión con tu cuenta (Google, Microsoft, GitHub…).")
		dice("En el celu vas a entrar con " + negrita("esa misma cuenta") + ".")
		iniciarSesionTailscale(ts)
		estado = leerTailscale(ts, 10*time.Second)
		if estado.BackendState != "Running" {
			errorFatal("Tailscale no quedó conectado (estado: " + estado.BackendState + ")")
		}
	}
	if cuenta := estado.cuenta(); cuenta != "" {
		ok("conectada a Tailscale con la cuenta " + cuenta)
	} else {
		ok("conectada a Tailscale")
	}
}

var reEnlace = regexp.MustCompile(`https://\S+`)

// iniciarSesionTailscale: «tailscale up» escribe el enlace para iniciar sesión y espera a que
// se complete. «Unattended»: Tailscale sigue conectado aunque se cierre la sesión de Windows.
func iniciarSesionTailscale(ts string) {
	cmd := exec.Command(ts, "up", "--unattended")
	lector, escritor, err := os.Pipe()
	if err != nil {
		errorFatal(err.Error())
	}
	cmd.Stdout, cmd.Stderr = escritor, escritor
	if err := cmd.Start(); err != nil {
		errorFatal("no pude abrir Tailscale: " + err.Error())
	}
	escritor.Close()
	abierto := false
	lineas := leerLineas(lector)
	for linea := range lineas {
		if strings.TrimSpace(linea) == "" {
			continue
		}
		dice("  " + strings.TrimSpace(linea))
		if enlace := reEnlace.FindString(linea); enlace != "" && !abierto {
			abierto = true
			abrirEnNavegador(enlace)
			dice("  (si no se abrió solo, copiá ese enlace en el navegador)")
		}
	}
	if err := cmd.Wait(); err != nil {
		errorFatal("Tailscale no pudo iniciar sesión: " + err.Error())
	}
}

// ------------------------------------------------------------------ 2. Sunshine

var (
	dirSunshine = filepath.Join(carpetaProgramas(), "Sunshine")
	claveNueva  string
)

func pasoSunshine() {
	titulo("2. Sunshine (el que manda la imagen al celu)")
	exe := filepath.Join(dirSunshine, "sunshine.exe")
	if !existe(exe) {
		falta("Sunshine no está instalado")
		if revisando() {
			return
		}
		dice("Lo bajo y lo instalo…")
		instalarMSI(bajar(ultimoSunshine(), "sunshine.msi"), "Sunshine")
		if !existe(exe) {
			errorFatal("Sunshine no quedó instalado")
		}
	}
	ok("Sunshine instalado")

	if !servicioAndando("SunshineService") || !servicioAutomatico("SunshineService") {
		falta("Sunshine no arranca solo con Windows")
		if revisando() {
			return
		}
		arrancarServicio("SunshineService")
	}
	ok("Sunshine arranca solo con Windows (servicio SunshineService)")

	// Acceso a la API de Sunshine para el panel: con él la app se empareja sola (el PIN).
	if existe(filepath.Join(datos, "sunshine.json")) && guardarAcceso("", "", false) {
		ok("el panel tiene acceso a Sunshine")
		return
	}
	falta("el panel no tiene acceso a Sunshine")
	if revisando() {
		return
	}
	if sunshineTieneUsuario() {
		dice("Sunshine ya tiene usuario y contraseña (los de su página, https://localhost:47990).")
		for range 3 {
			usuario := preguntar("Usuario de Sunshine (vacío para crear uno nuevo): ")
			if usuario == "" {
				break
			}
			clave := preguntarClave("Contraseña: ")
			if guardarAcceso(usuario, clave, true) {
				ok("el panel tiene acceso a Sunshine")
				return
			}
		}
		dice("Creo un usuario nuevo para Sunshine (reemplaza al anterior).")
	}
	usuario, clave := "cloudgaming", claveAlAzar()
	cmd := exec.Command(exe, "--creds", usuario, clave)
	cmd.Dir = dirSunshine
	if err := cmd.Run(); err != nil {
		errorFatal("Sunshine no aceptó el usuario nuevo: " + err.Error())
	}
	reiniciarServicio("SunshineService") // lee el usuario al arrancar
	if !guardarAcceso(usuario, clave, true) {
		anotarFinal(filepath.Join(dirSunshine, "config", "sunshine.log"), 40)
		errorFatal("Sunshine no aceptó el usuario nuevo (¿arrancó bien? mirá " + filepath.Join(dirSunshine, "config", "sunshine.log") + ")")
	}
	claveNueva = clave
	ok("el panel tiene acceso a Sunshine (usuario «" + usuario + "»)")
}

// ultimoSunshine: el instalador de Windows de la última versión publicada.
func ultimoSunshine() string {
	cliente := &http.Client{Timeout: 30 * time.Second}
	respuesta, err := cliente.Get("https://api.github.com/repos/" + sunshineRepo + "/releases/latest")
	if err != nil {
		errorFatal("no pude consultar la última versión de Sunshine: " + err.Error())
	}
	defer respuesta.Body.Close()
	var version struct {
		Assets []struct {
			Name string `json:"name"`
			URL  string `json:"browser_download_url"`
		} `json:"assets"`
	}
	if respuesta.StatusCode != 200 || json.NewDecoder(respuesta.Body).Decode(&version) != nil {
		errorFatal("no pude consultar la última versión de Sunshine (GitHub contestó " + respuesta.Status + ")")
	}
	for _, archivo := range version.Assets {
		if strings.HasSuffix(archivo.Name, "-Windows-AMD64-installer.msi") {
			return archivo.URL
		}
	}
	errorFatal("la última versión de Sunshine no trae instalador para Windows: instalalo a mano (https://github.com/" + sunshineRepo + ") y volvé a abrir esto")
	return ""
}

// Sunshine guarda su usuario en sunshine_state.json (solo el usuario y la clave cifrada).
func sunshineTieneUsuario() bool {
	crudo, err := os.ReadFile(filepath.Join(dirSunshine, "config", "sunshine_state.json"))
	return err == nil && bytes.Contains(crudo, []byte(`"username"`))
}

// guardarAcceso: el host prueba el usuario y la clave contra Sunshine y, si andan, los
// guarda (sin usuario, prueba lo que ya tenía guardado).
func guardarAcceso(usuario, clave string, mostrar bool) bool {
	host := filepath.Join(tmp, "cloudgaming-host.exe")
	if err := os.WriteFile(host, hostAdentro, 0o700); err != nil {
		errorFatal("no puedo escribir " + host + ": " + err.Error())
	}
	cmd := exec.Command(host, "-guardar-sunshine")
	cmd.Stdin = strings.NewReader(usuario + "\n" + clave + "\n")
	var salida bytes.Buffer
	cmd.Stderr = &salida
	err := cmd.Run()
	if mostrar {
		for _, linea := range strings.Split(strings.TrimSpace(salida.String()), "\n") {
			if linea != "" {
				dice(linea)
			}
		}
	}
	return err == nil
}

func claveAlAzar() string {
	crudo := make([]byte, 15)
	rand.Read(crudo)
	return hex.EncodeToString(crudo)
}

// ------------------------------------------------------------------ mando virtual

// pasoMando: ViGEmBus, el mando virtual. Sunshine para Windows no lo trae, y sin él el
// joystick del celu no llega a los juegos (la otra opción de Sunshine es paga).
func pasoMando() {
	titulo("Mando virtual (ViGEmBus)")
	if servicioExiste("ViGEmBus") {
		ok("ViGEmBus instalado: el joystick llega a los juegos como uno de Xbox")
		return
	}
	falta("falta ViGEmBus: sin él, el joystick no llega a los juegos")
	if revisando() {
		return
	}
	dice("Lo bajo y lo instalo…")
	instalador := bajar(vigemURL, "ViGEmBus.exe")
	if huella := sha256De(instalador); huella != vigemSHA256 {
		errorFatal("el instalador de ViGEmBus que bajé no es el esperado (huella " + huella + ")")
	}
	codigo, _ := correr(instalador, "/qn", "/norestart")
	if codigo == 3010 || codigo == 1641 {
		reinicioPendiente = true
	} else if codigo != 0 {
		errorFatal(fmt.Sprintf("el instalador de ViGEmBus falló (código %d)", codigo))
	}
	if !servicioExiste("ViGEmBus") {
		errorFatal("ViGEmBus no quedó instalado")
	}
	if servicioAndando("SunshineService") {
		reiniciarServicio("SunshineService") // busca el mando al arrancar
	}
	ok("ViGEmBus instalado: el joystick llega a los juegos como uno de Xbox")
}

// ------------------------------------------------------------------ cortafuegos

// Sunshine pone sus propias reglas al instalarse. El panel necesita la suya: lo usan la app
// (por Tailscale) y el celu de la casa para bajar la app. El host mismo rechaza al resto.
func pasoCortafuegos() {
	titulo("Cortafuegos de Windows")
	exe := filepath.Join(destino, "cloudgaming-host.exe")
	if codigo, salida := correr("netsh", "advfirewall", "firewall", "show", "rule", "name="+regla, "verbose"); codigo == 0 &&
		strings.Contains(strings.ToLower(salida), strings.ToLower(exe)) {
		ok("el panel puede recibir conexiones (regla «" + regla + "»)")
		return
	}
	falta("el cortafuegos no deja entrar al panel")
	if revisando() {
		return
	}
	correr("netsh", "advfirewall", "firewall", "delete", "rule", "name="+regla)
	codigo, salida := correr("netsh", "advfirewall", "firewall", "add", "rule", "name="+regla,
		"dir=in", "action=allow", "protocol=TCP", fmt.Sprintf("localport=%d", puerto),
		"program="+exe, "profile=any", "enable=yes")
	if codigo != 0 {
		errorFatal("no pude agregar la regla del cortafuegos: " + strings.TrimSpace(salida))
	}
	ok("el panel puede recibir conexiones (regla «" + regla + "»)")
}

// ------------------------------------------------------------------ 3. CloudGaming Host

func pasoHost() {
	titulo("3. CloudGaming Host (el panel que usa la app)")
	exe := filepath.Join(destino, "cloudgaming-host.exe")
	alDia := hostAlDia(exe)
	if alDia && tareaExiste() && hostResponde(time.Second) {
		ok("CloudGaming Host instalado y andando")
		return
	}
	if alDia {
		falta("CloudGaming Host no está andando")
	} else {
		falta("CloudGaming Host no está instalado (o es otra versión)")
	}
	if revisando() {
		return
	}

	cerrarHost()
	if err := os.MkdirAll(filepath.Join(destino, "descargas"), 0o755); err != nil {
		errorFatal("no puedo crear " + destino + ": " + err.Error())
	}
	if err := os.WriteFile(exe, hostAdentro, 0o755); err != nil {
		errorFatal("no puedo copiar el host a " + destino + ": " + err.Error())
	}
	apks, _ := fs.Glob(apksAdentro, "adentro/descargas/*.apk")
	for _, apk := range apks {
		crudo, _ := apksAdentro.ReadFile(apk)
		if err := os.WriteFile(filepath.Join(destino, "descargas", filepath.Base(apk)), crudo, 0o644); err != nil {
			errorFatal("no puedo copiar la app a " + destino + ": " + err.Error())
		}
	}
	crearTarea(exe)
	if codigo, salida := correr("schtasks", "/Run", "/TN", tarea); codigo != 0 {
		errorFatal("no pude abrir CloudGaming Host: " + strings.TrimSpace(salida))
	}
	if !hostResponde(15 * time.Second) {
		anotarFinal(filepath.Join(datos, "host.log"), 20)
		errorFatal("CloudGaming Host no arrancó (mirá " + filepath.Join(datos, "host.log") + ")")
	}
	ok(fmt.Sprintf("CloudGaming Host instalado y andando (puerto %d)", puerto))
}

func hostAlDia(exe string) bool {
	crudo, err := os.ReadFile(exe)
	if err != nil || !bytes.Equal(crudo, hostAdentro) {
		return false
	}
	apks, _ := fs.Glob(apksAdentro, "adentro/descargas/*.apk")
	for _, apk := range apks {
		info, err := os.Stat(filepath.Join(destino, "descargas", filepath.Base(apk)))
		dentro, _ := fs.Stat(apksAdentro, apk)
		if err != nil || info.Size() != dentro.Size() {
			return false
		}
	}
	return true
}

func hostResponde(espera time.Duration) bool {
	cliente := &http.Client{Timeout: 2 * time.Second}
	limite := time.Now().Add(espera)
	for {
		if respuesta, err := cliente.Get(fmt.Sprintf("http://127.0.0.1:%d/api/hola", puerto)); err == nil {
			var cuerpo bytes.Buffer
			cuerpo.ReadFrom(respuesta.Body)
			respuesta.Body.Close()
			if strings.Contains(cuerpo.String(), `"cloudgaming"`) {
				return true
			}
		}
		if time.Now().After(limite) {
			return false
		}
		time.Sleep(500 * time.Millisecond)
	}
}

func tareaExiste() bool {
	codigo, _ := correr("schtasks", "/Query", "/TN", tarea)
	return codigo == 0
}

func cerrarHost() {
	correr("schtasks", "/End", "/TN", tarea)
	correr("taskkill", "/F", "/IM", "cloudgaming-host.exe")
	time.Sleep(time.Second) // que Windows suelte el archivo
}

// crearTarea: el host arranca al iniciar sesión, con la cuenta de quien instala y sin
// permisos de administrador; si se cae, Windows lo vuelve a abrir.
func crearTarea(exe string) {
	yo, err := user.Current()
	if err != nil {
		errorFatal("no sé con qué cuenta de Windows estás: " + err.Error())
	}
	xml := fmt.Sprintf(`<?xml version="1.0" encoding="UTF-16"?>
<Task version="1.2" xmlns="http://schemas.microsoft.com/windows/2004/02/mit/task">
  <RegistrationInfo>
    <Description>El panel de esta PC para la app CloudGaming.</Description>
  </RegistrationInfo>
  <Triggers>
    <LogonTrigger>
      <Enabled>true</Enabled>
      <UserId>%[1]s</UserId>
    </LogonTrigger>
  </Triggers>
  <Principals>
    <Principal id="Usuario">
      <UserId>%[1]s</UserId>
      <LogonType>InteractiveToken</LogonType>
      <RunLevel>LeastPrivilege</RunLevel>
    </Principal>
  </Principals>
  <Settings>
    <MultipleInstancesPolicy>IgnoreNew</MultipleInstancesPolicy>
    <DisallowStartIfOnBatteries>false</DisallowStartIfOnBatteries>
    <StopIfGoingOnBatteries>false</StopIfGoingOnBatteries>
    <ExecutionTimeLimit>PT0S</ExecutionTimeLimit>
    <RestartOnFailure>
      <Interval>PT1M</Interval>
      <Count>999</Count>
    </RestartOnFailure>
    <Priority>6</Priority>
  </Settings>
  <Actions Context="Usuario">
    <Exec>
      <Command>%[2]s</Command>
    </Exec>
  </Actions>
</Task>
`, yo.Uid, exe)
	archivo := filepath.Join(tmp, "tarea.xml")
	if err := os.WriteFile(archivo, utf16ConBOM(xml), 0o600); err != nil {
		errorFatal(err.Error())
	}
	defer os.Remove(archivo)
	if codigo, salida := correr("schtasks", "/Create", "/TN", tarea, "/XML", archivo, "/F"); codigo != 0 {
		errorFatal("no pude crear la tarea que abre el host al iniciar sesión: " + strings.TrimSpace(salida))
	}
}

// ------------------------------------------------------------------ desinstalar

func quitar() {
	ponerTitulo("CloudGaming: quitar")
	titulo("Quitar CloudGaming Host")
	cerrarHost()
	correr("schtasks", "/Delete", "/TN", tarea, "/F")
	correr("netsh", "advfirewall", "firewall", "delete", "rule", "name="+regla)
	if err := os.RemoveAll(destino); err != nil {
		errorFatal("no pude borrar " + destino + ": " + err.Error())
	}
	ok("CloudGaming Host quitado. Tailscale, Sunshine y ViGEmBus quedan instalados.")
	dice("Tus datos (consumo, sesiones, acceso a Sunshine) siguen en " + datos + ": borralos a mano si querés.")
}
