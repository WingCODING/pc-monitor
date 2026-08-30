package api

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"pc-monitor-agent/internal/model"
	"pc-monitor-agent/internal/service"
)

type processCollectorFalso struct {
	processos []model.ProcessMetrics
	err       error
}

func (f processCollectorFalso) Collect(context.Context) ([]model.ProcessMetrics, error) {
	return f.processos, f.err
}

func routerComProcessos(c processCollectorFalso) http.Handler {
	return NewRouter(Deps{Process: service.NewProcessService(c)})
}

func pedirProcessos(t *testing.T, c processCollectorFalso, alvo string) *httptest.ResponseRecorder {
	t.Helper()

	req := httptest.NewRequest(http.MethodGet, alvo, nil)
	rec := httptest.NewRecorder()

	routerComProcessos(c).ServeHTTP(rec, req)

	return rec
}

func decodificarProcessos(t *testing.T, rec *httptest.ResponseRecorder) []model.ProcessMetrics {
	t.Helper()

	var processos []model.ProcessMetrics
	if err := json.NewDecoder(rec.Body).Decode(&processos); err != nil {
		t.Fatalf("resposta não é JSON válido: %v", err)
	}

	return processos
}

func amostraDeProcessos() []model.ProcessMetrics {
	return []model.ProcessMetrics{
		{PID: 1, Name: "systemd", CPUPercent: 0.4, Memory: 12 << 20},
		{PID: 900, Name: "java", CPUPercent: 180, Memory: 1 << 30},
		{PID: 4321, Name: "firefox", CPUPercent: 12.5, Memory: 2 << 30},
	}
}

// TestHandleProcessesOrdenacaoPadrao fixa o padrão do endpoint: maior uso de
// CPU primeiro, que é o que um monitor mostra ao abrir.
func TestHandleProcessesOrdenacaoPadrao(t *testing.T) {
	rec := pedirProcessos(t, processCollectorFalso{processos: amostraDeProcessos()}, "/api/v1/processes")

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, esperado %d", rec.Code, http.StatusOK)
	}

	obtido := decodificarProcessos(t, rec)

	if len(obtido) != 3 || obtido[0].Name != "java" || obtido[1].Name != "firefox" {
		t.Errorf("ordem = %+v, esperado java, firefox, systemd", obtido)
	}
}

func TestHandleProcessesOrdenaPorMemoria(t *testing.T) {
	rec := pedirProcessos(t, processCollectorFalso{processos: amostraDeProcessos()}, "/api/v1/processes?sort=memory")

	obtido := decodificarProcessos(t, rec)

	if len(obtido) != 3 || obtido[0].Name != "firefox" {
		t.Errorf("ordem = %+v, esperado firefox primeiro", obtido)
	}
}

func TestHandleProcessesAplicaLimite(t *testing.T) {
	rec := pedirProcessos(t, processCollectorFalso{processos: amostraDeProcessos()}, "/api/v1/processes?limit=1")

	if obtido := decodificarProcessos(t, rec); len(obtido) != 1 {
		t.Errorf("len = %d, esperado 1", len(obtido))
	}
}

// TestHandleProcessesLimitePadrao garante que uma máquina com centenas de
// processos não devolve todos por omissão.
func TestHandleProcessesLimitePadrao(t *testing.T) {
	muitos := make([]model.ProcessMetrics, 0, 300)
	for i := range 300 {
		muitos = append(muitos, model.ProcessMetrics{
			PID:        int32(i + 1),
			Name:       fmt.Sprintf("processo-%d", i),
			CPUPercent: float64(i),
		})
	}

	rec := pedirProcessos(t, processCollectorFalso{processos: muitos}, "/api/v1/processes")

	if obtido := decodificarProcessos(t, rec); len(obtido) != service.DefaultProcessLimit {
		t.Errorf("len = %d, esperado %d", len(obtido), service.DefaultProcessLimit)
	}
}

// TestHandleProcessesParametroInvalido cobre o 400: parâmetro fora do domínio
// é erro do cliente, não métrica indisponível — repetir não adianta.
func TestHandleProcessesParametroInvalido(t *testing.T) {
	casos := []string{
		"/api/v1/processes?sort=disco",
		"/api/v1/processes?limit=0",
		"/api/v1/processes?limit=abc",
		"/api/v1/processes?limit=501",
		"/api/v1/processes?limit=-3",
	}

	for _, alvo := range casos {
		t.Run(alvo, func(t *testing.T) {
			rec := pedirProcessos(t, processCollectorFalso{processos: amostraDeProcessos()}, alvo)

			if rec.Code != http.StatusBadRequest {
				t.Fatalf("status = %d, esperado %d", rec.Code, http.StatusBadRequest)
			}

			var corpo errorResponse
			if err := json.NewDecoder(rec.Body).Decode(&corpo); err != nil {
				t.Fatalf("erro não é JSON válido: %v", err)
			}

			if corpo.Error != "invalid_request" {
				t.Errorf("error = %q, esperado invalid_request", corpo.Error)
			}
		})
	}
}

func TestHandleProcessesListaVaziaEArray(t *testing.T) {
	rec := pedirProcessos(t, processCollectorFalso{processos: nil}, "/api/v1/processes")

	if corpo := rec.Body.String(); corpo != "[]" {
		t.Errorf("corpo = %s, esperado []", corpo)
	}
}

func TestHandleProcessesIndisponivel(t *testing.T) {
	rec := pedirProcessos(t, processCollectorFalso{err: errors.New("/proc ilegível")}, "/api/v1/processes")

	if rec.Code != http.StatusServiceUnavailable {
		t.Errorf("status = %d, esperado %d", rec.Code, http.StatusServiceUnavailable)
	}
}
