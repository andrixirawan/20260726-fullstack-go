package database

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/shendrong/fullstack-go/server/internal/config"
)

// New creates a new PostgreSQL connection pool with retry logic.
func New(ctx context.Context, cfg *config.DatabaseConfig, logger *slog.Logger) (*pgxpool.Pool, error) {
	poolCfg, err := pgxpool.ParseConfig(cfg.DSN())
	if err != nil {
		return nil, fmt.Errorf("parsing database config: %w", err)
	}

	poolCfg.MaxConns = int32(cfg.MaxConns)
	poolCfg.MinConns = int32(cfg.MinConns)

	var pool *pgxpool.Pool

	// Retry connection with backoff.
	maxRetries := 10
	for i := range maxRetries {
		pool, err = pgxpool.NewWithConfig(ctx, poolCfg)
		if err == nil {
			if pingErr := pool.Ping(ctx); pingErr == nil {
				logger.Info("connected to database",
					slog.String("host", cfg.Host),
					slog.Int("port", cfg.Port),
					slog.String("database", cfg.Name),
				)
				return pool, nil
			}
			pool.Close()
		}

		wait := time.Duration(i+1) * time.Second
		logger.Warn("failed to connect to database, retrying...",
			slog.Int("attempt", i+1),
			slog.Int("max_retries", maxRetries),
			slog.Duration("wait", wait),
			slog.Any("error", err),
		)
		time.Sleep(wait)
	}

	return nil, fmt.Errorf("failed to connect to database after %d attempts: %w", maxRetries, err)
}
