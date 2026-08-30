package collector

import (
	"context"
	"errors"
	"fmt"

	"github.com/shirou/gopsutil/v4/mem"

	"pc-monitor-agent/internal/model"
)

// memoryCollector lê o uso da memória RAM.
//
// Diferente do de CPU, não guarda estado: os valores de memória são
// instantâneos, não contadores acumulados.
type memoryCollector struct{}

// NewMemoryCollector devolve um MemoryCollector baseado no gopsutil.
func NewMemoryCollector() MemoryCollector {
	return memoryCollector{}
}

func (memoryCollector) Collect(ctx context.Context) (model.MemoryMetrics, error) {
	stat, err := mem.VirtualMemoryWithContext(ctx)
	if err != nil {
		return model.MemoryMetrics{}, fmt.Errorf("memória virtual: %w", err)
	}

	if stat == nil {
		return model.MemoryMetrics{}, errors.New("memória virtual: leitura vazia")
	}

	if stat.Total == 0 {
		return model.MemoryMetrics{}, errors.New("memória virtual: total zerado")
	}

	return model.MemoryMetrics{
		Total:     stat.Total,
		Used:      stat.Used,
		Available: stat.Available,

		// O percentual é recalculado a partir de Used e Total em vez de usar
		// stat.UsedPercent, garantindo que o número publicado seja sempre
		// coerente com os bytes publicados ao lado dele.
		UsagePercent: percentOf(stat.Used, stat.Total),
	}, nil
}
