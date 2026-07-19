package handler

import (
	"errors"
	"log/slog"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"

	"github.com/shendrong/fullstack-go/server/internal/middleware"
	"github.com/shendrong/fullstack-go/server/internal/model"
	"github.com/shendrong/fullstack-go/server/internal/service"
	"github.com/shendrong/fullstack-go/server/internal/validator"
)

// CommentHandler handles comment HTTP endpoints.
type CommentHandler struct {
	commentService *service.CommentService
	validator      *validator.Validator
	logger         *slog.Logger
}

// NewCommentHandler creates a new CommentHandler.
func NewCommentHandler(commentService *service.CommentService, v *validator.Validator, logger *slog.Logger) *CommentHandler {
	return &CommentHandler{commentService: commentService, validator: v, logger: logger}
}

// ListComments returns threaded comments for a post.
//
//	@Summary		List comments for a post
//	@Description	Returns a threaded (nested) list of comments for a given post
//	@Tags			comments
//	@Produce		json
//	@Param			id	path		string	true	"Post UUID"
//	@Success		200	{array}		model.CommentResponse
//	@Failure		400	{object}	model.ErrorResponse
//	@Failure		500	{object}	model.ErrorResponse
//	@Router			/posts/{id}/comments [get]
func (h *CommentHandler) ListComments(w http.ResponseWriter, r *http.Request) {
	postID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeJSON(w, http.StatusBadRequest, model.ErrorResponse{Error: "invalid post id"})
		return
	}

	comments, err := h.commentService.ListByPostID(r.Context(), postID)
	if err != nil {
		h.logger.Error("failed to list comments", slog.Any("error", err))
		writeJSON(w, http.StatusInternalServerError, model.ErrorResponse{Error: "internal server error"})
		return
	}

	writeJSON(w, http.StatusOK, comments)
}

// CreateComment adds a comment to a post.
//
//	@Summary		Add a comment
//	@Description	Adds a comment (or reply) to a blog post. Requires authentication.
//	@Tags			comments
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			id		path		string						true	"Post UUID"
//	@Param			request	body		model.CreateCommentRequest	true	"Comment payload"
//	@Success		201		{object}	model.CommentResponse
//	@Failure		400		{object}	model.ErrorResponse
//	@Failure		401		{object}	model.ErrorResponse
//	@Failure		500		{object}	model.ErrorResponse
//	@Router			/posts/{id}/comments [post]
func (h *CommentHandler) CreateComment(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		writeJSON(w, http.StatusUnauthorized, model.ErrorResponse{Error: "user not authenticated"})
		return
	}

	postID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeJSON(w, http.StatusBadRequest, model.ErrorResponse{Error: "invalid post id"})
		return
	}

	var req model.CreateCommentRequest
	if errResp := h.validator.DecodeAndValidate(r, &req); errResp != nil {
		writeJSON(w, http.StatusBadRequest, errResp)
		return
	}

	comment, err := h.commentService.Create(r.Context(), postID, userID, &req)
	if err != nil {
		h.logger.Error("failed to create comment", slog.Any("error", err))
		writeJSON(w, http.StatusInternalServerError, model.ErrorResponse{Error: "internal server error"})
		return
	}

	writeJSON(w, http.StatusCreated, comment)
}

// UpdateComment edits an existing comment.
//
//	@Summary		Update a comment
//	@Description	Edits the content of a comment. Only the original author can do this.
//	@Tags			comments
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			id		path		string						true	"Comment UUID"
//	@Param			request	body		model.UpdateCommentRequest	true	"Comment update payload"
//	@Success		200		{object}	model.CommentResponse
//	@Failure		400		{object}	model.ErrorResponse
//	@Failure		401		{object}	model.ErrorResponse
//	@Failure		403		{object}	model.ErrorResponse
//	@Failure		404		{object}	model.ErrorResponse
//	@Failure		500		{object}	model.ErrorResponse
//	@Router			/comments/{id} [put]
func (h *CommentHandler) UpdateComment(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		writeJSON(w, http.StatusUnauthorized, model.ErrorResponse{Error: "user not authenticated"})
		return
	}

	commentID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeJSON(w, http.StatusBadRequest, model.ErrorResponse{Error: "invalid comment id"})
		return
	}

	var req model.UpdateCommentRequest
	if errResp := h.validator.DecodeAndValidate(r, &req); errResp != nil {
		writeJSON(w, http.StatusBadRequest, errResp)
		return
	}

	comment, err := h.commentService.Update(r.Context(), commentID, userID, &req)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrCommentNotFound):
			writeJSON(w, http.StatusNotFound, model.ErrorResponse{Error: "comment not found"})
		case errors.Is(err, service.ErrCommentForbidden):
			writeJSON(w, http.StatusForbidden, model.ErrorResponse{Error: err.Error()})
		default:
			h.logger.Error("failed to update comment", slog.Any("error", err))
			writeJSON(w, http.StatusInternalServerError, model.ErrorResponse{Error: "internal server error"})
		}
		return
	}

	writeJSON(w, http.StatusOK, comment)
}

// DeleteComment soft-deletes a comment.
//
//	@Summary		Delete a comment
//	@Description	Soft-deletes a comment. Only the original author can do this.
//	@Tags			comments
//	@Produce		json
//	@Security		BearerAuth
//	@Param			id	path	string	true	"Comment UUID"
//	@Success		204	"No Content"
//	@Failure		401	{object}	model.ErrorResponse
//	@Failure		403	{object}	model.ErrorResponse
//	@Failure		404	{object}	model.ErrorResponse
//	@Failure		500	{object}	model.ErrorResponse
//	@Router			/comments/{id} [delete]
func (h *CommentHandler) DeleteComment(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		writeJSON(w, http.StatusUnauthorized, model.ErrorResponse{Error: "user not authenticated"})
		return
	}

	commentID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeJSON(w, http.StatusBadRequest, model.ErrorResponse{Error: "invalid comment id"})
		return
	}

	if err := h.commentService.Delete(r.Context(), commentID, userID); err != nil {
		switch {
		case errors.Is(err, service.ErrCommentNotFound):
			writeJSON(w, http.StatusNotFound, model.ErrorResponse{Error: "comment not found"})
		case errors.Is(err, service.ErrCommentForbidden):
			writeJSON(w, http.StatusForbidden, model.ErrorResponse{Error: err.Error()})
		default:
			h.logger.Error("failed to delete comment", slog.Any("error", err))
			writeJSON(w, http.StatusInternalServerError, model.ErrorResponse{Error: "internal server error"})
		}
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
