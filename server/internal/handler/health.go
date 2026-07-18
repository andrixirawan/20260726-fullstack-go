package handler

import (
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/shendrong/fullstack-go/server/internal/config"
	"github.com/shendrong/fullstack-go/server/internal/model"
)

// HealthHandler handles health check and utility endpoints.
type HealthHandler struct {
	pool      *pgxpool.Pool
	uploadCfg *config.UploadConfig
	logger    *slog.Logger
}

// NewHealthHandler creates a new HealthHandler.
func NewHealthHandler(pool *pgxpool.Pool, uploadCfg *config.UploadConfig, logger *slog.Logger) *HealthHandler {
	return &HealthHandler{
		pool:      pool,
		uploadCfg: uploadCfg,
		logger:    logger,
	}
}

// Health returns the health status of the server.
//
//	@Summary		Health check
//	@Description	Returns server and database connection status
//	@Tags			system
//	@Produce		json
//	@Success		200	{object}	map[string]string
//	@Router			/health [get]
func (h *HealthHandler) Health(w http.ResponseWriter, r *http.Request) {
	dbStatus := "up"
	if err := h.pool.Ping(r.Context()); err != nil {
		dbStatus = "down"
		h.logger.Error("database health check failed", slog.Any("error", err))
	}

	writeJSON(w, http.StatusOK, map[string]string{
		"status":   "ok",
		"database": dbStatus,
	})
}

// Upload handles file uploads.
//
//	@Summary		Upload a file
//	@Description	Upload a file (image or document). Returns the file URL. Max size 10MB.
//	@Tags			files
//	@Accept			multipart/form-data
//	@Produce		json
//	@Security		BearerAuth
//	@Param			file	formData	file				true	"File to upload"
//	@Success		201		{object}	model.UploadResponse
//	@Failure		400		{object}	model.ErrorResponse
//	@Failure		401		{object}	model.ErrorResponse
//	@Failure		500		{object}	model.ErrorResponse
//	@Router			/upload [post]
func (h *HealthHandler) Upload(w http.ResponseWriter, r *http.Request) {
	// Limit request body size.
	r.Body = http.MaxBytesReader(w, r.Body, h.uploadCfg.MaxSize)

	if err := r.ParseMultipartForm(h.uploadCfg.MaxSize); err != nil {
		writeJSON(w, http.StatusBadRequest, model.ErrorResponse{
			Error: fmt.Sprintf("file too large, max size is %d bytes", h.uploadCfg.MaxSize),
		})
		return
	}

	file, header, err := r.FormFile("file")
	if err != nil {
		writeJSON(w, http.StatusBadRequest, model.ErrorResponse{
			Error: "file field 'file' is required",
		})
		return
	}
	defer file.Close()

	// Validate file extension.
	ext := strings.ToLower(filepath.Ext(header.Filename))
	allowedExts := strings.Split(h.uploadCfg.AllowExt, ",")
	allowed := false
	for _, a := range allowedExts {
		if strings.TrimSpace(a) == ext {
			allowed = true
			break
		}
	}
	if !allowed {
		writeJSON(w, http.StatusBadRequest, model.ErrorResponse{
			Error: fmt.Sprintf("file extension '%s' is not allowed, allowed: %s", ext, h.uploadCfg.AllowExt),
		})
		return
	}

	// Ensure upload directory exists.
	if err := os.MkdirAll(h.uploadCfg.Dir, 0o755); err != nil {
		h.logger.Error("failed to create upload directory", slog.Any("error", err))
		writeJSON(w, http.StatusInternalServerError, model.ErrorResponse{
			Error: "internal server error",
		})
		return
	}

	// Generate unique filename to prevent collisions.
	newFilename := uuid.New().String() + ext
	destPath := filepath.Join(h.uploadCfg.Dir, newFilename)

	dst, err := os.Create(destPath)
	if err != nil {
		h.logger.Error("failed to create file", slog.Any("error", err))
		writeJSON(w, http.StatusInternalServerError, model.ErrorResponse{
			Error: "internal server error",
		})
		return
	}
	defer dst.Close()

	written, err := io.Copy(dst, file)
	if err != nil {
		h.logger.Error("failed to save file", slog.Any("error", err))
		writeJSON(w, http.StatusInternalServerError, model.ErrorResponse{
			Error: "internal server error",
		})
		return
	}

	h.logger.Info("file uploaded",
		slog.String("filename", newFilename),
		slog.Int64("size", written),
	)

	writeJSON(w, http.StatusCreated, model.UploadResponse{
		Filename: newFilename,
		URL:      "/uploads/" + newFilename,
		Size:     written,
	})
}
