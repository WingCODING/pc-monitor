package model

import "time"

// DashboardMetrics é o snapshot agregado consumido pelo dashboard, servido
// por GET /api/v1/metrics e pelo WebSocket.
//
// Todos os campos de métrica são opcionais: a falha de um collector individual
// não deve derrubar o snapshot inteiro, então a métrica indisponível é omitida
// do JSON em vez de aparecer zerada. Do lado KMP cada campo corresponde a um
// tipo anulável, o que permite à UI distinguir "0%" de "indisponível".
type DashboardMetrics struct {
	Timestamp time.Time `json:"timestamp"`

	CPU     *CPUMetrics     `json:"cpu,omitempty"`
	Memory  *MemoryMetrics  `json:"memory,omitempty"`
	Disk    *DiskSummary    `json:"disk,omitempty"`
	Network *NetworkSummary `json:"network,omitempty"`

	UptimeSeconds *uint64 `json:"uptimeSeconds,omitempty"`
}
