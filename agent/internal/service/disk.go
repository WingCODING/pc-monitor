package service

import (
	"context"
	"fmt"

	"pc-monitor-agent/internal/collector"
	"pc-monitor-agent/internal/model"
)

// DiskService entrega as métricas de disco para a camada HTTP.
type DiskService struct {
	collector collector.DiskCollector
}

// NewDiskService recebe o collector por dependência.
func NewDiskService(c collector.DiskCollector) *DiskService {
	return &DiskService{collector: c}
}

// List devolve uma entrada por partição.
func (s *DiskService) List(ctx context.Context) ([]model.DiskMetrics, error) {
	disks, err := s.collector.Collect(ctx)
	if err != nil {
		return nil, fmt.Errorf("%w: discos: %w", ErrUnavailable, err)
	}

	return disks, nil
}

// Summary condensa todas as partições num único indicador para o dashboard.
func (s *DiskService) Summary(ctx context.Context) (model.DiskSummary, error) {
	disks, err := s.List(ctx)
	if err != nil {
		return model.DiskSummary{}, err
	}

	return summarizeDisks(disks), nil
}

// summarizeDisks soma capacidade e uso das partições.
//
// A soma só é correta porque o collector já removeu montagens duplicadas do
// mesmo dispositivo; sem isso, um filesystem btrfs com quatro subvolumes seria
// contado quatro vezes.
func summarizeDisks(disks []model.DiskMetrics) model.DiskSummary {
	summary := model.DiskSummary{}

	for _, current := range disks {
		summary.Total += current.Total
		summary.Used += current.Used
	}

	summary.UsagePercent = percentOfBytes(summary.Used, summary.Total)

	return summary
}

// percentOfBytes repete a regra de percentOf do pacote collector: total zero
// devolve 0 em vez de NaN, que quebraria a desserialização no cliente.
func percentOfBytes(used, total uint64) float64 {
	if total == 0 {
		return 0
	}

	percent := float64(used) / float64(total) * 100

	switch {
	case percent < 0:
		return 0
	case percent > 100:
		return 100
	default:
		return percent
	}
}
