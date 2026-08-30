package model

// ProcessMetrics descreve um processo em execução.
type ProcessMetrics struct {
	PID  int32  `json:"pid"`
	Name string `json:"name"`

	// CPUPercent pode ultrapassar 100 em processos com várias threads.
	CPUPercent float64 `json:"cpuPercent"`

	// Memory é a memória residente em bytes.
	Memory uint64 `json:"memory"`
}
