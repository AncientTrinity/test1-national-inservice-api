package handlers_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"victortillett.net/test1-national-inservice-api/internal/handlers"
)

func TestMetricsHandler(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/metrics", nil)
	w := httptest.NewRecorder()
	h := handlers.NewHandler(nil)

	h.MetricsHandler(w, req)
	if w.Result().StatusCode != http.StatusOK {
		t.Fatalf("expected 200 OK, got %d", w.Result().StatusCode)
	}
}
