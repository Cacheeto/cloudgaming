package main

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"log"
	"path/filepath"
	"sync"
	"time"
)

// Cuánto dura una sesión del panel (la de usuario y contraseña).
const duracionSesion = 30 * 24 * time.Hour

// Sesiones del panel para los que entran con el usuario y la contraseña de Sunshine.
// Se guarda solo la huella de cada sesión, no la sesión misma.
type Sesiones struct {
	mu       sync.Mutex
	d        map[string]float64 // huella → vence (segundos desde 1970, como en Python)
	intentos map[string][]time.Time
	archivo  string
}

func nuevasSesiones() *Sesiones {
	s := &Sesiones{d: map[string]float64{}, intentos: map[string][]time.Time{},
		archivo: filepath.Join(lugar.estado, "sesiones.json")}
	leerJSON(s.archivo, &s.d)
	ahora := ahoraSeg()
	for huella, vence := range s.d {
		if vence <= ahora {
			delete(s.d, huella)
		}
	}
	return s
}

func ahoraSeg() float64 {
	return float64(time.Now().UnixNano()) / 1e9
}

func huella(token string) string {
	suma := sha256.Sum256([]byte(token))
	return hex.EncodeToString(suma[:])
}

func (s *Sesiones) crear() string {
	crudo := make([]byte, 32)
	rand.Read(crudo)
	token := base64.RawURLEncoding.EncodeToString(crudo)
	s.mu.Lock()
	defer s.mu.Unlock()
	s.d[huella(token)] = ahoraSeg() + duracionSesion.Seconds()
	s.guardar()
	return token
}

func (s *Sesiones) valida(token string) bool {
	if token == "" {
		return false
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.d[huella(token)] > ahoraSeg()
}

func (s *Sesiones) borrar(token string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.d, huella(token))
	s.guardar()
}

func (s *Sesiones) guardar() {
	if err := guardarJSON(s.archivo, s.d, true); err != nil {
		log.Printf("no pude guardar las sesiones: %v", err)
	}
}

// puedeIntentar: hasta 5 intentos de entrar por minuto desde cada IP.
func (s *Sesiones) puedeIntentar(ip string) bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	recientes := []time.Time{}
	for _, t := range s.intentos[ip] {
		if time.Since(t) < time.Minute {
			recientes = append(recientes, t)
		}
	}
	s.intentos[ip] = recientes
	return len(recientes) < 5
}

func (s *Sesiones) fallo(ip string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.intentos[ip] = append(s.intentos[ip], time.Now())
}
