package service

import (
	"context"
	"errors"
	"math"
	"testing"

	"pc-monitor-agent/internal/model"
)

type networkCollectorFalso struct {
	interfaces []model.NetworkMetrics
	err        error
}

func (f networkCollectorFalso) Collect(context.Context) ([]model.NetworkMetrics, error) {
	return f.interfaces, f.err
}

func TestNetworkServiceListConverteErro(t *testing.T) {
	falhaOriginal := errors.New("/proc/net/dev ilegível")

	_, err := NewNetworkService(networkCollectorFalso{err: falhaOriginal}).List(context.Background())
	if err == nil {
		t.Fatal("List() deveria falhar quando o collector falha")
	}

	if !errors.Is(err, ErrUnavailable) {
		t.Errorf("erro não classificado como ErrUnavailable: %v", err)
	}

	if !errors.Is(err, falhaOriginal) {
		t.Errorf("causa original perdida no encadeamento: %v", err)
	}
}

// TestSummarizeNetwork verifica a soma entre interfaces. Ao contrário dos
// discos, somar aqui é seguro: cada interface tem contadores próprios.
func TestSummarizeNetwork(t *testing.T) {
	casos := []struct {
		nome     string
		entrada  []model.NetworkMetrics
		download float64
		upload   float64
	}{
		{
			nome: "soma múltiplas interfaces",
			entrada: []model.NetworkMetrics{
				{InterfaceName: "enp13s0", DownloadBytesPerSecond: 5242880, UploadBytesPerSecond: 1048576},
				{InterfaceName: "proton0", DownloadBytesPerSecond: 1024, UploadBytesPerSecond: 512},
			},
			download: 5243904,
			upload:   1049088,
		},
		{
			nome:     "lista vazia",
			entrada:  nil,
			download: 0,
			upload:   0,
		},
		{
			nome: "interface única",
			entrada: []model.NetworkMetrics{
				{InterfaceName: "enp13s0", DownloadBytesPerSecond: 100, UploadBytesPerSecond: 50},
			},
			download: 100,
			upload:   50,
		},
	}

	for _, caso := range casos {
		t.Run(caso.nome, func(t *testing.T) {
			obtido := summarizeNetwork(caso.entrada)

			if math.Abs(obtido.DownloadBytesPerSecond-caso.download) > toleranciaPercentual {
				t.Errorf("download = %v, esperado %v", obtido.DownloadBytesPerSecond, caso.download)
			}

			if math.Abs(obtido.UploadBytesPerSecond-caso.upload) > toleranciaPercentual {
				t.Errorf("upload = %v, esperado %v", obtido.UploadBytesPerSecond, caso.upload)
			}
		})
	}
}
