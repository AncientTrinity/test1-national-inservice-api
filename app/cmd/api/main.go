package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"victortillett.net/test1-national-inservice-api/internal/config"
	"victortillett.net/test1-national-inservice-api/internal/db"
	"victortillett.net/test1-national-inservice-api/internal/handlers"
	"victortillett.net/test1-national-inservice-api/internal/middleware"
	"victortillett.net/test1-national-inservice-api/internal/services"

	"github.com/go-chi/chi/v5"
	chiCors "github.com/go-chi/cors"
)

func main() {
	// Load environment configuration
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("❌ config error: %v", err)
	}

	// Connect to PostgreSQL
	pool, err := db.Connect(cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("❌ database connection error: %v", err)
	}
	defer pool.Close()

	// Router setup
	r := chi.NewRouter()

	// Global middlewares
	r.Use(chiCors.Handler(chiCors.Options{
		AllowedOrigins:   []string{cfg.CORSAllowedOrigins},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token"},
		AllowCredentials: true,
		MaxAge:           300,
	}))
	r.Use(middleware.JSONMiddleware)

	// Rate limiter
	rl := middleware.NewRateLimiter(cfg.RateLimitRPS, cfg.RateLimitBurst)
	r.Use(rl.LimitMiddleware)

	// Initialize core handler
	h := handlers.NewHandler(pool)

	// Metrics endpoint
	r.Get("/metrics", h.MetricsHandler)

	// ========== API Routes ==========
	r.Route("/v1", func(r chi.Router) {

		// Persons
		r.Route("/persons", func(r chi.Router) {
			r.Get("/", h.ListPersons)
			r.Post("/", h.CreatePerson)
			r.Get("/{id}", h.GetPerson)
			r.Put("/{id}", h.UpdatePerson)
			r.Delete("/{id}", h.DeletePerson)
		})

		// Courses
		r.Route("/courses", func(r chi.Router) {
			r.Get("/", h.ListCourses)
			r.Post("/", h.CreateCourse)
			r.Get("/{id}", h.GetCourse)
			r.Put("/{id}", h.UpdateCourse)
			r.Delete("/{id}", h.DeleteCourse)
		})

		// Participants
		r.Route("/participants", func(r chi.Router) {
			r.Get("/", h.ListParticipants)
			r.Post("/", h.CreateParticipant)
			r.Get("/{id}", h.GetParticipant)
			r.Put("/{id}", h.UpdateParticipant)
			r.Delete("/{id}", h.DeleteParticipant)
		})

		// Auth setup
		emailer := services.NewSMTPEmailer(services.SMTPConfig{
			Host:     cfg.SMTPHost,
			Port:     cfg.SMTPPort,
			Username: cfg.SMTPUser,
			Password: cfg.SMTPPass,
			From:     cfg.EmailFrom,
		})
		authHandler := handlers.NewAuthHandler(pool, []byte(cfg.JWTSecret), cfg.JWTExpiryHours, emailer)

		// Auth routes
		r.Post("/auth/register", authHandler.Register)
		r.Post("/auth/login", authHandler.Login)
		r.Post("/auth/reset-password", authHandler.ResetPassword)

		// Protected routes
		r.Group(func(r chi.Router) {
			r.Use(middleware.RequireAuth([]byte(cfg.JWTSecret)))

			r.Get("/persons/{id}", h.GetPerson) // example protected route

			r.Group(func(r chi.Router) {
				r.Use(middleware.RequireRole("admin"))
				r.Post("/courses", h.CreateCourse)
			})
		})
	})

	// Start server
	startServer(r)
}

// startServer starts the HTTP server and supports graceful shutdown
func startServer(handler http.Handler) {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8081"
	}
	addr := ":" + port

	srv := &http.Server{
		Addr:         addr,
		Handler:      handler,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	go func() {
		log.Printf("🚀 Server running on %s\n", addr)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("❌ listen: %s\n", err)
		}
	}()

	// Graceful shutdown
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)
	<-stop
	log.Println("🛑 Shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Fatalf("Server forced to shutdown: %v", err)
	}

	log.Println("✅ Server exited gracefully")
}
