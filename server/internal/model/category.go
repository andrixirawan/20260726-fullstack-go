package model

import (
	"time"

	"github.com/google/uuid"
)

// Category represents a blog category.
type Category struct {
	ID          uuid.UUID `json:"id"`
	Name        string    `json:"name"`
	Slug        string    `json:"slug"`
	Description string    `json:"description"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// CreateCategoryRequest represents the request body for creating a category.
type CreateCategoryRequest struct {
	Name        string `json:"name"        validate:"required,min=1,max=100"`
	Description string `json:"description" validate:"omitempty,max=500"`
}

// UpdateCategoryRequest represents the request body for updating a category.
type UpdateCategoryRequest struct {
	Name        *string `json:"name"        validate:"omitempty,min=1,max=100"`
	Description *string `json:"description" validate:"omitempty,max=500"`
}
