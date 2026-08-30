package collector

import (
	"context"
	"testing"
)

// TestMemoryCollectorMaquinaReal valida os invariantes contra o sistema em
// execução. Não fixa valores absolutos, que variam por máquina.
func TestMemoryCollectorMaquinaReal(t *testing.T) {
	metricas, err := NewMemoryCollector().Collect(context.Background())
	if err != nil {
		t.Fatalf("Collect() falhou: %v", err)
	}

	if metricas.Total == 0 {
		t.Error("Total = 0, esperado > 0")
	}

	if metricas.Used > metricas.Total {
		t.Errorf("Used (%d) maior que Total (%d)", metricas.Used, metricas.Total)
	}

	if metricas.Available > metricas.Total {
		t.Errorf("Available (%d) maior que Total (%d)", metricas.Available, metricas.Total)
	}

	if metricas.UsagePercent < 0 || metricas.UsagePercent > 100 {
		t.Errorf("UsagePercent = %v, esperado entre 0 e 100", metricas.UsagePercent)
	}

	// O percentual publicado precisa ser coerente com os bytes publicados ao
	// lado dele — é o motivo de recalcular em vez de usar stat.UsedPercent.
	esperado := percentOf(metricas.Used, metricas.Total)
	if metricas.UsagePercent != esperado {
		t.Errorf("UsagePercent = %v, incoerente com Used/Total (%v)", metricas.UsagePercent, esperado)
	}
}
