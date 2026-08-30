package collector

import (
	"context"
	"math"
	"testing"
	"time"

	"pc-monitor-agent/internal/model"
)

const toleranciaTaxa = 1e-9

// TestByteRate cobre o cálculo da taxa isoladamente, incluindo o reinício de
// contador — que numa máquina real exigiria derrubar a interface.
func TestByteRate(t *testing.T) {
	casos := []struct {
		nome     string
		previous uint64
		current  uint64
		segundos float64
		esperado float64
	}{
		{"1 MB em 1 segundo", 0, 1048576, 1, 1048576},
		{"1 MB em 2 segundos", 0, 1048576, 2, 524288},
		{"sem tráfego", 1000, 1000, 1, 0},
		{"intervalo fracionário", 1000, 2000, 0.5, 2000},
		{
			// Sem a comparação antes da subtração, uint64 daria um valor
			// gigante por wraparound em vez de algo detectável como inválido.
			nome:     "contador reiniciado devolve zero",
			previous: 45406232521,
			current:  1024,
			segundos: 1,
			esperado: 0,
		},
	}

	for _, caso := range casos {
		t.Run(caso.nome, func(t *testing.T) {
			obtido := byteRate(caso.previous, caso.current, caso.segundos)

			if math.Abs(obtido-caso.esperado) > toleranciaTaxa {
				t.Errorf("byteRate(%d, %d, %v) = %v, esperado %v",
					caso.previous, caso.current, caso.segundos, obtido, caso.esperado)
			}

			if obtido < 0 {
				t.Errorf("taxa negativa: %v", obtido)
			}
		})
	}
}

// TestByteRateNuncaEstouraComWraparound é o teste que justifica a ordem das
// operações em byteRate. Se a subtração viesse antes da comparação, o
// resultado seria da ordem de 1e19 em vez de 0.
func TestByteRateNuncaEstouraComWraparound(t *testing.T) {
	const anterior = uint64(18446744073709551000) // próximo do máximo de uint64
	const atual = uint64(500)                     // contador reiniciado

	obtido := byteRate(anterior, atual, 1)

	if obtido != 0 {
		t.Errorf("byteRate() = %v, esperado 0; a subtração estourou", obtido)
	}
}

func TestSpeeds(t *testing.T) {
	base := time.Date(2026, 8, 30, 12, 0, 0, 0, time.UTC)

	t.Run("primeira leitura devolve zero", func(t *testing.T) {
		download, upload := speeds(
			counterSnapshot{}, // sem leitura anterior
			counterSnapshot{bytesReceived: 45406232521, bytesSent: 335118123, at: base},
		)

		if download != 0 || upload != 0 {
			t.Errorf("primeira leitura devolveu %v/%v, esperado 0/0", download, upload)
		}
	})

	t.Run("usa o intervalo real e não o nominal", func(t *testing.T) {
		anterior := counterSnapshot{bytesReceived: 1000, bytesSent: 500, at: base}
		// 2,5 s decorridos: um cálculo que assumisse 1 s erraria por 2,5x.
		atual := counterSnapshot{bytesReceived: 3500, bytesSent: 1000, at: base.Add(2500 * time.Millisecond)}

		download, upload := speeds(anterior, atual)

		if math.Abs(download-1000) > toleranciaTaxa {
			t.Errorf("download = %v, esperado 1000", download)
		}

		if math.Abs(upload-200) > toleranciaTaxa {
			t.Errorf("upload = %v, esperado 200", upload)
		}
	})

	t.Run("leituras no mesmo instante devolvem zero", func(t *testing.T) {
		anterior := counterSnapshot{bytesReceived: 1000, at: base}
		atual := counterSnapshot{bytesReceived: 2000, at: base}

		download, upload := speeds(anterior, atual)

		if download != 0 || upload != 0 {
			t.Errorf("intervalo zero devolveu %v/%v, esperado 0/0", download, upload)
		}
	})
}

// TestNetworkCollectorMaquinaReal valida o contrato contra as interfaces reais.
func TestNetworkCollectorMaquinaReal(t *testing.T) {
	coletor := NewNetworkCollector()
	ctx := context.Background()

	interfaces, err := coletor.Collect(ctx)
	if err != nil {
		t.Fatalf("Collect() falhou: %v", err)
	}

	if len(interfaces) == 0 {
		t.Fatal("nenhuma interface detectada")
	}
	loopbacks := loopbackInterfaces(ctx)

	for _, atual := range interfaces {
		if loopbacks[atual.InterfaceName] {
			t.Errorf("loopback %q não deveria aparecer nas métricas", atual.InterfaceName)
		}

		if atual.InterfaceName == "" {
			t.Error("nome de interface vazio")
		}

		verificarTaxasValidas(t, atual)
	}

	// A segunda coleta já tem baseline, então as velocidades passam a ser
	// calculadas — e ainda assim precisam respeitar o contrato.
	segunda, err := coletor.Collect(ctx)
	if err != nil {
		t.Fatalf("segunda coleta falhou: %v", err)
	}

	for _, atual := range segunda {
		verificarTaxasValidas(t, atual)
	}
}

func verificarTaxasValidas(t *testing.T, m model.NetworkMetrics) {
	t.Helper()

	if m.DownloadBytesPerSecond < 0 {
		t.Errorf("%s: download negativo (%v)", m.InterfaceName, m.DownloadBytesPerSecond)
	}

	if m.UploadBytesPerSecond < 0 {
		t.Errorf("%s: upload negativo (%v)", m.InterfaceName, m.UploadBytesPerSecond)
	}

	if math.IsNaN(m.DownloadBytesPerSecond) || math.IsInf(m.DownloadBytesPerSecond, 0) {
		t.Errorf("%s: download não serializável (%v)", m.InterfaceName, m.DownloadBytesPerSecond)
	}

	if math.IsNaN(m.UploadBytesPerSecond) || math.IsInf(m.UploadBytesPerSecond, 0) {
		t.Errorf("%s: upload não serializável (%v)", m.InterfaceName, m.UploadBytesPerSecond)
	}
}

// TestNetworkCollectorCalculaEntreColetas usa um relógio controlado para
// verificar o caminho completo do collector, sem depender de tráfego real.
func TestNetworkCollectorCalculaEntreColetas(t *testing.T) {
	instante := time.Date(2026, 8, 30, 12, 0, 0, 0, time.UTC)

	coletor := &networkCollector{
		previous: make(map[string]counterSnapshot),
		now:      func() time.Time { return instante },
	}

	if _, err := coletor.Collect(context.Background()); err != nil {
		t.Fatalf("primeira coleta falhou: %v", err)
	}

	// Avança o relógio: a segunda coleta calcula sobre 10 s decorridos.
	instante = instante.Add(10 * time.Second)

	interfaces, err := coletor.Collect(context.Background())
	if err != nil {
		t.Fatalf("segunda coleta falhou: %v", err)
	}

	for _, atual := range interfaces {
		verificarTaxasValidas(t, atual)
	}
}
