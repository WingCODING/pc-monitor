package api

import (
	"net/http"

	"pc-monitor-agent/internal/service"
)

// handleSystem serve GET /api/v1/system.
func handleSystem(svc *service.SystemService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		metrics, err := svc.Current(r.Context())
		if err != nil {
			writeError(w, r, err)
			return
		}

		writeJSON(w, http.StatusOK, metrics)
	}
}
