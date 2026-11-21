package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"io/fs"
	"log"
	"net/http"
	"net/netip"
	"os"
	"path"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

var tipos = map[string]string{
	".html": "text/html; charset=utf-8", ".css": "text/css; charset=utf-8",
	".js": "text/javascript; charset=utf-8", ".json": "application/json",
	".svg": "image/svg+xml", ".png": "image/png", ".webmanifest": "application/manifest+json",
	".apk": "application/vnd.android.package-archive",
}

const csp = "default-src 'self'; img-src 'self' data:; style-src 'self'; script-src 'self'; " +
	"connect-src 'self'; frame-ancestors 'none'; base-uri 'none'; form-action 'self'"

// Panel atiende la página y la API.
type Panel struct {
	web       fs.FS
	descargas string
}

func seguridad(w http.ResponseWriter) {
	h := w.Header()
	h.Set("X-Content-Type-Options", "nosniff")
	h.Set("Referrer-Policy", "no-referrer")
	h.Set("X-Frame-Options", "DENY")
	h.Set("Content-Security-Policy", csp)
}

func responder(w http.ResponseWriter, codigo int, dato any) {
	var cuerpo bytes.Buffer
	codificador := json.NewEncoder(&cuerpo)
	codificador.SetEscapeHTML(false)
	codificador.Encode(dato)
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.Header().Set("Cache-Control", "no-store")
	seguridad(w)
	w.Header().Set("Content-Length", strconv.Itoa(cuerpo.Len()))
	w.WriteHeader(codigo)
	w.Write(cuerpo.Bytes())
}

func errorJSON(w http.ResponseWriter, codigo int, mensaje string) {
	responder(w, codigo, map[string]string{"error": mensaje})
}

// estatico sirve un archivo de la página.
func (p *Panel) estatico(w http.ResponseWriter, ruta string) {
	nombre := strings.TrimLeft(ruta, "/")
	if nombre == "" {
		nombre = "index.html"
	}
	nombre = path.Clean(nombre)
	if !fs.ValidPath(nombre) {
		errorJSON(w, 404, "no_existe")
		return
	}
	if dir, archivo := path.Split(nombre); dir == "descargas/" {
		p.apk(w, archivo)
		return
	}
	cuerpo, err := fs.ReadFile(p.web, nombre)
	if err != nil {
		errorJSON(w, 404, "no_existe")
		return
	}
	enviar(w, nombre, cuerpo)
}

func enviar(w http.ResponseWriter, nombre string, cuerpo []byte) {
	tipo, ok := tipos[path.Ext(nombre)]
	if !ok {
		tipo = "application/octet-stream"
	}
	w.Header().Set("Content-Type", tipo)
	w.Header().Set("Cache-Control", "no-cache")
	seguridad(w)
	w.Header().Set("Content-Length", strconv.Itoa(len(cuerpo)))
	w.Write(cuerpo)
}

