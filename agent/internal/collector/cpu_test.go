package collector

import (
	"context"
	"testing"

	"github.com/shirou/gopsutil/v4/cpu"
)

// TestBusyPercent exercita o cálculo sem tocar no sistema operacional, o que
// permite cobrir cenários difíceis de reproduzir em uma máquina real.
func TestBusyPercent(t *testing.T) {
	casos := []struct {
		nome     string
		previous cpu.TimesStat
		current  cpu.TimesStat
		esperado float64
	}{
		{
			nome:     "metade do tempo ocupado",
			previous: cpu.TimesStat{User: 100, Idle: 100},
			current:  cpu.TimesStat{User: 150, Idle: 150},
			esperado: 50,
		},
		{
			nome:     "totalmente ocioso",
			previous: cpu.TimesStat{User: 100, Idle: 100},
			current:  cpu.TimesStat{User: 100, Idle: 200},
			esperado: 0,
		},
		{
			nome:     "totalmente ocupado",
			previous: cpu.TimesStat{User: 100, Idle: 100},
			current:  cpu.TimesStat{User: 200, Idle: 100},
			esperado: 100,
		},
		{
			nome:     "leitura repetida devolve zero",
			previous: cpu.TimesStat{User: 100, Idle: 100},
			current:  cpu.TimesStat{User: 100, Idle: 100},
			esperado: 0,
		},
		{
			nome:     "contador reiniciado não gera valor negativo",
			previous: cpu.TimesStat{User: 500, Idle: 500},
			current:  cpu.TimesStat{User: 10, Idle: 10},
			esperado: 0,
		},
		{
			nome:     "iowait conta como ocioso",
			previous: cpu.TimesStat{},
			current:  cpu.TimesStat{User: 50, Iowait: 50},
			esperado: 50,
		},
	}

	for _, caso := range casos {
		t.Run(caso.nome, func(t *testing.T) {
			obtido := busyPercent(caso.previous, caso.current)

			if obtido != caso.esperado {
				t.Errorf("busyPercent() = %v, esperado %v", obtido, caso.esperado)
			}
		})
	}
}

// TestCPUCollectorMaquinaReal valida o contrato contra o sistema em execução.
func TestCPUCollectorMaquinaReal(t *testing.T) {
	coletor := NewCPUCollector()

	metricas, err := coletor.Collect(context.Background())
	if err != nil {
		t.Fatalf("Collect() falhou: %v", err)
	}

	if metricas.Cores <= 0 {
		t.Errorf("Cores = %d, esperado > 0", metricas.Cores)
	}

	if metricas.Threads <= 0 {
		t.Errorf("Threads = %d, esperado > 0", metricas.Threads)
	}

	if metricas.Threads < metricas.Cores {
		t.Errorf("Threads (%d) menor que Cores (%d)", metricas.Threads, metricas.Cores)
	}

	if metricas.UsagePercent < 0 || metricas.UsagePercent > 100 {
		t.Errorf("UsagePercent = %v, esperado entre 0 e 100", metricas.UsagePercent)
	}
}

// TestCPUCollectorLeiturasConsecutivas garante que o baseline é mantido entre
// chamadas e que a segunda leitura continua dentro do contrato.
func TestCPUCollectorLeiturasConsecutivas(t *testing.T) {
	coletor := NewCPUCollector()
	ctx := context.Background()

	if _, err := coletor.Collect(ctx); err != nil {
		t.Fatalf("primeira coleta falhou: %v", err)
	}

	metricas, err := coletor.Collect(ctx)
	if err != nil {
		t.Fatalf("segunda coleta falhou: %v", err)
	}

	if metricas.UsagePercent < 0 || metricas.UsagePercent > 100 {
		t.Errorf("UsagePercent = %v, esperado entre 0 e 100", metricas.UsagePercent)
	}
}
