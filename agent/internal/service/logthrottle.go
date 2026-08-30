package service

import (
	"log/slog"
	"sync"
	"time"
)

// warnThrottleInterval é o intervalo mínimo entre dois registros da mesma
// falha.
const warnThrottleInterval = 30 * time.Second

// warnThrottle limita a frequência com que uma mesma falha é registrada.
//
// O loop do WebSocket coleta a cada segundo. Um collector quebrado — /proc/stat
// ilegível, montagem de rede fora do ar — produziria uma linha de WARN por
// segundo, 3 600 por hora, todas idênticas: o log deixaria de ser útil
// justamente quando fosse mais necessário. A primeira ocorrência é registrada
// na hora, e cada repetição espera o intervalo.
type warnThrottle struct {
	mu         sync.Mutex
	lastAt     map[string]time.Time
	suppressed map[string]int

	interval time.Duration

	// now é injetável para que o teste não dependa do relógio real.
	now func() time.Time
}

func newWarnThrottle(interval time.Duration) *warnThrottle {
	return &warnThrottle{
		lastAt:     make(map[string]time.Time),
		suppressed: make(map[string]int),
		interval:   interval,
		now:        time.Now,
	}
}

// warn registra a mensagem se a chave não tiver sido registrada há pouco.
//
// Quando registra depois de um período silencioso, informa quantas repetições
// foram omitidas — sem isso o log sugeriria uma falha esporádica onde havia
// uma falha contínua.
func (t *warnThrottle) warn(key, message string, args ...any) {
	t.mu.Lock()

	now := t.now()

	if last, seen := t.lastAt[key]; seen && now.Sub(last) < t.interval {
		t.suppressed[key]++
		t.mu.Unlock()

		return
	}

	omitted := t.suppressed[key]
	t.suppressed[key] = 0
	t.lastAt[key] = now

	t.mu.Unlock()

	if omitted > 0 {
		args = append(args, "repeticoesOmitidas", omitted)
	}

	slog.Warn(message, args...)
}
