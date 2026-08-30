package model

// SystemMetrics descreve informações gerais da máquina.
type SystemMetrics struct {
	Hostname      string `json:"hostname"`
	OS            string `json:"os"`
	OSVersion     string `json:"osVersion"`
	Architecture  string `json:"architecture"`
	UptimeSeconds uint64 `json:"uptimeSeconds"`
}
