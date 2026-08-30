package model

// DiskMetrics descreve o uso de uma partição montada.
//
// Todos os tamanhos são em bytes.
type DiskMetrics struct {
	Name       string `json:"name"`
	MountPoint string `json:"mountPoint"`
	Total      uint64 `json:"total"`
	Used       uint64 `json:"used"`
	Free       uint64 `json:"free"`

	// UsagePercent varia de 0 a 100.
	UsagePercent float64 `json:"usagePercent"`
}

// DiskSummary condensa o estado dos discos para o dashboard, que exibe um
// único indicador em vez da lista completa de partições.
type DiskSummary struct {
	Total uint64 `json:"total"`
	Used  uint64 `json:"used"`

	// UsagePercent varia de 0 a 100.
	UsagePercent float64 `json:"usagePercent"`
}
