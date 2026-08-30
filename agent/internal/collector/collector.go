// Package collector define os contratos de coleta de métricas do sistema.
//
// As interfaces isolam o restante da aplicação dos detalhes de cada sistema
// operacional: nem os services nem a camada HTTP devem conhecer /proc, APIs
// nativas ou bibliotecas de monitoramento.
//
// Todo Collect recebe um context.Context para que coletas custosas — a de
// processos, sobretudo — possam ser canceladas junto com a requisição HTTP ou
// com o loop de publicação do WebSocket.
package collector

import (
	"context"

	"pc-monitor-agent/internal/model"
)

// CPUCollector coleta modelo, topologia e uso do processador.
type CPUCollector interface {
	Collect(ctx context.Context) (model.CPUMetrics, error)
}

// MemoryCollector coleta o uso da memória RAM.
type MemoryCollector interface {
	Collect(ctx context.Context) (model.MemoryMetrics, error)
}

// DiskCollector coleta o uso das partições montadas.
type DiskCollector interface {
	Collect(ctx context.Context) ([]model.DiskMetrics, error)
}

// NetworkCollector coleta o tráfego das interfaces de rede.
//
// As velocidades derivam da diferença entre leituras consecutivas, portanto a
// implementação guarda estado entre chamadas.
type NetworkCollector interface {
	Collect(ctx context.Context) ([]model.NetworkMetrics, error)
}

// ProcessCollector coleta os processos em execução.
type ProcessCollector interface {
	Collect(ctx context.Context) ([]model.ProcessMetrics, error)
}

// SystemCollector coleta informações gerais da máquina.
type SystemCollector interface {
	Collect(ctx context.Context) (model.SystemMetrics, error)
}
