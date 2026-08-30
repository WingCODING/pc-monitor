package collector

import (
	"context"
	"testing"

	"pc-monitor-agent/internal/model"
)

// TestDeduplicateByDevice cobre o cenário que motivou a função: subvolumes
// btrfs e bind mounts expondo o mesmo dispositivo em vários pontos de
// montagem, cada um reportando a capacidade inteira do filesystem.
func TestDeduplicateByDevice(t *testing.T) {
	casos := []struct {
		nome     string
		entrada  []model.DiskMetrics
		esperado []model.DiskMetrics
	}{
		{
			nome: "subvolumes btrfs colapsam na raiz",
			entrada: []model.DiskMetrics{
				{Name: "/dev/dm-0", MountPoint: "/", Total: 998037782528},
				{Name: "/dev/dm-0", MountPoint: "/var/cache/pacman/pkg", Total: 998037782528},
				{Name: "/dev/dm-0", MountPoint: "/var/log", Total: 998037782528},
				{Name: "/dev/dm-0", MountPoint: "/home", Total: 998037782528},
			},
			esperado: []model.DiskMetrics{
				{Name: "/dev/dm-0", MountPoint: "/", Total: 998037782528},
			},
		},
		{
			nome: "dispositivos distintos são preservados",
			entrada: []model.DiskMetrics{
				{Name: "/dev/dm-0", MountPoint: "/", Total: 998037782528},
				{Name: "/dev/nvme0n1p1", MountPoint: "/boot", Total: 2143281152},
				{Name: "/dev/sdb1", MountPoint: "/run/media/will/Ventoy", Total: 15559819264},
			},
			esperado: []model.DiskMetrics{
				{Name: "/dev/dm-0", MountPoint: "/", Total: 998037782528},
				{Name: "/dev/nvme0n1p1", MountPoint: "/boot", Total: 2143281152},
				{Name: "/dev/sdb1", MountPoint: "/run/media/will/Ventoy", Total: 15559819264},
			},
		},
		{
			nome: "vence o caminho mais curto mesmo fora de ordem",
			entrada: []model.DiskMetrics{
				{Name: "/dev/dm-0", MountPoint: "/home/will/dados", Total: 100},
				{Name: "/dev/dm-0", MountPoint: "/", Total: 100},
			},
			esperado: []model.DiskMetrics{
				{Name: "/dev/dm-0", MountPoint: "/", Total: 100},
			},
		},
		{
			nome:     "lista vazia",
			entrada:  []model.DiskMetrics{},
			esperado: []model.DiskMetrics{},
		},
	}

	for _, caso := range casos {
		t.Run(caso.nome, func(t *testing.T) {
			obtido := deduplicateByDevice(caso.entrada)

			if len(obtido) != len(caso.esperado) {
				t.Fatalf("len = %d, esperado %d (%+v)", len(obtido), len(caso.esperado), obtido)
			}

			for i := range obtido {
				if obtido[i] != caso.esperado[i] {
					t.Errorf("posição %d = %+v, esperado %+v", i, obtido[i], caso.esperado[i])
				}
			}
		})
	}
}

// TestDiskCollectorMaquinaReal valida o contrato contra o sistema em execução.
func TestDiskCollectorMaquinaReal(t *testing.T) {
	disks, err := NewDiskCollector().Collect(context.Background())
	if err != nil {
		t.Fatalf("Collect() falhou: %v", err)
	}

	if len(disks) == 0 {
		t.Fatal("nenhuma partição detectada; ao menos a principal deveria aparecer")
	}

	dispositivos := make(map[string]string, len(disks))

	for _, atual := range disks {
		if atual.Total == 0 {
			t.Errorf("%s: Total = 0, partição sem capacidade não deveria ser publicada", atual.MountPoint)
		}

		if atual.Used > atual.Total {
			t.Errorf("%s: Used (%d) maior que Total (%d)", atual.MountPoint, atual.Used, atual.Total)
		}

		if atual.UsagePercent < 0 || atual.UsagePercent > 100 {
			t.Errorf("%s: UsagePercent = %v, esperado entre 0 e 100", atual.MountPoint, atual.UsagePercent)
		}

		if atual.MountPoint == "" {
			t.Error("MountPoint vazio")
		}

		if anterior, repetido := dispositivos[atual.Name]; repetido {
			t.Errorf("dispositivo %s repetido em %q e %q; a deduplicação falhou",
				atual.Name, anterior, atual.MountPoint)
		}

		dispositivos[atual.Name] = atual.MountPoint
	}
}
