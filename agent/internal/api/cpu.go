package api

import (
	"net/http"

	"pc-monitor-agent/internal/service"
)

// handleCPU serve GET /api/v1/cpu.
func handleCPU(svc *service.CPUService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		metrics, err := svc.Current(r.Context())
		if err != nil {
			writeError(w, r, err)
			return
		}

		writeJSON(w, http.StatusOK, metrics)
	}
}
