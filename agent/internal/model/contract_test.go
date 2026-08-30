package model

import (
	"encoding/json"
	"testing"
	"time"
)

// TestContratoJSON trava os nomes dos campos serializados. Os valores
// esperados vêm dos exemplos do plan.md, e o lado KMP espelha exatamente estes
// nomes — uma mudança aqui exige mudança lá.
func TestContratoJSON(t *testing.T) {
	casos := []struct {
		nome     string
		valor    any
		esperado string
	}{
		{
			nome: "CPUMetrics",
			valor: CPUMetrics{
				Model:        "AMD Ryzen 7 7800X3D",
				Cores:        8,
				Threads:      16,
				UsagePercent: 24.8,
			},
			esperado: `{"model":"AMD Ryzen 7 7800X3D","cores":8,"threads":16,"usagePercent":24.8}`,
		},
		{
			nome: "MemoryMetrics",
			valor: MemoryMetrics{
				Total:        34359738368,
				Used:         17179869184,
				Available:    17179869184,
				UsagePercent: 50,
			},
			esperado: `{"total":34359738368,"used":17179869184,"available":17179869184,"usagePercent":50}`,
		},
		{
			nome: "DiskMetrics",
			valor: DiskMetrics{
				Name:         "/dev/nvme0n1p2",
				MountPoint:   "/",
				Total:        1000204886016,
				Used:         520204886016,
				Free:         480000000000,
				UsagePercent: 52,
			},
			esperado: `{"name":"/dev/nvme0n1p2","mountPoint":"/","total":1000204886016,"used":520204886016,"free":480000000000,"usagePercent":52}`,
		},
		{
			nome: "NetworkMetrics",
			valor: NetworkMetrics{
				InterfaceName:          "eth0",
				BytesReceived:          1024,
				BytesSent:              512,
				DownloadBytesPerSecond: 5242880,
				UploadBytesPerSecond:   1048576,
			},
			esperado: `{"interfaceName":"eth0","bytesReceived":1024,"bytesSent":512,"downloadBytesPerSecond":5242880,"uploadBytesPerSecond":1048576}`,
		},
		{
			nome: "ProcessMetrics",
			valor: ProcessMetrics{
				PID:        4321,
				Name:       "idea",
				CPUPercent: 8.2,
				Memory:     1879048192,
			},
			esperado: `{"pid":4321,"name":"idea","cpuPercent":8.2,"memory":1879048192}`,
		},
		{
			nome: "SystemMetrics",
			valor: SystemMetrics{
				Hostname:      "desktop",
				OS:            "linux",
				OSVersion:     "6.1.0",
				Architecture:  "amd64",
				UptimeSeconds: 58232,
			},
			esperado: `{"hostname":"desktop","os":"linux","osVersion":"6.1.0","architecture":"amd64","uptimeSeconds":58232}`,
		},
	}

	for _, caso := range casos {
		t.Run(caso.nome, func(t *testing.T) {
			obtido, err := json.Marshal(caso.valor)
			if err != nil {
				t.Fatalf("marshal falhou: %v", err)
			}

			if string(obtido) != caso.esperado {
				t.Errorf("JSON divergente\nobtido:   %s\nesperado: %s", obtido, caso.esperado)
			}
		})
	}
}

func TestDashboardMetricsCompleto(t *testing.T) {
	uptime := uint64(53214)

	snapshot := DashboardMetrics{
		Timestamp:     time.Date(2026, 8, 30, 16, 0, 0, 0, time.UTC),
		CPU:           &CPUMetrics{Model: "Ryzen", Cores: 8, Threads: 16, UsagePercent: 32.8},
		Memory:        &MemoryMetrics{Total: 34359738368, Used: 17179869184, Available: 17179869184, UsagePercent: 50},
		Disk:          &DiskSummary{Total: 1000204886016, Used: 520204886016, UsagePercent: 52},
		Network:       &NetworkSummary{DownloadBytesPerSecond: 5242880, UploadBytesPerSecond: 1048576},
		UptimeSeconds: &uptime,
	}

	obtido, err := json.Marshal(snapshot)
	if err != nil {
		t.Fatalf("marshal falhou: %v", err)
	}

	var campos map[string]json.RawMessage
	if err := json.Unmarshal(obtido, &campos); err != nil {
		t.Fatalf("unmarshal falhou: %v", err)
	}

	for _, campo := range []string{"timestamp", "cpu", "memory", "disk", "network", "uptimeSeconds"} {
		if _, ok := campos[campo]; !ok {
			t.Errorf("campo %q ausente no snapshot completo", campo)
		}
	}
}

// TestDashboardMetricsParcial garante que uma métrica indisponível some do
// JSON em vez de ser serializada como zero — a UI precisa distinguir "0%" de
// "indisponível".
func TestDashboardMetricsParcial(t *testing.T) {
	snapshot := DashboardMetrics{
		Timestamp: time.Date(2026, 8, 30, 16, 0, 0, 0, time.UTC),
		CPU:       &CPUMetrics{UsagePercent: 32.8},
	}

	obtido, err := json.Marshal(snapshot)
	if err != nil {
		t.Fatalf("marshal falhou: %v", err)
	}

	var campos map[string]json.RawMessage
	if err := json.Unmarshal(obtido, &campos); err != nil {
		t.Fatalf("unmarshal falhou: %v", err)
	}

	for _, campo := range []string{"memory", "disk", "network", "uptimeSeconds"} {
		if _, ok := campos[campo]; ok {
			t.Errorf("campo %q deveria ser omitido quando indisponível", campo)
		}
	}

	if _, ok := campos["cpu"]; !ok {
		t.Error(`campo "cpu" deveria estar presente`)
	}

	if _, ok := campos["timestamp"]; !ok {
		t.Error(`campo "timestamp" é obrigatório e deveria estar sempre presente`)
	}
}
