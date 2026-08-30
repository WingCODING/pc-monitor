package model

// CPUMetrics descreve o processador e seu uso instantâneo.
type CPUMetrics struct {
	Model   string `json:"model"`
	Cores   int    `json:"cores"`
	Threads int    `json:"threads"`

	// UsagePercent varia de 0 a 100.
	UsagePercent float64 `json:"usagePercent"`
}
