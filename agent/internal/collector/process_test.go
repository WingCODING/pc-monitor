package collector

import (
	"context"
	"math"
	"os"
	"testing"
	"time"

	"pc-monitor-agent/internal/model"
)

// TestSampleCPU cobre o cálculo isolado, incluindo os dois casos que só
// apareceriam numa máquina real depois de muito tempo: PID reaproveitado e
// tempo de CPU regredido.
func TestSampleCPU(t *testing.T) {
	base := time.Date(2026, 8, 30, 12, 0, 0, 0, time.UTC)
	criadoEm := base.Add(-100 * time.Second).UnixMilli()

	casos := []struct {
		nome     string
		previous processSnapshot
		current  processSnapshot
		esperado float64
	}{
		{
			// 20 s de CPU em 100 s de vida: a estimativa disponível antes de
			// existir uma leitura anterior.
			nome:     "primeira observação usa a média desde a criação",
			previous: processSnapshot{},
			current:  processSnapshot{createdAt: criadoEm, cpuSeconds: 20, at: base},
			esperado: 20,
		},
		{
			nome:     "delta entre leituras",
			previous: processSnapshot{createdAt: criadoEm, cpuSeconds: 20, at: base},
			current:  processSnapshot{createdAt: criadoEm, cpuSeconds: 20.5, at: base.Add(time.Second)},
			esperado: 50,
		},
		{
			// Quatro threads saturadas por um segundo. Limitar a 100
			// esconderia justamente o processo que mais pesa.
			nome:     "processo multithread passa de 100",
			previous: processSnapshot{createdAt: criadoEm, cpuSeconds: 10, at: base},
			current:  processSnapshot{createdAt: criadoEm, cpuSeconds: 14, at: base.Add(time.Second)},
			esperado: 400,
		},
		{
			// Sem a checagem de createdAt, o tempo de CPU do processo morto
			// seria comparado com o do novo.
			nome:     "PID reaproveitado recomeça pela média",
			previous: processSnapshot{createdAt: criadoEm, cpuSeconds: 900, at: base},
			current:  processSnapshot{createdAt: base.Add(-10 * time.Second).UnixMilli(), cpuSeconds: 1, at: base},
			esperado: 10,
		},
		{
			nome:     "tempo de CPU regredido devolve zero",
			previous: processSnapshot{createdAt: criadoEm, cpuSeconds: 30, at: base},
			current:  processSnapshot{createdAt: criadoEm, cpuSeconds: 10, at: base.Add(time.Second)},
			esperado: 0,
		},
		{
			// Duas chamadas em sequência caem dentro de um mesmo tick de 10 ms
			// do relógio de CPU; medir aí devolveria 0 % para tudo.
			nome:     "janela curta demais repete a última medição",
			previous: processSnapshot{createdAt: criadoEm, cpuSeconds: 10, at: base, percent: 42},
			current:  processSnapshot{createdAt: criadoEm, cpuSeconds: 10, at: base.Add(20 * time.Millisecond)},
			esperado: 42,
		},
		{
			nome:     "leituras no mesmo instante repetem a última medição",
			previous: processSnapshot{createdAt: criadoEm, cpuSeconds: 10, at: base, percent: 7},
			current:  processSnapshot{createdAt: criadoEm, cpuSeconds: 12, at: base},
			esperado: 7,
		},
		{
			nome:     "sem CreateTime a primeira observação devolve zero",
			previous: processSnapshot{},
			current:  processSnapshot{createdAt: 0, cpuSeconds: 42, at: base},
			esperado: 0,
		},
	}

	for _, caso := range casos {
		t.Run(caso.nome, func(t *testing.T) {
			obtido, guardado := sampleCPU(caso.previous, caso.current)

			if math.Abs(obtido-caso.esperado) > 1e-9 {
				t.Errorf("sampleCPU() = %v, esperado %v", obtido, caso.esperado)
			}

			if obtido < 0 {
				t.Errorf("percentual negativo: %v", obtido)
			}

			if math.Abs(guardado.percent-obtido) > 1e-9 {
				t.Errorf("snapshot guardado com percent %v, esperado %v", guardado.percent, obtido)
			}
		})
	}
}

