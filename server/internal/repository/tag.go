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

// Common tag repository errors.
var (
	ErrTagNotFound    = errors.New("tag not found")
	ErrTagNameConflict = errors.New("a tag with this name already exists")
)

// TagRepository handles tag database operations.
type TagRepository struct {
	pool *pgxpool.Pool
}

// NewTagRepository creates a new TagRepository.
func NewTagRepository(pool *pgxpool.Pool) *TagRepository {
	return &TagRepository{pool: pool}
}

// Create inserts a new tag.
func (r *TagRepository) Create(ctx context.Context, tag *model.Tag) error {
	query := `
		INSERT INTO tags (name, slug)
		VALUES ($1, $2)
		RETURNING id, created_at`

	err := r.pool.QueryRow(ctx, query, tag.Name, tag.Slug).Scan(&tag.ID, &tag.CreatedAt)
	if err != nil {
		if isDuplicateKeyError(err) {
			return ErrTagNameConflict
		}
		return fmt.Errorf("creating tag: %w", err)
	}
	return nil
}

// GetByID retrieves a tag by UUID.
func (r *TagRepository) GetByID(ctx context.Context, id uuid.UUID) (*model.Tag, error) {
	query := `SELECT id, name, slug, created_at FROM tags WHERE id = $1`
	tag := &model.Tag{}
	err := r.pool.QueryRow(ctx, query, id).Scan(&tag.ID, &tag.Name, &tag.Slug, &tag.CreatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrTagNotFound
		}
		return nil, fmt.Errorf("getting tag by id: %w", err)
	}
	return tag, nil
}

// List returns all tags ordered by name.
func (r *TagRepository) List(ctx context.Context) ([]model.Tag, error) {
	query := `SELECT id, name, slug, created_at FROM tags ORDER BY name`
	rows, err := r.pool.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("listing tags: %w", err)
	}
	defer rows.Close()

	tags := []model.Tag{}
	for rows.Next() {
		t := model.Tag{}
		if err := rows.Scan(&t.ID, &t.Name, &t.Slug, &t.CreatedAt); err != nil {
			return nil, fmt.Errorf("scanning tag: %w", err)
		}
		tags = append(tags, t)
	}
	return tags, nil
}

// Delete removes a tag by ID.
func (r *TagRepository) Delete(ctx context.Context, id uuid.UUID) error {
	query := `DELETE FROM tags WHERE id = $1`
	ct, err := r.pool.Exec(ctx, query, id)
	if err != nil {
		return fmt.Errorf("deleting tag: %w", err)
	}
	if ct.RowsAffected() == 0 {
		return ErrTagNotFound
	}
	return nil
}
