package main

import (
	"fmt"
	"log"
	"math"
	"net/netip"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"
)

const (
	gb             = 1_000_000_000 // GB "de compañía" (1000 MB): cuenta un poco antes que 1024
	diasHistorial  = 70
	cicloTailscale = 5 * time.Second // entre mediciones de tráfico
	formatoDia     = "2006-01-02"
)

type configDatos struct {
	LimiteGB float64 `json:"limite_gb"`
	AvisoPct int     `json:"aviso_pct"`
	Accion   string  `json:"accion"` // «avisar» o «cortar»
	DiaCorte int     `json:"dia_corte"`
}

type ajusteDatos struct {
	Inicio string `json:"inicio"`
	Bytes  int64  `json:"bytes"`
}

type avisosDatos struct {
	Inicio string `json:"inicio"`
	Aviso  bool   `json:"aviso"`
	Limite bool   `json:"limite"`
}

// Lo que se guarda en datos.json (el mismo archivo que el panel en Python).
type archivoDatos struct {
	Config configDatos      `json:"config"`
	Dias   map[string]int64 `json:"dias"`
	Ajuste *ajusteDatos     `json:"ajuste"`
	Avisos avisosDatos      `json:"avisos"`
}

type celuDatos struct {
	Nombre    string `json:"nombre"`
	Online    bool   `json:"online"`
	Via       string `json:"via"`
	Direccion any    `json:"direccion"`
	Rebote    any    `json:"rebote"`
	IP        any    `json:"ip"`
}

// Contador suma el tráfico de Tailscale con el celu cuando está fuera de casa.
//
// En el wifi de casa Tailscale va por la red local (dirección privada): eso no gasta
// datos y no se cuenta.
type Contador struct {
	mu             sync.Mutex
	d              archivoDatos
	archivo        string
	previos        map[string][2]int64
	sesion         int64
	ocupado        bool
	celu           *celuDatos
	ultimoGuardado time.Time
	sucio          bool
	ultimoCorte    time.Time
}

func nuevoContador() *Contador {
	c := &Contador{
		archivo: filepath.Join(lugar.estado, "datos.json"),
		d:       archivoDatos{Config: configDatos{AvisoPct: 80, Accion: "avisar", DiaCorte: 1}},
		previos: map[string][2]int64{},
	}
	leerJSON(c.archivo, &c.d)
	if c.d.Dias == nil {
		c.d.Dias = map[string]int64{}
	}
	return c
}

// ------------------------------------------------------------------ cuentas

func inicioCiclo(hoy time.Time, dia int) time.Time {
	if hoy.Day() >= dia {
		return time.Date(hoy.Year(), hoy.Month(), dia, 0, 0, 0, 0, time.Local)
	}
	return time.Date(hoy.Year(), hoy.Month()-1, dia, 0, 0, 0, 0, time.Local)
}

func (c *Contador) ciclo() (time.Time, time.Time) {
	inicio := inicioCiclo(time.Now(), c.d.Config.DiaCorte)
	return inicio, inicio.AddDate(0, 1, 0) // el día de corte llega hasta 28: no se pasa de mes
}

func (c *Contador) bytesEntre(desde, hasta time.Time) int64 {
	d, h := desde.Format(formatoDia), hasta.Format(formatoDia)
	var total int64
	for dia, bytes := range c.d.Dias {
		if d <= dia && dia < h {
			total += bytes
		}
	}
	return total
}

func (c *Contador) bytesCiclo() int64 {
	inicio, fin := c.ciclo()
	total := c.bytesEntre(inicio, fin)
	if a := c.d.Ajuste; a != nil && a.Inicio == inicio.Format(formatoDia) {
		total -= a.Bytes
	}
	return max(total, 0)
}

func (c *Contador) avisos() *avisosDatos {
	inicio, _ := c.ciclo()
	if c.d.Avisos.Inicio != inicio.Format(formatoDia) {
		c.d.Avisos = avisosDatos{Inicio: inicio.Format(formatoDia)}
	}
	return &c.d.Avisos
}

// ------------------------------------------------------- medición periódica

func clasificar(par *parRed, delta int64) string {
	if par.CurAddr != "" {
		direccion, err := netip.ParseAddrPort(par.CurAddr)
		if err != nil {
			return "datos"
		}
		ip := direccion.Addr().Unmap()
		if ip.IsPrivate() || ip.IsLoopback() || ip.IsLinkLocalUnicast() {
			return "casa"
		}
		return "datos"
	}
	if par.Relay != "" && (par.Active || delta > 0) {
		return "rebote"
	}
	return "inactivo"
}

func (c *Contador) paso() {
	estado := leerEstadoRed(true)
	if estado == nil {
		return
	}
	info := leerInfoServidor()
	c.mu.Lock()
	defer c.mu.Unlock()
	ocupado := info != nil && info.ocupado
	if ocupado && !c.ocupado {
		c.sesion = 0
	}
	c.ocupado = ocupado
	var remoto int64
	var celu *celuDatos
	for _, par := range estado.Peer {
		clave := par.PublicKey
		if clave == "" {
			clave = par.ID
		}
		previo, habia := c.previos[clave]
		c.previos[clave] = [2]int64{par.TxBytes, par.RxBytes}
		var delta int64
		if habia {
			// Tailscale pone los contadores en cero cuando el par queda inactivo.
			delta = diferencia(par.TxBytes, previo[0]) + diferencia(par.RxBytes, previo[1])
		}
		via := clasificar(par, delta)
		if via == "datos" || via == "rebote" {
			remoto += delta
		}
		// Puede haber más de un celu en la red: vale el que está conectado.
		if (par.OS == "android" || par.OS == "iOS") && (celu == nil || (par.Online && !celu.Online)) {
			celu = &celuDatos{Nombre: par.HostName, Online: par.Online, Via: via,
				Direccion: oNulo(par.CurAddr), Rebote: oNulo(par.Relay)}
			if len(par.TailscaleIPs) > 0 {
				celu.IP = par.TailscaleIPs[0]
			}
		}
	}
	c.celu = celu
	if remoto > 0 {
		hoy := time.Now().Format(formatoDia)
		c.d.Dias[hoy] += remoto
		if c.ocupado {
			c.sesion += remoto
		}
		c.sucio = true
	}
	c.revisarLimite(remoto)
	if c.sucio && time.Since(c.ultimoGuardado) > 15*time.Second {
		c.escribir()
	}
}

