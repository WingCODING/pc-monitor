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

type systemCollectorFalso struct {
	metricas model.SystemMetrics
	err      error
}

func (f systemCollectorFalso) Collect(context.Context) (model.SystemMetrics, error) {
	return f.metricas, f.err
}

func routerComSistema(c systemCollectorFalso) http.Handler {
	return NewRouter(Deps{System: service.NewSystemService(c)})
}

func TestHandleSystemSucesso(t *testing.T) {
	esperado := model.SystemMetrics{
		Hostname:      "desktop",
		OS:            "linux",
		OSVersion:     "4.0.1",
		Architecture:  "amd64",
		UptimeSeconds: 58232,
	}

	req := httptest.NewRequest(http.MethodGet, "/api/v1/system", nil)
	rec := httptest.NewRecorder()

	routerComSistema(systemCollectorFalso{metricas: esperado}).ServeHTTP(rec, req)

	res := rec.Result()
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		t.Errorf("status = %d, esperado %d", res.StatusCode, http.StatusOK)
	}

	var obtido model.SystemMetrics
	if err := json.NewDecoder(res.Body).Decode(&obtido); err != nil {
		t.Fatalf("resposta não é JSON válido: %v", err)
	}

	if obtido != esperado {
		t.Errorf("corpo = %+v, esperado %+v", obtido, esperado)
	}
}

func TestHandleSystemIndisponivel(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/api/v1/system", nil)
	rec := httptest.NewRecorder()

	routerComSistema(systemCollectorFalso{err: errors.New("host inacessível")}).ServeHTTP(rec, req)

	if rec.Code != http.StatusServiceUnavailable {
		t.Errorf("status = %d, esperado %d", rec.Code, http.StatusServiceUnavailable)
	}
}
