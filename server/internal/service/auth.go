package service

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"

	"github.com/shendrong/fullstack-go/server/internal/config"
	"github.com/shendrong/fullstack-go/server/internal/model"
	"github.com/shendrong/fullstack-go/server/internal/repository"
)

// Common service errors.
var (
	ErrInvalidCredentials = errors.New("invalid email or password")
	ErrUserInactive       = errors.New("user account is inactive")
)

// AuthService handles authentication business logic.
type AuthService struct {
	userRepo  *repository.UserRepository
	jwtCfg    *config.JWTConfig
	uploadDir string
}

// NewAuthService creates a new AuthService.
func NewAuthService(userRepo *repository.UserRepository, jwtCfg *config.JWTConfig, uploadDir string) *AuthService {
	return &AuthService{
		userRepo:  userRepo,
		jwtCfg:    jwtCfg,
		uploadDir: uploadDir,
	}
}

// Register creates a new user with a hashed password.
func (s *AuthService) Register(ctx context.Context, req *model.RegisterRequest) (*model.AuthResponse, error) {
	// Hash the password with bcrypt.
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, fmt.Errorf("hashing password: %w", err)
	}

	user := &model.User{
		Email:    req.Email,
		Password: string(hashedPassword),
		FullName: req.FullName,
		IsActive: true,
	}

	if err := s.userRepo.Create(ctx, user); err != nil {
		if errors.Is(err, repository.ErrUserAlreadyExists) {
			return nil, repository.ErrUserAlreadyExists
		}
		return nil, fmt.Errorf("creating user: %w", err)
	}

	// Generate JWT token.
	token, err := s.generateToken(user.ID)
	if err != nil {
		return nil, fmt.Errorf("generating token: %w", err)
	}

	return &model.AuthResponse{
		Token: token,
		User:  user.ToResponse(),
	}, nil
}

// Login authenticates a user and returns a JWT token.
func (s *AuthService) Login(ctx context.Context, req *model.LoginRequest) (*model.AuthResponse, error) {
	user, err := s.userRepo.GetByEmail(ctx, req.Email)
	if err != nil {
		if errors.Is(err, repository.ErrUserNotFound) {
			return nil, ErrInvalidCredentials
		}
		return nil, fmt.Errorf("getting user: %w", err)
	}

	if !user.IsActive {
		return nil, ErrUserInactive
	}

	// Compare password with hash.
	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password)); err != nil {
		return nil, ErrInvalidCredentials
	}

	// Generate JWT token.
	token, err := s.generateToken(user.ID)
	if err != nil {
		return nil, fmt.Errorf("generating token: %w", err)
	}

	return &model.AuthResponse{
		Token: token,
		User:  user.ToResponse(),
	}, nil
}

// GetCurrentUser retrieves the authenticated user by ID.
func (s *AuthService) GetCurrentUser(ctx context.Context, userID uuid.UUID) (*model.UserResponse, error) {
	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		if errors.Is(err, repository.ErrUserNotFound) {
			return nil, repository.ErrUserNotFound
		}
		return nil, fmt.Errorf("getting user: %w", err)
	}

	resp := user.ToResponse()
	return &resp, nil
}

// UpdateProfile updates the profile details of the authenticated user.
func (s *AuthService) UpdateProfile(ctx context.Context, userID uuid.UUID, req *model.UpdateProfileRequest) (*model.UserResponse, error) {
	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return nil, err
	}

	oldAvatarURL := user.AvatarURL

	if req.FullName != nil {
		user.FullName = *req.FullName
	}

	avatarChanged := false
	if req.AvatarURL != nil && *req.AvatarURL != oldAvatarURL {
		user.AvatarURL = *req.AvatarURL
		avatarChanged = true
	}

	if err := s.userRepo.Update(ctx, user); err != nil {
		return nil, fmt.Errorf("updating user profile: %w", err)
	}

	// If the avatar was changed/deleted and there was an old avatar, delete the old file from disk.
	if avatarChanged && oldAvatarURL != "" {
		if filename := extractFilename(oldAvatarURL); filename != "" {
			oldFilePath := filepath.Join(s.uploadDir, filename)
			// Ignore error if file doesn't exist
			_ = os.Remove(oldFilePath)
		}
	}

	resp := user.ToResponse()
	return &resp, nil
}

// extractFilename gets the physical filename from the avatar URL path.
func extractFilename(avatarURL string) string {
	if avatarURL == "" {
		return ""
	}
	const separator = "/uploads/"
	idx := strings.Index(avatarURL, separator)
	if idx == -1 {
		return ""
	}
	return avatarURL[idx+len(separator):]
}

// generateToken creates a signed JWT token for the given user ID.
func (s *AuthService) generateToken(userID uuid.UUID) (string, error) {
	now := time.Now()

	claims := jwt.RegisteredClaims{
		Subject:   userID.String(),
		Issuer:    s.jwtCfg.Issuer,
		IssuedAt:  jwt.NewNumericDate(now),
		ExpiresAt: jwt.NewNumericDate(now.Add(s.jwtCfg.Expiration)),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	tokenStr, err := token.SignedString([]byte(s.jwtCfg.Secret))
	if err != nil {
		return "", fmt.Errorf("signing token: %w", err)
	}

	return tokenStr, nil
}

// ValidateToken parses and validates a JWT token, returning the user ID.
func (s *AuthService) ValidateToken(tokenStr string) (uuid.UUID, error) {
	token, err := jwt.ParseWithClaims(tokenStr, &jwt.RegisteredClaims{}, func(token *jwt.Token) (any, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(s.jwtCfg.Secret), nil
	})

	if err != nil {
		return uuid.Nil, fmt.Errorf("parsing token: %w", err)
	}

	claims, ok := token.Claims.(*jwt.RegisteredClaims)
	if !ok || !token.Valid {
		return uuid.Nil, errors.New("invalid token claims")
	}

	userID, err := uuid.Parse(claims.Subject)
	if err != nil {
		return uuid.Nil, fmt.Errorf("parsing user ID from token: %w", err)
	}

	return userID, nil
}
