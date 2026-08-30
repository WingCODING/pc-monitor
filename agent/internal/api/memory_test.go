package api

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"pc-monitor-agent/internal/model"
	"pc-monitor-agent/internal/service"
)

type memoryCollectorFalso struct {
	metricas model.MemoryMetrics
	err      error
}

func (f memoryCollectorFalso) Collect(context.Context) (model.MemoryMetrics, error) {
	return f.metricas, f.err
}

func routerComMemoria(c memoryCollectorFalso) http.Handler {
	return NewRouter(Deps{Memory: service.NewMemoryService(c)})
}

func TestHandleMemorySucesso(t *testing.T) {
	esperado := model.MemoryMetrics{
		Total:        34359738368,
		Used:         17179869184,
		Available:    17179869184,
		UsagePercent: 50,
	}

	req := httptest.NewRequest(http.MethodGet, "/api/v1/memory", nil)
	rec := httptest.NewRecorder()

	routerComMemoria(memoryCollectorFalso{metricas: esperado}).ServeHTTP(rec, req)

	res := rec.Result()
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		t.Errorf("status = %d, esperado %d", res.StatusCode, http.StatusOK)
	}

	if got := res.Header.Get("Content-Type"); got != contentTypeJSON {
		t.Errorf("Content-Type = %q, esperado %q", got, contentTypeJSON)
	}

	var obtido model.MemoryMetrics
	if err := json.NewDecoder(res.Body).Decode(&obtido); err != nil {
		t.Fatalf("resposta não é JSON válido: %v", err)
	}

	if obtido != esperado {
		t.Errorf("corpo = %+v, esperado %+v", obtido, esperado)
	}
}

func TestHandleMemoryIndisponivel(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/api/v1/memory", nil)
	rec := httptest.NewRecorder()

	routerComMemoria(memoryCollectorFalso{err: errors.New("/proc/meminfo ilegível")}).ServeHTTP(rec, req)

	if rec.Code != http.StatusServiceUnavailable {
		t.Errorf("status = %d, esperado %d", rec.Code, http.StatusServiceUnavailable)
	}
}
