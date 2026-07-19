package handler

import (
	"errors"
	"log/slog"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"

	"github.com/shendrong/fullstack-go/server/internal/middleware"
	"github.com/shendrong/fullstack-go/server/internal/model"
	"github.com/shendrong/fullstack-go/server/internal/service"
	"github.com/shendrong/fullstack-go/server/internal/validator"
)

// PostHandler handles post HTTP endpoints.
type PostHandler struct {
	postService *service.PostService
	validator   *validator.Validator
	logger      *slog.Logger
}

// NewPostHandler creates a new PostHandler.
func NewPostHandler(postService *service.PostService, v *validator.Validator, logger *slog.Logger) *PostHandler {
	return &PostHandler{postService: postService, validator: v, logger: logger}
}

// ListPosts returns a paginated, filtered list of posts.
//
//	@Summary		List posts
//	@Description	Returns a paginated list of blog posts with optional filtering by status, category, tag, author, and search
//	@Tags			posts
//	@Produce		json
//	@Param			page        query	int		false	"Page number (default: 1)"
//	@Param			page_size   query	int		false	"Items per page (default: 10, max: 100)"
//	@Param			status      query	string	false	"Filter by status (draft or published)"
//	@Param			category_id query	string	false	"Filter by category UUID"
//	@Param			tag_id      query	string	false	"Filter by tag UUID"
//	@Param			author_id   query	string	false	"Filter by author UUID"
//	@Param			search      query	string	false	"Search in title and excerpt"
//	@Success		200	{object}	model.PostListResponse
//	@Failure		500	{object}	model.ErrorResponse
//	@Router			/posts [get]
func (h *PostHandler) ListPosts(w http.ResponseWriter, r *http.Request) {
	q := model.PostListQuery{
		Page:     parseIntQuery(r, "page", 1),
		PageSize: parseIntQuery(r, "page_size", 10),
		Status:   model.PostStatus(r.URL.Query().Get("status")),
		Search:   r.URL.Query().Get("search"),
	}

	if categoryIDStr := r.URL.Query().Get("category_id"); categoryIDStr != "" {
		id, err := uuid.Parse(categoryIDStr)
		if err == nil {
			q.CategoryID = &id
		}
	}
	if tagIDStr := r.URL.Query().Get("tag_id"); tagIDStr != "" {
		id, err := uuid.Parse(tagIDStr)
		if err == nil {
			q.TagID = &id
		}
	}
	if authorIDStr := r.URL.Query().Get("author_id"); authorIDStr != "" {
		id, err := uuid.Parse(authorIDStr)
		if err == nil {
			q.AuthorID = &id
		}
	}

	resp, err := h.postService.List(r.Context(), q)
	if err != nil {
		h.logger.Error("failed to list posts", slog.Any("error", err))
		writeJSON(w, http.StatusInternalServerError, model.ErrorResponse{Error: "internal server error"})
		return
	}

	writeJSON(w, http.StatusOK, resp)
}

// CreatePost creates a new blog post.
//
//	@Summary		Create a post
//	@Description	Creates a new blog post. Requires authentication.
//	@Tags			posts
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			request	body		model.CreatePostRequest	true	"Post payload"
//	@Success		201		{object}	model.PostResponse
//	@Failure		400		{object}	model.ErrorResponse
//	@Failure		401		{object}	model.ErrorResponse
//	@Failure		409		{object}	model.ErrorResponse
//	@Failure		500		{object}	model.ErrorResponse
//	@Router			/posts [post]
func (h *PostHandler) CreatePost(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		writeJSON(w, http.StatusUnauthorized, model.ErrorResponse{Error: "user not authenticated"})
		return
	}

	var req model.CreatePostRequest
	if errResp := h.validator.DecodeAndValidate(r, &req); errResp != nil {
		writeJSON(w, http.StatusBadRequest, errResp)
		return
	}

	resp, err := h.postService.Create(r.Context(), userID, &req)
	if err != nil {
		if errors.Is(err, service.ErrPostSlugConflict) {
			writeJSON(w, http.StatusConflict, model.ErrorResponse{Error: err.Error()})
			return
		}
		h.logger.Error("failed to create post", slog.Any("error", err))
		writeJSON(w, http.StatusInternalServerError, model.ErrorResponse{Error: "internal server error"})
		return
	}

	writeJSON(w, http.StatusCreated, resp)
}

// GetPostByID returns a post by its UUID.
//
//	@Summary		Get post by ID
//	@Description	Returns a single blog post by its UUID, incrementing the view count
//	@Tags			posts
//	@Produce		json
//	@Param			id	path		string	true	"Post UUID"
//	@Success		200	{object}	model.PostResponse
//	@Failure		404	{object}	model.ErrorResponse
//	@Failure		500	{object}	model.ErrorResponse
//	@Router			/posts/{id} [get]
func (h *PostHandler) GetPostByID(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeJSON(w, http.StatusBadRequest, model.ErrorResponse{Error: "invalid post id"})
		return
	}

	resp, err := h.postService.GetByID(r.Context(), id)
	if err != nil {
		if errors.Is(err, service.ErrPostNotFound) {
			writeJSON(w, http.StatusNotFound, model.ErrorResponse{Error: "post not found"})
			return
		}
		h.logger.Error("failed to get post", slog.Any("error", err))
		writeJSON(w, http.StatusInternalServerError, model.ErrorResponse{Error: "internal server error"})
		return
	}

	writeJSON(w, http.StatusOK, resp)
}

