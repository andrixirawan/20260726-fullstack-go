package service

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"
	"unicode"

	"github.com/google/uuid"

	"github.com/shendrong/fullstack-go/server/internal/model"
	"github.com/shendrong/fullstack-go/server/internal/repository"
)

// Common post service errors.
var (
	ErrPostNotFound      = errors.New("post not found")
	ErrPostForbidden     = errors.New("you are not allowed to modify this post")
	ErrPostSlugConflict  = errors.New("a post with this slug already exists")
)

// PostService handles post business logic.
type PostService struct {
	postRepo     *repository.PostRepository
	categoryRepo *repository.CategoryRepository
	tagRepo      *repository.TagRepository
	userRepo     *repository.UserRepository
	uploadDir    string
}

// NewPostService creates a new PostService.
func NewPostService(
	postRepo *repository.PostRepository,
	categoryRepo *repository.CategoryRepository,
	tagRepo *repository.TagRepository,
	userRepo *repository.UserRepository,
	uploadDir string,
) *PostService {
	return &PostService{
		postRepo:     postRepo,
		categoryRepo: categoryRepo,
		tagRepo:      tagRepo,
		userRepo:     userRepo,
		uploadDir:    uploadDir,
	}
}

// Create creates a new blog post.
func (s *PostService) Create(ctx context.Context, authorID uuid.UUID, req *model.CreatePostRequest) (*model.PostResponse, error) {
	// Generate a unique slug from the title.
	slug, err := s.generateUniqueSlug(ctx, req.Title, nil)
	if err != nil {
		return nil, fmt.Errorf("generating slug: %w", err)
	}

	status := model.PostStatusDraft
	if req.Status != "" {
		status = req.Status
	}

	var publishedAt *time.Time
	if status == model.PostStatusPublished {
		now := time.Now()
		publishedAt = &now
	}

	post := &model.Post{
		AuthorID:    authorID,
		CategoryID:  req.CategoryID,
		Title:       req.Title,
		Slug:        slug,
		Excerpt:     req.Excerpt,
		Content:     req.Content,
		CoverImage:  req.CoverImage,
		Status:      status,
		PublishedAt: publishedAt,
	}

	if err := s.postRepo.Create(ctx, post); err != nil {
		if errors.Is(err, repository.ErrSlugConflict) {
			return nil, ErrPostSlugConflict
		}
		return nil, fmt.Errorf("creating post: %w", err)
	}

	// Associate tags.
	if len(req.TagIDs) > 0 {
		if err := s.postRepo.SetTags(ctx, post.ID, req.TagIDs); err != nil {
			return nil, fmt.Errorf("setting post tags: %w", err)
		}
	}

	return s.buildPostResponse(ctx, post)
}

// GetByID returns a post by its UUID, incrementing the view count.
func (s *PostService) GetByID(ctx context.Context, id uuid.UUID) (*model.PostResponse, error) {
	post, err := s.postRepo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, repository.ErrPostNotFound) {
			return nil, ErrPostNotFound
		}
		return nil, fmt.Errorf("getting post: %w", err)
	}

	// Async-like: ignore error on view count increment.
	_ = s.postRepo.IncrementViewCount(ctx, id)

	return s.buildPostResponse(ctx, post)
}

// GetBySlug returns a post by its slug, incrementing the view count.
func (s *PostService) GetBySlug(ctx context.Context, slug string) (*model.PostResponse, error) {
	post, err := s.postRepo.GetBySlug(ctx, slug)
	if err != nil {
		if errors.Is(err, repository.ErrPostNotFound) {
			return nil, ErrPostNotFound
		}
		return nil, fmt.Errorf("getting post by slug: %w", err)
	}

	_ = s.postRepo.IncrementViewCount(ctx, post.ID)

	return s.buildPostResponse(ctx, post)
}

// List returns a paginated list of posts.
func (s *PostService) List(ctx context.Context, q model.PostListQuery) (*model.PostListResponse, error) {
	if q.Page < 1 {
		q.Page = 1
	}
	if q.PageSize < 1 || q.PageSize > 100 {
		q.PageSize = 10
	}

	posts, total, err := s.postRepo.List(ctx, q)
	if err != nil {
		return nil, fmt.Errorf("listing posts: %w", err)
	}

	items := make([]model.PostListItem, 0, len(posts))
	for _, post := range posts {
		item, err := s.buildPostListItem(ctx, post)
		if err != nil {
			return nil, err
		}
		items = append(items, *item)
	}

	totalPages := int(total) / q.PageSize
	if int(total)%q.PageSize != 0 {
		totalPages++
	}

	return &model.PostListResponse{
		Data:       items,
		Total:      total,
		Page:       q.Page,
		PageSize:   q.PageSize,
		TotalPages: totalPages,
	}, nil
}

