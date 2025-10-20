package handlers

import (
	"encoding/json"
	"net/http"
	"runtime"
	"time"
)

type Metrics struct {
	Timestamp   time.Time `json:"timestamp"`
	NumGoroutine int      `json:"num_goroutine"`
	NumCPU       int      `json:"num_cpu"`
}

func (h *Handler) MetricsHandler(w http.ResponseWriter, r *http.Request) {
	m := Metrics{
		Timestamp:    time.Now(),
		NumGoroutine: runtime.NumGoroutine(),
		NumCPU:       runtime.NumCPU(),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(m)
}
