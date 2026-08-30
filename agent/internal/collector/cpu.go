package collector

import (
	"context"
	"errors"
	"fmt"
	"runtime"
	"sync"

	"github.com/shirou/gopsutil/v4/cpu"

	"pc-monitor-agent/internal/model"
)

// cpuCollector calcula o uso de CPU a partir da diferença entre duas leituras
// dos contadores acumulados do sistema.
//
// O baseline é guardado na própria instância, e não no estado global de
// cpu.Percent: o loop de publicação do WebSocket e as requisições REST
// convivem, e com o estado compartilhado do gopsutil uma chamada consumiria o
// baseline da outra, produzindo leituras erráticas.

type cpuCollector struct {
	mu        sync.Mutex
	lastTimes *cpu.TimesStat
}

// NewCPUCollector devolve um CPUCollector baseado no gopsutil.
func NewCPUCollector() CPUCollector {
	return &cpuCollector{}
}

func (c *cpuCollector) Collect(ctx context.Context) (model.CPUMetrics, error) {
	times, err := cpu.TimesWithContext(ctx, false)
	if err != nil {
		return model.CPUMetrics{}, fmt.Errorf("tempos de CPU: %w", err)
	}

	if len(times) == 0 {
		return model.CPUMetrics{}, errors.New("tempos de CPU: leitura vazia")
	}

	threads, err := cpu.CountsWithContext(ctx, true)
	if err != nil {
		return model.CPUMetrics{}, fmt.Errorf("contagem de threads: %w", err)
	}

	if threads <= 0 {
		return model.CPUMetrics{}, errors.New("contagem de threads: valor inválido")
	}

	// Núcleos físicos não são expostos em todo ambiente — contêineres e
	// algumas VMs só reportam os lógicos. Nesse caso o total lógico é a
	// melhor aproximação disponível.
	cores, err := cpu.CountsWithContext(ctx, false)
	if err != nil || cores <= 0 {
		cores = threads
	}

	return model.CPUMetrics{
		Model:        modelName(ctx),
		Cores:        cores,
		Threads:      threads,
		UsagePercent: c.usageSince(times[0]),
	}, nil
}

// usageSince consome o baseline anterior e registra o atual.
//
// Na primeira leitura o baseline é o zero absoluto, o que faz o resultado ser
// a média desde o boot — um valor real, em vez de um 0% enganoso.
func (c *cpuCollector) usageSince(current cpu.TimesStat) float64 {
	c.mu.Lock()
	defer c.mu.Unlock()

	previous := cpu.TimesStat{}
	if c.lastTimes != nil {
		previous = *c.lastTimes
	}

	c.lastTimes = &current

	return busyPercent(previous, current)
}

// modelName devolve o nome comercial do processador, ou string vazia quando o
// sistema não o expõe. A falha não é fatal: o resto das métricas continua útil.
func modelName(ctx context.Context) string {
	infos, err := cpu.InfoWithContext(ctx)
	if err != nil || len(infos) == 0 {
		return ""
	}

	return infos[0].ModelName
}

// busyPercent devolve o percentual de tempo ocupado entre duas leituras.
func busyPercent(previous, current cpu.TimesStat) float64 {
	totalPrevious, busyPrevious := busyTimes(previous)
	totalCurrent, busyCurrent := busyTimes(current)

	totalDelta := totalCurrent - totalPrevious
	busyDelta := busyCurrent - busyPrevious

	// Contadores só crescem; um delta não positivo significa leitura repetida
	// no mesmo tick do relógio, ou reinício de contador.
	if totalDelta <= 0 {
		return 0
	}

	return clampPercent(busyDelta / totalDelta * 100)
}

// busyTimes separa o tempo total do tempo ocupado de uma leitura.
func busyTimes(times cpu.TimesStat) (total, busy float64) {
	total = times.Total()

	if runtime.GOOS == "linux" {
		// No Linux, Guest e GuestNice já estão contabilizados dentro de User;
		// mantê-los no total os contaria duas vezes.
		total -= times.Guest
		total -= times.GuestNice
	}

	busy = total - times.Idle - times.Iowait

	return total, busy
}
