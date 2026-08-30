package model

// MemoryMetrics descreve o uso da memória RAM.
//
// Todos os tamanhos são em bytes; a formatação para MB/GB/TB é
// responsabilidade da UI.
type MemoryMetrics struct {
	Total     uint64 `json:"total"`
	Used      uint64 `json:"used"`
	Available uint64 `json:"available"`

	// UsagePercent varia de 0 a 100.
	UsagePercent float64 `json:"usagePercent"`
}
