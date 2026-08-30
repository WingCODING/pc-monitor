package api

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"pc-monitor-agent/internal/model"
	"pc-monitor-agent/internal/service"
)

type cpuCollectorFalso struct {
	metricas model.CPUMetrics
	err      error
}

func (f cpuCollectorFalso) Collect(context.Context) (model.CPUMetrics, error) {
	return f.metricas, f.err
}

func routerComCPU(c cpuCollectorFalso) http.Handler {
	return NewRouter(Deps{CPU: service.NewCPUService(c)})
}

func TestHandleCPUSucesso(t *testing.T) {
	esperado := model.CPUMetrics{
		Model:        "AMD Ryzen 7 7800X3D",
		Cores:        8,
		Threads:      16,
		UsagePercent: 24.8,
	}

	req := httptest.NewRequest(http.MethodGet, "/api/v1/cpu", nil)
	rec := httptest.NewRecorder()

	routerComCPU(cpuCollectorFalso{metricas: esperado}).ServeHTTP(rec, req)

	res := rec.Result()
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		t.Errorf("status = %d, esperado %d", res.StatusCode, http.StatusOK)
	}

	if got := res.Header.Get("Content-Type"); got != contentTypeJSON {
		t.Errorf("Content-Type = %q, esperado %q", got, contentTypeJSON)
	}

	var obtido model.CPUMetrics
	if err := json.NewDecoder(res.Body).Decode(&obtido); err != nil {
		t.Fatalf("resposta não é JSON válido: %v", err)
	}

	if obtido != esperado {
		t.Errorf("corpo = %+v, esperado %+v", obtido, esperado)
	}
}

// TestHandleCPUIndisponivel garante que uma falha de collector vira 503, e não
// 500: o cliente precisa saber que vale tentar de novo.
func TestHandleCPUIndisponivel(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/api/v1/cpu", nil)
	rec := httptest.NewRecorder()

	routerComCPU(cpuCollectorFalso{err: errors.New("/proc/stat ilegível")}).ServeHTTP(rec, req)

	res := rec.Result()
	defer res.Body.Close()

	if res.StatusCode != http.StatusServiceUnavailable {
		t.Errorf("status = %d, esperado %d", res.StatusCode, http.StatusServiceUnavailable)
	}

	var corpo errorResponse
	if err := json.NewDecoder(res.Body).Decode(&corpo); err != nil {
		t.Fatalf("resposta de erro não é JSON válido: %v", err)
	}

	if corpo.Error != "collector_unavailable" {
		t.Errorf("error = %q, esperado %q", corpo.Error, "collector_unavailable")
	}

	if corpo.Message == "" {
		t.Error("mensagem de erro vazia")
	}
}

// TestHandleCPUNaoVazaDetalheInterno protege o contrato de segurança: a causa
// real é registrada no log, nunca devolvida ao cliente.
func TestHandleCPUNaoVazaDetalheInterno(t *testing.T) {
	const detalheInterno = "/proc/stat ilegível"

	req := httptest.NewRequest(http.MethodGet, "/api/v1/cpu", nil)
	rec := httptest.NewRecorder()

	routerComCPU(cpuCollectorFalso{err: errors.New(detalheInterno)}).ServeHTTP(rec, req)

	if corpo := rec.Body.String(); strings.Contains(corpo, detalheInterno) {
		t.Errorf("detalhe interno vazou na resposta: %s", corpo)
	}
}
