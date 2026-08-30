package api

import (
	"net/http"

	"pc-monitor-agent/internal/model"
	"pc-monitor-agent/internal/service"
)

// handleNetwork serve GET /api/v1/network.
func handleNetwork(svc *service.NetworkService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		interfaces, err := svc.List(r.Context())
		if err != nil {
			writeError(w, r, err)
			return
		}

		if interfaces == nil {
			interfaces = []model.NetworkMetrics{}
		}

		writeJSON(w, http.StatusOK, interfaces)
	}
}
