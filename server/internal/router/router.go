package router

import (
	"log/slog"
	"net/http"

	"github.com/go-chi/chi/v5"
	chimiddleware "github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"

	"github.com/shendrong/fullstack-go/server/internal/handler"
	"github.com/shendrong/fullstack-go/server/internal/middleware"
	"github.com/shendrong/fullstack-go/server/internal/service"
)

// New creates and configures the chi router with all routes and middleware.
func New(
	authHandler *handler.AuthHandler,
	healthHandler *handler.HealthHandler,
	authService *service.AuthService,
	logger *slog.Logger,
) http.Handler {
	r := chi.NewRouter()

	// Global middleware stack.
	r.Use(chimiddleware.RequestID)
	r.Use(chimiddleware.RealIP)
	r.Use(middleware.Logger(logger))
	r.Use(chimiddleware.Recoverer)

	// CORS configuration.
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{"http://localhost:3000", "http://localhost:5173"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-Requested-With"},
		ExposedHeaders:   []string{"Link"},
		AllowCredentials: true,
		MaxAge:           300,
	}))

	// API v1 routes.
	r.Route("/api/v1", func(r chi.Router) {
		// Health check (public).
		r.Get("/health", healthHandler.Health)

		// Auth routes (public).
		r.Route("/auth", func(r chi.Router) {
			r.Post("/register", authHandler.Register)
			r.Post("/login", authHandler.Login)

			// Protected auth routes.
			r.Group(func(r chi.Router) {
				r.Use(middleware.Auth(authService))
				r.Get("/me", authHandler.Me)
			})
		})

		// Protected routes.
		r.Group(func(r chi.Router) {
			r.Use(middleware.Auth(authService))
			r.Post("/upload", healthHandler.Upload)
		})
	})

	// Serve uploaded files statically.
	fileServer := http.StripPrefix("/uploads/", http.FileServer(http.Dir("./uploads")))
	r.Handle("/uploads/*", fileServer)

	return r
}
