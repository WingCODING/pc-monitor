package service

import (
	"context"
	"time"

	"pc-monitor-agent/internal/model"
)

// MetricsService monta o snapshot agregado consumido pelo dashboard.
type MetricsService struct {
	cpu     *CPUService
	memory  *MemoryService
	system  *SystemService
	disk    *DiskService
	network *NetworkService

	// warn evita que um collector quebrado escreva uma linha por segundo no
	// log enquanto o loop do WebSocket estiver publicando.
	warn *warnThrottle
}

// NewMetricsService compõe os services já existentes em vez de falar
// diretamente com os collectors, reaproveitando a classificação de erros.
func NewMetricsService(cpu *CPUService, memory *MemoryService, system *SystemService, disk *DiskService, network *NetworkService) *MetricsService {
	return &MetricsService{
		cpu:     cpu,
		memory:  memory,
		system:  system,
		disk:    disk,
		network: network,
		warn:    newWarnThrottle(warnThrottleInterval),
	}
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
			s.warn.warn("cpu", "cpu indisponível no snapshot", "error", err)
		} else {
			snapshot.CPU = &metrics
		}
	}

	if s.memory != nil {
		if metrics, err := s.memory.Current(ctx); err != nil {
			s.warn.warn("memory", "memória indisponível no snapshot", "error", err)
		} else {
			snapshot.Memory = &metrics
		}
	}

	if s.disk != nil {
		if summary, err := s.disk.Summary(ctx); err != nil {
			s.warn.warn("disk", "disco indisponível no snapshot", "error", err)
		} else {
			snapshot.Disk = &summary
		}
	}

	if s.network != nil {
		if summary, err := s.network.Summary(ctx); err != nil {
			s.warn.warn("network", "rede indisponível no snapshot", "error", err)
		} else {
			snapshot.Network = &summary
		}
	}

	if s.system != nil {
		if metrics, err := s.system.Current(ctx); err != nil {
			s.warn.warn("system", "sistema indisponível no snapshot", "error", err)
		} else {
			uptime := metrics.UptimeSeconds
			snapshot.UptimeSeconds = &uptime
		}
	}

	return snapshot
}
