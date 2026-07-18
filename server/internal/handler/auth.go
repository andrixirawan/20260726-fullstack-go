package handler

import (
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"

	"github.com/shendrong/fullstack-go/server/internal/middleware"
	"github.com/shendrong/fullstack-go/server/internal/model"
	"github.com/shendrong/fullstack-go/server/internal/repository"
	"github.com/shendrong/fullstack-go/server/internal/service"
	"github.com/shendrong/fullstack-go/server/internal/validator"
)

// AuthHandler handles authentication HTTP endpoints.
type AuthHandler struct {
	authService *service.AuthService
	validator   *validator.Validator
	logger      *slog.Logger
}

// NewAuthHandler creates a new AuthHandler.
func NewAuthHandler(authService *service.AuthService, v *validator.Validator, logger *slog.Logger) *AuthHandler {
	return &AuthHandler{
		authService: authService,
		validator:   v,
		logger:      logger,
	}
}

// Register handles user registration.
//
//	@Summary		Register a new user
//	@Description	Create a new user account with email and password
//	@Tags			auth
//	@Accept			json
//	@Produce		json
//	@Param			request	body		model.RegisterRequest	true	"Registration payload"
//	@Success		201		{object}	model.AuthResponse
//	@Failure		400		{object}	model.ErrorResponse
//	@Failure		409		{object}	model.ErrorResponse
//	@Failure		500		{object}	model.ErrorResponse
//	@Router			/auth/register [post]
func (h *AuthHandler) Register(w http.ResponseWriter, r *http.Request) {
	var req model.RegisterRequest
	if errResp := h.validator.DecodeAndValidate(r, &req); errResp != nil {
		writeJSON(w, http.StatusBadRequest, errResp)
		return
	}

	resp, err := h.authService.Register(r.Context(), &req)
	if err != nil {
		if errors.Is(err, repository.ErrUserAlreadyExists) {
			writeJSON(w, http.StatusConflict, model.ErrorResponse{
				Error: "a user with this email already exists",
			})
			return
		}
		h.logger.Error("failed to register user", slog.Any("error", err))
		writeJSON(w, http.StatusInternalServerError, model.ErrorResponse{
			Error: "internal server error",
		})
		return
	}

	h.logger.Info("user registered", slog.String("email", req.Email))
	writeJSON(w, http.StatusCreated, resp)
}

// Login handles user login.
//
//	@Summary		Login user
//	@Description	Authenticate with email and password, returns JWT token
//	@Tags			auth
//	@Accept			json
//	@Produce		json
//	@Param			request	body		model.LoginRequest	true	"Login payload"
//	@Success		200		{object}	model.AuthResponse
//	@Failure		400		{object}	model.ErrorResponse
//	@Failure		401		{object}	model.ErrorResponse
//	@Failure		403		{object}	model.ErrorResponse
//	@Failure		500		{object}	model.ErrorResponse
//	@Router			/auth/login [post]
func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	var req model.LoginRequest
	if errResp := h.validator.DecodeAndValidate(r, &req); errResp != nil {
		writeJSON(w, http.StatusBadRequest, errResp)
		return
	}

	resp, err := h.authService.Login(r.Context(), &req)
	if err != nil {
		if errors.Is(err, service.ErrInvalidCredentials) {
			writeJSON(w, http.StatusUnauthorized, model.ErrorResponse{
				Error: "invalid email or password",
			})
			return
		}
		if errors.Is(err, service.ErrUserInactive) {
			writeJSON(w, http.StatusForbidden, model.ErrorResponse{
				Error: "user account is inactive",
			})
			return
		}
		h.logger.Error("failed to login user", slog.Any("error", err))
		writeJSON(w, http.StatusInternalServerError, model.ErrorResponse{
			Error: "internal server error",
		})
		return
	}

	h.logger.Info("user logged in", slog.String("email", req.Email))
	writeJSON(w, http.StatusOK, resp)
}

// Me returns the authenticated user's profile.
//
//	@Summary		Get current user
//	@Description	Returns the profile of the currently authenticated user
//	@Tags			auth
//	@Produce		json
//	@Security		BearerAuth
//	@Success		200	{object}	model.UserResponse
//	@Failure		401	{object}	model.ErrorResponse
//	@Failure		404	{object}	model.ErrorResponse
//	@Failure		500	{object}	model.ErrorResponse
//	@Router			/auth/me [get]
func (h *AuthHandler) Me(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		writeJSON(w, http.StatusUnauthorized, model.ErrorResponse{
			Error: "user not authenticated",
		})
		return
	}

	user, err := h.authService.GetCurrentUser(r.Context(), userID)
	if err != nil {
		if errors.Is(err, repository.ErrUserNotFound) {
			writeJSON(w, http.StatusNotFound, model.ErrorResponse{
				Error: "user not found",
			})
			return
		}
		h.logger.Error("failed to get current user", slog.Any("error", err))
		writeJSON(w, http.StatusInternalServerError, model.ErrorResponse{
			Error: "internal server error",
		})
		return
	}

	writeJSON(w, http.StatusOK, user)
}

// UpdateProfile handles updating the authenticated user's profile.
//
//	@Summary		Update current user profile
//	@Description	Allows updating full_name and/or avatar_url of the currently authenticated user
//	@Tags			auth
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			request	body		model.UpdateProfileRequest	true	"Profile update fields"
//	@Success		200		{object}	model.UserResponse
//	@Failure		400		{object}	model.ErrorResponse
//	@Failure		401		{object}	model.ErrorResponse
//	@Failure		500		{object}	model.ErrorResponse
//	@Router			/auth/me [patch]
func (h *AuthHandler) UpdateProfile(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		writeJSON(w, http.StatusUnauthorized, model.ErrorResponse{
			Error: "user not authenticated",
		})
		return
	}

	var req model.UpdateProfileRequest
	if errResp := h.validator.DecodeAndValidate(r, &req); errResp != nil {
		writeJSON(w, http.StatusBadRequest, errResp)
		return
	}

	user, err := h.authService.UpdateProfile(r.Context(), userID, &req)
	if err != nil {
		if errors.Is(err, repository.ErrUserNotFound) {
			writeJSON(w, http.StatusNotFound, model.ErrorResponse{
				Error: "user not found",
			})
			return
		}
		h.logger.Error("failed to update user profile", slog.Any("error", err))
		writeJSON(w, http.StatusInternalServerError, model.ErrorResponse{
			Error: "internal server error",
		})
		return
	}

	h.logger.Info("user profile updated", slog.String("id", userID.String()))
	writeJSON(w, http.StatusOK, user)
}

// writeJSON writes a JSON response with the given status code.
func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(v); err != nil {
		slog.Error("failed to encode JSON response", slog.Any("error", err))
	}
}
