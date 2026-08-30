package collector

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/shirou/gopsutil/v4/net"

	"pc-monitor-agent/internal/model"
)

// counterSnapshot guarda os contadores de uma interface e o instante da
// leitura. O instante é essencial: a velocidade precisa do intervalo real
// decorrido, não do intervalo nominal de coleta.
type counterSnapshot struct {
	bytesReceived uint64
	bytesSent     uint64
	at            time.Time
}

// networkCollector deriva velocidades da diferença entre leituras sucessivas
// dos contadores acumulados de cada interface.
type networkCollector struct {
	mu       sync.Mutex
	previous map[string]counterSnapshot

	// now é injetável para que os testes controlem o tempo decorrido.
	now func() time.Time
}

// NewNetworkCollector devolve um NetworkCollector baseado no gopsutil.
func NewNetworkCollector() NetworkCollector {
	return &networkCollector{
		previous: make(map[string]counterSnapshot),
		now:      time.Now,
	}
}

func (c *networkCollector) Collect(ctx context.Context) ([]model.NetworkMetrics, error) {
	counters, err := net.IOCountersWithContext(ctx, true)
	if err != nil {
		return nil, fmt.Errorf("contadores de rede: %w", err)
	}

	loopbacks := loopbackInterfaces(ctx)

	readAt := c.now()

	c.mu.Lock()
	defer c.mu.Unlock()

	metrics := make([]model.NetworkMetrics, 0, len(counters))

	for _, counter := range counters {
		if loopbacks[counter.Name] {
			continue
		}

		current := counterSnapshot{
			bytesReceived: counter.BytesRecv,
			bytesSent:     counter.BytesSent,
			at:            readAt,
		}

		download, upload := speeds(c.previous[counter.Name], current)
		c.previous[counter.Name] = current

		metrics = append(metrics, model.NetworkMetrics{
			InterfaceName:          counter.Name,
			BytesReceived:          counter.BytesRecv,
			BytesSent:              counter.BytesSent,
			DownloadBytesPerSecond: download,
			UploadBytesPerSecond:   upload,
		})
	}

	return metrics, nil
}

// loopbackInterfaces identifica as interfaces pelo sinalizador fornecido pelo
// sistema operacional. Comparar apenas com "lo" funcionava no Linux, mas
// incluía a interface de loopback do Windows nas métricas do dashboard.
func loopbackInterfaces(ctx context.Context) map[string]bool {
	result := map[string]bool{"lo": true}

	interfaces, err := net.InterfacesWithContext(ctx)
	if err != nil {
		return result
	}

	for _, iface := range interfaces {
		for _, flag := range iface.Flags {
			if flag == "loopback" {
				result[iface.Name] = true
				break
			}
		}
	}

	return result
}

// speeds calcula download e upload em bytes por segundo entre duas leituras.
//
// A leitura anterior zerada — primeira coleta da interface — devolve
// velocidade zero: não há intervalo sobre o qual calcular, e inventar um valor
// a partir do contador acumulado reportaria a média desde o boot como se fosse
// a velocidade instantânea.
func speeds(previous, current counterSnapshot) (download, upload float64) {
	if previous.at.IsZero() {
		return 0, 0
	}

	elapsed := current.at.Sub(previous.at).Seconds()
	if elapsed <= 0 {
		return 0, 0
	}

	return byteRate(previous.bytesReceived, current.bytesReceived, elapsed),
		byteRate(previous.bytesSent, current.bytesSent, elapsed)
}

// byteRate devolve a taxa entre dois contadores acumulados.
//
// A comparação precede a subtração de propósito: os contadores são uint64, e
// current - previous com o contador reiniciado não produz um número negativo
// detectável — produz um valor gigante por wraparound. Reinícios acontecem ao
// derrubar e subir a interface, ou ao recarregar o driver.
func byteRate(previous, current uint64, elapsedSeconds float64) float64 {
	if current < previous {
		return 0
	}

	return float64(current-previous) / elapsedSeconds
}
