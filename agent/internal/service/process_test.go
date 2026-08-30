package service

import (
	"context"
	"errors"
	"testing"

	"pc-monitor-agent/internal/model"
)

type processCollectorFalso struct {
	processos []model.ProcessMetrics
	err       error
}

func (f processCollectorFalso) Collect(context.Context) ([]model.ProcessMetrics, error) {
	return f.processos, f.err
}

func processosDeTeste() []model.ProcessMetrics {
	return []model.ProcessMetrics{
		{PID: 4321, Name: "firefox", CPUPercent: 12.5, Memory: 2 << 30},
		{PID: 1, Name: "systemd", CPUPercent: 0, Memory: 12 << 20},
		{PID: 900, Name: "java", CPUPercent: 180, Memory: 1 << 30},
		{PID: 12, Name: "kworker/0:1", CPUPercent: 0, Memory: 0},
	}
}

func pids(processos []model.ProcessMetrics) []int32 {
	obtidos := make([]int32, 0, len(processos))
	for _, atual := range processos {
		obtidos = append(obtidos, atual.PID)
	}

	return obtidos
}

func exigirPIDs(t *testing.T, obtidos []model.ProcessMetrics, esperados ...int32) {
	t.Helper()

	if len(obtidos) != len(esperados) {
		t.Fatalf("PIDs = %v, esperado %v", pids(obtidos), esperados)
	}

	for i, esperado := range esperados {
		if obtidos[i].PID != esperado {
			t.Fatalf("PIDs = %v, esperado %v", pids(obtidos), esperados)
		}
	}
}

func TestSortProcessesPorCPU(t *testing.T) {
	processos := processosDeTeste()

	sortProcesses(processos, ProcessSortByCPU)

	// java (180) > firefox (12,5) > empatados em 0, desempatados pelo PID.
	exigirPIDs(t, processos, 900, 4321, 1, 12)
}

func TestSortProcessesPorMemoria(t *testing.T) {
	processos := processosDeTeste()

	sortProcesses(processos, ProcessSortByMemory)

	exigirPIDs(t, processos, 4321, 900, 1, 12)
}

// TestSortProcessesDesempateEstavel documenta por que o PID entra na
// comparação: sem ele, as dezenas de daemons parados em 0% trocariam de
// posição entre coletas e a tabela da UI ficaria piscando.
func TestSortProcessesDesempateEstavel(t *testing.T) {
	primeira := []model.ProcessMetrics{
		{PID: 30, CPUPercent: 0}, {PID: 10, CPUPercent: 0}, {PID: 20, CPUPercent: 0},
	}
	segunda := []model.ProcessMetrics{
		{PID: 20, CPUPercent: 0}, {PID: 30, CPUPercent: 0}, {PID: 10, CPUPercent: 0},
	}

	sortProcesses(primeira, ProcessSortByCPU)
	sortProcesses(segunda, ProcessSortByCPU)

	exigirPIDs(t, primeira, 10, 20, 30)
	exigirPIDs(t, segunda, 10, 20, 30)
}

func TestLimitProcesses(t *testing.T) {
	casos := []struct {
		nome     string
		limite   int
		esperado int
	}{
		{"limite menor que a lista", 2, 2},
		{"limite maior que a lista", 100, 4},
		{"limite igual ao tamanho", 4, 4},
		{"limite zero devolve tudo", 0, 4},
		{"limite negativo devolve tudo", -1, 4},
	}

	for _, caso := range casos {
		t.Run(caso.nome, func(t *testing.T) {
			obtido := limitProcesses(processosDeTeste(), caso.limite)

			if len(obtido) != caso.esperado {
				t.Errorf("len = %d, esperado %d", len(obtido), caso.esperado)
			}
		})
	}
}

func TestProcessServiceListOrdenaELimita(t *testing.T) {
	svc := NewProcessService(processCollectorFalso{processos: processosDeTeste()})

	obtido, err := svc.List(context.Background(), ProcessOptions{SortBy: ProcessSortByMemory, Limit: 2})
	if err != nil {
		t.Fatalf("List() falhou: %v", err)
	}

	exigirPIDs(t, obtido, 4321, 900)
}

func TestProcessServiceListConverteErro(t *testing.T) {
	falhaOriginal := errors.New("/proc ilegível")

	_, err := NewProcessService(processCollectorFalso{err: falhaOriginal}).
		List(context.Background(), ProcessOptions{})
	if err == nil {
		t.Fatal("List() deveria falhar quando o collector falha")
	}

	if !errors.Is(err, ErrUnavailable) {
		t.Errorf("erro não classificado como ErrUnavailable: %v", err)
	}

	if !errors.Is(err, falhaOriginal) {
		t.Errorf("causa original perdida no encadeamento: %v", err)
	}
}
