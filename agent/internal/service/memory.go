package service

import (
	"context"
	"fmt"

	"pc-monitor-agent/internal/collector"
	"pc-monitor-agent/internal/model"
)

// MemoryService entrega as métricas de memória RAM para a camada HTTP.
type MemoryService struct {
	collector collector.MemoryCollector
}

// NewMemoryService recebe o collector por dependência.
func NewMemoryService(c collector.MemoryCollector) *MemoryService {
	return &MemoryService{collector: c}
}

// Current devolve o uso atual da memória.
func (s *MemoryService) Current(ctx context.Context) (model.MemoryMetrics, error) {
	metrics, err := s.collector.Collect(ctx)
	if err != nil {
		return model.MemoryMetrics{}, fmt.Errorf("%w: memória: %w", ErrUnavailable, err)
	}

	return metrics, nil
}