// Update updates a post, only if the requestor is the author.
func (s *PostService) Update(ctx context.Context, postID uuid.UUID, requesterID uuid.UUID, req *model.UpdatePostRequest) (*model.PostResponse, error) {
	post, err := s.postRepo.GetByID(ctx, postID)
	if err != nil {
		if errors.Is(err, repository.ErrPostNotFound) {
			return nil, ErrPostNotFound
		}
		return nil, fmt.Errorf("getting post: %w", err)
	}

	if post.AuthorID != requesterID {
		return nil, ErrPostForbidden
	}

	oldCoverImage := post.CoverImage

	if req.Title != nil && *req.Title != post.Title {
		newSlug, err := s.generateUniqueSlug(ctx, *req.Title, &postID)
		if err != nil {
			return nil, fmt.Errorf("generating slug: %w", err)
		}
		post.Title = *req.Title
		post.Slug = newSlug
	}

	if req.Excerpt != nil {
		post.Excerpt = *req.Excerpt
	}
	if req.Content != nil {
		post.Content = *req.Content
	}
	coverChanged := false
	if req.CoverImage != nil {
		if *req.CoverImage != oldCoverImage {
			post.CoverImage = *req.CoverImage
			coverChanged = true
		}
	}
	if req.CategoryID != nil {
		post.CategoryID = req.CategoryID
	}
	if req.Status != nil {
		oldStatus := post.Status
		post.Status = *req.Status
		if *req.Status == model.PostStatusPublished && oldStatus != model.PostStatusPublished && post.PublishedAt == nil {
			now := time.Now()
			post.PublishedAt = &now
		}
	}

	if err := s.postRepo.Update(ctx, post); err != nil {
		if errors.Is(err, repository.ErrSlugConflict) {
			return nil, ErrPostSlugConflict
		}
		return nil, fmt.Errorf("updating post: %w", err)
	}

	// If the cover was changed/deleted and there was an old cover, delete the old file from disk.
	if coverChanged && oldCoverImage != "" {
		if filename := extractFilename(oldCoverImage); filename != "" {
			oldFilePath := filepath.Join(s.uploadDir, filename)
			_ = os.Remove(oldFilePath)
		}
	}

	// Update tags if provided.
	if req.TagIDs != nil {
		if err := s.postRepo.SetTags(ctx, post.ID, req.TagIDs); err != nil {
			return nil, fmt.Errorf("updating post tags: %w", err)
		}
	}

	return s.buildPostResponse(ctx, post)
}

// Delete removes a post if the requester is the author.
func (s *PostService) Delete(ctx context.Context, postID uuid.UUID, requesterID uuid.UUID) error {
	post, err := s.postRepo.GetByID(ctx, postID)
	if err != nil {
		if errors.Is(err, repository.ErrPostNotFound) {
			return ErrPostNotFound
		}
		return fmt.Errorf("getting post: %w", err)
	}

	if post.AuthorID != requesterID {
		return ErrPostForbidden
	}

	oldCoverImage := post.CoverImage

	if err := s.postRepo.Delete(ctx, postID); err != nil {
		return err
	}

	// Delete cover image file from disk if it was stored locally.
	if oldCoverImage != "" {
		if filename := extractFilename(oldCoverImage); filename != "" {
			oldFilePath := filepath.Join(s.uploadDir, filename)
			_ = os.Remove(oldFilePath)
		}
	}

	return nil
}

// TogglePublish toggles the publish status of a post.
func (s *PostService) TogglePublish(ctx context.Context, postID uuid.UUID, requesterID uuid.UUID) (*model.PostResponse, error) {
	post, err := s.postRepo.GetByID(ctx, postID)
	if err != nil {
		if errors.Is(err, repository.ErrPostNotFound) {
			return nil, ErrPostNotFound
		}
		return nil, fmt.Errorf("getting post: %w", err)
	}

	if post.AuthorID != requesterID {
		return nil, ErrPostForbidden
	}

	if post.Status == model.PostStatusPublished {
		post.Status = model.PostStatusDraft
	} else {
		post.Status = model.PostStatusPublished
		if post.PublishedAt == nil {
			now := time.Now()
			post.PublishedAt = &now
		}
	}

	if err := s.postRepo.Update(ctx, post); err != nil {
		return nil, fmt.Errorf("updating post status: %w", err)
	}

	return s.buildPostResponse(ctx, post)
}

