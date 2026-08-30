package service

import (
	"context"
	"fmt"
	"slices"

	"pc-monitor-agent/internal/collector"
	"pc-monitor-agent/internal/model"
)

// ProcessSortBy identifica o critério de ordenação da lista de processos.
type ProcessSortBy string

const (
	// ProcessSortByCPU ordena do maior para o menor uso de CPU.
	ProcessSortByCPU ProcessSortBy = "cpu"
	// ProcessSortByMemory ordena da maior para a menor memória residente.
	ProcessSortByMemory ProcessSortBy = "memory"
)

const (
	// DefaultProcessSort reflete o que um monitor mostra por padrão: quem está
	// consumindo CPU agora.
	DefaultProcessSort = ProcessSortByCPU

	// DefaultProcessLimit corta a lista num tamanho que cabe numa tela. Uma
	// máquina de desenvolvimento tem centenas de processos, e a cauda é quase
	// toda de daemons ociosos.
	DefaultProcessLimit = 50

	// MaxProcessLimit protege o cliente de uma resposta desnecessariamente
	// grande num endpoint consultado a cada poucos segundos.
	MaxProcessLimit = 500
)

// ProcessOptions controla ordenação e tamanho da lista devolvida.
type ProcessOptions struct {
	SortBy ProcessSortBy

	// Limit zero ou negativo devolve a lista inteira; quem expõe o endpoint
	// aplica o padrão antes de chamar.
	Limit int
}

// ProcessService entrega a lista de processos para a camada HTTP.
type ProcessService struct {
	collector collector.ProcessCollector
}

// NewProcessService recebe o collector por dependência.
func NewProcessService(c collector.ProcessCollector) *ProcessService {
	return &ProcessService{collector: c}
}

// List devolve os processos já ordenados e cortados conforme as opções.
func (s *ProcessService) List(ctx context.Context, options ProcessOptions) ([]model.ProcessMetrics, error) {
	processes, err := s.collector.Collect(ctx)
	if err != nil {
		return nil, fmt.Errorf("%w: processos: %w", ErrUnavailable, err)
	}

	sortProcesses(processes, options.SortBy)

	return limitProcesses(processes, options.Limit), nil
}

// sortProcesses ordena in-place pelo critério pedido.
//
// O PID desempata: sem isso, processos com o mesmo consumo — dezenas de
// daemons ociosos em 0% — trocariam de posição entre coletas e a tabela da UI
// ficaria piscando.
func sortProcesses(processes []model.ProcessMetrics, sortBy ProcessSortBy) {
	slices.SortFunc(processes, func(a, b model.ProcessMetrics) int {
		var order int

		if sortBy == ProcessSortByMemory {
			order = compareDescending(float64(a.Memory), float64(b.Memory))
		} else {
			order = compareDescending(a.CPUPercent, b.CPUPercent)
		}

		if order != 0 {
			return order
		}

		return int(a.PID - b.PID)
	})
}

// compareDescending devolve a ordem de a e b do maior para o menor.
func compareDescending(a, b float64) int {
	switch {
	case a > b:
		return -1
	case a < b:
		return 1
	default:
		return 0
	}
}

// limitProcesses corta a lista mantendo a slice original intacta a partir do
// início — os elementos já estão ordenados, então o corte é a cauda.
func limitProcesses(processes []model.ProcessMetrics, limit int) []model.ProcessMetrics {
	if limit <= 0 || limit >= len(processes) {
		return processes
	}

	return processes[:limit]
}
