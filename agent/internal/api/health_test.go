package api

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestHandleHealth(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	rec := httptest.NewRecorder()

	NewRouter(Deps{}).ServeHTTP(rec, req)

	res := rec.Result()
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		t.Errorf("status = %d, esperado %d", res.StatusCode, http.StatusOK)
	}

	if got := res.Header.Get("Content-Type"); got != contentTypeJSON {
		t.Errorf("Content-Type = %q, esperado %q", got, contentTypeJSON)
	}

	var body healthResponse
	if err := json.NewDecoder(res.Body).Decode(&body); err != nil {
		t.Fatalf("resposta não é JSON válido: %v", err)
	}

	if body.Status != "ok" {
		t.Errorf("status = %q, esperado %q", body.Status, "ok")
	}
}

func TestHandleHealthRejeitaMetodoInvalido(t *testing.T) {
	req := httptest.NewRequest(http.MethodPost, "/health", nil)
	rec := httptest.NewRecorder()

	NewRouter(Deps{}).ServeHTTP(rec, req)

	if rec.Code != http.StatusMethodNotAllowed {
		t.Errorf("status = %d, esperado %d", rec.Code, http.StatusMethodNotAllowed)
	}
}
