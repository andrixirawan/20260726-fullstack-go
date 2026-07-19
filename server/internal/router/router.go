package router

import (
	"log/slog"
	"net/http"

	"github.com/go-chi/chi/v5"
	chimiddleware "github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	httpSwagger "github.com/swaggo/http-swagger/v2"

	"github.com/shendrong/fullstack-go/server/internal/handler"
	"github.com/shendrong/fullstack-go/server/internal/middleware"
	"github.com/shendrong/fullstack-go/server/internal/service"
)

// New creates and configures the chi router with all routes and middleware.
func New(
	authHandler *handler.AuthHandler,
	healthHandler *handler.HealthHandler,
	postHandler *handler.PostHandler,
	categoryHandler *handler.CategoryHandler,
	tagHandler *handler.TagHandler,
	commentHandler *handler.CommentHandler,
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
		AllowedOrigins:   []string{"http://localhost:3000", "http://localhost:3001", "http://localhost:5173"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-Requested-With"},
		ExposedHeaders:   []string{"Link"},
		AllowCredentials: true,
		MaxAge:           300,
	}))

	// Swagger UI — served directly from the Go server.
	// Access at: http://localhost:8080/swagger/index.html
	r.Get("/swagger/*", httpSwagger.Handler(
		httpSwagger.URL("/swagger/doc.json"),
	))

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
				r.Patch("/me", authHandler.UpdateProfile)
			})
		})

		// Protected routes.
		r.Group(func(r chi.Router) {
			r.Use(middleware.Auth(authService))
			r.Post("/upload", healthHandler.Upload)
		})

		// ── Blog: Posts ──────────────────────────────────────────────────────
		r.Route("/posts", func(r chi.Router) {
			// Public endpoints.
			r.Get("/", postHandler.ListPosts)
			r.Get("/slug/{slug}", postHandler.GetPostBySlug)
			r.Get("/{id}", postHandler.GetPostByID)
			r.Get("/{id}/comments", commentHandler.ListComments)

			// Protected endpoints.
			r.Group(func(r chi.Router) {
				r.Use(middleware.Auth(authService))
				r.Post("/", postHandler.CreatePost)
				r.Put("/{id}", postHandler.UpdatePost)
				r.Delete("/{id}", postHandler.DeletePost)
				r.Patch("/{id}/publish", postHandler.TogglePublish)
				r.Post("/{id}/comments", commentHandler.CreateComment)
			})
		})

		// ── Blog: Comments ───────────────────────────────────────────────────
		r.Route("/comments", func(r chi.Router) {
			r.Group(func(r chi.Router) {
				r.Use(middleware.Auth(authService))
				r.Put("/{id}", commentHandler.UpdateComment)
				r.Delete("/{id}", commentHandler.DeleteComment)
			})
		})

		// ── Blog: Categories ─────────────────────────────────────────────────
		r.Route("/categories", func(r chi.Router) {
			r.Get("/", categoryHandler.ListCategories)
			r.Group(func(r chi.Router) {
				r.Use(middleware.Auth(authService))
				r.Post("/", categoryHandler.CreateCategory)
				r.Put("/{id}", categoryHandler.UpdateCategory)
				r.Delete("/{id}", categoryHandler.DeleteCategory)
			})
		})

		// ── Blog: Tags ───────────────────────────────────────────────────────
		r.Route("/tags", func(r chi.Router) {
			r.Get("/", tagHandler.ListTags)
			r.Group(func(r chi.Router) {
				r.Use(middleware.Auth(authService))
				r.Post("/", tagHandler.CreateTag)
				r.Delete("/{id}", tagHandler.DeleteTag)
			})
		})
	})

	// Serve uploaded files statically.
	fileServer := http.StripPrefix("/uploads/", http.FileServer(http.Dir("./uploads")))
	r.Handle("/uploads/*", fileServer)

	return r
}
