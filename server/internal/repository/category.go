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

// Common category repository errors.
var (
	ErrCategoryNotFound    = errors.New("category not found")
	ErrCategorySlugConflict = errors.New("a category with this name already exists")
)

// CategoryRepository handles category database operations.
type CategoryRepository struct {
	pool *pgxpool.Pool
}

// NewCategoryRepository creates a new CategoryRepository.
func NewCategoryRepository(pool *pgxpool.Pool) *CategoryRepository {
	return &CategoryRepository{pool: pool}
}

// Create inserts a new category.
func (r *CategoryRepository) Create(ctx context.Context, cat *model.Category) error {
	query := `
		INSERT INTO categories (name, slug, description)
		VALUES ($1, $2, $3)
		RETURNING id, created_at, updated_at`

	err := r.pool.QueryRow(ctx, query, cat.Name, cat.Slug, cat.Description).
		Scan(&cat.ID, &cat.CreatedAt, &cat.UpdatedAt)

	if err != nil {
		if isDuplicateKeyError(err) {
			return ErrCategorySlugConflict
		}
		return fmt.Errorf("creating category: %w", err)
	}
	return nil
}

// GetByID retrieves a category by UUID.
func (r *CategoryRepository) GetByID(ctx context.Context, id uuid.UUID) (*model.Category, error) {
	query := `SELECT id, name, slug, description, created_at, updated_at FROM categories WHERE id = $1`
	cat := &model.Category{}
	err := r.pool.QueryRow(ctx, query, id).Scan(
		&cat.ID, &cat.Name, &cat.Slug, &cat.Description, &cat.CreatedAt, &cat.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrCategoryNotFound
		}
		return nil, fmt.Errorf("getting category by id: %w", err)
	}
	return cat, nil
}

// List returns all categories ordered by name.
func (r *CategoryRepository) List(ctx context.Context) ([]model.Category, error) {
	query := `SELECT id, name, slug, description, created_at, updated_at FROM categories ORDER BY name`
	rows, err := r.pool.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("listing categories: %w", err)
	}
	defer rows.Close()

	cats := []model.Category{}
	for rows.Next() {
		c := model.Category{}
		if err := rows.Scan(&c.ID, &c.Name, &c.Slug, &c.Description, &c.CreatedAt, &c.UpdatedAt); err != nil {
			return nil, fmt.Errorf("scanning category: %w", err)
		}
		cats = append(cats, c)
	}
	return cats, nil
}

// Update updates an existing category.
func (r *CategoryRepository) Update(ctx context.Context, cat *model.Category) error {
	query := `
		UPDATE categories SET name = $1, slug = $2, description = $3, updated_at = NOW()
		WHERE id = $4
		RETURNING updated_at`

	err := r.pool.QueryRow(ctx, query, cat.Name, cat.Slug, cat.Description, cat.ID).
		Scan(&cat.UpdatedAt)
	if err != nil {
		if isDuplicateKeyError(err) {
			return ErrCategorySlugConflict
		}
		return fmt.Errorf("updating category: %w", err)
	}
	return nil
}

// Delete removes a category by ID.
func (r *CategoryRepository) Delete(ctx context.Context, id uuid.UUID) error {
	query := `DELETE FROM categories WHERE id = $1`
	ct, err := r.pool.Exec(ctx, query, id)
	if err != nil {
		return fmt.Errorf("deleting category: %w", err)
	}
	if ct.RowsAffected() == 0 {
		return ErrCategoryNotFound
	}
	return nil
}
