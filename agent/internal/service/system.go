package service

import (
	"context"
	"fmt"

	"pc-monitor-agent/internal/collector"
	"pc-monitor-agent/internal/model"
)

// SystemService entrega as informações gerais da máquina.
type SystemService struct {
	collector collector.SystemCollector
}

// NewSystemService recebe o collector por dependência.
func NewSystemService(c collector.SystemCollector) *SystemService {
	return &SystemService{collector: c}
}

// Current devolve as informações do sistema.
func (s *SystemService) Current(ctx context.Context) (model.SystemMetrics, error) {
	metrics, err := s.collector.Collect(ctx)
	if err != nil {
		return model.SystemMetrics{}, fmt.Errorf("%w: sistema: %w", ErrUnavailable, err)
	}

	return metrics, nil
}
