package handlers

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestMetricsHandler(t *testing.T) {
	h := NewHandler(nil) // NewHandler should accept nil pool in your code or use a test pool
	req := httptest.NewRequest(http.MethodGet, "/metrics", nil)
	w := httptest.NewRecorder()
	h.MetricsHandler(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200 OK, got %d; body: %s", w.Code, w.Body.String())
	}
}
