package api

import (
	"encoding/json"
	"log/slog"
	"net/http"
)

const contentTypeJSON = "application/json"

// writeJSON serializa v e escreve a resposta com o status informado.
//
// A serialização acontece antes de qualquer escrita para que uma falha ainda
// possa ser reportada como 500 — depois de WriteHeader o status é definitivo.
func writeJSON(w http.ResponseWriter, status int, v any) {
	body, err := json.Marshal(v)
	if err != nil {
		slog.Error("falha ao serializar resposta", "error", err)
		w.Header().Set("Content-Type", contentTypeJSON)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", contentTypeJSON)
	w.WriteHeader(status)

	if _, err := w.Write(body); err != nil {
		slog.Error("falha ao escrever resposta", "error", err)
	}
}
