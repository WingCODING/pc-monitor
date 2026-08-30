package api

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"pc-monitor-agent/internal/model"
	"pc-monitor-agent/internal/service"
)

func routerComMetrics(cpu cpuCollectorFalso, memoria memoryCollectorFalso) http.Handler {
	sistema := systemCollectorFalso{metricas: model.SystemMetrics{Hostname: "desktop", UptimeSeconds: 53214}}

	disco := diskCollectorFalso{discos: []model.DiskMetrics{{Name: "/dev/dm-0", MountPoint: "/", Total: 1000, Used: 400, UsagePercent: 40}}}

	metrics := service.NewMetricsService(
		service.NewCPUService(cpu),
		service.NewMemoryService(memoria),
		service.NewSystemService(sistema),
		service.NewDiskService(disco),
	)

	return NewRouter(Deps{Metrics: metrics})
}

func TestHandleMetricsSucesso(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/api/v1/metrics", nil)
	rec := httptest.NewRecorder()

	routerComMetrics(
		cpuCollectorFalso{metricas: model.CPUMetrics{Cores: 8, Threads: 16, UsagePercent: 32.8}},
		memoryCollectorFalso{metricas: model.MemoryMetrics{Total: 34359738368, Used: 17179869184, UsagePercent: 50}},
	).ServeHTTP(rec, req)

	res := rec.Result()
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		t.Errorf("status = %d, esperado %d", res.StatusCode, http.StatusOK)
	}

	var campos map[string]json.RawMessage
	if err := json.NewDecoder(res.Body).Decode(&campos); err != nil {
		t.Fatalf("resposta não é JSON válido: %v", err)
	}

	for _, campo := range []string{"timestamp", "cpu", "memory", "disk", "uptimeSeconds"} {
		if _, ok := campos[campo]; !ok {
			t.Errorf("campo %q ausente na resposta", campo)
		}
	}
}

// TestHandleMetricsParcialResponde200 fixa a decisão de desenho do epic: um
// snapshot incompleto é resposta legítima, não erro. A UI distingue
// "indisponível" de "zero" pela ausência do campo.
func TestHandleMetricsParcialResponde200(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/api/v1/metrics", nil)
	rec := httptest.NewRecorder()

	routerComMetrics(
		cpuCollectorFalso{err: errors.New("/proc/stat ilegível")},
		memoryCollectorFalso{metricas: model.MemoryMetrics{Total: 34359738368, Used: 17179869184, UsagePercent: 50}},
	).ServeHTTP(rec, req)

	res := rec.Result()
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		t.Errorf("status = %d, esperado %d mesmo com métrica faltando", res.StatusCode, http.StatusOK)
	}

	var campos map[string]json.RawMessage
	if err := json.NewDecoder(res.Body).Decode(&campos); err != nil {
		t.Fatalf("resposta não é JSON válido: %v", err)
	}

	if _, ok := campos["cpu"]; ok {
		t.Error(`campo "cpu" deveria ser omitido quando indisponível`)
	}

	if _, ok := campos["memory"]; !ok {
		t.Error(`campo "memory" deveria continuar presente`)
	}
}
