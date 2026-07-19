package service

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"

	"github.com/shendrong/fullstack-go/server/internal/model"
	"github.com/shendrong/fullstack-go/server/internal/repository"
)

// Common category service errors.
var (
	ErrCategoryNotFound    = errors.New("category not found")
	ErrCategoryNameConflict = errors.New("a category with this name already exists")
)

// CategoryService handles category business logic.
type CategoryService struct {
	categoryRepo *repository.CategoryRepository
}

// NewCategoryService creates a new CategoryService.
func NewCategoryService(categoryRepo *repository.CategoryRepository) *CategoryService {
	return &CategoryService{categoryRepo: categoryRepo}
}

// Create creates a new category.
func (s *CategoryService) Create(ctx context.Context, req *model.CreateCategoryRequest) (*model.Category, error) {
	cat := &model.Category{
		Name:        req.Name,
		Slug:        slugify(req.Name),
		Description: req.Description,
	}

	if err := s.categoryRepo.Create(ctx, cat); err != nil {
		if errors.Is(err, repository.ErrCategorySlugConflict) {
			return nil, ErrCategoryNameConflict
		}
		return nil, fmt.Errorf("creating category: %w", err)
	}

	return cat, nil
}

// List returns all categories.
func (s *CategoryService) List(ctx context.Context) ([]model.Category, error) {
	cats, err := s.categoryRepo.List(ctx)
	if err != nil {
		return nil, fmt.Errorf("listing categories: %w", err)
	}
	return cats, nil
}

// Update updates an existing category.
func (s *CategoryService) Update(ctx context.Context, id uuid.UUID, req *model.UpdateCategoryRequest) (*model.Category, error) {
	cat, err := s.categoryRepo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, repository.ErrCategoryNotFound) {
			return nil, ErrCategoryNotFound
		}
		return nil, fmt.Errorf("getting category: %w", err)
	}

	if req.Name != nil {
		cat.Name = *req.Name
		cat.Slug = slugify(*req.Name)
	}
	if req.Description != nil {
		cat.Description = *req.Description
	}

	if err := s.categoryRepo.Update(ctx, cat); err != nil {
		if errors.Is(err, repository.ErrCategorySlugConflict) {
			return nil, ErrCategoryNameConflict
		}
		return nil, fmt.Errorf("updating category: %w", err)
	}

	return cat, nil
}

// Delete removes a category.
func (s *CategoryService) Delete(ctx context.Context, id uuid.UUID) error {
	if err := s.categoryRepo.Delete(ctx, id); err != nil {
		if errors.Is(err, repository.ErrCategoryNotFound) {
			return ErrCategoryNotFound
		}
		return fmt.Errorf("deleting category: %w", err)
	}
	return nil
}
