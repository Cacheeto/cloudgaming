package main

import (
	"bytes"
	"crypto/tls"
	"encoding/base64"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"
)

const (
	sunshineAPI    = "https://127.0.0.1:47990"
	sunshineOrigen = "https://localhost:47990"
	sunshineInfo   = "http://127.0.0.1:47989/serverinfo"
	// Lo primero que anota Sunshine al abrir una app (y el principio de una sesión).
	marcaApp = "Display device configuration is disabled"
)

var (
	erroresCSRF = map[string]bool{"Missing CSRF token": true, "Invalid CSRF token": true, "CSRF token expired": true}
	reHora      = regexp.MustCompile(`^\[(\d{4}-\d\d-\d\d \d\d:\d\d:\d\d)`)
	reANSI      = regexp.MustCompile(`\x1b\[[0-9;]*m`)
)

// ErrorSunshine: lo que se le contesta a la página cuando Sunshine falla.
type ErrorSunshine struct {
	codigo  int
	mensaje string
}

func (e *ErrorSunshine) Error() string { return e.mensaje }

type credSunshine struct {
	Usuario string `json:"usuario"`
	Clave   string `json:"clave"`
}

// Sunshine: su API, con el usuario y la contraseña de Sunshine guardados en el host.
type Sunshine struct {
	mu      sync.Mutex
	cred    *credSunshine
	token   string
	archivo string
	cliente *http.Client
}

func nuevoSunshine() *Sunshine {
	s := &Sunshine{
		archivo: filepath.Join(lugar.config, "sunshine.json"),
		cliente: &http.Client{
			Timeout: 8 * time.Second,
			// Certificado propio de Sunshine, y solo hacia 127.0.0.1.
			Transport: &http.Transport{TLSClientConfig: &tls.Config{InsecureSkipVerify: true}},
		},
	}
	var cred credSunshine
	if leerJSON(s.archivo, &cred) {
		s.cred = &cred
	}
	return s
}

func (s *Sunshine) tieneCred() bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.cred != nil
}

// pedir llama a la API de Sunshine. Devuelve el JSON decodificado (o el texto, si no es JSON).
func (s *Sunshine) pedir(metodo, ruta string, cuerpo any) (any, error) {
	s.mu.Lock()
	cred := s.cred
	s.mu.Unlock()
	return s.pedirCon(cred, metodo, ruta, cuerpo, true)
}

func (s *Sunshine) pedirCon(cred *credSunshine, metodo, ruta string, cuerpo any, reintento bool) (any, error) {
	if cred == nil {
		return nil, &ErrorSunshine{409, "credenciales_sunshine"}
	}
	var datos io.Reader
	if metodo != "GET" {
		crudo := []byte{}
		if cuerpo != nil {
			crudo, _ = json.Marshal(cuerpo)
		}
		datos = bytes.NewReader(crudo)
	}
	pedido, err := http.NewRequest(metodo, sunshineAPI+ruta, datos)
	if err != nil {
		return nil, err
	}
	par := base64.StdEncoding.EncodeToString([]byte(cred.Usuario + ":" + cred.Clave))
	pedido.Header.Set("Authorization", "Basic "+par)
	// Sunshine rechaza lo que no venga de su propia página (CSRF).
	pedido.Header.Set("Origin", sunshineOrigen)
	pedido.Header.Set("Referer", sunshineOrigen+"/")
	if metodo != "GET" {
		pedido.Header.Set("Content-Type", "application/json")
	}
	s.mu.Lock()
	if s.token != "" {
		pedido.Header.Set("X-CSRF-Token", s.token)
	}
	s.mu.Unlock()

	respuesta, err := s.cliente.Do(pedido)
	if err != nil {
		return nil, &ErrorSunshine{503, "sunshine_apagado"}
	}
	defer respuesta.Body.Close()
	crudo, _ := io.ReadAll(io.LimitReader(respuesta.Body, 4<<20))
	dato := decodificar(crudo)
	if respuesta.StatusCode < 400 {
		return dato, nil
	}
	mensaje := ""
	if m, ok := dato.(map[string]any); ok {
		mensaje, _ = m["error"].(string)
	}
	switch {
	case respuesta.StatusCode == 400 && erroresCSRF[mensaje] && reintento:
		token := s.pedirToken(cred)
		s.mu.Lock()
		s.token = token
		s.mu.Unlock()
		return s.pedirCon(cred, metodo, ruta, cuerpo, false)
	case respuesta.StatusCode == 401:
		return nil, &ErrorSunshine{409, "credenciales_sunshine"}
	case mensaje == "":
		mensaje = "sunshine_http_" + strconv.Itoa(respuesta.StatusCode)
	}
	return nil, &ErrorSunshine{502, mensaje}
}

