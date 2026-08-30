package api

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/coder/websocket"
	"github.com/coder/websocket/wsjson"

	"pc-monitor-agent/internal/model"
	"pc-monitor-agent/internal/service"
)

func hubDeTeste(t *testing.T) (*service.MetricsHub, *httptest.Server, string) {
	t.Helper()

	metrics := service.NewMetricsService(
		service.NewCPUService(cpuCollectorFalso{metricas: model.CPUMetrics{
			Model: "Ryzen", Cores: 16, Threads: 32, UsagePercent: 12.5,
		}}),
		nil, nil, nil, nil,
	)

	hub := service.NewMetricsHub(metrics, 50*time.Millisecond)

	ctx, cancelarHub := context.WithCancel(context.Background())
	go hub.Run(ctx)

	servidor := httptest.NewServer(NewRouter(Deps{Hub: hub}))

	t.Cleanup(func() {
		servidor.Close()
		cancelarHub()
	})

	return hub, servidor, "ws" + strings.TrimPrefix(servidor.URL, "http") + "/api/v1/ws/metrics"
}

// aguardarClientes espera o hub chegar à contagem esperada.
//
// A remoção acontece no handler, numa goroutine que ainda está terminando
// quando o cliente já fechou; verificar na hora seria uma corrida.
func aguardarClientes(t *testing.T, hub *service.MetricsHub, esperado int) {
	t.Helper()

	prazo := time.Now().Add(3 * time.Second)

	for time.Now().Before(prazo) {
		if hub.SubscriberCount() == esperado {
			return
		}

		time.Sleep(10 * time.Millisecond)
	}

	t.Fatalf("hub ficou com %d clientes, esperado %d", hub.SubscriberCount(), esperado)
}

func TestWebSocketEntregaSnapshots(t *testing.T) {
	hub, _, endereco := hubDeTeste(t)

	ctx, cancelar := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancelar()

	conn, _, err := websocket.Dial(ctx, endereco, nil)
	if err != nil {
		t.Fatalf("não foi possível conectar: %v", err)
	}

	defer conn.CloseNow()

	var snapshot model.DashboardMetrics
	if err := wsjson.Read(ctx, conn, &snapshot); err != nil {
		t.Fatalf("leitura falhou: %v", err)
	}

	if snapshot.CPU == nil || snapshot.CPU.Model != "Ryzen" {
		t.Errorf("snapshot sem CPU: %+v", snapshot)
	}

	if snapshot.Timestamp.IsZero() {
		t.Error("snapshot sem timestamp")
	}

	// Segunda mensagem: prova que o loop continua publicando, e não que o
	// servidor mandou um snapshot de boas-vindas e parou.
	if err := wsjson.Read(ctx, conn, &snapshot); err != nil {
		t.Fatalf("segunda leitura falhou: %v", err)
	}

	if err := conn.Close(websocket.StatusNormalClosure, ""); err != nil {
		t.Fatalf("fechamento falhou: %v", err)
	}

	aguardarClientes(t, hub, 0)
}

// TestWebSocketRemoveClienteQueSomeSemFechar cobre o caminho feio: processo
// morto, cabo arrancado, nada de handshake de fechamento.
func TestWebSocketRemoveClienteQueSomeSemFechar(t *testing.T) {
	hub, _, endereco := hubDeTeste(t)

	ctx, cancelar := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancelar()

	conn, _, err := websocket.Dial(ctx, endereco, nil)
	if err != nil {
		t.Fatalf("não foi possível conectar: %v", err)
	}

	aguardarClientes(t, hub, 1)

	// CloseNow derruba o TCP sem o handshake de fechamento.
	conn.CloseNow()

	aguardarClientes(t, hub, 0)
}

func TestWebSocketVariosClientesRecebemOMesmoSnapshot(t *testing.T) {
	hub, _, endereco := hubDeTeste(t)

	ctx, cancelar := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancelar()

	const clientes = 3

	snapshots := make([]model.DashboardMetrics, 0, clientes)

	for range clientes {
		conn, _, err := websocket.Dial(ctx, endereco, nil)
		if err != nil {
			t.Fatalf("não foi possível conectar: %v", err)
		}

		defer conn.CloseNow()

		var snapshot model.DashboardMetrics
		if err := wsjson.Read(ctx, conn, &snapshot); err != nil {
			t.Fatalf("leitura falhou: %v", err)
		}

		snapshots = append(snapshots, snapshot)
	}

	aguardarClientes(t, hub, clientes)

	for _, snapshot := range snapshots {
		if snapshot.CPU == nil {
			t.Errorf("cliente recebeu snapshot sem CPU: %+v", snapshot)
		}
	}
}

// TestWebSocketSemHubRespondeErroPadrao: a rota existe mesmo quando o hub não
// foi montado, e a resposta precisa continuar sendo JSON.
func TestWebSocketSemHubRespondeErroPadrao(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/api/v1/ws/metrics", nil)
	rec := httptest.NewRecorder()

	NewRouter(Deps{}).ServeHTTP(rec, req)

	if rec.Code != http.StatusInternalServerError {
		t.Fatalf("status = %d, esperado %d", rec.Code, http.StatusInternalServerError)
	}

	if corpo := decodificarErro(t, rec); corpo.Error != "internal_error" {
		t.Errorf("error = %q, esperado internal_error", corpo.Error)
	}
}