// TestSampleCPUPreservaBaseEmJanelaCurta prova a outra metade da regra da
// janela mínima: além de repetir a medição, a base antiga continua guardada,
// para que a próxima coleta meça sobre um intervalo utilizável em vez de
// recomeçar de uma base recém-gravada.
func TestSampleCPUPreservaBaseEmJanelaCurta(t *testing.T) {
	base := time.Date(2026, 8, 30, 12, 0, 0, 0, time.UTC)
	criadoEm := base.Add(-time.Minute).UnixMilli()

	anterior := processSnapshot{createdAt: criadoEm, cpuSeconds: 10, at: base, percent: 42}
	atual := processSnapshot{createdAt: criadoEm, cpuSeconds: 10.001, at: base.Add(50 * time.Millisecond)}

	_, guardado := sampleCPU(anterior, atual)

	if !guardado.at.Equal(anterior.at) {
		t.Errorf("base movida para %v; deveria continuar em %v", guardado.at, anterior.at)
	}

	// Meio segundo depois da base antiga a medição volta a acontecer.
	depois := processSnapshot{createdAt: criadoEm, cpuSeconds: 10.5, at: base.Add(500 * time.Millisecond)}

	percent, _ := sampleCPU(guardado, depois)
	if math.Abs(percent-100) > 1e-9 {
		t.Errorf("percentual = %v, esperado 100", percent)
	}
}

// TestProcessCollectorMaquinaReal valida o contrato contra os processos reais.
func TestProcessCollectorMaquinaReal(t *testing.T) {
	coletor := NewProcessCollector()

	processos, err := coletor.Collect(context.Background())
	if err != nil {
		t.Fatalf("Collect() falhou: %v", err)
	}

	if len(processos) == 0 {
		t.Fatal("nenhum processo detectado")
	}

	for _, atual := range processos {
		if atual.PID <= 0 {
			t.Errorf("PID inválido: %d", atual.PID)
		}

		if atual.Name == "" {
			t.Errorf("processo %d sem nome", atual.PID)
		}

		if atual.CPUPercent < 0 || math.IsNaN(atual.CPUPercent) || math.IsInf(atual.CPUPercent, 0) {
			t.Errorf("processo %d: cpuPercent não serializável (%v)", atual.PID, atual.CPUPercent)
		}
	}

	// O próprio processo de teste tem de aparecer na lista.
	if !contemPID(processos, int32(os.Getpid())) {
		t.Error("o processo de teste não apareceu na própria coleta")
	}
}

// TestProcessCollectorNaoBloqueia protege o requisito de que o endpoint não
// pode travar a API. O limite é generoso porque a máquina pode estar sob carga.
func TestProcessCollectorNaoBloqueia(t *testing.T) {
	coletor := NewProcessCollector()

	inicio := time.Now()
	if _, err := coletor.Collect(context.Background()); err != nil {
		t.Fatalf("Collect() falhou: %v", err)
	}

	if decorrido := time.Since(inicio); decorrido > 3*time.Second {
		t.Errorf("coleta levou %v; a varredura de /proc está lenta demais para uma requisição HTTP", decorrido)
	}
}

// TestProcessCollectorDescartaProcessosMortos verifica que o estado interno
// não cresce indefinidamente: o mapa é substituído a cada coleta, então só
// guarda os processos vistos na última.
func TestProcessCollectorDescartaProcessosMortos(t *testing.T) {
	coletor := &processCollector{
		previous: map[int32]processSnapshot{
			// Um PID que não existe mais, sobrevivente de uma coleta anterior.
			999999: {createdAt: 1, cpuSeconds: 10, at: time.Now()},
		},
		now: time.Now,
	}

	processos, err := coletor.Collect(context.Background())
	if err != nil {
		t.Fatalf("Collect() falhou: %v", err)
	}

	if _, ainda := coletor.previous[999999]; ainda {
		t.Error("processo morto continuou no estado interno")
	}

	if len(coletor.previous) != len(processos) {
		t.Errorf("estado interno com %d entradas para %d processos coletados",
			len(coletor.previous), len(processos))
	}
}

// TestProcessCollectorRespeitaCancelamento garante que a varredura desiste
// quando o cliente desiste.
func TestProcessCollectorRespeitaCancelamento(t *testing.T) {
	ctx, cancelar := context.WithCancel(context.Background())
	cancelar()

	if _, err := NewProcessCollector().Collect(ctx); err == nil {
		t.Error("Collect() com contexto cancelado deveria falhar")
	}
}

func contemPID(processos []model.ProcessMetrics, pid int32) bool {
	for _, atual := range processos {
		if atual.PID == pid {
			return true
		}
	}

	return false
}
