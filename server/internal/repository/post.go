package repository

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/shendrong/fullstack-go/server/internal/model"
)

// Common post repository errors.
var (
	ErrPostNotFound = errors.New("post not found")
	ErrSlugConflict = errors.New("a post with this slug already exists")
)

// PostRepository handles post database operations.
type PostRepository struct {
	pool *pgxpool.Pool
}

// NewPostRepository creates a new PostRepository.
func NewPostRepository(pool *pgxpool.Pool) *PostRepository {
	return &PostRepository{pool: pool}
}

// Create inserts a new post into the database.
func (r *PostRepository) Create(ctx context.Context, post *model.Post) error {
	query := `
		INSERT INTO posts (author_id, category_id, title, slug, excerpt, content, cover_image, status, published_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
		RETURNING id, view_count, created_at, updated_at`

	err := r.pool.QueryRow(ctx, query,
		post.AuthorID,
		post.CategoryID,
		post.Title,
		post.Slug,
		post.Excerpt,
		post.Content,
		post.CoverImage,
		post.Status,
		post.PublishedAt,
	).Scan(&post.ID, &post.ViewCount, &post.CreatedAt, &post.UpdatedAt)

	if err != nil {
		if isDuplicateKeyError(err) {
			return ErrSlugConflict
		}
		return fmt.Errorf("creating post: %w", err)
	}

	return nil
}

// GetByID retrieves a post by its UUID.
func (r *PostRepository) GetByID(ctx context.Context, id uuid.UUID) (*model.Post, error) {
	query := `
		SELECT id, author_id, category_id, title, slug, excerpt, content, cover_image,
		       status, published_at, view_count, created_at, updated_at
		FROM posts
		WHERE id = $1`

	post := &model.Post{}
	err := r.pool.QueryRow(ctx, query, id).Scan(
		&post.ID, &post.AuthorID, &post.CategoryID, &post.Title, &post.Slug,
		&post.Excerpt, &post.Content, &post.CoverImage, &post.Status,
		&post.PublishedAt, &post.ViewCount, &post.CreatedAt, &post.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrPostNotFound
		}
		return nil, fmt.Errorf("getting post by id: %w", err)
	}

	return post, nil
}

// GetBySlug retrieves a post by its slug.
func (r *PostRepository) GetBySlug(ctx context.Context, slug string) (*model.Post, error) {
	query := `
		SELECT id, author_id, category_id, title, slug, excerpt, content, cover_image,
		       status, published_at, view_count, created_at, updated_at
		FROM posts
		WHERE slug = $1`

	post := &model.Post{}
	err := r.pool.QueryRow(ctx, query, slug).Scan(
		&post.ID, &post.AuthorID, &post.CategoryID, &post.Title, &post.Slug,
		&post.Excerpt, &post.Content, &post.CoverImage, &post.Status,
		&post.PublishedAt, &post.ViewCount, &post.CreatedAt, &post.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrPostNotFound
		}
		return nil, fmt.Errorf("getting post by slug: %w", err)
	}

	return post, nil
}

// List returns a paginated, filtered list of posts.
func (r *PostRepository) List(ctx context.Context, q model.PostListQuery) ([]*model.Post, int64, error) {
	args := []any{}
	idx := 1

	conditions := []string{"1=1"}

	if q.Status != "" {
		conditions = append(conditions, fmt.Sprintf("p.status = $%d", idx))
		args = append(args, q.Status)
		idx++
	}
	if q.CategoryID != nil {
		conditions = append(conditions, fmt.Sprintf("p.category_id = $%d", idx))
		args = append(args, q.CategoryID)
		idx++
	}
	if q.AuthorID != nil {
		conditions = append(conditions, fmt.Sprintf("p.author_id = $%d", idx))
		args = append(args, q.AuthorID)
		idx++
	}
	if q.Search != "" {
		conditions = append(conditions, fmt.Sprintf("(p.title ILIKE $%d OR p.excerpt ILIKE $%d)", idx, idx))
		args = append(args, "%"+q.Search+"%")
		idx++
	}
	if q.TagID != nil {
		conditions = append(conditions, fmt.Sprintf("EXISTS (SELECT 1 FROM post_tags pt WHERE pt.post_id = p.id AND pt.tag_id = $%d)", idx))
		args = append(args, q.TagID)
		idx++
	}

	where := strings.Join(conditions, " AND ")

	// Count query.
	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM posts p WHERE %s", where)
	var total int64
	if err := r.pool.QueryRow(ctx, countQuery, args...).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("counting posts: %w", err)
	}

	// List query with pagination.
	offset := (q.Page - 1) * q.PageSize
	listQuery := fmt.Sprintf(`
		SELECT p.id, p.author_id, p.category_id, p.title, p.slug, p.excerpt, p.content,
		       p.cover_image, p.status, p.published_at, p.view_count, p.created_at, p.updated_at
		FROM posts p
		WHERE %s
		ORDER BY p.created_at DESC
		LIMIT $%d OFFSET $%d`, where, idx, idx+1)

	args = append(args, q.PageSize, offset)

	rows, err := r.pool.Query(ctx, listQuery, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("listing posts: %w", err)
	}
	defer rows.Close()

	posts := []*model.Post{}
	for rows.Next() {
		p := &model.Post{}
		if err := rows.Scan(
			&p.ID, &p.AuthorID, &p.CategoryID, &p.Title, &p.Slug, &p.Excerpt, &p.Content,
			&p.CoverImage, &p.Status, &p.PublishedAt, &p.ViewCount, &p.CreatedAt, &p.UpdatedAt,
		); err != nil {
			return nil, 0, fmt.Errorf("scanning post row: %w", err)
		}
		posts = append(posts, p)
	}

	return posts, total, nil
}