// GetPostBySlug returns a post by its slug.
//
//	@Summary		Get post by slug
//	@Description	Returns a single blog post by its URL slug, incrementing the view count
//	@Tags			posts
//	@Produce		json
//	@Param			slug	path		string	true	"Post slug"
//	@Success		200		{object}	model.PostResponse
//	@Failure		404		{object}	model.ErrorResponse
//	@Failure		500		{object}	model.ErrorResponse
//	@Router			/posts/slug/{slug} [get]
func (h *PostHandler) GetPostBySlug(w http.ResponseWriter, r *http.Request) {
	slug := chi.URLParam(r, "slug")

	resp, err := h.postService.GetBySlug(r.Context(), slug)
	if err != nil {
		if errors.Is(err, service.ErrPostNotFound) {
			writeJSON(w, http.StatusNotFound, model.ErrorResponse{Error: "post not found"})
			return
		}
		h.logger.Error("failed to get post by slug", slog.Any("error", err))
		writeJSON(w, http.StatusInternalServerError, model.ErrorResponse{Error: "internal server error"})
		return
	}

	writeJSON(w, http.StatusOK, resp)
}

// UpdatePost updates an existing post.
//
//	@Summary		Update a post
//	@Description	Updates a blog post. Only the original author can update it.
//	@Tags			posts
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			id		path		string					true	"Post UUID"
//	@Param			request	body		model.UpdatePostRequest	true	"Post update payload"
//	@Success		200		{object}	model.PostResponse
//	@Failure		400		{object}	model.ErrorResponse
//	@Failure		401		{object}	model.ErrorResponse
//	@Failure		403		{object}	model.ErrorResponse
//	@Failure		404		{object}	model.ErrorResponse
//	@Failure		409		{object}	model.ErrorResponse
//	@Failure		500		{object}	model.ErrorResponse
//	@Router			/posts/{id} [put]
func (h *PostHandler) UpdatePost(w http.ResponseWriter, r *http.Request) {
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

	var req model.UpdatePostRequest
	if errResp := h.validator.DecodeAndValidate(r, &req); errResp != nil {
		writeJSON(w, http.StatusBadRequest, errResp)
		return
	}

	resp, err := h.postService.Update(r.Context(), postID, userID, &req)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrPostNotFound):
			writeJSON(w, http.StatusNotFound, model.ErrorResponse{Error: "post not found"})
		case errors.Is(err, service.ErrPostForbidden):
			writeJSON(w, http.StatusForbidden, model.ErrorResponse{Error: err.Error()})
		case errors.Is(err, service.ErrPostSlugConflict):
			writeJSON(w, http.StatusConflict, model.ErrorResponse{Error: err.Error()})
		default:
			h.logger.Error("failed to update post", slog.Any("error", err))
			writeJSON(w, http.StatusInternalServerError, model.ErrorResponse{Error: "internal server error"})
		}
		return
	}

	writeJSON(w, http.StatusOK, resp)
}

// DeletePost deletes a post by ID.
//
//	@Summary		Delete a post
//	@Description	Deletes a blog post. Only the original author can delete it.
//	@Tags			posts
//	@Produce		json
//	@Security		BearerAuth
//	@Param			id	path	string	true	"Post UUID"
//	@Success		204	"No Content"
//	@Failure		401	{object}	model.ErrorResponse
//	@Failure		403	{object}	model.ErrorResponse
//	@Failure		404	{object}	model.ErrorResponse
//	@Failure		500	{object}	model.ErrorResponse
//	@Router			/posts/{id} [delete]
func (h *PostHandler) DeletePost(w http.ResponseWriter, r *http.Request) {
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

	if err := h.postService.Delete(r.Context(), postID, userID); err != nil {
		switch {
		case errors.Is(err, service.ErrPostNotFound):
			writeJSON(w, http.StatusNotFound, model.ErrorResponse{Error: "post not found"})
		case errors.Is(err, service.ErrPostForbidden):
			writeJSON(w, http.StatusForbidden, model.ErrorResponse{Error: err.Error()})
		default:
			h.logger.Error("failed to delete post", slog.Any("error", err))
			writeJSON(w, http.StatusInternalServerError, model.ErrorResponse{Error: "internal server error"})
		}
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// TogglePublish toggles the publish status of a post.
//
//	@Summary		Toggle post publish status
//	@Description	Toggles a post between draft and published. Only the author can do this.
//	@Tags			posts
//	@Produce		json
//	@Security		BearerAuth
//	@Param			id	path		string	true	"Post UUID"
//	@Success		200	{object}	model.PostResponse
//	@Failure		401	{object}	model.ErrorResponse
//	@Failure		403	{object}	model.ErrorResponse
//	@Failure		404	{object}	model.ErrorResponse
//	@Failure		500	{object}	model.ErrorResponse
//	@Router			/posts/{id}/publish [patch]
func (h *PostHandler) TogglePublish(w http.ResponseWriter, r *http.Request) {
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

	resp, err := h.postService.TogglePublish(r.Context(), postID, userID)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrPostNotFound):
			writeJSON(w, http.StatusNotFound, model.ErrorResponse{Error: "post not found"})
		case errors.Is(err, service.ErrPostForbidden):
			writeJSON(w, http.StatusForbidden, model.ErrorResponse{Error: err.Error()})
		default:
			h.logger.Error("failed to toggle post publish", slog.Any("error", err))
			writeJSON(w, http.StatusInternalServerError, model.ErrorResponse{Error: "internal server error"})
		}
		return
	}

	writeJSON(w, http.StatusOK, resp)
}

// parseIntQuery reads a query param as int with a fallback default.
func parseIntQuery(r *http.Request, key string, defaultVal int) int {
	s := r.URL.Query().Get(key)
	if s == "" {
		return defaultVal
	}
	n, err := strconv.Atoi(s)
	if err != nil || n < 1 {
		return defaultVal
	}
	return n
}
