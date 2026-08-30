package service

import (
	"context"
	"errors"
	"testing"
	"time"

	"pc-monitor-agent/internal/model"
)

func metricsServiceComDubles(cpu cpuCollectorFalso, memoria memoryCollectorFalso, sistema systemCollectorFalso) *MetricsService {
	disco := diskCollectorFalso{discos: []model.DiskMetrics{{Name: "/dev/dm-0", MountPoint: "/", Total: 1000, Used: 400, UsagePercent: 40}}}

	return NewMetricsService(NewCPUService(cpu), NewMemoryService(memoria), NewSystemService(sistema), NewDiskService(disco))
}

// sistemaOK devolve um dublê de sistema saudável, para os testes que não estão
// exercitando o caminho de uptime.
func sistemaOK() systemCollectorFalso {
	return systemCollectorFalso{metricas: model.SystemMetrics{Hostname: "desktop", UptimeSeconds: 53214}}
}

func TestMetricsServiceSnapshotCompleto(t *testing.T) {
	cpuEsperado := model.CPUMetrics{Model: "Ryzen", Cores: 8, Threads: 16, UsagePercent: 32.8}
	memoriaEsperada := model.MemoryMetrics{Total: 34359738368, Used: 17179869184, Available: 17179869184, UsagePercent: 50}

	antes := time.Now().UTC()

	snapshot := metricsServiceComDubles(
		cpuCollectorFalso{metricas: cpuEsperado},
		memoryCollectorFalso{metricas: memoriaEsperada},
		sistemaOK(),
	).Snapshot(context.Background())

	depois := time.Now().UTC()

	if snapshot.CPU == nil {
		t.Fatal("CPU ausente no snapshot completo")
	}

	if *snapshot.CPU != cpuEsperado {
		t.Errorf("CPU = %+v, esperado %+v", *snapshot.CPU, cpuEsperado)
	}

	if snapshot.Memory == nil {
		t.Fatal("Memory ausente no snapshot completo")
	}

	if *snapshot.Memory != memoriaEsperada {
		t.Errorf("Memory = %+v, esperado %+v", *snapshot.Memory, memoriaEsperada)
	}

	if snapshot.Timestamp.Before(antes) || snapshot.Timestamp.After(depois) {
		t.Errorf("Timestamp %v fora da janela da coleta (%v–%v)", snapshot.Timestamp, antes, depois)
	}
}

// TestMetricsServiceSnapshotParcial é o teste central deste epic: a falha de
// uma métrica não pode derrubar as outras nem o snapshot inteiro.
func TestMetricsServiceSnapshotParcial(t *testing.T) {
	memoriaEsperada := model.MemoryMetrics{Total: 34359738368, Used: 17179869184, UsagePercent: 50}

	snapshot := metricsServiceComDubles(
		cpuCollectorFalso{err: errors.New("/proc/stat ilegível")},
		memoryCollectorFalso{metricas: memoriaEsperada},
		sistemaOK(),
	).Snapshot(context.Background())

	if snapshot.CPU != nil {
		t.Errorf("CPU deveria ser omitida quando o collector falha, veio %+v", *snapshot.CPU)
	}

	if snapshot.Memory == nil {
		t.Fatal("memória foi perdida por causa da falha de CPU")
	}

	if *snapshot.Memory != memoriaEsperada {
		t.Errorf("Memory = %+v, esperado %+v", *snapshot.Memory, memoriaEsperada)
	}

	if snapshot.Timestamp.IsZero() {
		t.Error("Timestamp ausente no snapshot parcial")
	}
}

// TestMetricsServiceTodasAsMetricasFalham garante que o snapshot continua
// sendo produzido mesmo sem nenhuma métrica disponível.
func TestMetricsServiceTodasAsMetricasFalham(t *testing.T) {
	falha := errors.New("sistema inacessível")

	snapshot := metricsServiceComDubles(
		cpuCollectorFalso{err: falha},
		memoryCollectorFalso{err: falha},
		systemCollectorFalso{err: falha},
	).Snapshot(context.Background())

	if snapshot.CPU != nil || snapshot.Memory != nil || snapshot.UptimeSeconds != nil {
		t.Error("nenhuma métrica deveria estar presente")
	}

	if snapshot.Timestamp.IsZero() {
		t.Error("Timestamp deveria estar presente mesmo sem métricas")
	}
}
