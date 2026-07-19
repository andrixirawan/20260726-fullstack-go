package model

import (
	"time"

	"github.com/google/uuid"
)

// PostStatus represents the publication status of a post.
type PostStatus string

const (
	PostStatusDraft     PostStatus = "draft"
	PostStatusPublished PostStatus = "published"
)

// Post represents a blog post domain model.
type Post struct {
	ID          uuid.UUID  `json:"id"`
	AuthorID    uuid.UUID  `json:"author_id"`
	CategoryID  *uuid.UUID `json:"category_id"`
	Title       string     `json:"title"`
	Slug        string     `json:"slug"`
	Excerpt     string     `json:"excerpt"`
	Content     string     `json:"content"`
	CoverImage  string     `json:"cover_image"`
	Status      PostStatus `json:"status"`
	PublishedAt *time.Time `json:"published_at"`
	ViewCount   int64      `json:"view_count"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
}

// PostResponse represents the public-facing post data returned in API responses.
type PostResponse struct {
	ID          uuid.UUID      `json:"id"`
	Author      UserResponse   `json:"author"`
	Category    *Category      `json:"category,omitempty"`
	Tags        []Tag          `json:"tags"`
	Title       string         `json:"title"`
	Slug        string         `json:"slug"`
	Excerpt     string         `json:"excerpt"`
	Content     string         `json:"content"`
	CoverImage  string         `json:"cover_image"`
	Status      PostStatus     `json:"status"`
	PublishedAt *time.Time     `json:"published_at"`
	ViewCount   int64          `json:"view_count"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
}

// PostListItem represents a lightweight post for list views.
type PostListItem struct {
	ID          uuid.UUID      `json:"id"`
	Author      UserResponse   `json:"author"`
	Category    *Category      `json:"category,omitempty"`
	Tags        []Tag          `json:"tags"`
	Title       string         `json:"title"`
	Slug        string         `json:"slug"`
	Excerpt     string         `json:"excerpt"`
	CoverImage  string         `json:"cover_image"`
	Status      PostStatus     `json:"status"`
	PublishedAt *time.Time     `json:"published_at"`
	ViewCount   int64          `json:"view_count"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
}

// PostListResponse is the paginated list response.
type PostListResponse struct {
	Data       []PostListItem `json:"data"`
	Total      int64          `json:"total"`
	Page       int            `json:"page"`
	PageSize   int            `json:"page_size"`
	TotalPages int            `json:"total_pages"`
}

// CreatePostRequest represents the request body for creating a post.
type CreatePostRequest struct {
	Title      string     `json:"title"       validate:"required,min=1,max=255"`
	Excerpt    string     `json:"excerpt"     validate:"omitempty,max=500"`
	Content    string     `json:"content"     validate:"required,min=1"`
	CoverImage string     `json:"cover_image" validate:"omitempty,max=500"`
	CategoryID *uuid.UUID `json:"category_id" validate:"omitempty"`
	TagIDs     []uuid.UUID `json:"tag_ids"    validate:"omitempty"`
	Status     PostStatus `json:"status"      validate:"omitempty,oneof=draft published"`
}

// UpdatePostRequest represents the request body for updating a post.
type UpdatePostRequest struct {
	Title      *string    `json:"title"       validate:"omitempty,min=1,max=255"`
	Excerpt    *string    `json:"excerpt"     validate:"omitempty,max=500"`
	Content    *string    `json:"content"     validate:"omitempty,min=1"`
	CoverImage *string    `json:"cover_image" validate:"omitempty,max=500"`
	CategoryID *uuid.UUID `json:"category_id" validate:"omitempty"`
	TagIDs     []uuid.UUID `json:"tag_ids"    validate:"omitempty"`
	Status     *PostStatus `json:"status"     validate:"omitempty,oneof=draft published"`
}

// PostListQuery represents query parameters for listing posts.
type PostListQuery struct {
	Page       int        `json:"page"`
	PageSize   int        `json:"page_size"`
	Status     PostStatus `json:"status"`
	CategoryID *uuid.UUID `json:"category_id"`
	TagID      *uuid.UUID `json:"tag_id"`
	AuthorID   *uuid.UUID `json:"author_id"`
	Search     string     `json:"search"`
}
