package api

import (
	"bufio"
	"encoding/json"
	"errors"
	"log/slog"
	"net"
	"net/http"
	"runtime/debug"
	"strconv"
	"time"
)

// withMiddleware envolve o roteador com as três garantias que valem para
// qualquer rota: nenhum erro sai da API fora do formato padrão, um panic não
// derruba a resposta, e toda requisição é registrada.
//
// As três vivem no mesmo lugar porque compartilham o mesmo recurso — o
// ResponseWriter instrumentado. Separá-las exigiria uma cadeia de wrappers,
// cada um reimplementando o rastreamento de status.
func withMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		started := time.Now()
		recorder := &instrumentedWriter{ResponseWriter: w}

		defer func() {
			if recovered := recover(); recovered != nil {
				recoverFromPanic(recorder, r, recovered)
			}

			// Debug, não Info: o dashboard consulta o agente a cada segundo, e
			// uma linha por requisição encheria o log sem dizer nada. As
			// falhas continuam sendo registradas em writeError, com a causa.
			slog.Debug("requisição atendida",
				"method", r.Method,
				"path", r.URL.Path,
				"status", recorder.statusOrDefault(),
				"durationMs", time.Since(started).Milliseconds(),
			)
		}()

		next.ServeHTTP(recorder, r)
	})
}

// recoverFromPanic converte um panic de handler em 500 no formato padrão.
//
// Sem isso o servidor HTTP do Go aborta a conexão sem resposta: o cliente vê
// um erro de rede em vez de um erro da API, e o dashboard trata como agente
// fora do ar. O stack fica no log e não atravessa a fronteira HTTP.
func recoverFromPanic(w *instrumentedWriter, r *http.Request, recovered any) {
	// ErrAbortHandler é o sinal documentado para abortar a resposta em
	// silêncio; repassá-lo é o comportamento esperado.
	if err, ok := recovered.(error); ok && errors.Is(err, http.ErrAbortHandler) {
		panic(recovered)
	}

	slog.Error("panic no handler",
		"method", r.Method,
		"path", r.URL.Path,
		"panic", recovered,
		"stack", string(debug.Stack()),
	)

	// Se o handler já começou a responder, o status está definido e qualquer
	// escrita adicional só corromperia o corpo.
	if w.wrote {
		return
	}

	writeJSON(w, http.StatusInternalServerError, errorResponse{
		Error:   "internal_error",
		Message: "erro interno do agente",
	})
}

// instrumentedWriter registra o status da resposta e padroniza os erros que
// não passaram por writeError — 404 e 405 vêm do próprio ServeMux, em texto
// puro.
type instrumentedWriter struct {
	http.ResponseWriter

	status int
	wrote  bool

	// replaced indica que o corpo original foi substituído pelo JSON padrão e
	// que as escritas seguintes devem ser descartadas.
	replaced bool
}

func (w *instrumentedWriter) WriteHeader(status int) {
	if w.wrote {
		return
	}

	w.status = status
	w.wrote = true

	// Handlers do agente já respondem em JSON; só o que vem de fora deles
	// precisa ser convertido.
	if status < 400 || w.Header().Get("Content-Type") == contentTypeJSON {
		w.ResponseWriter.WriteHeader(status)

		return
	}

	body, err := json.Marshal(errorBodyFor(status))
	if err != nil {
		slog.Error("falha ao serializar erro padrão", "error", err)
		w.ResponseWriter.WriteHeader(status)

		return
	}

	w.replaced = true

	w.Header().Set("Content-Type", contentTypeJSON)
	// O ServeMux já anunciou o tamanho do corpo em texto puro.
	w.Header().Set("Content-Length", strconv.Itoa(len(body)))
	w.ResponseWriter.WriteHeader(status)

	if _, err := w.ResponseWriter.Write(body); err != nil {
		slog.Error("falha ao escrever erro padrão", "error", err)
	}
}

func (w *instrumentedWriter) Write(body []byte) (int, error) {
	if w.replaced {
		// Descarta o texto puro do ServeMux, já substituído pelo JSON.
		return len(body), nil
	}

	if !w.wrote {
		w.status = http.StatusOK
		w.wrote = true
	}

	return w.ResponseWriter.Write(body)
}

// Hijack repassa a tomada da conexão, usada pelo upgrade do WebSocket, e
// registra o 101 que o próprio protocolo escreve fora deste writer.
func (w *instrumentedWriter) Hijack() (net.Conn, *bufio.ReadWriter, error) {
	hijacker, ok := w.ResponseWriter.(http.Hijacker)
	if !ok {
		return nil, nil, errors.New("conexão não permite hijack")
	}

	conn, buffered, err := hijacker.Hijack()
	if err == nil {
		w.status = http.StatusSwitchingProtocols
		w.wrote = true
	}

	return conn, buffered, err
}

// Unwrap expõe o writer original para o http.ResponseController, por onde
// bibliotecas de WebSocket pedem Flush e Hijack.
func (w *instrumentedWriter) Unwrap() http.ResponseWriter {
	return w.ResponseWriter
}

// statusOrDefault devolve 200 quando nada foi escrito explicitamente, que é o
// que o servidor HTTP assume.
func (w *instrumentedWriter) statusOrDefault() int {
	if w.status == 0 {
		return http.StatusOK
	}

	return w.status
}

// errorBodyFor traduz os erros gerados fora dos handlers para o formato padrão
// da API.
func errorBodyFor(status int) errorResponse {
	switch status {
	case http.StatusNotFound:
		return errorResponse{
			Error:   "not_found",
			Message: "recurso não encontrado",
		}

	case http.StatusMethodNotAllowed:
		return errorResponse{
			Error:   "method_not_allowed",
			Message: "método não permitido para este recurso",
		}

	default:
		return errorResponse{
			Error:   "request_failed",
			Message: "requisição não pôde ser atendida",
		}
	}
}
