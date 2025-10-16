package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"time"

	"victortillett.net/test1-national-inservice-api/internal/config"
	"victortillett.net/test1-national-inservice-api/internal/db"
	"victortillett.net/test1-national-inservice-api/internal/handlers"
	"victortillett.net/test1-national-inservice-api/internal/middleware"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/cors"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		fmt.Println("failed to load config:", err)
		os.Exit(1)
	}

	pool, err := db.Connect(cfg.DatabaseURL)
	if err != nil {
		fmt.Println("db connect:", err)
		os.Exit(1)
	}

	defer pool.Close()

	r := chi.NewRouter()

	// CORS
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{cfg.CORSAllowedOrigins},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token"},
		AllowCredentials: true,
		MaxAge:           300,
	}))

	// Rate limiting
	rl := middleware.NewRateLimiter(cfg.RateLimitRPS, cfg.RateLimitBurst)
	r.Use(rl.LimitMiddleware)

	// JSON content type
	r.Use(middleware.JSONMiddleware)

	// Handlers
	h := handlers.NewHandler(pool)

	r.Route("/v1", func(r chi.Router) {
		r.Route("/persons", func(r chi.Router) {
			r.Get("/", h.ListPersons)       // list with pagination & sorting
			r.Post("/", h.CreatePerson)     // create
			r.Get("/{id}", h.GetPerson)     // read
			r.Put("/{id}", h.UpdatePerson)  // update
			r.Delete("/{id}", h.DeletePerson) // delete
		})
	})

	srv := &http.Server{
		Addr:    fmt.Sprintf(":%s", cfg.Port),
		Handler: r,
	}

	// Graceful shutdown
	idleConnsClosed := make(chan struct{})
	go func() {
		c := make(chan os.Signal, 1)
		signal.Notify(c, os.Interrupt)
		<-c
		ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
		defer cancel()
		_ = srv.Shutdown(ctx)
		close(idleConnsClosed)
	}()

	fmt.Println("server listening on", cfg.Port)
	if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		fmt.Println("server error:", err)
	}

	<-idleConnsClosed
	fmt.Println("server stopped gracefully")
}
