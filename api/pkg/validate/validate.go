package validate

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/go-playground/validator/v10"
)

var validate *validator.Validate

func init() {
	validate = validator.New(validator.WithRequiredStructEnabled())
}

// ValidationError represents a single field validation error
type ValidationError struct {
	Field   string `json:"field"`
	Message string `json:"message"`
}

// ValidationErrors is a collection of validation errors
type ValidationErrors struct {
	Errors []ValidationError `json:"errors"`
}

func (v ValidationErrors) Error() string {
	var msgs []string
	for _, e := range v.Errors {
		msgs = append(msgs, fmt.Sprintf("%s: %s", e.Field, e.Message))
	}
	return strings.Join(msgs, "; ")
}

// Struct validates a struct and returns formatted errors
func Struct(s any) error {
	err := validate.Struct(s)
	if err == nil {
		return nil
	}

	var validationErrs validator.ValidationErrors
	if errors.As(err, &validationErrs) {
		return formatErrors(validationErrs)
	}

	return err
}

// Request decodes JSON body and validates the struct
func Request(r *http.Request, dst any) error {
	if err := json.NewDecoder(r.Body).Decode(dst); err != nil {
		return fmt.Errorf("invalid JSON: %w", err)
	}
	return Struct(dst)
}

// formatErrors converts validator errors to user-friendly messages
func formatErrors(errs validator.ValidationErrors) ValidationErrors {
	var result ValidationErrors
	for _, e := range errs {
		result.Errors = append(result.Errors, ValidationError{
			Field:   toSnakeCase(e.Field()),
			Message: formatMessage(e),
		})
	}
	return result
}

// formatMessage creates a human-readable error message
func formatMessage(e validator.FieldError) string {
	switch e.Tag() {
	case "required":
		return "is required"
	case "email":
		return "must be a valid email address"
	case "min":
		return fmt.Sprintf("must be at least %s characters", e.Param())
	case "max":
		return fmt.Sprintf("must be at most %s characters", e.Param())
	case "oneof":
		return fmt.Sprintf("must be one of: %s", e.Param())
	case "url":
		return "must be a valid URL"
	case "uuid":
		return "must be a valid UUID"
	case "gt":
		return fmt.Sprintf("must be greater than %s", e.Param())
	case "gte":
		return fmt.Sprintf("must be at least %s", e.Param())
	case "lt":
		return fmt.Sprintf("must be less than %s", e.Param())
	case "lte":
		return fmt.Sprintf("must be at most %s", e.Param())
	default:
		return fmt.Sprintf("failed %s validation", e.Tag())
	}
}

// toSnakeCase converts PascalCase/camelCase to snake_case
func toSnakeCase(s string) string {
	var result strings.Builder
	for i, r := range s {
		if i > 0 && r >= 'A' && r <= 'Z' {
			result.WriteRune('_')
		}
		result.WriteRune(r)
	}
	return strings.ToLower(result.String())
}

// FirstError returns the first validation error message, useful for simple error responses
func FirstError(err error) string {
	var ve ValidationErrors
	if errors.As(err, &ve) && len(ve.Errors) > 0 {
		return fmt.Sprintf("%s %s", ve.Errors[0].Field, ve.Errors[0].Message)
	}
	return err.Error()
}