func (s *Sunshine) pedirToken(cred *credSunshine) string {
	dato, err := s.pedirCon(cred, "GET", "/api/csrf-token", nil, false)
	if m, ok := dato.(map[string]any); ok && err == nil {
		for clave, valor := range m {
			if t, ok := valor.(string); ok && strings.Contains(clave, "token") {
				return t
			}
		}
	}
	return ""
}

func decodificar(crudo []byte) any {
	var dato any
	if json.Unmarshal(crudo, &dato) == nil {
		return dato
	}
	return string(crudo)
}

// probar: ¿sirven este usuario y contraseña?
func (s *Sunshine) probar(cred *credSunshine) error {
	_, err := s.pedirCon(cred, "GET", "/api/apps", nil, true)
	return err
}

func (s *Sunshine) guardarCred(cred *credSunshine) {
	s.mu.Lock()
	s.cred = cred
	s.token = ""
	s.mu.Unlock()
	if err := guardarJSON(s.archivo, cred, true); err != nil {
		log.Printf("no pude guardar el acceso a Sunshine: %v", err)
	}
}

// cerrarApp: termina la sesión (Sunshine cierra la app y corta el stream).
func (s *Sunshine) cerrarApp() (bool, error) {
	dato, err := s.pedir("POST", "/api/apps/close", nil)
	if err != nil {
		return false, err
	}
	m, _ := dato.(map[string]any)
	return m["status"] == true, nil
}

// -------------------------------------------------------- estado público

type infoServidor struct {
	ocupado bool
	appID   int
	// Moonlight identifica la PC con estos dos datos (para abrir juegos desde el celu).
	uuid   string
	nombre string
}

var (
	reEstado    = regexp.MustCompile(`<state>([^<]*)</state>`)
	reJuego     = regexp.MustCompile(`<currentgame>(\d+)</currentgame>`)
	reUUID      = regexp.MustCompile(`<uniqueid>([^<]*)</uniqueid>`)
	reNombre    = regexp.MustCompile(`<hostname>([^<]*)</hostname>`)
	clienteInfo = &http.Client{Timeout: 3 * time.Second}
)

// leerInfoServidor: el estado público de Sunshine (sin usuario), o nil si está apagado.
func leerInfoServidor() *infoServidor {
	respuesta, err := clienteInfo.Get(sunshineInfo)
	if err != nil {
		return nil
	}
	defer respuesta.Body.Close()
	crudo, err := io.ReadAll(io.LimitReader(respuesta.Body, 1<<20))
	if err != nil {
		return nil
	}
	xml := string(crudo)
	info := &infoServidor{}
	if m := reEstado.FindStringSubmatch(xml); m != nil {
		info.ocupado = strings.HasSuffix(m[1], "BUSY")
	}
	if m := reJuego.FindStringSubmatch(xml); m != nil {
		info.appID, _ = strconv.Atoi(m[1])
	}
	if m := reUUID.FindStringSubmatch(xml); m != nil {
		info.uuid = m[1]
	}
	if m := reNombre.FindStringSubmatch(xml); m != nil {
		info.nombre = m[1]
	}
	return info
}

