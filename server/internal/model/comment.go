package model

import (
	"time"

	"github.com/google/uuid"
)

// Comment represents a blog comment (supports threading via ParentID).
type Comment struct {
	ID        uuid.UUID  `json:"id"`
	PostID    uuid.UUID  `json:"post_id"`
	AuthorID  uuid.UUID  `json:"author_id"`
	ParentID  *uuid.UUID `json:"parent_id"`
	Content   string     `json:"content"`
	IsDeleted bool       `json:"is_deleted"`
	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt time.Time  `json:"updated_at"`
}

// CommentResponse represents the public comment response with author info and nested replies.
type CommentResponse struct {
	ID        uuid.UUID         `json:"id"`
	PostID    uuid.UUID         `json:"post_id"`
	Author    UserResponse      `json:"author"`
	ParentID  *uuid.UUID        `json:"parent_id"`
	Content   string            `json:"content"`
	IsDeleted bool              `json:"is_deleted"`
	Replies   []CommentResponse `json:"replies,omitempty"`
	CreatedAt time.Time         `json:"created_at"`
	UpdatedAt time.Time         `json:"updated_at"`
}

// CreateCommentRequest represents the request body for adding a comment.
type CreateCommentRequest struct {
	Content  string     `json:"content"   validate:"required,min=1,max=2000"`
	ParentID *uuid.UUID `json:"parent_id" validate:"omitempty"`
}

// UpdateCommentRequest represents the request body for editing a comment.
type UpdateCommentRequest struct {
	Content string `json:"content" validate:"required,min=1,max=2000"`
}
