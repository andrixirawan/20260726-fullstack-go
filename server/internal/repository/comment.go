package repository

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/shendrong/fullstack-go/server/internal/model"
)

// Common comment repository errors.
var (
	ErrCommentNotFound = errors.New("comment not found")
)

// CommentRepository handles comment database operations.
type CommentRepository struct {
	pool *pgxpool.Pool
}

// NewCommentRepository creates a new CommentRepository.
func NewCommentRepository(pool *pgxpool.Pool) *CommentRepository {
	return &CommentRepository{pool: pool}
}

// Create inserts a new comment.
func (r *CommentRepository) Create(ctx context.Context, comment *model.Comment) error {
	query := `
		INSERT INTO comments (post_id, author_id, parent_id, content)
		VALUES ($1, $2, $3, $4)
		RETURNING id, is_deleted, created_at, updated_at`

	err := r.pool.QueryRow(ctx, query,
		comment.PostID,
		comment.AuthorID,
		comment.ParentID,
		comment.Content,
	).Scan(&comment.ID, &comment.IsDeleted, &comment.CreatedAt, &comment.UpdatedAt)

	if err != nil {
		return fmt.Errorf("creating comment: %w", err)
	}
	return nil
}

// GetByID retrieves a comment by UUID.
func (r *CommentRepository) GetByID(ctx context.Context, id uuid.UUID) (*model.Comment, error) {
	query := `
		SELECT id, post_id, author_id, parent_id, content, is_deleted, created_at, updated_at
		FROM comments
		WHERE id = $1`

	c := &model.Comment{}
	err := r.pool.QueryRow(ctx, query, id).Scan(
		&c.ID, &c.PostID, &c.AuthorID, &c.ParentID, &c.Content,
		&c.IsDeleted, &c.CreatedAt, &c.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrCommentNotFound
		}
		return nil, fmt.Errorf("getting comment by id: %w", err)
	}
	return c, nil
}

// ListByPostID retrieves all non-deleted comments for a post (flat list, ordered for nesting).
func (r *CommentRepository) ListByPostID(ctx context.Context, postID uuid.UUID) ([]model.Comment, error) {
	query := `
		SELECT id, post_id, author_id, parent_id, content, is_deleted, created_at, updated_at
		FROM comments
		WHERE post_id = $1
		ORDER BY created_at ASC`

	rows, err := r.pool.Query(ctx, query, postID)
	if err != nil {
		return nil, fmt.Errorf("listing comments: %w", err)
	}
	defer rows.Close()

	comments := []model.Comment{}
	for rows.Next() {
		c := model.Comment{}
		if err := rows.Scan(
			&c.ID, &c.PostID, &c.AuthorID, &c.ParentID, &c.Content,
			&c.IsDeleted, &c.CreatedAt, &c.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("scanning comment: %w", err)
		}
		comments = append(comments, c)
	}
	return comments, nil
}

// Update updates a comment's content.
func (r *CommentRepository) Update(ctx context.Context, comment *model.Comment) error {
	query := `
		UPDATE comments SET content = $1, updated_at = NOW()
		WHERE id = $2
		RETURNING updated_at`

	err := r.pool.QueryRow(ctx, query, comment.Content, comment.ID).Scan(&comment.UpdatedAt)
	if err != nil {
		return fmt.Errorf("updating comment: %w", err)
	}
	return nil
}

// SoftDelete marks a comment as deleted without removing it (preserves thread structure).
func (r *CommentRepository) SoftDelete(ctx context.Context, id uuid.UUID) error {
	query := `UPDATE comments SET is_deleted = true, content = '[deleted]', updated_at = NOW() WHERE id = $1`
	ct, err := r.pool.Exec(ctx, query, id)
	if err != nil {
		return fmt.Errorf("soft deleting comment: %w", err)
	}
	if ct.RowsAffected() == 0 {
		return ErrCommentNotFound
	}
	return nil
}
