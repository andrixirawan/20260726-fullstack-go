package model

import (
	"time"

	"github.com/google/uuid"
)

// Tag represents a blog tag.
type Tag struct {
	ID        uuid.UUID `json:"id"`
	Name      string    `json:"name"`
	Slug      string    `json:"slug"`
	CreatedAt time.Time `json:"created_at"`
}

// CreateTagRequest represents the request body for creating a tag.
type CreateTagRequest struct {
	Name string `json:"name" validate:"required,min=1,max=50"`
}
