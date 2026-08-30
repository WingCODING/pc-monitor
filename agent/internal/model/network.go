package model

// NetworkMetrics descreve o tráfego de uma interface de rede.
//
// BytesReceived e BytesSent são contadores acumulados desde o boot; as
// velocidades são derivadas da diferença entre duas leituras.
type NetworkMetrics struct {
	InterfaceName string `json:"interfaceName"`
	BytesReceived uint64 `json:"bytesReceived"`
	BytesSent     uint64 `json:"bytesSent"`

	// As velocidades são float64 porque o intervalo real entre coletas nunca
	// é exatamente o intervalo nominal.
	DownloadBytesPerSecond float64 `json:"downloadBytesPerSecond"`
	UploadBytesPerSecond   float64 `json:"uploadBytesPerSecond"`
}

// NetworkSummary agrega as velocidades de todas as interfaces para o
// dashboard, que exibe um único par download/upload.
type NetworkSummary struct {
	DownloadBytesPerSecond float64 `json:"downloadBytesPerSecond"`
	UploadBytesPerSecond   float64 `json:"uploadBytesPerSecond"`
}
