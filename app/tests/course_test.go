package tests

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"app/internal/config"
	"app/internal/db"
	"app/internal/handlers"

	"github.com/go-chi/chi/v5"
)

func setupCourse(t *testing.T) (*chi.Mux, func()) {
	cfg, _ := config.Load()
	pool, err := db.Connect(cfg.DatabaseURL)
	if err != nil { t.Fatalf("db connect: %v", err) }
	cleanup := func() { pool.Close(nil) }
	h := handlers.NewHandler(pool)
	r := chi.NewRouter()
	r.Post("/v1/courses", h.CreateCourse)
	return r, cleanup
}

func TestCreateCourse(t *testing.T) {
	r, cleanup := setupCourse(t)
	defer cleanup()
	body := map[string]interface{}{"code":"TST101","title":"Test Course","credit_hours":3}
	b, _ := json.Marshal(body)
	req := httptest.NewRequest(http.MethodPost, "/v1/courses", bytes.NewReader(b))
	req.Header.Set("Content-Type","application/json")
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)
	if rec.Code != http.StatusCreated {
		t.Fatalf("expected 201 got %d body %s", rec.Code, rec.Body.String())
	}
}

func TestCreateCourseInvalidBody(t *testing.T) {
	r, cleanup := setupCourse(t)
	defer cleanup()
	req := httptest.NewRequest(http.MethodPost, "/v1/courses", bytes.NewReader([]byte("invalid json")))
	req.Header.Set("Content-Type","application/json")
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 got %d body %s", rec.Code, rec.Body.String())
	}
}