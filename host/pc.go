package main

import (
	"strconv"
	"strings"
	"sync"
	"time"
)

// MedidorPC: uso de procesador, placa de video y memoria, para «Sistema» en el panel.
// Lo que depende del sistema (tiemposCPU, memoria, encendida) está en *_linux.go / *_windows.go.
type MedidorPC struct {
	mu        sync.Mutex
	cpuPrevio [2]uint64 // inactivo, total
	cpu       any
	gpu       any
	gpuHora   time.Time
}

func (m *MedidorPC) medirCPU() any {
	inactivo, total, ok := tiemposCPU()
	if !ok {
		return m.cpu
	}
	if m.cpuPrevio[1] != 0 && total > m.cpuPrevio[1] {
		di, dt := float64(inactivo-m.cpuPrevio[0]), float64(total-m.cpuPrevio[1])
		m.cpu = int(100*(1-di/dt) + 0.5)
	}
	m.cpuPrevio = [2]uint64{inactivo, total}
	return m.cpu
}

// medirGPU: solo placas NVIDIA (nvidia-smi, en Linux y en Windows); cada 4 segundos.
func (m *MedidorPC) medirGPU() any {
	if time.Since(m.gpuHora) < 4*time.Second {
		return m.gpu
	}
	m.gpuHora = time.Now()
	m.gpu = nil
	_, salida := correr(5*time.Second, "nvidia-smi",
		"--query-gpu=name,temperature.gpu,utilization.gpu,memory.used,memory.total",
		"--format=csv,noheader,nounits")
	campos := strings.Split(strings.SplitN(salida, "\n", 2)[0], ",")
	if len(campos) != 5 {
		return nil
	}
	valores := make([]float64, 4)
	for i, campo := range campos[1:] {
		v, err := strconv.ParseFloat(strings.TrimSpace(campo), 64)
		if err != nil {
			return nil
		}
		valores[i] = v
	}
	m.gpu = map[string]any{
		"nombre": strings.Replace(strings.TrimSpace(campos[0]), "NVIDIA GeForce ", "", 1),
		"temp":   valores[0], "uso": valores[1],
		"vram_usada": valores[2] * (1 << 20), "vram_total": valores[3] * (1 << 20),
	}
	return m.gpu
}

func (m *MedidorPC) medir() map[string]any {
	m.mu.Lock()
	defer m.mu.Unlock()
	total, disponible := memoria()
	var encendidaJSON any
	if seg, ok := encendida(); ok {
		encendidaJSON = seg
	}
	return map[string]any{
		"cpu": m.medirCPU(), "gpu": m.medirGPU(),
		"ram_total": total, "ram_usada": total - disponible,
		"encendida": encendidaJSON,
	}
}

// ---------------------------------------------------------------------- WoW
// Solo si está instalado (Battle.net). Lo que depende del sistema (wowInstalado,
// wowAbierto, battlenetAbierto, terminarWow, abrirBattlenet) está en *_linux.go / *_windows.go.

// estadoJuego: para el panel; nil si en esta PC no hay WoW (la categoría no aparece).
func estadoJuego() any {
	if !wowInstalado() {
		return nil
	}
	return map[string]bool{"wow": wowAbierto(), "battlenet": battlenetAbierto()}
}

// cerrarWow: primero le pide que se cierre; si en 6 segundos sigue, lo cierra a la fuerza.
func cerrarWow() bool {
	if !wowAbierto() {
		return false
	}
	terminarWow(false)
	for range 12 {
		time.Sleep(500 * time.Millisecond)
		if !wowAbierto() {
			return true
		}
	}
	terminarWow(true)
	return true
}