// buildPostResponse assembles a PostResponse from raw post data.
func (s *PostService) buildPostResponse(ctx context.Context, post *model.Post) (*model.PostResponse, error) {
	author, err := s.userRepo.GetByID(ctx, post.AuthorID)
	if err != nil {
		return nil, fmt.Errorf("getting post author: %w", err)
	}

	tags, err := s.postRepo.GetTagsByPostID(ctx, post.ID)
	if err != nil {
		return nil, fmt.Errorf("getting post tags: %w", err)
	}

	resp := &model.PostResponse{
		ID:          post.ID,
		Author:      author.ToResponse(),
		Tags:        tags,
		Title:       post.Title,
		Slug:        post.Slug,
		Excerpt:     post.Excerpt,
		Content:     post.Content,
		CoverImage:  post.CoverImage,
		Status:      post.Status,
		PublishedAt: post.PublishedAt,
		ViewCount:   post.ViewCount,
		CreatedAt:   post.CreatedAt,
		UpdatedAt:   post.UpdatedAt,
	}

	if post.CategoryID != nil {
		cat, err := s.categoryRepo.GetByID(ctx, *post.CategoryID)
		if err == nil {
			resp.Category = cat
		}
	}

	return resp, nil
}

// buildPostListItem assembles a PostListItem (no full content) from raw post data.
func (s *PostService) buildPostListItem(ctx context.Context, post *model.Post) (*model.PostListItem, error) {
	author, err := s.userRepo.GetByID(ctx, post.AuthorID)
	if err != nil {
		return nil, fmt.Errorf("getting post author: %w", err)
	}

	tags, err := s.postRepo.GetTagsByPostID(ctx, post.ID)
	if err != nil {
		return nil, fmt.Errorf("getting post tags: %w", err)
	}

	item := &model.PostListItem{
		ID:          post.ID,
		Author:      author.ToResponse(),
		Tags:        tags,
		Title:       post.Title,
		Slug:        post.Slug,
		Excerpt:     post.Excerpt,
		CoverImage:  post.CoverImage,
		Status:      post.Status,
		PublishedAt: post.PublishedAt,
		ViewCount:   post.ViewCount,
		CreatedAt:   post.CreatedAt,
		UpdatedAt:   post.UpdatedAt,
	}

	if post.CategoryID != nil {
		cat, err := s.categoryRepo.GetByID(ctx, *post.CategoryID)
		if err == nil {
			item.Category = cat
		}
	}

	return item, nil
}

// generateUniqueSlug creates a URL-safe slug from the title, appending a UUID suffix if needed.
func (s *PostService) generateUniqueSlug(ctx context.Context, title string, excludeID *uuid.UUID) (string, error) {
	base := slugify(title)
	slug := base

	exists, err := s.postRepo.SlugExists(ctx, slug, excludeID)
	if err != nil {
		return "", err
	}

	if exists {
		// Append a short unique suffix to avoid collision.
		suffix := uuid.New().String()[:8]
		slug = fmt.Sprintf("%s-%s", base, suffix)
	}

	return slug, nil
}

// slugify converts a string to a URL-safe slug.
var nonAlphanumeric = regexp.MustCompile(`[^a-z0-9-]+`)

func slugify(s string) string {
	s = strings.ToLower(s)
	// Replace spaces and non-letter characters with hyphen.
	var b strings.Builder
	for _, r := range s {
		if unicode.IsLetter(r) || unicode.IsDigit(r) {
			b.WriteRune(r)
		} else {
			b.WriteRune('-')
		}
	}
	result := nonAlphanumeric.ReplaceAllString(b.String(), "-")
	result = strings.Trim(result, "-")
	// Collapse multiple hyphens.
	for strings.Contains(result, "--") {
		result = strings.ReplaceAll(result, "--", "-")
	}
	if len(result) > 200 {
		result = result[:200]
	}
	return result
}


