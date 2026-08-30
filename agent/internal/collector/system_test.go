package collector

import (
	"context"
	"os"
	"runtime"
	"testing"

	"github.com/shirou/gopsutil/v4/host"
)

// TestSystemCollectorMaquinaReal compara o resultado com fontes independentes
// do gopsutil, para detectar um mapeamento de campo trocado.
func TestSystemCollectorMaquinaReal(t *testing.T) {
	metricas, err := NewSystemCollector().Collect(context.Background())
	if err != nil {
		t.Fatalf("Collect() falhou: %v", err)
	}

	hostnameReal, err := os.Hostname()
	if err != nil {
		t.Fatalf("os.Hostname() falhou: %v", err)
	}

	if metricas.Hostname != hostnameReal {
		t.Errorf("Hostname = %q, esperado %q", metricas.Hostname, hostnameReal)
	}

	if metricas.OS != runtime.GOOS {
		t.Errorf("OS = %q, esperado %q", metricas.OS, runtime.GOOS)
	}

	if metricas.Architecture != runtime.GOARCH {
		t.Errorf("Architecture = %q, esperado %q", metricas.Architecture, runtime.GOARCH)
	}

	if metricas.UptimeSeconds == 0 {
		t.Error("UptimeSeconds = 0, esperado > 0 numa máquina em execução")
	}

	if metricas.OSVersion == "" {
		t.Error("OSVersion vazia; nem PlatformVersion nem KernelVersion foram preenchidas")
	}
}

// TestOSVersionPrefereePlatformVersion cobre a escolha entre as duas fontes de
// versão sem depender do que a máquina de teste expõe.
func TestOSVersionPreferePlatformVersion(t *testing.T) {
	casos := []struct {
		nome     string
		info     host.InfoStat
		esperado string
	}{
		{
			nome:     "usa PlatformVersion quando disponível",
			info:     host.InfoStat{PlatformVersion: "4.0.1", KernelVersion: "7.1.9-arch1-2"},
			esperado: "4.0.1",
		},
		{
			nome:     "cai para KernelVersion quando a distribuição não informa",
			info:     host.InfoStat{PlatformVersion: "", KernelVersion: "7.1.9-arch1-2"},
			esperado: "7.1.9-arch1-2",
		},
		{
			nome:     "ambas vazias devolve vazio",
			info:     host.InfoStat{},
			esperado: "",
		},
	}

	for _, caso := range casos {
		t.Run(caso.nome, func(t *testing.T) {
			if obtido := osVersion(&caso.info); obtido != caso.esperado {
				t.Errorf("osVersion() = %q, esperado %q", obtido, caso.esperado)
			}
		})
	}
}
