package api

import (
	"net/http"

	"pc-monitor-agent/internal/service"
)

// Deps reúne os services que a camada HTTP consome. Um campo por categoria de
// métrica, preenchido no bootstrap do servidor.
type Deps struct {
	CPU *service.CPUService
}

// NewRouter registra as rotas do agente e devolve o handler HTTP raiz.
func NewRouter(deps Deps) http.Handler {
	mux := http.NewServeMux()

	mux.HandleFunc("GET /health", handleHealth)
	mux.HandleFunc("GET /api/v1/cpu", handleCPU(deps.CPU))

	return mux
}