// Update updates an existing post.
func (r *PostRepository) Update(ctx context.Context, post *model.Post) error {
	query := `
		UPDATE posts
		SET category_id = $1, title = $2, slug = $3, excerpt = $4, content = $5,
		    cover_image = $6, status = $7, published_at = $8, updated_at = NOW()
		WHERE id = $9
		RETURNING updated_at`

	err := r.pool.QueryRow(ctx, query,
		post.CategoryID,
		post.Title,
		post.Slug,
		post.Excerpt,
		post.Content,
		post.CoverImage,
		post.Status,
		post.PublishedAt,
		post.ID,
	).Scan(&post.UpdatedAt)

	if err != nil {
		if isDuplicateKeyError(err) {
			return ErrSlugConflict
		}
		return fmt.Errorf("updating post: %w", err)
	}

	return nil
}

// Delete removes a post by ID.
func (r *PostRepository) Delete(ctx context.Context, id uuid.UUID) error {
	query := `DELETE FROM posts WHERE id = $1`
	ct, err := r.pool.Exec(ctx, query, id)
	if err != nil {
		return fmt.Errorf("deleting post: %w", err)
	}
	if ct.RowsAffected() == 0 {
		return ErrPostNotFound
	}
	return nil
}

// IncrementViewCount atomically increments the view count for a post.
func (r *PostRepository) IncrementViewCount(ctx context.Context, id uuid.UUID) error {
	query := `UPDATE posts SET view_count = view_count + 1 WHERE id = $1`
	_, err := r.pool.Exec(ctx, query, id)
	if err != nil {
		return fmt.Errorf("incrementing view count: %w", err)
	}
	return nil
}

// SetTags replaces all tags for a post (delete + insert in a transaction).
func (r *PostRepository) SetTags(ctx context.Context, postID uuid.UUID, tagIDs []uuid.UUID) error {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("beginning transaction: %w", err)
	}
	defer tx.Rollback(ctx) //nolint:errcheck

	if _, err := tx.Exec(ctx, `DELETE FROM post_tags WHERE post_id = $1`, postID); err != nil {
		return fmt.Errorf("clearing post tags: %w", err)
	}

	for _, tagID := range tagIDs {
		if _, err := tx.Exec(ctx, `INSERT INTO post_tags (post_id, tag_id) VALUES ($1, $2) ON CONFLICT DO NOTHING`, postID, tagID); err != nil {
			return fmt.Errorf("inserting post tag: %w", err)
		}
	}

	return tx.Commit(ctx)
}

// GetTagsByPostID returns all tags for a given post.
func (r *PostRepository) GetTagsByPostID(ctx context.Context, postID uuid.UUID) ([]model.Tag, error) {
	query := `
		SELECT t.id, t.name, t.slug, t.created_at
		FROM tags t
		INNER JOIN post_tags pt ON pt.tag_id = t.id
		WHERE pt.post_id = $1
		ORDER BY t.name`

	rows, err := r.pool.Query(ctx, query, postID)
	if err != nil {
		return nil, fmt.Errorf("getting tags by post id: %w", err)
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

// SlugExists checks if a slug is already taken, optionally excluding a specific post ID.
func (r *PostRepository) SlugExists(ctx context.Context, slug string, excludeID *uuid.UUID) (bool, error) {
	query := `SELECT EXISTS(SELECT 1 FROM posts WHERE slug = $1 AND ($2::uuid IS NULL OR id != $2))`
	var exists bool
	if err := r.pool.QueryRow(ctx, query, slug, excludeID).Scan(&exists); err != nil {
		return false, fmt.Errorf("checking slug existence: %w", err)
	}
	return exists, nil
}
