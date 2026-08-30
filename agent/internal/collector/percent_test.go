package collector

import (
	"math"
	"testing"
)

func TestClampPercent(t *testing.T) {
	casos := []struct {
		nome     string
		entrada  float64
		esperado float64
	}{
		{"valor normal", 42.5, 42.5},
		{"limite inferior", 0, 0},
		{"limite superior", 100, 100},
		{"negativo vira zero", -3.2, 0},
		{"acima de cem satura", 104.7, 100},
	}

	for _, caso := range casos {
		t.Run(caso.nome, func(t *testing.T) {
			if obtido := clampPercent(caso.entrada); obtido != caso.esperado {
				t.Errorf("clampPercent(%v) = %v, esperado %v", caso.entrada, obtido, caso.esperado)
			}
		})
	}
}

func TestPercentOf(t *testing.T) {
	casos := []struct {
		nome     string
		used     uint64
		total    uint64
		esperado float64
	}{
		{"metade", 50, 100, 50},
		{"nada usado", 0, 100, 0},
		{"tudo usado", 100, 100, 100},
		{"total zero não gera NaN", 10, 0, 0},
		{"used acima do total satura", 150, 100, 100},
		{"valores reais de RAM", 17179869184, 34359738368, 50},
	}

	for _, caso := range casos {
		t.Run(caso.nome, func(t *testing.T) {
			obtido := percentOf(caso.used, caso.total)

			if math.IsNaN(obtido) || math.IsInf(obtido, 0) {
				t.Fatalf("percentOf(%d, %d) = %v, valor não serializável em JSON", caso.used, caso.total, obtido)
			}

			if obtido != caso.esperado {
				t.Errorf("percentOf(%d, %d) = %v, esperado %v", caso.used, caso.total, obtido, caso.esperado)
			}
		})
	}
}
