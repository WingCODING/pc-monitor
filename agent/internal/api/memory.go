package api

import (
	"net/http"

	"pc-monitor-agent/internal/service"
)

// handleMemory serve GET /api/v1/memory.
func handleMemory(svc *service.MemoryService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		metrics, err := svc.Current(r.Context())
		if err != nil {
			writeError(w, r, err)
			return
		}

		writeJSON(w, http.StatusOK, metrics)
	}
}
