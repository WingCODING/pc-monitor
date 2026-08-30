package service

import (
	"context"
	"errors"
	"testing"

	"pc-monitor-agent/internal/model"
)

// cpuCollectorFalso substitui o collector real, isolando o service do sistema
// operacional e permitindo testar o caminho de erro sem provocar uma falha
// verdadeira de hardware.
type cpuCollectorFalso struct {
	metricas model.CPUMetrics
	err      error
}

func (f cpuCollectorFalso) Collect(context.Context) (model.CPUMetrics, error) {
	return f.metricas, f.err
}

func TestCPUServiceCurrent(t *testing.T) {
	esperado := model.CPUMetrics{
		Model:        "AMD Ryzen 7 7800X3D",
		Cores:        8,
		Threads:      16,
		UsagePercent: 24.8,
	}

	servico := NewCPUService(cpuCollectorFalso{metricas: esperado})

	obtido, err := servico.Current(context.Background())
	if err != nil {
		t.Fatalf("Current() falhou: %v", err)
	}

	if obtido != esperado {
		t.Errorf("Current() = %+v, esperado %+v", obtido, esperado)
	}
}

func TestCPUServiceCurrentConverteErro(t *testing.T) {
	falhaOriginal := errors.New("/proc/stat ilegível")

	servico := NewCPUService(cpuCollectorFalso{err: falhaOriginal})

	_, err := servico.Current(context.Background())
	if err == nil {
		t.Fatal("Current() deveria falhar quando o collector falha")
	}

	if !errors.Is(err, ErrUnavailable) {
		t.Errorf("erro não classificado como ErrUnavailable: %v", err)
	}

	// A causa original precisa sobreviver ao encadeamento para aparecer no log
	// do servidor, mesmo que não seja exposta ao cliente.
	if !errors.Is(err, falhaOriginal) {
		t.Errorf("causa original perdida no encadeamento: %v", err)
	}
}
