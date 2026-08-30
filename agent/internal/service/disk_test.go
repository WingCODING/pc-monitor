package service

import (
	"context"
	"errors"
	"math"
	"testing"

	"pc-monitor-agent/internal/model"
)

// toleranciaPercentual absorve a diferença de arredondamento entre a
// aritmética de constantes do teste e a de float64 em tempo de execução.
const toleranciaPercentual = 1e-9

type diskCollectorFalso struct {
	discos []model.DiskMetrics
	err    error
}

func (f diskCollectorFalso) Collect(context.Context) ([]model.DiskMetrics, error) {
	return f.discos, f.err
}

func TestDiskServiceList(t *testing.T) {
	esperado := []model.DiskMetrics{
		{Name: "/dev/dm-0", MountPoint: "/", Total: 1000, Used: 400, Free: 600, UsagePercent: 40},
	}

	obtido, err := NewDiskService(diskCollectorFalso{discos: esperado}).List(context.Background())
	if err != nil {
		t.Fatalf("List() falhou: %v", err)
	}

	if len(obtido) != 1 || obtido[0] != esperado[0] {
		t.Errorf("List() = %+v, esperado %+v", obtido, esperado)
	}
}

func TestDiskServiceListConverteErro(t *testing.T) {
	falhaOriginal := errors.New("mountinfo ilegível")

	_, err := NewDiskService(diskCollectorFalso{err: falhaOriginal}).List(context.Background())
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

func TestSummarizeDisks(t *testing.T) {
	casos := []struct {
		nome     string
		entrada  []model.DiskMetrics
		esperado model.DiskSummary
	}{
		{
			nome: "soma partições distintas",
			entrada: []model.DiskMetrics{
				{Name: "/dev/dm-0", Total: 1000, Used: 400},
				{Name: "/dev/nvme0n1p1", Total: 500, Used: 100},
			},
			esperado: model.DiskSummary{Total: 1500, Used: 500, UsagePercent: 500.0 / 1500.0 * 100},
		},
		{
			nome:     "lista vazia não gera NaN",
			entrada:  nil,
			esperado: model.DiskSummary{Total: 0, Used: 0, UsagePercent: 0},
		},
		{
			nome: "partição única",
			entrada: []model.DiskMetrics{
				{Name: "/dev/dm-0", Total: 1000, Used: 250},
			},
			esperado: model.DiskSummary{Total: 1000, Used: 250, UsagePercent: 25},
		},
	}

	for _, caso := range casos {
		t.Run(caso.nome, func(t *testing.T) {
			obtido := summarizeDisks(caso.entrada)

			if obtido.Total != caso.esperado.Total || obtido.Used != caso.esperado.Used {
				t.Errorf("summarizeDisks() = %+v, esperado %+v", obtido, caso.esperado)
			}

			// Comparação com tolerância: a expressão do valor esperado é
			// avaliada como constante, em precisão arbitrária e com um único
			// arredondamento, enquanto a função arredonda a cada operação em
			// float64. Exigir igualdade exata falharia por 1 ulp.
			if math.Abs(obtido.UsagePercent-caso.esperado.UsagePercent) > toleranciaPercentual {
				t.Errorf("UsagePercent = %v, esperado %v", obtido.UsagePercent, caso.esperado.UsagePercent)
			}
		})
	}
}

// TestSummarizeDisksNaoInflaComSubvolumes documenta por que a deduplicação
// acontece no collector e não aqui: a soma assume entradas já únicas.
func TestSummarizeDisksNaoInflaComSubvolumes(t *testing.T) {
	const capacidade = 998037782528

	// Como o collector deduplica antes, o resumo recebe uma única entrada.
	resumo := summarizeDisks([]model.DiskMetrics{
		{Name: "/dev/dm-0", MountPoint: "/", Total: capacidade, Used: 382323040256},
	})

	if resumo.Total != capacidade {
		t.Errorf("Total = %d, esperado %d", resumo.Total, capacidade)
	}

	if resumo.Total > capacidade {
		t.Error("capacidade inflada; a deduplicação do collector não foi respeitada")
	}
}