// apk sirve uno de los APK de la app (de la carpeta de descargas, fuera de la página).
func (p *Panel) apk(w http.ResponseWriter, nombre string) {
	if path.Ext(nombre) != ".apk" || strings.ContainsAny(nombre, `/\`) || p.descargas == "" {
		errorJSON(w, 404, "no_existe")
		return
	}
	cuerpo, err := os.ReadFile(filepath.Join(p.descargas, nombre))
	if err != nil {
		errorJSON(w, 404, "no_existe")
		return
	}
	enviar(w, nombre, cuerpo)
}

func esAPK(ruta string) bool {
	dir, archivo := path.Split(path.Clean(ruta))
	return dir == "/descargas/" && path.Ext(archivo) == ".apk"
}

func ipDe(r *http.Request) (netip.Addr, bool) {
	direccion, err := netip.ParseAddrPort(r.RemoteAddr)
	if err != nil {
		return netip.Addr{}, false
	}
	return direccion.Addr().Unmap(), true
}

func (p *Panel) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	ip, ok := ipDe(r)
	ruta := r.URL.Path
	if !ok || !dentroDe(ip, redesPermitidas) {
		apk := ok && r.Method == "GET" && dentroDe(ip, redesCasa) && esAPK(ruta)
		resultado := "prohibido"
		if apk {
			resultado = "APK"
		}
		log.Printf("desde %v: %s %s → %s", ip, r.Method, ruta, resultado)
		if apk {
			p.estatico(w, ruta)
			return
		}
		w.Header().Set("Connection", "close")
		errorJSON(w, 403, "prohibido")
		return
	}
	if !strings.HasPrefix(ruta, "/api/") {
		if r.Method != "GET" {
			w.Header().Set("Connection", "close")
			errorJSON(w, 405, "metodo")
			return
		}
		p.estatico(w, ruta)
		return
	}
	if r.Method != "GET" && r.Method != "POST" {
		errorJSON(w, 405, "metodo")
		return
	}
	cuerpo := map[string]any{}
	if r.Method == "POST" {
		// Encabezado propio: otra página no lo puede mandar sin permiso (CSRF).
		if r.Header.Get("X-Panel") != "1" {
			w.Header().Set("Connection", "close")
			errorJSON(w, 403, "csrf")
			return
		}
		crudo, _ := io.ReadAll(io.LimitReader(r.Body, 65536))
		if len(bytes.TrimSpace(crudo)) > 0 && json.Unmarshal(crudo, &cuerpo) != nil {
			w.Header().Set("Connection", "close")
			errorJSON(w, 400, "json")
			return
		}
		if cuerpo == nil { // vino «null»
			errorJSON(w, 400, "json")
			return
		}
	}
	err := p.api(w, r, ip, r.Method, strings.TrimPrefix(ruta, "/api/"), cuerpo)
	var errSun *ErrorSunshine
	switch {
	case errors.As(err, &errSun):
		errorJSON(w, errSun.codigo, errSun.mensaje)
	case err != nil:
		errorJSON(w, 400, "datos_invalidos")
	}
}

// deMiCuenta: la conexión viene directo de un equipo de la cuenta dueña de la PC (ver Cuenta).
//
// Lo que llega por tailscale serve o desde la propia PC (127.0.0.1) sigue con usuario y
// contraseña: ahí la IP de origen no dice de qué equipo viene.
func deMiCuenta(r *http.Request, ip netip.Addr) bool {
	return !ip.IsLoopback() && hostPropio(r.Host) && cuenta.esMio(ip.String())
}

func token(r *http.Request) string {
	galleta, err := r.Cookie("panel")
	if err != nil {
		return ""
	}
	return galleta.Value
}

// api: devuelve un error solo para que ServeHTTP conteste (Sunshine o datos inválidos).
func (p *Panel) api(w http.ResponseWriter, r *http.Request, ip netip.Addr, metodo, ruta string, cuerpo map[string]any) error {
	pedido := metodo + " " + ruta
	if pedido == "GET hola" {
		// Para que la app encuentre esta PC entre los equipos de la red (y sepa su teclado).
		nombre, _ := os.Hostname()
		responder(w, 200, map[string]any{"cloudgaming": 1, "nombre": nombre, "host": version, "teclado": distribucionTeclado()})
		return nil
	}
	porCuenta := deMiCuenta(r, ip)
	switch pedido {
	case "GET sesion":
		responder(w, 200, map[string]bool{"logueado": porCuenta || sesiones.valida(token(r)), "por_cuenta": porCuenta})
		return nil
	case "POST login":
		return login(w, ip.String(), cuerpo)
	}
	if !porCuenta && !sesiones.valida(token(r)) {
		errorJSON(w, 401, "sesion")
		return nil
	}

	switch pedido {
	case "GET estado":
		responder(w, 200, estado())
	case "POST logout":
		sesiones.borrar(token(r))
		w.Header().Set("Set-Cookie", "panel=; Path=/; Max-Age=0; HttpOnly; SameSite=Strict")
		responder(w, 200, map[string]bool{"ok": true})
	case "POST accion":
		accion, _ := cuerpo["accion"].(string)
		return hacer(w, accion)
	case "GET clientes":
		dato, err := sunshine.pedir("GET", "/api/clients/list", nil)
		if err != nil {
			return err
		}
		responder(w, 200, map[string]any{"clientes": lista(dato, "named_certs")})
	case "POST clientes/desemparejar":
		uuid, err := texto(cuerpo["uuid"])
		if err != nil {
			return err
		}
		dato, err := sunshine.pedir("POST", "/api/clients/unpair", map[string]string{"uuid": uuid})
		if err != nil {
			return err
		}
		responder(w, 200, map[string]bool{"ok": estadoOk(dato)})
	case "GET pin":
		dato, err := sunshine.pedir("GET", "/api/pin", nil)
		if err != nil {
			return err
		}
		responder(w, 200, map[string]any{"pendientes": lista(dato, "pairings")})
	case "POST pin":
		id, err1 := texto(cuerpo["pairing_id"])
		pin, err2 := texto(cuerpo["pin"])
		if err := errors.Join(err1, err2); err != nil {
			return err
		}
		nombre, _ := texto(cuerpo["name"])
		dato, err := sunshine.pedir("POST", "/api/pin", map[string]string{"pairing_id": id, "pin": pin, "name": nombre})
		if err != nil {
			return err
		}
		responder(w, 200, map[string]bool{"ok": estadoOk(dato)})
	case "GET registros":
		responder(w, 200, map[string]any{"lineas": registrosSunshine(120)})
	case "POST datos/config":
		if err := contador.configurar(cuerpo); err != nil {
			return err
		}
		responder(w, 200, contador.resumen())
	case "POST datos/cero":
		contador.ponerEnCero()
		responder(w, 200, contador.resumen())
	default:
		errorJSON(w, 404, "no_existe")
	}
	return nil
}

func lista(dato any, clave string) any {
	if m, ok := dato.(map[string]any); ok {
		if l, ok := m[clave].([]any); ok {
			return l
		}
	}
	return []any{}
}

func estadoOk(dato any) bool {
	m, _ := dato.(map[string]any)
	return m["status"] == true
}

func login(w http.ResponseWriter, ip string, cuerpo map[string]any) error {
	if !sesiones.puedeIntentar(ip) {
		errorJSON(w, 429, "demasiados_intentos")
		return nil
	}
	usuario, _ := texto(cuerpo["usuario"])
	clave, _ := texto(cuerpo["clave"])
	cred := &credSunshine{Usuario: usuario, Clave: clave}
	if err := sunshine.probar(cred); err != nil {
		var errSun *ErrorSunshine
		if errors.As(err, &errSun) && errSun.mensaje == "credenciales_sunshine" {
			sesiones.fallo(ip)
			time.Sleep(time.Second)
		}
		return err
	}
	sunshine.guardarCred(cred)
	w.Header().Set("Set-Cookie", "panel="+sesiones.crear()+"; Path=/; Max-Age="+
		strconv.Itoa(int(duracionSesion.Seconds()))+"; HttpOnly; SameSite=Strict")
	responder(w, 200, map[string]bool{"ok": true})
	return nil
}

func estado() map[string]any {
	info := leerInfoServidor()
	ocupado := info != nil && info.ocupado
	sun := map[string]any{"activo": info != nil, "ocupado": ocupado, "credenciales": sunshine.tieneCred(),
		"app": nil, "desde": nil, "cliente": false}
	if ocupado {
		sesion := leerSesion()
		sun["app"], sun["desde"], sun["cliente"] = sesion.App, sesion.Desde, sesion.Cliente
	}
	moonlight := map[string]any{"uuid": nil, "pc": nil, "apps": nombresApps()}
	if info != nil {
		moonlight["uuid"], moonlight["pc"] = oNulo(info.uuid), oNulo(info.nombre)
	}
	return map[string]any{
		"hora":      time.Now().Format("2006-01-02T15:04:05"),
		"sunshine":  sun,
		"juego":     estadoJuego(),
		"moonlight": moonlight,
		"recientes": recientes.lista(),
		"pc":        medidor.medir(),
		"datos":     contador.resumen(),
		"ip_casa":   ipCasa(),
	}
}

func hacer(w http.ResponseWriter, accion string) error {
	switch accion {
	case "cerrar_app":
		ok, err := sunshine.cerrarApp()
		if err != nil {
			return err
		}
		responder(w, 200, map[string]bool{"ok": ok})
	case "reiniciar_sunshine":
		go reiniciarSunshine()
		responder(w, 200, map[string]bool{"ok": true})
	case "cerrar_wow":
		responder(w, 200, map[string]bool{"ok": true, "estaba": cerrarWow()})
	case "abrir_battlenet":
		abierto := !battlenetAbierto() && wowInstalado()
		if abierto {
			abierto = abrirBattlenet() == nil
		}
		responder(w, 200, map[string]bool{"ok": true, "abierto": abierto})
	case "probar_aviso":
		notificar("Panel CloudGaming", "Así se ven los avisos en la pantalla del stream.", false)
		responder(w, 200, map[string]bool{"ok": true})
	default:
		errorJSON(w, 400, "accion_desconocida")
	}
	return nil
}
