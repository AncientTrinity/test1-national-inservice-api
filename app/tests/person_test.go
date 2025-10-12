package tests

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"victortillett.net/test1-national-inservice-api/internal/config"
	"victortillett.net/test1-national-inservice-api/internal/db"
	"victortillett.net/test1-national-inservice-api/internal/handlers"
	"github.com/go-chi/chi/v5"
)

func setupServer(t *testing.T) (*chi.Mux, func()) {
	cfg, _ := config.Load()
	pool, err := db.Connect(cfg.DatabaseURL)
	if err != nil {
		t.Fatalf("db connect: %v", err)
	}
	cleanup := func() {
		_ = pool.Close(context.Background())
	}
	h := handlers.NewHandler(pool)
	r := chi.NewRouter()
	r.Post("/v1/persons", h.CreatePerson)
	r.Get("/v1/persons/{id}", h.GetPerson)
	return r, cleanup
}

func TestCreateAndGetPerson(t *testing.T) {
	r, cleanup := setupServer(t)
	defer cleanup()
	// create person
	body := map[string]interface{}{
		"first_name": "Test",
		"last_name":  "User",
		"email":      "test-user@example.org",
		"is_active":  true,
	}
	b, _ := json.Marshal(body)
	req := httptest.NewRequest("POST", "/v1/persons", bytes.NewReader(b))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)
	if rec.Code != http.StatusCreated {
		t.Fatalf("expected 201 created got %d body %s", rec.Code, rec.Body.String())
	}
	var resp map[string]int64
	_ = json.Unmarshal(rec.Body.Bytes(), &resp)
	id := resp["person_id"]
	// get
	req = httptest.NewRequest("GET", "/v1/persons/"+strconv.FormatInt(id,10), nil)
	rec = httptest.NewRecorder()
	r.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200 got %d", rec.Code)
	}
}
