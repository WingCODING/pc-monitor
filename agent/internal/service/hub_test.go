package service

import (
	"context"
	"sync"
	"testing"
	"time"

	"pc-monitor-agent/internal/model"
)

// cpuCollectorContador conta as coletas, que é como se verifica a promessa
// central do hub: uma coleta por intervalo, não uma por cliente.
type cpuCollectorContador struct {
	mu       sync.Mutex
	chamadas int
}

func (c *cpuCollectorContador) Collect(context.Context) (model.CPUMetrics, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.chamadas++

	return model.CPUMetrics{Model: "Ryzen", Cores: 16, Threads: 32, UsagePercent: 12.5}, nil
}

func (c *cpuCollectorContador) total() int {
	c.mu.Lock()
	defer c.mu.Unlock()

	return c.chamadas
}

func hubDeTeste(contador *cpuCollectorContador, interval time.Duration) *MetricsHub {
	metrics := NewMetricsService(NewCPUService(contador), nil, nil, nil, nil)

	return NewMetricsHub(metrics, interval)
}

func receberEm(t *testing.T, canal <-chan model.DashboardMetrics, prazo time.Duration) (model.DashboardMetrics, bool) {
	t.Helper()

	select {
	case snapshot, aberto := <-canal:
		return snapshot, aberto
	case <-time.After(prazo):
		t.Fatal("nenhum snapshot recebido dentro do prazo")

		return model.DashboardMetrics{}, false
	}
}

// TestMetricsHubUmaColetaAlimentaTodos é o teste que justifica a existência do
// hub: dez janelas abertas não podem virar dez varreduras por segundo.
func TestMetricsHubUmaColetaAlimentaTodos(t *testing.T) {
	contador := &cpuCollectorContador{}
	hub := hubDeTeste(contador, time.Hour)

	primeiro, _ := hub.Subscribe()
	segundo, _ := hub.Subscribe()
	terceiro, _ := hub.Subscribe()

	hub.publish(context.Background())

	for _, canal := range []<-chan model.DashboardMetrics{primeiro, segundo, terceiro} {
		snapshot, aberto := receberEm(t, canal, time.Second)

		if !aberto || snapshot.CPU == nil {
			t.Fatalf("cliente não recebeu o snapshot: %+v", snapshot)
		}
	}

	if total := contador.total(); total != 1 {
		t.Errorf("%d coletas para 3 clientes, esperado 1", total)
	}
}

// TestMetricsHubNaoColetaSemClientes: ler /proc sem ninguém ouvindo é gasto
// puro, e o agente pode ficar aberto o dia inteiro sem nenhuma janela.
func TestMetricsHubNaoColetaSemClientes(t *testing.T) {
	contador := &cpuCollectorContador{}
	hub := hubDeTeste(contador, time.Hour)

	hub.publish(context.Background())

	if total := contador.total(); total != 0 {
		t.Errorf("%d coletas sem clientes conectados, esperado 0", total)
	}
}

// TestMetricsHubClienteLentoNaoTravaOsOutros: o cliente que parou de ler tem a
// sua fila cheia, e o snapshot dele é descartado em vez de bloquear o loop.
func TestMetricsHubClienteLentoNaoTravaOsOutros(t *testing.T) {
	contador := &cpuCollectorContador{}
	hub := hubDeTeste(contador, time.Hour)

	lento, _ := hub.Subscribe()
	rapido, _ := hub.Subscribe()

	concluiu := make(chan struct{})

	go func() {
		defer close(concluiu)

		// Três publicações contra um cliente que nunca lê: a segunda já
		// encontra a fila dele cheia.
		for range 3 {
			hub.publish(context.Background())
		}
	}()

	select {
	case <-concluiu:
	case <-time.After(2 * time.Second):
		t.Fatal("publicação travou por causa de um cliente lento")
	}

	if _, aberto := receberEm(t, rapido, time.Second); !aberto {
		t.Error("cliente rápido deixou de receber")
	}

	// O lento fica com o primeiro snapshot; os outros dois foram descartados.
	if len(lento) != 1 {
		t.Errorf("fila do cliente lento com %d itens, esperado 1", len(lento))
	}
}

func TestMetricsHubUnsubscribeEIdempotente(t *testing.T) {
	hub := hubDeTeste(&cpuCollectorContador{}, time.Hour)

	_, cancelar := hub.Subscribe()

	cancelar()
	cancelar()

	if total := hub.SubscriberCount(); total != 0 {
		t.Errorf("%d clientes após cancelar, esperado 0", total)
	}
}

// TestMetricsHubPublicaAoConectar cobre o acordar do loop: sem ele, o primeiro
// snapshot de uma janela recém-aberta demoraria um intervalo inteiro.
func TestMetricsHubPublicaAoConectar(t *testing.T) {
	hub := hubDeTeste(&cpuCollectorContador{}, time.Hour)

	ctx, cancelar := context.WithCancel(context.Background())
	defer cancelar()

	go hub.Run(ctx)

	snapshots, _ := hub.Subscribe()

	if _, aberto := receberEm(t, snapshots, 2*time.Second); !aberto {
		t.Error("canal fechado em vez de receber o primeiro snapshot")
	}
}

// TestMetricsHubEncerraComOContexto garante que desligar o agente fecha os
// canais — é assim que os handlers de WebSocket sabem que devem sair.
func TestMetricsHubEncerraComOContexto(t *testing.T) {
	hub := hubDeTeste(&cpuCollectorContador{}, time.Hour)

	ctx, cancelar := context.WithCancel(context.Background())

	encerrou := make(chan struct{})

	go func() {
		hub.Run(ctx)
		close(encerrou)
	}()

	snapshots, _ := hub.Subscribe()

	// Consome o snapshot de boas-vindas para que o canal fique vazio.
	receberEm(t, snapshots, 2*time.Second)

	cancelar()

	select {
	case <-encerrou:
	case <-time.After(2 * time.Second):
		t.Fatal("Run não retornou depois do cancelamento")
	}

	if _, aberto := receberEm(t, snapshots, time.Second); aberto {
		t.Error("canal continuou aberto depois do encerramento")
	}

	if total := hub.SubscriberCount(); total != 0 {
		t.Errorf("%d clientes após o encerramento, esperado 0", total)
	}
}

// TestMetricsHubSubscribeDepoisDeEncerrado: uma conexão que chega no meio do
// desligamento recebe um canal já fechado em vez de ficar pendurada.
func TestMetricsHubSubscribeDepoisDeEncerrado(t *testing.T) {
	hub := hubDeTeste(&cpuCollectorContador{}, time.Hour)

	hub.closeAll()

	snapshots, cancelar := hub.Subscribe()
	defer cancelar()

	if _, aberto := <-snapshots; aberto {
		t.Error("canal aberto depois do encerramento do hub")
	}
}
