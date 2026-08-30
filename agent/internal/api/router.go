package api

import "net/http"

// NewRouter registra as rotas do agente e devolve o handler HTTP raiz.
func NewRouter() http.Handler {
	mux := http.NewServeMux()

	mux.HandleFunc("GET /health", handleHealth)

	return mux
}
