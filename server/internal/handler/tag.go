package handler

import (
	"errors"
	"log/slog"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"

	"github.com/shendrong/fullstack-go/server/internal/model"
	"github.com/shendrong/fullstack-go/server/internal/service"
	"github.com/shendrong/fullstack-go/server/internal/validator"
)

// TagHandler handles tag HTTP endpoints.
type TagHandler struct {
	tagService *service.TagService
	validator  *validator.Validator
	logger     *slog.Logger
}

// NewTagHandler creates a new TagHandler.
func NewTagHandler(tagService *service.TagService, v *validator.Validator, logger *slog.Logger) *TagHandler {
	return &TagHandler{tagService: tagService, validator: v, logger: logger}
}

// ListTags returns all tags.
//
//	@Summary		List tags
//	@Description	Returns all blog tags
//	@Tags			tags
//	@Produce		json
//	@Success		200	{array}		model.Tag
//	@Failure		500	{object}	model.ErrorResponse
//	@Router			/tags [get]
func (h *TagHandler) ListTags(w http.ResponseWriter, r *http.Request) {
	tags, err := h.tagService.List(r.Context())
	if err != nil {
		h.logger.Error("failed to list tags", slog.Any("error", err))
		writeJSON(w, http.StatusInternalServerError, model.ErrorResponse{Error: "internal server error"})
		return
	}
	writeJSON(w, http.StatusOK, tags)
}

// CreateTag creates a new tag.
//
//	@Summary		Create a tag
//	@Description	Creates a new blog tag. Requires authentication.
//	@Tags			tags
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			request	body		model.CreateTagRequest	true	"Tag payload"
//	@Success		201		{object}	model.Tag
//	@Failure		400		{object}	model.ErrorResponse
//	@Failure		401		{object}	model.ErrorResponse
//	@Failure		409		{object}	model.ErrorResponse
//	@Failure		500		{object}	model.ErrorResponse
//	@Router			/tags [post]
func (h *TagHandler) CreateTag(w http.ResponseWriter, r *http.Request) {
	var req model.CreateTagRequest
	if errResp := h.validator.DecodeAndValidate(r, &req); errResp != nil {
		writeJSON(w, http.StatusBadRequest, errResp)
		return
	}

	tag, err := h.tagService.Create(r.Context(), &req)
	if err != nil {
		if errors.Is(err, service.ErrTagNameConflict) {
			writeJSON(w, http.StatusConflict, model.ErrorResponse{Error: err.Error()})
			return
		}
		h.logger.Error("failed to create tag", slog.Any("error", err))
		writeJSON(w, http.StatusInternalServerError, model.ErrorResponse{Error: "internal server error"})
		return
	}

	writeJSON(w, http.StatusCreated, tag)
}

// DeleteTag deletes a tag.
//
//	@Summary		Delete a tag
//	@Description	Deletes a blog tag by ID. Requires authentication.
//	@Tags			tags
//	@Produce		json
//	@Security		BearerAuth
//	@Param			id	path	string	true	"Tag UUID"
//	@Success		204	"No Content"
//	@Failure		401	{object}	model.ErrorResponse
//	@Failure		404	{object}	model.ErrorResponse
//	@Failure		500	{object}	model.ErrorResponse
//	@Router			/tags/{id} [delete]
func (h *TagHandler) DeleteTag(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeJSON(w, http.StatusBadRequest, model.ErrorResponse{Error: "invalid tag id"})
		return
	}

	if err := h.tagService.Delete(r.Context(), id); err != nil {
		if errors.Is(err, service.ErrTagNotFound) {
			writeJSON(w, http.StatusNotFound, model.ErrorResponse{Error: "tag not found"})
			return
		}
		h.logger.Error("failed to delete tag", slog.Any("error", err))
		writeJSON(w, http.StatusInternalServerError, model.ErrorResponse{Error: "internal server error"})
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
