package service

import (
	"context"
	"fmt"

	"pc-monitor-agent/internal/collector"
	"pc-monitor-agent/internal/model"
)

// NetworkService entrega as métricas de rede para a camada HTTP.
type NetworkService struct {
	collector collector.NetworkCollector
}

// NewNetworkService recebe o collector por dependência.
func NewNetworkService(c collector.NetworkCollector) *NetworkService {
	return &NetworkService{collector: c}
}

// List devolve uma entrada por interface de rede.
func (s *NetworkService) List(ctx context.Context) ([]model.NetworkMetrics, error) {
	interfaces, err := s.collector.Collect(ctx)
	if err != nil {
		return nil, fmt.Errorf("%w: rede: %w", ErrUnavailable, err)
	}

	return interfaces, nil
}

// Summary condensa todas as interfaces num único par download/upload.
func (s *NetworkService) Summary(ctx context.Context) (model.NetworkSummary, error) {
	interfaces, err := s.List(ctx)
	if err != nil {
		return model.NetworkSummary{}, err
	}

	return summarizeNetwork(interfaces), nil
}

// summarizeNetwork soma as velocidades de todas as interfaces.
//
// Somar velocidades é seguro, ao contrário do que acontece com discos: cada
// interface tem contadores próprios, então não há risco de contagem dupla.
func summarizeNetwork(interfaces []model.NetworkMetrics) model.NetworkSummary {
	summary := model.NetworkSummary{}

	for _, current := range interfaces {
		summary.DownloadBytesPerSecond += current.DownloadBytesPerSecond
		summary.UploadBytesPerSecond += current.UploadBytesPerSecond
	}

	return summary
}
