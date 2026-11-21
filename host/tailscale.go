package main

import (
	"encoding/json"
	"sync"
	"time"
)

// Lo que se usa del «tailscale status --json».
type estadoRed struct {
	Self *parRed
	Peer map[string]*parRed
}

type parRed struct {
	ID           string
	PublicKey    string
	HostName     string
	OS           string
	UserID       int64
	TailscaleIPs []string
	Online       bool
	Active       bool
	CurAddr      string
	Relay        string
	RxBytes      int64
	TxBytes      int64
}

// leerEstadoRed: el estado de la red según el Tailscale de la PC, o nil si no anda.
func leerEstadoRed(conPares bool) *estadoRed {
	args := []string{"status", "--json"}
	if !conPares {
		args = append(args, "--peers=false")
	}
	codigo, salida := correr(6*time.Second, programaTailscale(), args...)
	var estado estadoRed
	if codigo != 0 || json.Unmarshal([]byte(salida), &estado) != nil {
		return nil
	}
	return &estado
}

// Cuenta reconoce los equipos de la misma cuenta de Tailscale que la PC.
//
// Esos entran al panel sin usuario ni contraseña: ya iniciaron sesión en la red con la
// cuenta del dueño, y Tailscale asegura de qué equipo viene cada conexión. Así la app
// queda lista con un solo inicio de sesión (el de la red).
type Cuenta struct {
	mu      sync.Mutex
	cache   map[string]respuestaCuenta
	usuario int64
}

type respuestaCuenta struct {
	esMio bool
	vence time.Time
}

// Cuánto se recuerda la respuesta para cada IP.
const vigenciaCuenta = time.Minute

func (c *Cuenta) esMio(ip string) bool {
	c.mu.Lock()
	dato, ok := c.cache[ip]
	c.mu.Unlock()
	if ok && time.Now().Before(dato.vence) {
		return dato.esMio
	}
	esMio := c.consultar(ip)
	c.mu.Lock()
	c.cache[ip] = respuestaCuenta{esMio, time.Now().Add(vigenciaCuenta)}
	c.mu.Unlock()
	return esMio
}

func (c *Cuenta) miUsuario() int64 {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.usuario == 0 {
		if estado := leerEstadoRed(false); estado != nil && estado.Self != nil {
			c.usuario = estado.Self.UserID
		}
	}
	return c.usuario
}

func (c *Cuenta) consultar(ip string) bool {
	usuario := c.miUsuario()
	if usuario == 0 {
		return false
	}
	codigo, salida := correr(6*time.Second, programaTailscale(), "whois", "--json", ip)
	var quien struct {
		Node struct {
			User int64
			Tags []string
		}
	}
	if codigo != 0 || json.Unmarshal([]byte(salida), &quien) != nil {
		return false
	}
	// Un equipo con etiquetas (tags) no es de una persona, aunque lo haya sumado el dueño.
	return len(quien.Node.Tags) == 0 && quien.Node.User == usuario
}
