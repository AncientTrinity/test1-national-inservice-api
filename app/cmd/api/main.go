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
	chiCors "github.com/go-chi/cors"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		fmt.Println("config error:", err)
		os.Exit(1)
	}

	pool, err := db.Connect(cfg.DatabaseURL)
	if err != nil {
		fmt.Println("db connect:", err)
		os.Exit(1)
	}
	defer pool.Close(context.Background())

	r := chi.NewRouter()

	// CORS
	r.Use(chiCors.Handler(chiCors.Options{
		AllowedOrigins:   []string{cfg.CORSAllowedOrigins},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token"},
		AllowCredentials: true,
		MaxAge:           300,
	}))

	// JSON header
	r.Use(middleware.JSONMiddleware)

	// Rate limiter
	rl := middleware.NewRateLimiter(cfg.RateLimitRPS, cfg.RateLimitBurst)
	r.Use(rl.LimitMiddleware)

	h := handlers.NewHandler(pool)

	r.Route("/v1", func(r chi.Router) {
		r.Route("/persons", func(r chi.Router) {
			r.Get("/", h.ListPersons)
			r.Post("/", h.CreatePerson)
			r.Get("/{id}", h.GetPerson)
			r.Put("/{id}", h.UpdatePerson)
			r.Delete("/{id}", h.DeletePerson)
		})

		r.Route("/courses", func(r chi.Router) {
			r.Get("/", h.ListCourses)
			r.Post("/", h.CreateCourse)
			r.Get("/{id}", h.GetCourse)
			r.Put("/{id}", h.UpdateCourse)
			r.Delete("/{id}", h.DeleteCourse)
		})

		r.Route("/participants", func(r chi.Router) {
			r.Get("/", h.ListParticipants)
			r.Post("/", h.CreateParticipant)
			r.Get("/{id}", h.GetParticipant)
			r.Put("/{id}", h.UpdateParticipant)
			r.Delete("/{id}", h.DeleteParticipant)
		})

		// add more resource routes (facilitators, formations, etc.) similarly
	})

	srv := &http.Server{
		Addr:    fmt.Sprintf(":%s", cfg.Port),
		Handler: r,
	}

	// graceful shutdown
	idle := make(chan struct{})
	go func() {
		stop := make(chan os.Signal, 1)
		signal.Notify(stop, os.Interrupt)
		<-stop
		ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
		defer cancel()
		_ = srv.Shutdown(ctx)
		close(idle)
	}()

	fmt.Printf("listening on %s\n", cfg.Port)
	if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		fmt.Println("server error:", err)
	}
	<-idle
	fmt.Println("server stopped")
}
