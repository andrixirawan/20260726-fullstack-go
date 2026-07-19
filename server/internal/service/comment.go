package service

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"

	"github.com/shendrong/fullstack-go/server/internal/model"
	"github.com/shendrong/fullstack-go/server/internal/repository"
)

// Common comment service errors.
var (
	ErrCommentNotFound  = errors.New("comment not found")
	ErrCommentForbidden = errors.New("you are not allowed to modify this comment")
)

// CommentService handles comment business logic.
type CommentService struct {
	commentRepo *repository.CommentRepository
	userRepo    *repository.UserRepository
}

// NewCommentService creates a new CommentService.
func NewCommentService(commentRepo *repository.CommentRepository, userRepo *repository.UserRepository) *CommentService {
	return &CommentService{commentRepo: commentRepo, userRepo: userRepo}
}

// Create adds a new comment to a post.
func (s *CommentService) Create(ctx context.Context, postID uuid.UUID, authorID uuid.UUID, req *model.CreateCommentRequest) (*model.CommentResponse, error) {
	comment := &model.Comment{
		PostID:   postID,
		AuthorID: authorID,
		ParentID: req.ParentID,
		Content:  req.Content,
	}

	if err := s.commentRepo.Create(ctx, comment); err != nil {
		return nil, fmt.Errorf("creating comment: %w", err)
	}

	return s.buildCommentResponse(ctx, comment)
}

// ListByPostID returns a threaded list of comments for a post.
func (s *CommentService) ListByPostID(ctx context.Context, postID uuid.UUID) ([]model.CommentResponse, error) {
	flat, err := s.commentRepo.ListByPostID(ctx, postID)
	if err != nil {
		return nil, fmt.Errorf("listing comments: %w", err)
	}

	// Build a map of all comments by ID.
	responseMap := make(map[uuid.UUID]*model.CommentResponse)
	for i := range flat {
		c := &flat[i]
		author, err := s.userRepo.GetByID(ctx, c.AuthorID)
		if err != nil {
			return nil, fmt.Errorf("getting comment author: %w", err)
		}
		cr := &model.CommentResponse{
			ID:        c.ID,
			PostID:    c.PostID,
			Author:    author.ToResponse(),
			ParentID:  c.ParentID,
			Content:   c.Content,
			IsDeleted: c.IsDeleted,
			Replies:   []model.CommentResponse{},
			CreatedAt: c.CreatedAt,
			UpdatedAt: c.UpdatedAt,
		}
		responseMap[c.ID] = cr
	}

	// Build tree structure.
	roots := []model.CommentResponse{}
	for i := range flat {
		c := &flat[i]
		cr := responseMap[c.ID]
		if c.ParentID == nil {
			roots = append(roots, *cr)
		} else {
			if parent, ok := responseMap[*c.ParentID]; ok {
				parent.Replies = append(parent.Replies, *cr)
			}
		}
	}

	return roots, nil
}

// Update edits a comment's content, only if the requester is the author.
func (s *CommentService) Update(ctx context.Context, commentID uuid.UUID, requesterID uuid.UUID, req *model.UpdateCommentRequest) (*model.CommentResponse, error) {
	comment, err := s.commentRepo.GetByID(ctx, commentID)
	if err != nil {
		if errors.Is(err, repository.ErrCommentNotFound) {
			return nil, ErrCommentNotFound
		}
		return nil, fmt.Errorf("getting comment: %w", err)
	}

	if comment.AuthorID != requesterID {
		return nil, ErrCommentForbidden
	}

	comment.Content = req.Content
	if err := s.commentRepo.Update(ctx, comment); err != nil {
		return nil, fmt.Errorf("updating comment: %w", err)
	}

	return s.buildCommentResponse(ctx, comment)
}

// Delete soft-deletes a comment, only if the requester is the author.
func (s *CommentService) Delete(ctx context.Context, commentID uuid.UUID, requesterID uuid.UUID) error {
	comment, err := s.commentRepo.GetByID(ctx, commentID)
	if err != nil {
		if errors.Is(err, repository.ErrCommentNotFound) {
			return ErrCommentNotFound
		}
		return fmt.Errorf("getting comment: %w", err)
	}

	if comment.AuthorID != requesterID {
		return ErrCommentForbidden
	}

	return s.commentRepo.SoftDelete(ctx, commentID)
}

// buildCommentResponse assembles a CommentResponse with author info.
func (s *CommentService) buildCommentResponse(ctx context.Context, comment *model.Comment) (*model.CommentResponse, error) {
	author, err := s.userRepo.GetByID(ctx, comment.AuthorID)
	if err != nil {
		return nil, fmt.Errorf("getting comment author: %w", err)
	}

	return &model.CommentResponse{
		ID:        comment.ID,
		PostID:    comment.PostID,
		Author:    author.ToResponse(),
		ParentID:  comment.ParentID,
		Content:   comment.Content,
		IsDeleted: comment.IsDeleted,
		Replies:   []model.CommentResponse{},
		CreatedAt: comment.CreatedAt,
		UpdatedAt: comment.UpdatedAt,
	}, nil
}
