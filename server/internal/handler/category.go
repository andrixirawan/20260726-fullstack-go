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

// CategoryHandler handles category HTTP endpoints.
type CategoryHandler struct {
	categoryService *service.CategoryService
	validator       *validator.Validator
	logger          *slog.Logger
}

// NewCategoryHandler creates a new CategoryHandler.
func NewCategoryHandler(categoryService *service.CategoryService, v *validator.Validator, logger *slog.Logger) *CategoryHandler {
	return &CategoryHandler{categoryService: categoryService, validator: v, logger: logger}
}

// ListCategories returns all categories.
//
//	@Summary		List categories
//	@Description	Returns all blog categories
//	@Tags			categories
//	@Produce		json
//	@Success		200	{array}		model.Category
//	@Failure		500	{object}	model.ErrorResponse
//	@Router			/categories [get]
func (h *CategoryHandler) ListCategories(w http.ResponseWriter, r *http.Request) {
	cats, err := h.categoryService.List(r.Context())
	if err != nil {
		h.logger.Error("failed to list categories", slog.Any("error", err))
		writeJSON(w, http.StatusInternalServerError, model.ErrorResponse{Error: "internal server error"})
		return
	}
	writeJSON(w, http.StatusOK, cats)
}

// CreateCategory creates a new category.
//
//	@Summary		Create a category
//	@Description	Creates a new blog category. Requires authentication.
//	@Tags			categories
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			request	body		model.CreateCategoryRequest	true	"Category payload"
//	@Success		201		{object}	model.Category
//	@Failure		400		{object}	model.ErrorResponse
//	@Failure		401		{object}	model.ErrorResponse
//	@Failure		409		{object}	model.ErrorResponse
//	@Failure		500		{object}	model.ErrorResponse
//	@Router			/categories [post]
func (h *CategoryHandler) CreateCategory(w http.ResponseWriter, r *http.Request) {
	var req model.CreateCategoryRequest
	if errResp := h.validator.DecodeAndValidate(r, &req); errResp != nil {
		writeJSON(w, http.StatusBadRequest, errResp)
		return
	}

	cat, err := h.categoryService.Create(r.Context(), &req)
	if err != nil {
		if errors.Is(err, service.ErrCategoryNameConflict) {
			writeJSON(w, http.StatusConflict, model.ErrorResponse{Error: err.Error()})
			return
		}
		h.logger.Error("failed to create category", slog.Any("error", err))
		writeJSON(w, http.StatusInternalServerError, model.ErrorResponse{Error: "internal server error"})
		return
	}

	writeJSON(w, http.StatusCreated, cat)
}

// UpdateCategory updates an existing category.
//
//	@Summary		Update a category
//	@Description	Updates a blog category by ID. Requires authentication.
//	@Tags			categories
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			id		path		string						true	"Category UUID"
//	@Param			request	body		model.UpdateCategoryRequest	true	"Category update payload"
//	@Success		200		{object}	model.Category
//	@Failure		400		{object}	model.ErrorResponse
//	@Failure		401		{object}	model.ErrorResponse
//	@Failure		404		{object}	model.ErrorResponse
//	@Failure		409		{object}	model.ErrorResponse
//	@Failure		500		{object}	model.ErrorResponse
//	@Router			/categories/{id} [put]
func (h *CategoryHandler) UpdateCategory(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeJSON(w, http.StatusBadRequest, model.ErrorResponse{Error: "invalid category id"})
		return
	}

	var req model.UpdateCategoryRequest
	if errResp := h.validator.DecodeAndValidate(r, &req); errResp != nil {
		writeJSON(w, http.StatusBadRequest, errResp)
		return
	}

	cat, err := h.categoryService.Update(r.Context(), id, &req)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrCategoryNotFound):
			writeJSON(w, http.StatusNotFound, model.ErrorResponse{Error: "category not found"})
		case errors.Is(err, service.ErrCategoryNameConflict):
			writeJSON(w, http.StatusConflict, model.ErrorResponse{Error: err.Error()})
		default:
			h.logger.Error("failed to update category", slog.Any("error", err))
			writeJSON(w, http.StatusInternalServerError, model.ErrorResponse{Error: "internal server error"})
		}
		return
	}

	writeJSON(w, http.StatusOK, cat)
}

// DeleteCategory deletes a category.
//
//	@Summary		Delete a category
//	@Description	Deletes a blog category by ID. Requires authentication.
//	@Tags			categories
//	@Produce		json
//	@Security		BearerAuth
//	@Param			id	path	string	true	"Category UUID"
//	@Success		204	"No Content"
//	@Failure		401	{object}	model.ErrorResponse
//	@Failure		404	{object}	model.ErrorResponse
//	@Failure		500	{object}	model.ErrorResponse
//	@Router			/categories/{id} [delete]
func (h *CategoryHandler) DeleteCategory(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeJSON(w, http.StatusBadRequest, model.ErrorResponse{Error: "invalid category id"})
		return
	}

	if err := h.categoryService.Delete(r.Context(), id); err != nil {
		if errors.Is(err, service.ErrCategoryNotFound) {
			writeJSON(w, http.StatusNotFound, model.ErrorResponse{Error: "category not found"})
			return
		}
		h.logger.Error("failed to delete category", slog.Any("error", err))
		writeJSON(w, http.StatusInternalServerError, model.ErrorResponse{Error: "internal server error"})
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
