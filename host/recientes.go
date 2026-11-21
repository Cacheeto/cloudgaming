package main

import (
	"log"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"time"
)

// Cada cuánto se mira qué app está abierta en Sunshine.
const cicloRecientes = 10 * time.Second

// Recientes recuerda cuándo se jugó por última vez cada app de Sunshine (para «Jugar»).
//
// Mira cada tanto qué app está abierta. Así cuenta también las sesiones que no se
// abrieron desde el panel.
type Recientes struct {
	mu       sync.Mutex
	d        map[string]string // app → cuándo (ISO, hora local)
	guardado time.Time
	archivo  string
}

type reciente struct {
	App    string `json:"app"`
	Cuando string `json:"cuando"`
}

func nuevosRecientes() *Recientes {
	r := &Recientes{d: map[string]string{}, archivo: filepath.Join(lugar.estado, "recientes.json")}
	leerJSON(r.archivo, &r.d)
	return r
}

func (r *Recientes) anotar(app string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	_, vieja := r.d[app]
	r.d[app] = time.Now().Format("2006-01-02T15:04:05")
	// Mientras se juega se actualiza la hora, pero al disco va de a ratos.
	if !vieja || time.Since(r.guardado) > time.Minute {
		r.guardar()
	}
}

func (r *Recientes) guardar() {
	if err := guardarJSON(r.archivo, r.d, false); err != nil {
		log.Printf("no pude guardar los recientes: %v", err)
	}
	r.guardado = time.Now()
}

func (r *Recientes) lista() []reciente {
	r.mu.Lock()
	lista := make([]reciente, 0, len(r.d))
	for app, cuando := range r.d {
		lista = append(lista, reciente{app, cuando})
	}
	r.mu.Unlock()
	sort.Slice(lista, func(i, j int) bool { return lista[i].Cuando > lista[j].Cuando })
	return lista
}

func (r *Recientes) bucle() {
	r.mu.Lock()
	if len(r.d) == 0 {
		// Primera vez: arrancar con la última sesión que quedó en el registro de Sunshine.
		ultima := leerSesion()
		app, _ := ultima.App.(string)
		desde, _ := ultima.Desde.(string)
		if app != "" && desde != "" {
			r.d[app] = strings.Replace(desde, " ", "T", 1)
			r.guardar()
		}
	}
	r.mu.Unlock()
	appID, app := 0, ""
	for {
		if info := leerInfoServidor(); info != nil && info.ocupado {
			if info.appID != appID || app == "" {
				appID = info.appID
				app, _ = leerSesion().App.(string)
			}
			if app != "" {
				r.anotar(app)
			}
		} else {
			appID, app = 0, ""
		}
		time.Sleep(cicloRecientes)
	}
}
