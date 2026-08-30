package api

import (
	"net/http"

	"pc-monitor-agent/internal/service"
)

// Deps reúne os services que a camada HTTP consome. Um campo por categoria de
// métrica, preenchido no bootstrap do servidor.
type Deps struct {
	CPU     *service.CPUService
	Memory  *service.MemoryService
	System  *service.SystemService
	Disk    *service.DiskService
	Network *service.NetworkService
	Process *service.ProcessService
	Metrics *service.MetricsService
}

// NewRouter registra as rotas do agente e devolve o handler HTTP raiz.
func NewRouter(deps Deps) http.Handler {
	mux := http.NewServeMux()

	mux.HandleFunc("GET /health", handleHealth)
	mux.HandleFunc("GET /api/v1/cpu", handleCPU(deps.CPU))
	mux.HandleFunc("GET /api/v1/memory", handleMemory(deps.Memory))
	mux.HandleFunc("GET /api/v1/system", handleSystem(deps.System))
	mux.HandleFunc("GET /api/v1/disks", handleDisks(deps.Disk))
	mux.HandleFunc("GET /api/v1/network", handleNetwork(deps.Network))
	mux.HandleFunc("GET /api/v1/processes", handleProcesses(deps.Process))
	mux.HandleFunc("GET /api/v1/metrics", handleMetrics(deps.Metrics))

	return withMiddleware(mux)
}
