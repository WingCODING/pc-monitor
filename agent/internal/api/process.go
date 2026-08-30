package api

import (
	"fmt"
	"net/http"
	"net/url"
	"strconv"

	"pc-monitor-agent/internal/model"
	"pc-monitor-agent/internal/service"
)

// handleProcesses serve GET /api/v1/processes.
//
// Aceita `sort` (cpu|memory) e `limit`; sem parâmetros usa os padrões do
// service, que são os de um monitor: maior uso de CPU primeiro, 50 linhas.
func handleProcesses(svc *service.ProcessService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		options, err := parseProcessOptions(r.URL.Query())
		if err != nil {
			writeError(w, r, err)
			return
		}

		processes, err := svc.List(r.Context(), options)
		if err != nil {
			writeError(w, r, err)
			return
		}

		if processes == nil {
			processes = []model.ProcessMetrics{}
		}

		writeJSON(w, http.StatusOK, processes)
	}
}

// parseProcessOptions valida a query string. Parâmetro inválido é erro do
// cliente (400) e não uma métrica indisponível (503), então precisa ser
// distinguido antes de chegar ao service.
func parseProcessOptions(query url.Values) (service.ProcessOptions, error) {
	options := service.ProcessOptions{
		SortBy: service.DefaultProcessSort,
		Limit:  service.DefaultProcessLimit,
	}

	if raw := query.Get("sort"); raw != "" {
		switch service.ProcessSortBy(raw) {
		case service.ProcessSortByCPU, service.ProcessSortByMemory:
			options.SortBy = service.ProcessSortBy(raw)
		default:
			return service.ProcessOptions{}, fmt.Errorf("%w: sort=%q", service.ErrInvalidArgument, raw)
		}
	}

	if raw := query.Get("limit"); raw != "" {
		limit, err := strconv.Atoi(raw)
		if err != nil || limit < 1 || limit > service.MaxProcessLimit {
			return service.ProcessOptions{}, fmt.Errorf("%w: limit=%q", service.ErrInvalidArgument, raw)
		}

		options.Limit = limit
	}

	return options, nil
}
