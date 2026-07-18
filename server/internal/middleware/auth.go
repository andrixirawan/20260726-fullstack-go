package middleware

import (
	"context"
	"net/http"
	"strings"

	"github.com/google/uuid"

	"github.com/shendrong/fullstack-go/server/internal/model"
	"github.com/shendrong/fullstack-go/server/internal/service"
)

// contextKey is an unexported type for context keys to avoid collisions.
type contextKey string

const userIDKey contextKey = "userID"

// Auth returns a middleware that validates JWT tokens from the Authorization header.
func Auth(authService *service.AuthService) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			authHeader := r.Header.Get("Authorization")
			if authHeader == "" {
				writeJSON(w, http.StatusUnauthorized, model.ErrorResponse{
					Error: "authorization header is required",
				})
				return
			}

			// Expect "Bearer <token>" format.
			parts := strings.SplitN(authHeader, " ", 2)
			if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
				writeJSON(w, http.StatusUnauthorized, model.ErrorResponse{
					Error: "authorization header must be in 'Bearer <token>' format",
				})
				return
			}

			tokenStr := parts[1]

			userID, err := authService.ValidateToken(tokenStr)
			if err != nil {
				writeJSON(w, http.StatusUnauthorized, model.ErrorResponse{
					Error: "invalid or expired token",
				})
				return
			}

			// Store user ID in context for downstream handlers.
			ctx := context.WithValue(r.Context(), userIDKey, userID)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// UserIDFromContext extracts the user ID from the request context.
func UserIDFromContext(ctx context.Context) (uuid.UUID, bool) {
	userID, ok := ctx.Value(userIDKey).(uuid.UUID)
	return userID, ok
}
