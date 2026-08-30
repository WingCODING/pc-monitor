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

type networkCollectorFalso struct {
	interfaces []model.NetworkMetrics
	err        error
}

func (f networkCollectorFalso) Collect(context.Context) ([]model.NetworkMetrics, error) {
	return f.interfaces, f.err
}

func routerComRede(c networkCollectorFalso) http.Handler {
	return NewRouter(Deps{Network: service.NewNetworkService(c)})
}

func TestHandleNetworkSucesso(t *testing.T) {
	esperado := []model.NetworkMetrics{
		{
			InterfaceName:          "enp13s0",
			BytesReceived:          45406232521,
			BytesSent:              335118123,
			DownloadBytesPerSecond: 5242880,
			UploadBytesPerSecond:   1048576,
		},
	}

	req := httptest.NewRequest(http.MethodGet, "/api/v1/network", nil)
	rec := httptest.NewRecorder()

	routerComRede(networkCollectorFalso{interfaces: esperado}).ServeHTTP(rec, req)

	res := rec.Result()
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		t.Errorf("status = %d, esperado %d", res.StatusCode, http.StatusOK)
	}

	var obtido []model.NetworkMetrics
	if err := json.NewDecoder(res.Body).Decode(&obtido); err != nil {
		t.Fatalf("resposta não é JSON válido: %v", err)
	}

	if len(obtido) != 1 || obtido[0] != esperado[0] {
		t.Errorf("corpo = %+v, esperado %+v", obtido, esperado)
	}
}

func TestHandleNetworkListaVaziaEArray(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/api/v1/network", nil)
	rec := httptest.NewRecorder()

	routerComRede(networkCollectorFalso{interfaces: nil}).ServeHTTP(rec, req)

	if corpo := rec.Body.String(); corpo != "[]" {
		t.Errorf("corpo = %s, esperado []", corpo)
	}
}

func TestHandleNetworkIndisponivel(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/api/v1/network", nil)
	rec := httptest.NewRecorder()

	routerComRede(networkCollectorFalso{err: errors.New("/proc/net/dev ilegível")}).ServeHTTP(rec, req)

	if rec.Code != http.StatusServiceUnavailable {
		t.Errorf("status = %d, esperado %d", rec.Code, http.StatusServiceUnavailable)
	}
}