func diferencia(ahora, antes int64) int64 {
	if ahora >= antes {
		return ahora - antes
	}
	return ahora
}

func oNulo(s string) any {
	if s == "" {
		return nil
	}
	return s
}

func (c *Contador) revisarLimite(remoto int64) {
	conf := c.d.Config
	limite := conf.LimiteGB * gb
	if limite <= 0 {
		return
	}
	usado := float64(c.bytesCiclo())
	avisos := c.avisos()
	texto := strings.ReplaceAll(fmt.Sprintf("Llevás %.2f GB de %s GB este ciclo",
		usado/gb, strconv.FormatFloat(conf.LimiteGB, 'g', -1, 64)), ".", ",")
	if usado >= limite*float64(conf.AvisoPct)/100 && !avisos.Aviso {
		avisos.Aviso = true
		c.sucio = true
		notificar("Ojo con los datos", texto, false)
	}
	if usado < limite {
		return
	}
	if !avisos.Limite {
		avisos.Limite = true
		c.sucio = true
		notificar("Llegaste al límite de datos", texto, true)
	}
	if conf.Accion == "cortar" && remoto > 0 && c.ocupado && time.Since(c.ultimoCorte) > time.Minute {
		c.ultimoCorte = time.Now()
		go func() { // sin frenar la medición mientras Sunshine contesta
			if _, err := sunshine.cerrarApp(); err != nil {
				notificar("No pude cortar la sesión", "Entrá al panel para que guarde el acceso a Sunshine.", true)
			} else {
				notificar("Sesión cortada", "Se cortó el stream para no gastar más datos.", true)
			}
		}()
	}
}

// escribir guarda datos.json; con el candado ya tomado.
func (c *Contador) escribir() {
	limite := time.Now().AddDate(0, 0, -diasHistorial).Format(formatoDia)
	for dia := range c.d.Dias {
		if dia < limite {
			delete(c.d.Dias, dia)
		}
	}
	if err := guardarJSON(c.archivo, c.d, false); err != nil {
		log.Printf("no pude guardar el consumo de datos: %v", err)
	}
	c.ultimoGuardado = time.Now()
	c.sucio = false
}

func (c *Contador) guardar() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.escribir()
}

func (c *Contador) bucle() {
	for {
		c.paso()
		time.Sleep(cicloTailscale)
	}
}

// ------------------------------------------------------------- para la API

func (c *Contador) resumen() map[string]any {
	c.mu.Lock()
	defer c.mu.Unlock()
	conf := c.d.Config
	inicio, fin := c.ciclo()
	anterior := inicioCiclo(inicio.AddDate(0, 0, -1), conf.DiaCorte)
	hoy := time.Now()
	dias := []map[string]any{}
	for i := 13; i >= 0; i-- {
		dia := hoy.AddDate(0, 0, -i).Format(formatoDia)
		dias = append(dias, map[string]any{"fecha": dia, "bytes": c.d.Dias[dia]})
	}
	ciclo := c.bytesCiclo()
	limite := conf.LimiteGB * gb
	var sesion, limiteJSON, pct any
	if c.ocupado {
		sesion = c.sesion
	}
	if limite > 0 {
		limiteJSON = limite
		pct = math.Round(1000*float64(ciclo)/limite) / 10
	}
	var celu any
	if c.celu != nil {
		celu = c.celu
	}
	return map[string]any{
		"ciclo": ciclo, "hoy": c.d.Dias[hoy.Format(formatoDia)],
		"sesion": sesion, "limite": limiteJSON, "pct": pct,
		"inicio": inicio.Format(formatoDia), "fin": fin.Format(formatoDia),
		"anterior": c.bytesEntre(anterior, inicio),
		"config":   conf, "dias": dias, "celu": celu,
	}
}

func (c *Contador) configurar(nuevo map[string]any) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	conf := c.d.Config // se aplica recién si todo vino bien
	if v, ok := nuevo["limite_gb"]; ok {
		n, err := numero(v)
		if err != nil {
			return err
		}
		conf.LimiteGB = limitar(n, 0, 1000)
	}
	if v, ok := nuevo["aviso_pct"]; ok {
		n, err := numero(v)
		if err != nil {
			return err
		}
		conf.AvisoPct = limitar(int(n), 10, 99)
	}
	if accion, _ := nuevo["accion"].(string); accion == "avisar" || accion == "cortar" {
		conf.Accion = accion
	}
	if v, ok := nuevo["dia_corte"]; ok {
		n, err := numero(v)
		if err != nil {
			return err
		}
		conf.DiaCorte = limitar(int(n), 1, 28)
	}
	c.d.Config = conf
	c.d.Avisos = avisosDatos{} // con un límite nuevo se vuelve a avisar
	c.escribir()
	return nil
}

func (c *Contador) ponerEnCero() {
	c.mu.Lock()
	defer c.mu.Unlock()
	inicio, fin := c.ciclo()
	c.d.Ajuste = &ajusteDatos{Inicio: inicio.Format(formatoDia), Bytes: c.bytesEntre(inicio, fin)}
	c.d.Avisos = avisosDatos{}
	c.escribir()
}