// --------------------------------------------- apps y registro de Sunshine

type appSunshine struct {
	Nombre   string            `json:"name"`
	Cmd      string            `json:"cmd"`
	Detached []string          `json:"detached"`
	Prep     []json.RawMessage `json:"prep-cmd"`
}

// appsSunshine: las apps configuradas en Sunshine (su apps.json).
func appsSunshine() []appSunshine {
	var dato struct {
		Apps []appSunshine `json:"apps"`
	}
	leerJSON(filepath.Join(dirSunshine(), "apps.json"), &dato)
	return dato.Apps
}

func nombresApps() []string {
	nombres := []string{}
	for _, app := range appsSunshine() {
		nombres = append(nombres, app.Nombre)
	}
	return nombres
}

// colaDelRegistro: las últimas líneas del registro de Sunshine (su sunshine.log).
func colaDelRegistro() []string {
	archivo, err := os.Open(filepath.Join(dirSunshine(), "sunshine.log"))
	if err != nil {
		return nil
	}
	defer archivo.Close()
	const cola = 512 << 10
	if info, err := archivo.Stat(); err == nil && info.Size() > cola {
		archivo.Seek(-cola, io.SeekEnd)
	}
	crudo, _ := io.ReadAll(archivo)
	return strings.Split(strings.ReplaceAll(string(crudo), "\r\n", "\n"), "\n")
}

type sesionSunshine struct {
	App     any  `json:"app"`
	Desde   any  `json:"desde"`
	Cliente bool `json:"cliente"`
}

// leerSesion: qué app está abierta, desde cuándo y si hay alguien conectado. Sunshine no
// lo dice por la API: sale de su registro, a partir de la última vez que abrió una app.
func leerSesion() sesionSunshine {
	lineas := colaDelRegistro()
	inicio := -1
	for i := len(lineas) - 1; i >= 0; i-- {
		if strings.Contains(lineas[i], marcaApp) {
			inicio = i
			break
		}
	}
	if inicio < 0 {
		return sesionSunshine{}
	}
	ventana := lineas[inicio:]
	sesion := sesionSunshine{}
	if m := reHora.FindStringSubmatch(lineas[inicio]); m != nil {
		sesion.Desde = m[1]
	}
	if app := adivinarApp(ventana); app != "" {
		sesion.App = app
	}
	for _, linea := range ventana {
		if strings.Contains(linea, "CLIENT CONNECTED") {
			sesion.Cliente = true
		} else if strings.Contains(linea, "CLIENT DISCONNECTED") {
			sesion.Cliente = false
		}
	}
	return sesion
}

// adivinarApp: Sunshine anota el comando que corre, no el nombre de la app. Las que no
// tienen comando («Executing [Desktop]») se distinguen por tener o no comandos previos.
func adivinarApp(ventana []string) string {
	apps := appsSunshine()
	for _, app := range apps {
		firma := app.Cmd
		if firma == "" && len(app.Detached) > 0 {
			firma = app.Detached[0]
		}
		if len(firma) > 50 {
			firma = firma[:50]
		}
		for _, linea := range ventana {
			if firma != "" && strings.Contains(linea, firma) {
				return app.Nombre
			}
		}
	}
	conPrevios := false
	for _, linea := range ventana {
		conPrevios = conPrevios || strings.Contains(linea, "Executing Do Cmd")
	}
	for _, app := range apps {
		if app.Cmd == "" && len(app.Detached) == 0 && (len(app.Prep) > 0) == conPrevios {
			return app.Nombre
		}
	}
	return ""
}

// registrosSunshine: los últimos mensajes del registro, para verlos en el panel.
func registrosSunshine(cantidad int) []string {
	lineas := []string{}
	for _, linea := range colaDelRegistro() {
		if reHora.MatchString(linea) {
			lineas = append(lineas, reANSI.ReplaceAllString(linea, ""))
		}
	}
	return lineas[max(0, len(lineas)-cantidad):]
}
