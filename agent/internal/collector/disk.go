package collector

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/shirou/gopsutil/v4/disk"

	"pc-monitor-agent/internal/model"
)

// diskCollector lê o uso das partições montadas.
type diskCollector struct{}

// NewDiskCollector devolve um DiskCollector baseado no gopsutil.
func NewDiskCollector() DiskCollector {
	return diskCollector{}
}

func (diskCollector) Collect(ctx context.Context) ([]model.DiskMetrics, error) {
	// all=false já descarta as pseudo-filesystems (tmpfs, devtmpfs, efivarfs),
	// que ocupariam a lista sem representar armazenamento real.
	partitions, err := disk.PartitionsWithContext(ctx, false)
	if err != nil {
		return nil, fmt.Errorf("partições: %w", err)
	}

	disks := make([]model.DiskMetrics, 0, len(partitions))

	for _, partition := range partitions {
		usage, err := disk.UsageWithContext(ctx, partition.Mountpoint)
		if err != nil {
			// Um ponto de montagem ilegível — permissão negada, mídia
			// removida, montagem de rede fora do ar — não invalida os demais.
			slog.Warn("partição ignorada",
				"mountPoint", partition.Mountpoint,
				"error", err,
			)

			continue
		}

		// Total zero indica filesystem sem capacidade própria; publicá-lo só
		// poluiria a lista com uma linha de "0 B".
		if usage == nil || usage.Total == 0 {
			continue
		}

		disks = append(disks, model.DiskMetrics{
			Name:         partition.Device,
			MountPoint:   partition.Mountpoint,
			Total:        usage.Total,
			Used:         usage.Used,
			Free:         usage.Free,
			UsagePercent: percentOf(usage.Used, usage.Total),
		})
	}

	return deduplicateByDevice(disks), nil
}

// deduplicateByDevice mantém uma entrada por dispositivo físico.
//
// Subvolumes btrfs e bind mounts expõem o mesmo dispositivo em vários pontos
// de montagem, cada um reportando a capacidade total do filesystem. Nesta
// máquina, /dev/dm-0 aparece em /, /home, /var/log e /var/cache/pacman/pkg com
// os mesmos 998 GB — somá-los indicaria 4 TB num disco de 1 TB.
//
// Entre montagens do mesmo dispositivo vence a de caminho mais curto, que é a
// raiz do filesystem. A ordem original é preservada para que a saída seja
// determinística.
func deduplicateByDevice(disks []model.DiskMetrics) []model.DiskMetrics {
	positionOf := make(map[string]int, len(disks))
	unique := make([]model.DiskMetrics, 0, len(disks))

	for _, current := range disks {
		position, seen := positionOf[current.Name]
		if !seen {
			positionOf[current.Name] = len(unique)
			unique = append(unique, current)

			continue
		}

		if len(current.MountPoint) < len(unique[position].MountPoint) {
			unique[position] = current
		}
	}

	return unique
}
