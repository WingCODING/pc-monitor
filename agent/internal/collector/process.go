package collector

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/shirou/gopsutil/v4/process"

	"pc-monitor-agent/internal/model"
)

// processSnapshot guarda o tempo de CPU já consumido por um processo e o
// instante da leitura.
//
// createdAt identifica a encarnação do PID: o sistema reaproveita números de
// processo, e sem essa checagem o tempo de CPU de um processo morto seria
// comparado com o de outro recém-criado, produzindo um percentual negativo ou
// absurdo na primeira leitura do processo novo.
type processSnapshot struct {
	createdAt  int64
	cpuSeconds float64
	at         time.Time

	// percent guarda a última medição feita sobre uma janela válida, para que
	// coletas em sequência rápida repitam esse valor em vez de reportarem
	// ruído de quantização.
	percent float64
}

// minSampleWindow é o intervalo mínimo entre duas amostras que ainda produz um
// percentual confiável.
//
// O tempo de CPU em /proc avança em ticks de 10 ms. Medir sobre uma janela de
// 20 ms devolve 0 % ou 50 % conforme o tick caia dentro ou fora dela — e o
// dashboard consulta o endpoint quando quer, não em cadência fixa.
const minSampleWindow = 500 * time.Millisecond

// processCollector lista os processos em execução e deriva o uso de CPU da
// diferença entre leituras consecutivas, como o collector de rede faz com os
// contadores de bytes.
type processCollector struct {
	mu       sync.Mutex
	previous map[int32]processSnapshot

	// now é injetável para que os testes controlem o tempo decorrido.
	now func() time.Time
}

// NewProcessCollector devolve um ProcessCollector baseado no gopsutil.
func NewProcessCollector() ProcessCollector {
	return &processCollector{
		previous: make(map[int32]processSnapshot),
		now:      time.Now,
	}
}

func (c *processCollector) Collect(ctx context.Context) ([]model.ProcessMetrics, error) {
	processes, err := process.ProcessesWithContext(ctx)
	if err != nil {
		return nil, fmt.Errorf("lista de processos: %w", err)
	}

	readAt := c.now()

	c.mu.Lock()
	defer c.mu.Unlock()

	metrics := make([]model.ProcessMetrics, 0, len(processes))
	current := make(map[int32]processSnapshot, len(processes))

	for _, proc := range processes {
		// A coleta percorre centenas de processos, cada um custando algumas
		// leituras em /proc. Se o cliente desistir da requisição no meio, não
		// há motivo para terminar a varredura.
		if err := ctx.Err(); err != nil {
			return nil, fmt.Errorf("coleta de processos cancelada: %w", err)
		}

		metric, snapshot, ok := c.measure(ctx, proc, readAt)
		if !ok {
			continue
		}

		current[proc.Pid] = snapshot
		metrics = append(metrics, metric)
	}

	// Substituir o mapa inteiro descarta os processos que morreram. Mantê-lo
	// crescendo seria um vazamento lento: uma máquina de desenvolvimento cria
	// e destrói milhares de processos por dia.
	c.previous = current

	return metrics, nil
}

// measure lê um processo. O terceiro retorno é falso quando o processo deve
// ficar de fora da lista.
//
// Um processo pode terminar entre a listagem e a leitura dos seus dados; nesse
// caso os arquivos em /proc desaparecem e cada chamada falha. Isso é rotina em
// qualquer máquina, não uma anomalia: o processo é apenas ignorado, sem log —
// registrar cada um encheria a saída a cada coleta.
func (c *processCollector) measure(ctx context.Context, proc *process.Process, readAt time.Time) (model.ProcessMetrics, processSnapshot, bool) {
	name, err := proc.NameWithContext(ctx)
	if err != nil || name == "" {
		return model.ProcessMetrics{}, processSnapshot{}, false
	}

	times, err := proc.TimesWithContext(ctx)
	if err != nil || times == nil {
		return model.ProcessMetrics{}, processSnapshot{}, false
	}

	memory, err := proc.MemoryInfoWithContext(ctx)
	if err != nil || memory == nil {
		return model.ProcessMetrics{}, processSnapshot{}, false
	}

	// CreateTime pode falhar sem invalidar o processo; sem ele o percentual
	// cai no caminho conservador de processCPUPercent.
	createdAt, err := proc.CreateTimeWithContext(ctx)
	if err != nil {
		createdAt = 0
	}

	current := processSnapshot{
		createdAt: createdAt,
		// Só user e system: o Total() do gopsutil soma também idle e iowait,
		// que não representam CPU consumida pelo processo.
		cpuSeconds: times.User + times.System,
		at:         readAt,
	}

	percent, keep := sampleCPU(c.previous[proc.Pid], current)

	return model.ProcessMetrics{
		PID:        proc.Pid,
		Name:       name,
		CPUPercent: percent,
		Memory:     memory.RSS,
	}, keep, true
}

// sampleCPU devolve o uso de CPU do processo em percentual e o snapshot que
// deve ficar guardado como base da próxima coleta.
//
// O percentual é relativo a um núcleo, como no top: um processo com quatro
// threads saturadas passa de 100. Limitá-lo a 100 esconderia justamente os
// processos que mais pesam na máquina.
func sampleCPU(previous, current processSnapshot) (percent float64, keep processSnapshot) {
	// Sem leitura anterior — ou com o PID reaproveitado por outro processo — a
	// média desde a criação é a melhor estimativa disponível. Devolver zero
	// deixaria a primeira resposta inteira zerada, e a ordenação por CPU não
	// teria o que ordenar.
	if previous.at.IsZero() || previous.createdAt != current.createdAt {
		current.percent = averageSinceStart(current)

		return current.percent, current
	}

	// Janela curta demais para medir: repete-se a última medição válida e
	// mantém-se a base anterior, de modo que a próxima coleta compare com um
	// intervalo utilizável em vez de recomeçar de uma base recém-gravada.
	if current.at.Sub(previous.at) < minSampleWindow {
		return previous.percent, previous
	}

	// O tempo de CPU é monotônico; uma diferença negativa significa leitura
	// inconsistente, não uso negativo.
	delta := current.cpuSeconds - previous.cpuSeconds
	if delta < 0 {
		return 0, current
	}

	current.percent = delta / current.at.Sub(previous.at).Seconds() * 100

	return current.percent, current
}

// averageSinceStart divide o tempo de CPU acumulado pela idade do processo.
func averageSinceStart(current processSnapshot) float64 {
	if current.createdAt <= 0 {
		return 0
	}

	lifetime := current.at.Sub(time.UnixMilli(current.createdAt)).Seconds()
	if lifetime <= 0 {
		return 0
	}

	return current.cpuSeconds / lifetime * 100
}
