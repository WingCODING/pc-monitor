package api

import (
	"errors"
	"log/slog"
	"net/http"

	"pc-monitor-agent/internal/service"
)

// errorResponse é o corpo padrão de erro da API.
type errorResponse struct {
	Error   string `json:"error"`
	Message string `json:"message"`
}

// writeError registra a causa completa no log e devolve ao cliente apenas uma
// descrição genérica — detalhes internos não atravessam a fronteira HTTP.
func writeError(w http.ResponseWriter, r *http.Request, err error) {
	status := http.StatusInternalServerError
	body := errorResponse{
		Error:   "internal_error",
		Message: "erro interno do agente",
	}

	if errors.Is(err, service.ErrUnavailable) {
		status = http.StatusServiceUnavailable
		body = errorResponse{
			Error:   "collector_unavailable",
			Message: "métrica temporariamente indisponível",
		}
	}

	slog.Error("requisição falhou",
		"method", r.Method,
		"path", r.URL.Path,
		"status", status,
		"error", err,
	)

	writeJSON(w, status, body)
}
