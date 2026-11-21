package main

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net"
	"net/netip"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

// Dónde guarda el host lo suyo (ver lugaresDelSistema en *_linux.go / *_windows.go).
type lugares struct {
	config string // credenciales de Sunshine
	estado string // sesiones del panel, consumo de datos, juegos recientes
}

var (
	// Propia PC y red de Tailscale. Tailscale descarta los paquetes con esas direcciones
	// que no lleguen por su interfaz.
	redesPermitidas = prefijos("127.0.0.0/8", "100.64.0.0/10", "::1/128", "fd7a:115c:a1e0::/48")
	// Desde la red de casa (sin Tailscale) solo se pueden bajar los APK: así se instala la
	// app la primera vez (o en la tele), o se sale de una versión que no se actualiza sola.
	redesCasa = prefijos("192.168.0.0/16", "10.0.0.0/8", "172.16.0.0/12")
)

func prefijos(textos ...string) []netip.Prefix {
	lista := make([]netip.Prefix, len(textos))
	for i, t := range textos {
		lista[i] = netip.MustParsePrefix(t)
	}
	return lista
}

func dentroDe(ip netip.Addr, redes []netip.Prefix) bool {
	for _, red := range redes {
		if red.Contains(ip) {
			return true
		}
	}
	return false
}

func leerJSON(ruta string, dato any) bool {
	crudo, err := os.ReadFile(ruta)
	return err == nil && json.Unmarshal(crudo, dato) == nil
}

// guardarJSON escribe de una vez (archivo temporal y después renombrar), para que un
// corte a mitad de camino no deje el archivo roto.
func guardarJSON(ruta string, dato any, privado bool) error {
	crudo, err := json.MarshalIndent(dato, "", " ")
	if err != nil {
		return err
	}
	permiso := os.FileMode(0o644)
	if privado {
		permiso = 0o600
	}
	if err := os.MkdirAll(filepath.Dir(ruta), 0o700); err != nil {
		return err
	}
	tmp := ruta + ".tmp"
	if err := os.WriteFile(tmp, crudo, permiso); err != nil {
		return err
	}
	return os.Rename(tmp, ruta)
}

// correr ejecuta un programa y devuelve su código de salida y lo que escribió
// (-1 si no se pudo correr o tardó más de la cuenta).
func correr(espera time.Duration, programa string, args ...string) (int, string) {
	ctx, cancelar := context.WithTimeout(context.Background(), espera)
	defer cancelar()
	cmd := exec.CommandContext(ctx, programa, args...)
	ocultarVentana(cmd)
	var salida bytes.Buffer
	cmd.Stdout = &salida
	err := cmd.Run()
	var errSalida *exec.ExitError
	switch {
	case err == nil:
		return 0, salida.String()
	case errors.As(err, &errSalida) && ctx.Err() == nil:
		return errSalida.ExitCode(), salida.String()
	default:
		return -1, ""
	}
}

// ipCasa: la IP de la PC en la red de casa (la de la salida a internet), o nil.
func ipCasa() any {
	// No manda nada: solo elige la interfaz de salida.
	conexion, err := net.Dial("udp4", "192.0.2.1:9")
	if err != nil {
		return nil
	}
	defer conexion.Close()
	ip, err := netip.ParseAddrPort(conexion.LocalAddr().String())
	if err != nil || !dentroDe(ip.Addr(), redesCasa) {
		return nil
	}
	return ip.Addr().String()
}

// hostPropio: el encabezado Host nombra a esta PC de una forma que un sitio de afuera
// no puede usar.
//
// Contra el «DNS rebinding»: una página abierta en un equipo de la red podría hacer que
// su dominio apunte a la IP de Tailscale de la PC y hablarle al panel desde ese equipo,
// que es de la cuenta. Una IP escrita tal cual, localhost o un nombre de Tailscale, no.
func hostPropio(host string) bool {
	nombre := strings.ToLower(strings.TrimSpace(host))
	if strings.HasPrefix(nombre, "[") { // IPv6 con puerto: [fd7a::1]:47777
		fin := strings.Index(nombre, "]")
		if fin < 0 {
			return false
		}
		nombre = nombre[1:fin]
	} else if strings.Count(nombre, ":") == 1 {
		nombre = strings.Split(nombre, ":")[0]
	}
	nombre = strings.TrimSuffix(nombre, ".")
	if _, err := netip.ParseAddr(nombre); err == nil {
		return true
	}
	propio, _ := os.Hostname()
	return nombre == "localhost" || nombre == strings.ToLower(propio) || strings.HasSuffix(nombre, ".ts.net")
}

// Los datos que llegan en JSON, como los convertía Python (str(), int(), float()).

var errDatos = errors.New("datos_invalidos")

func texto(v any) (string, error) {
	switch x := v.(type) {
	case string:
		return x, nil
	case float64:
		return strconv.FormatFloat(x, 'f', -1, 64), nil
	case bool:
		return strconv.FormatBool(x), nil
	}
	return "", errDatos
}

func numero(v any) (float64, error) {
	switch x := v.(type) {
	case float64:
		return x, nil
	case string:
		n, err := strconv.ParseFloat(strings.TrimSpace(x), 64)
		if err != nil {
			return 0, errDatos
		}
		return n, nil
	}
	return 0, errDatos
}

func limitar[T int | float64](v, minimo, maximo T) T {
	return max(minimo, min(v, maximo))
}
