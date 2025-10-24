package handlers_test

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"

	"victortillett.net/test1-national-inservice-api/internal/handlers"
)

func TestRegister_InvalidBody(t *testing.T) {
	h := handlers.NewAuthHandler(nil, []byte("secret"), 1, nil)
	req := httptest.NewRequest(http.MethodPost, "/v1/auth/register", bytes.NewBufferString("invalid"))
	w := httptest.NewRecorder()

	h.Register(w, req)
	if w.Result().StatusCode != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", w.Result().StatusCode)
	}
}
