package service

import (
	"context"
	"errors"
	"testing"

	"pc-monitor-agent/internal/model"
)

type memoryCollectorFalso struct {
	metricas model.MemoryMetrics
	err      error
}

func (f memoryCollectorFalso) Collect(context.Context) (model.MemoryMetrics, error) {
	return f.metricas, f.err
}

func TestMemoryServiceCurrent(t *testing.T) {
	esperado := model.MemoryMetrics{
		Total:        34359738368,
		Used:         17179869184,
		Available:    17179869184,
		UsagePercent: 50,
	}

	obtido, err := NewMemoryService(memoryCollectorFalso{metricas: esperado}).Current(context.Background())
	if err != nil {
		t.Fatalf("Current() falhou: %v", err)
	}

	if obtido != esperado {
		t.Errorf("Current() = %+v, esperado %+v", obtido, esperado)
	}
}

func TestMemoryServiceCurrentConverteErro(t *testing.T) {
	falhaOriginal := errors.New("/proc/meminfo ilegível")

	_, err := NewMemoryService(memoryCollectorFalso{err: falhaOriginal}).Current(context.Background())
	if err == nil {
		t.Fatal("Current() deveria falhar quando o collector falha")
	}

	if !errors.Is(err, ErrUnavailable) {
		t.Errorf("erro não classificado como ErrUnavailable: %v", err)
	}

	if !errors.Is(err, falhaOriginal) {
		t.Errorf("causa original perdida no encadeamento: %v", err)
	}
}
