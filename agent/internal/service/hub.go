package service

import (
	"context"
	"sync"
	"time"

	"pc-monitor-agent/internal/model"
)

// subscriberBuffer é o tamanho da fila de cada cliente conectado.
//
// Um lugar basta: se o cliente ainda não consumiu o snapshot anterior, mandar
// um mais novo por cima não ajuda — ele está atrasado, e o que interessa é
// sempre a leitura mais recente.
const subscriberBuffer = 1

// MetricsHub coleta uma vez por intervalo e entrega o mesmo snapshot a todos
// os clientes conectados.
//
// É o que impede que dez janelas abertas virem dez varreduras de /proc por
// segundo — e, pior, que cada uma consuma o baseline de delta da outra nos
// collectors de CPU, rede e processos.
type MetricsHub struct {
	metrics  *MetricsService
	interval time.Duration

	mu          sync.Mutex
	subscribers map[chan model.DashboardMetrics]struct{}
	closed      bool

	// wake pede uma publicação imediata quando alguém se conecta, para que o
	// primeiro snapshot não demore um intervalo inteiro para chegar.
	wake chan struct{}
}

// NewMetricsHub prepara o hub. O loop só roda depois de Run.
func NewMetricsHub(metrics *MetricsService, interval time.Duration) *MetricsHub {
	return &MetricsHub{
		metrics:     metrics,
		interval:    interval,
		subscribers: make(map[chan model.DashboardMetrics]struct{}),
		wake:        make(chan struct{}, 1),
	}
}

// Run publica snapshots até o contexto ser cancelado.
//
// Ao terminar, fecha os canais dos clientes: é assim que os handlers de
// WebSocket sabem que devem encerrar suas conexões, sem que o hub precise
// conhecê-los.
func (h *MetricsHub) Run(ctx context.Context) {
	ticker := time.NewTicker(h.interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			h.closeAll()

			return

		case <-h.wake:
			h.publish(ctx)

		case <-ticker.C:
			h.publish(ctx)
		}
	}
}

// Subscribe registra um cliente e devolve o canal de snapshots junto com a
// função que o remove. A função é idempotente e pode ser chamada com defer.
func (h *MetricsHub) Subscribe() (<-chan model.DashboardMetrics, func()) {
	subscriber := make(chan model.DashboardMetrics, subscriberBuffer)

	h.mu.Lock()

	if h.closed {
		h.mu.Unlock()
		close(subscriber)

		return subscriber, func() {}
	}

	h.subscribers[subscriber] = struct{}{}
	h.mu.Unlock()

	// Não bloqueia: se já existe um pedido de publicação pendente, este entra
	// junto nele.
	select {
	case h.wake <- struct{}{}:
	default:
	}

	return subscriber, func() { h.remove(subscriber) }
}

// SubscriberCount devolve quantos clientes estão conectados.
func (h *MetricsHub) SubscriberCount() int {
	h.mu.Lock()
	defer h.mu.Unlock()

	return len(h.subscribers)
}

// publish coleta uma vez e distribui.
func (h *MetricsHub) publish(ctx context.Context) {
	// Sem ninguém ouvindo não há por que ler /proc.
	if h.SubscriberCount() == 0 {
		return
	}

	// A coleta acontece fora do mutex: ela lê o sistema inteiro e prender o
	// hub durante isso bloquearia quem tenta conectar ou desconectar.
	snapshot := h.metrics.Snapshot(ctx)

	h.mu.Lock()
	defer h.mu.Unlock()

	for subscriber := range h.subscribers {
		select {
		case subscriber <- snapshot:
		default:
			// Cliente lento: descartar este snapshot é melhor do que travar o
			// loop e atrasar todos os outros. O próximo vem no tique seguinte.
		}
	}
}

// remove desconecta um cliente. Chamar duas vezes é seguro — o segundo não
// encontra o canal e não tenta fechá-lo de novo.
func (h *MetricsHub) remove(subscriber chan model.DashboardMetrics) {
	h.mu.Lock()
	defer h.mu.Unlock()

	if _, registered := h.subscribers[subscriber]; !registered {
		return
	}

	delete(h.subscribers, subscriber)
	close(subscriber)
}

// closeAll encerra todos os clientes e impede novas inscrições.
func (h *MetricsHub) closeAll() {
	h.mu.Lock()
	defer h.mu.Unlock()

	h.closed = true

	for subscriber := range h.subscribers {
		delete(h.subscribers, subscriber)
		close(subscriber)
	}
}
