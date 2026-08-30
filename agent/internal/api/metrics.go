package api

import (
	"net/http"

	"pc-monitor-agent/internal/service"
)

// handleMetrics serve GET /api/v1/metrics.
//
// Responde 200 mesmo com métricas faltando: o snapshot parcial é uma resposta
// legítima, e o cliente distingue "indisponível" de "zero" pela ausência do
// campo no JSON.
func handleMetrics(svc *service.MetricsService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, http.StatusOK, svc.Snapshot(r.Context()))
	}
}
