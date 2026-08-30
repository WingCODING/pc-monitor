package api

import "net/http"

// healthResponse é o corpo devolvido por GET /health.
type healthResponse struct {
	Status string `json:"status"`
}

// handleHealth informa que o agente está no ar.
func handleHealth(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, healthResponse{Status: "ok"})
}
