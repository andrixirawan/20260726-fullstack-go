// Package main is the entry point for the Fullstack Go API server.
//
//	@title						Fullstack Go API
//	@version					1.0
//	@description				Production-grade REST API with JWT authentication, file upload, and PostgreSQL.
//	@termsOfService				http://swagger.io/terms/
//
//	@contact.name				API Support
//	@contact.email				support@example.com
//
//	@license.name				MIT
//	@license.url				https://opensource.org/licenses/MIT
//
//	@host						localhost:8080
//	@BasePath					/api/v1
//
//	@securityDefinitions.apikey	BearerAuth
//	@in							header
//	@name						Authorization
//	@description				JWT token. Format: "Bearer {token}"
package main

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/pgx/v5"
	_ "github.com/golang-migrate/migrate/v4/source/file"

	"github.com/shendrong/fullstack-go/server/internal/config"
	"github.com/shendrong/fullstack-go/server/internal/database"
	"github.com/shendrong/fullstack-go/server/internal/handler"
	"github.com/shendrong/fullstack-go/server/internal/repository"
	"github.com/shendrong/fullstack-go/server/internal/router"
	"github.com/shendrong/fullstack-go/server/internal/service"
	"github.com/shendrong/fullstack-go/server/internal/validator"

	// Import generated swagger docs.
	_ "github.com/shendrong/fullstack-go/server/docs"
)

func main() {
	// Initialize structured logger.
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}))
	slog.SetDefault(logger)

	if err := run(logger); err != nil {
		logger.Error("application error", slog.Any("error", err))
		os.Exit(1)
	}
}

func run(logger *slog.Logger) error {
	// Load configuration.
	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("loading config: %w", err)
	}

	logger.Info("configuration loaded",
		slog.Int("server_port", cfg.Server.Port),
		slog.String("db_host", cfg.Database.Host),
		slog.Int("db_port", cfg.Database.Port),
		slog.String("db_name", cfg.Database.Name),
	)

	// Connect to database.
	ctx := context.Background()
	pool, err := database.New(ctx, &cfg.Database, logger)
	if err != nil {
		return fmt.Errorf("connecting to database: %w", err)
	}
	defer pool.Close()

	// Run database migrations.
	if err := runMigrations(cfg, logger); err != nil {
		return fmt.Errorf("running migrations: %w", err)
	}

	// Ensure upload directory exists.
	if err := os.MkdirAll(cfg.Upload.Dir, 0o755); err != nil {
		return fmt.Errorf("creating upload directory: %w", err)
	}

	// Initialize dependencies.
	v := validator.New()

	// User.
	userRepo := repository.NewUserRepository(pool)
	authService := service.NewAuthService(userRepo, &cfg.JWT, cfg.Upload.Dir)

	// Blog: repositories.
	postRepo := repository.NewPostRepository(pool)
	categoryRepo := repository.NewCategoryRepository(pool)
	tagRepo := repository.NewTagRepository(pool)
	commentRepo := repository.NewCommentRepository(pool)

	// Blog: services.
	postService := service.NewPostService(postRepo, categoryRepo, tagRepo, userRepo, cfg.Upload.Dir)
	categoryService := service.NewCategoryService(categoryRepo)
	tagService := service.NewTagService(tagRepo)
	commentService := service.NewCommentService(commentRepo, userRepo)

	// Initialize handlers.
	authHandler := handler.NewAuthHandler(authService, v, logger)
	healthHandler := handler.NewHealthHandler(pool, &cfg.Upload, logger)
	postHandler := handler.NewPostHandler(postService, v, logger)
	categoryHandler := handler.NewCategoryHandler(categoryService, v, logger)
	tagHandler := handler.NewTagHandler(tagService, v, logger)
	commentHandler := handler.NewCommentHandler(commentService, v, logger)

	// Setup router.
	r := router.New(authHandler, healthHandler, postHandler, categoryHandler, tagHandler, commentHandler, authService, logger)

	// Create HTTP server.
	srv := &http.Server{
		Addr:         fmt.Sprintf(":%d", cfg.Server.Port),
		Handler:      r,
		ReadTimeout:  cfg.Server.ReadTimeout,
		WriteTimeout: cfg.Server.WriteTimeout,
		IdleTimeout:  cfg.Server.IdleTimeout,
	}

	// Graceful shutdown.
	shutdownCh := make(chan os.Signal, 1)
	signal.Notify(shutdownCh, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		logger.Info("starting server",
			slog.String("addr", srv.Addr),
			slog.String("swagger", fmt.Sprintf("http://localhost:%d/swagger/index.html", cfg.Server.Port)),
		)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Error("server error", slog.Any("error", err))
			os.Exit(1)
		}
	}()

	// Wait for shutdown signal.
	sig := <-shutdownCh
	logger.Info("shutdown signal received", slog.String("signal", sig.String()))

	// Give outstanding requests 30 seconds to complete.
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		return fmt.Errorf("server shutdown: %w", err)
	}

	logger.Info("server stopped gracefully")
	return nil
}

// runMigrations applies database migrations.
func runMigrations(cfg *config.Config, logger *slog.Logger) error {
	dsn := fmt.Sprintf(
		"pgx5://%s:%s@%s:%d/%s?sslmode=%s",
		cfg.Database.User, cfg.Database.Password, cfg.Database.Host,
		cfg.Database.Port, cfg.Database.Name, cfg.Database.SSLMode,
	)

	m, err := migrate.New("file://migrations", dsn)
	if err != nil {
		return fmt.Errorf("creating migration instance: %w", err)
	}
	defer m.Close()

	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		return fmt.Errorf("applying migrations: %w", err)
	}

	logger.Info("database migrations applied successfully")
	return nil
}
