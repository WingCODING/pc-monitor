package api

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func decodificarErro(t *testing.T, rec *httptest.ResponseRecorder) errorResponse {
	t.Helper()

	if tipo := rec.Header().Get("Content-Type"); tipo != contentTypeJSON {
		t.Errorf("Content-Type = %q, esperado %q", tipo, contentTypeJSON)
	}

	var corpo errorResponse
	if err := json.NewDecoder(rec.Body).Decode(&corpo); err != nil {
		t.Fatalf("erro não é JSON válido: %v", err)
	}

	return corpo
}

// TestRotaInexistenteRespondeJSON cobre o 404 do ServeMux, que sem o
// middleware sairia como texto puro e quebraria o padrão da API.
func TestRotaInexistenteRespondeJSON(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/api/v1/inexistente", nil)
	rec := httptest.NewRecorder()

	NewRouter(Deps{}).ServeHTTP(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Fatalf("status = %d, esperado %d", rec.Code, http.StatusNotFound)
	}

	if corpo := decodificarErro(t, rec); corpo.Error != "not_found" {
		t.Errorf("error = %q, esperado not_found", corpo.Error)
	}
}

// TestMetodoNaoPermitidoRespondeJSON cobre o 405, também gerado pelo ServeMux.
func TestMetodoNaoPermitidoRespondeJSON(t *testing.T) {
	req := httptest.NewRequest(http.MethodPost, "/health", nil)
	rec := httptest.NewRecorder()

	NewRouter(Deps{}).ServeHTTP(rec, req)

	if rec.Code != http.StatusMethodNotAllowed {
		t.Fatalf("status = %d, esperado %d", rec.Code, http.StatusMethodNotAllowed)
	}

	if corpo := decodificarErro(t, rec); corpo.Error != "method_not_allowed" {
		t.Errorf("error = %q, esperado method_not_allowed", corpo.Error)
	}
}

// TestRespostaDeSucessoPassaIntacta garante que a padronização de erros não
// mexe no caminho feliz.
func TestRespostaDeSucessoPassaIntacta(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	rec := httptest.NewRecorder()

	NewRouter(Deps{}).ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, esperado %d", rec.Code, http.StatusOK)
	}

	var corpo map[string]string
	if err := json.NewDecoder(rec.Body).Decode(&corpo); err != nil {
		t.Fatalf("resposta não é JSON válido: %v", err)
	}

	if corpo["status"] != "ok" {
		t.Errorf("status = %q, esperado ok", corpo["status"])
	}
}

// TestPanicViraErroPadrao documenta o motivo do recover: sem ele o servidor
// aborta a conexão sem resposta, e o dashboard trataria como agente fora do ar
// em vez de erro da API.
func TestPanicViraErroPadrao(t *testing.T) {
	explosivo := http.HandlerFunc(func(http.ResponseWriter, *http.Request) {
		panic("collector explodiu")
	})

	req := httptest.NewRequest(http.MethodGet, "/api/v1/cpu", nil)
	rec := httptest.NewRecorder()

	withMiddleware(explosivo).ServeHTTP(rec, req)

	if rec.Code != http.StatusInternalServerError {
		t.Fatalf("status = %d, esperado %d", rec.Code, http.StatusInternalServerError)
	}

	corpo := decodificarErro(t, rec)

	if corpo.Error != "internal_error" {
		t.Errorf("error = %q, esperado internal_error", corpo.Error)
	}

	// O stack fica no log; nada dele atravessa a fronteira HTTP.
	if corpo.Message != "erro interno do agente" {
		t.Errorf("message = %q, esperado a mensagem genérica", corpo.Message)
	}
}

// TestPanicDepoisDeResponderNaoCorrompeOCorpo cobre o caso em que o handler já
// escreveu: sobrescrever ali só produziria um corpo inválido.
func TestPanicDepoisDeResponderNaoCorrompeOCorpo(t *testing.T) {
	tardio := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
		panic("falha depois da resposta")
	})

	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	rec := httptest.NewRecorder()

	withMiddleware(tardio).ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("status = %d, esperado %d", rec.Code, http.StatusOK)
	}

	if corpo := rec.Body.String(); corpo != `{"status":"ok"}` {
		t.Errorf("corpo = %s, esperado a resposta original intacta", corpo)
	}
}
