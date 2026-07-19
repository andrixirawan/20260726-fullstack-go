package config

import (
	"fmt"
	"time"

	"github.com/kelseyhightower/envconfig"
)

// Config holds all configuration for the application.
type Config struct {
	Server   ServerConfig
	Database DatabaseConfig
	JWT      JWTConfig
	Upload   UploadConfig
}

// ServerConfig holds HTTP server configuration.
type ServerConfig struct {
	Port         int           `envconfig:"SERVER_PORT" default:"8080"`
	ReadTimeout  time.Duration `envconfig:"SERVER_READ_TIMEOUT" default:"15s"`
	WriteTimeout time.Duration `envconfig:"SERVER_WRITE_TIMEOUT" default:"15s"`
	IdleTimeout  time.Duration `envconfig:"SERVER_IDLE_TIMEOUT" default:"60s"`
}

// DatabaseConfig holds PostgreSQL connection configuration.
type DatabaseConfig struct {
	URL      string `envconfig:"DATABASE_URL"`
	Host     string `envconfig:"DB_HOST" default:"localhost"`
	Port     int    `envconfig:"DB_PORT" default:"5432"`
	User     string `envconfig:"DB_USER" default:"postgres"`
	Password string `envconfig:"DB_PASSWORD" default:"postgres"`
	Name     string `envconfig:"DB_NAME" default:"fullstack"`
	SSLMode  string `envconfig:"DB_SSLMODE" default:"disable"`
	MaxConns int    `envconfig:"DB_MAX_CONNS" default:"25"`
	MinConns int    `envconfig:"DB_MIN_CONNS" default:"5"`
}

// JWTConfig holds JWT token configuration.
type JWTConfig struct {
	Secret     string        `envconfig:"JWT_SECRET" required:"true"`
	Expiration time.Duration `envconfig:"JWT_EXPIRATION" default:"24h"`
	Issuer     string        `envconfig:"JWT_ISSUER" default:"fullstack-go"`
}

// UploadConfig holds file upload configuration.
type UploadConfig struct {
	MaxSize  int64  `envconfig:"UPLOAD_MAX_SIZE" default:"10485760"` // 10MB
	Dir      string `envconfig:"UPLOAD_DIR" default:"./uploads"`
	AllowExt string `envconfig:"UPLOAD_ALLOW_EXT" default:".jpg,.jpeg,.png,.gif,.webp,.pdf"`
}

// DSN returns the PostgreSQL connection string.
func (d *DatabaseConfig) DSN() string {
	if d.URL != "" {
		return d.URL
	}
	return fmt.Sprintf(
		"postgres://%s:%s@%s:%d/%s?sslmode=%s",
		d.User, d.Password, d.Host, d.Port, d.Name, d.SSLMode,
	)
}

// Load reads configuration from environment variables.
func Load() (*Config, error) {
	var cfg Config

	if err := envconfig.Process("", &cfg.Server); err != nil {
		return nil, fmt.Errorf("loading server config: %w", err)
	}
	if err := envconfig.Process("", &cfg.Database); err != nil {
		return nil, fmt.Errorf("loading database config: %w", err)
	}
	if err := envconfig.Process("", &cfg.JWT); err != nil {
		return nil, fmt.Errorf("loading jwt config: %w", err)
	}
	if err := envconfig.Process("", &cfg.Upload); err != nil {
		return nil, fmt.Errorf("loading upload config: %w", err)
	}

	return &cfg, nil
}
