package validator

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/go-playground/validator/v10"

	"github.com/shendrong/fullstack-go/server/internal/model"
)

// Validator wraps the go-playground validator with helper methods.
type Validator struct {
	validate *validator.Validate
}

// New creates a new Validator instance.
func New() *Validator {
	return &Validator{
		validate: validator.New(validator.WithRequiredStructEnabled()),
	}
}

// DecodeAndValidate reads JSON from the request body, decodes it into dst,
// and validates the struct fields. Returns a map of field errors if validation fails.
func (v *Validator) DecodeAndValidate(r *http.Request, dst any) *model.ErrorResponse {
	if r.Body == nil {
		return &model.ErrorResponse{Error: "request body is required"}
	}
	defer r.Body.Close()

	// Limit body size to 1MB to prevent abuse.
	body := http.MaxBytesReader(nil, r.Body, 1<<20)

	decoder := json.NewDecoder(body)
	decoder.DisallowUnknownFields()

	if err := decoder.Decode(dst); err != nil {
		return &model.ErrorResponse{Error: fmt.Sprintf("invalid request body: %s", err.Error())}
	}

	// Check for extra data after the JSON object.
	if decoder.More() {
		return &model.ErrorResponse{Error: "request body must contain a single JSON object"}
	}

	if _, err := io.ReadAll(body); err != nil {
		// Ignore read errors after successful decode - body was already consumed.
	}

	if err := v.validate.Struct(dst); err != nil {
		validationErrors, ok := err.(validator.ValidationErrors)
		if !ok {
			return &model.ErrorResponse{Error: "validation failed"}
		}

		details := make(map[string]string, len(validationErrors))
		for _, fe := range validationErrors {
			details[fe.Field()] = formatValidationError(fe)
		}

		return &model.ErrorResponse{
			Error:   "validation failed",
			Details: details,
		}
	}

	return nil
}

// formatValidationError returns a human-readable error message for a validation error.
func formatValidationError(fe validator.FieldError) string {
	switch fe.Tag() {
	case "required":
		return "this field is required"
	case "email":
		return "must be a valid email address"
	case "min":
		return fmt.Sprintf("must be at least %s characters", fe.Param())
	case "max":
		return fmt.Sprintf("must be at most %s characters", fe.Param())
	default:
		return fmt.Sprintf("failed on '%s' validation", fe.Tag())
	}
}
