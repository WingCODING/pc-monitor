package api

import (
	"net/http"

	"pc-monitor-agent/internal/model"
	"pc-monitor-agent/internal/service"
)

// handleDisks serve GET /api/v1/disks.
func handleDisks(svc *service.DiskService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		disks, err := svc.List(r.Context())
		if err != nil {
			writeError(w, r, err)
			return
		}

		// Uma lista nil serializaria como null; o contrato promete um array,
		// e um cliente que itere sobre null quebra.
		if disks == nil {
			disks = []model.DiskMetrics{}
		}

		writeJSON(w, http.StatusOK, disks)
	}
}
