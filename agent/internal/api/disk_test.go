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

type diskCollectorFalso struct {
	discos []model.DiskMetrics
	err    error
}

func (f diskCollectorFalso) Collect(context.Context) ([]model.DiskMetrics, error) {
	return f.discos, f.err
}

func routerComDiscos(c diskCollectorFalso) http.Handler {
	return NewRouter(Deps{Disk: service.NewDiskService(c)})
}

func TestHandleDisksSucesso(t *testing.T) {
	esperado := []model.DiskMetrics{
		{Name: "/dev/dm-0", MountPoint: "/", Total: 1000, Used: 400, Free: 600, UsagePercent: 40},
		{Name: "/dev/nvme0n1p1", MountPoint: "/boot", Total: 500, Used: 100, Free: 400, UsagePercent: 20},
	}

	req := httptest.NewRequest(http.MethodGet, "/api/v1/disks", nil)
	rec := httptest.NewRecorder()

	routerComDiscos(diskCollectorFalso{discos: esperado}).ServeHTTP(rec, req)

	res := rec.Result()
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		t.Errorf("status = %d, esperado %d", res.StatusCode, http.StatusOK)
	}

	var obtido []model.DiskMetrics
	if err := json.NewDecoder(res.Body).Decode(&obtido); err != nil {
		t.Fatalf("resposta não é JSON válido: %v", err)
	}

	if len(obtido) != len(esperado) {
		t.Fatalf("len = %d, esperado %d", len(obtido), len(esperado))
	}

	for i := range obtido {
		if obtido[i] != esperado[i] {
			t.Errorf("posição %d = %+v, esperado %+v", i, obtido[i], esperado[i])
		}
	}
}

// TestHandleDisksListaVaziaEArray garante que a ausência de partições vira
// [] e não null: um cliente que itere sobre null quebra.
func TestHandleDisksListaVaziaEArray(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/api/v1/disks", nil)
	rec := httptest.NewRecorder()

	routerComDiscos(diskCollectorFalso{discos: nil}).ServeHTTP(rec, req)

	if corpo := rec.Body.String(); corpo != "[]" {
		t.Errorf("corpo = %s, esperado []", corpo)
	}
}

func TestHandleDisksIndisponivel(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/api/v1/disks", nil)
	rec := httptest.NewRecorder()

	routerComDiscos(diskCollectorFalso{err: errors.New("mountinfo ilegível")}).ServeHTTP(rec, req)

	if rec.Code != http.StatusServiceUnavailable {
		t.Errorf("status = %d, esperado %d", rec.Code, http.StatusServiceUnavailable)
	}
}
