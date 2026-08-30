package collector

import (
	"context"
	"errors"
	"fmt"
	"runtime"

	"github.com/shirou/gopsutil/v4/host"

	"pc-monitor-agent/internal/model"
)

// systemCollector lê informações gerais da máquina.
type systemCollector struct{}

// NewSystemCollector devolve um SystemCollector baseado no gopsutil.
func NewSystemCollector() SystemCollector {
	return systemCollector{}
}

func (systemCollector) Collect(ctx context.Context) (model.SystemMetrics, error) {
	info, err := host.InfoWithContext(ctx)
	if err != nil {
		return model.SystemMetrics{}, fmt.Errorf("informações do host: %w", err)
	}

	if info == nil {
		return model.SystemMetrics{}, errors.New("informações do host: leitura vazia")
	}

	return model.SystemMetrics{
		Hostname:  info.Hostname,
		OS:        info.OS,
		OSVersion: osVersion(info),

		// runtime.GOARCH em vez de info.KernelArch para respeitar o contrato
		// do plan.md, que exemplifica "amd64" e não "x86_64".
		Architecture:  runtime.GOARCH,
		UptimeSeconds: info.Uptime,
	}, nil
}

// osVersion escolhe a identificação de versão mais informativa disponível.
//
// PlatformVersion é preferido por descrever a distribuição, mas várias delas
// não o preenchem; nesses casos a versão do kernel ainda diz algo útil.
func osVersion(info *host.InfoStat) string {
	if info.PlatformVersion != "" {
		return info.PlatformVersion
	}

	return info.KernelVersion
}
