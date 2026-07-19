package service

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"

	"github.com/shendrong/fullstack-go/server/internal/model"
	"github.com/shendrong/fullstack-go/server/internal/repository"
)

// Common tag service errors.
var (
	ErrTagNotFound    = errors.New("tag not found")
	ErrTagNameConflict = errors.New("a tag with this name already exists")
)

// TagService handles tag business logic.
type TagService struct {
	tagRepo *repository.TagRepository
}

// NewTagService creates a new TagService.
func NewTagService(tagRepo *repository.TagRepository) *TagService {
	return &TagService{tagRepo: tagRepo}
}

// Create creates a new tag.
func (s *TagService) Create(ctx context.Context, req *model.CreateTagRequest) (*model.Tag, error) {
	tag := &model.Tag{
		Name: req.Name,
		Slug: slugify(req.Name),
	}

	if err := s.tagRepo.Create(ctx, tag); err != nil {
		if errors.Is(err, repository.ErrTagNameConflict) {
			return nil, ErrTagNameConflict
		}
		return nil, fmt.Errorf("creating tag: %w", err)
	}

	return tag, nil
}

// List returns all tags.
func (s *TagService) List(ctx context.Context) ([]model.Tag, error) {
	tags, err := s.tagRepo.List(ctx)
	if err != nil {
		return nil, fmt.Errorf("listing tags: %w", err)
	}
	return tags, nil
}

// Delete removes a tag by ID.
func (s *TagService) Delete(ctx context.Context, id uuid.UUID) error {
	if err := s.tagRepo.Delete(ctx, id); err != nil {
		if errors.Is(err, repository.ErrTagNotFound) {
			return ErrTagNotFound
		}
		return fmt.Errorf("deleting tag: %w", err)
	}
	return nil
}
