package service

import (
	"context"
	"log/slog"
	"time"

	"pc-monitor-agent/internal/model"
)

// MetricsService monta o snapshot agregado consumido pelo dashboard.
type MetricsService struct {
	cpu    *CPUService
	memory *MemoryService
}

// NewMetricsService compõe os services já existentes em vez de falar
// diretamente com os collectors, reaproveitando a classificação de erros.
func NewMetricsService(cpu *CPUService, memory *MemoryService) *MetricsService {
	return &MetricsService{cpu: cpu, memory: memory}
}

// Snapshot devolve o estado atual da máquina.
//
// Não devolve erro por decisão de desenho: a falha de um collector individual
// omite aquela métrica do snapshot, mas as demais continuam sendo publicadas.
// Um problema no /proc/stat não deve apagar a leitura de memória da tela.
//
// A coleta é sequencial: ler CPU e memória do /proc custa microssegundos, e
// paralelizar traria concorrência sem ganho mensurável.
func (s *MetricsService) Snapshot(ctx context.Context) model.DashboardMetrics {
	snapshot := model.DashboardMetrics{
		// UTC para que o timestamp seja comparável independentemente do fuso
		// da máquina que roda o agente.
		Timestamp: time.Now().UTC(),
	}

	if s.cpu != nil {
		if metrics, err := s.cpu.Current(ctx); err != nil {
			slog.Warn("cpu indisponível no snapshot", "error", err)
		} else {
			snapshot.CPU = &metrics
		}
	}

	if s.memory != nil {
		if metrics, err := s.memory.Current(ctx); err != nil {
			slog.Warn("memória indisponível no snapshot", "error", err)
		} else {
			snapshot.Memory = &metrics
		}
	}

	return snapshot
}
