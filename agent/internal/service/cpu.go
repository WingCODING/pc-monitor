package service

import (
	"context"
	"fmt"

	"pc-monitor-agent/internal/collector"
	"pc-monitor-agent/internal/model"
)

// CPUService entrega as métricas de CPU para a camada HTTP.
type CPUService struct {
	collector collector.CPUCollector
}

// NewCPUService recebe o collector por dependência, o que permite substituí-lo
// por um dublê nos testes.
func NewCPUService(c collector.CPUCollector) *CPUService {
	return &CPUService{collector: c}
}

// Current devolve o uso atual do processador.
func (s *CPUService) Current(ctx context.Context) (model.CPUMetrics, error) {
	metrics, err := s.collector.Collect(ctx)
	if err != nil {
		return model.CPUMetrics{}, fmt.Errorf("%w: cpu: %w", ErrUnavailable, err)
	}

	return metrics, nil
}
