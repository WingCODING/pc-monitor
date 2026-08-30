package service

import (
	"context"
	"errors"
	"testing"

	"pc-monitor-agent/internal/model"
)

type systemCollectorFalso struct {
	metricas model.SystemMetrics
	err      error
}

func (f systemCollectorFalso) Collect(context.Context) (model.SystemMetrics, error) {
	return f.metricas, f.err
}

func TestSystemServiceCurrent(t *testing.T) {
	esperado := model.SystemMetrics{
		Hostname:      "desktop",
		OS:            "linux",
		OSVersion:     "4.0.1",
		Architecture:  "amd64",
		UptimeSeconds: 58232,
	}

	obtido, err := NewSystemService(systemCollectorFalso{metricas: esperado}).Current(context.Background())
	if err != nil {
		t.Fatalf("Current() falhou: %v", err)
	}

	if obtido != esperado {
		t.Errorf("Current() = %+v, esperado %+v", obtido, esperado)
	}
}

func TestSystemServiceCurrentConverteErro(t *testing.T) {
	falhaOriginal := errors.New("host inacessível")

	_, err := NewSystemService(systemCollectorFalso{err: falhaOriginal}).Current(context.Background())
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
