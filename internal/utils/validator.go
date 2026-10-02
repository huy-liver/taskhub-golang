package utils

import (
	"fmt"
	"strings"
	"taskhub/internal/errors"

	"github.com/go-playground/validator/v10"
)

// FormatValidationError formats go-playground validator errors into FieldError slice
func FormatValidationError(err error) *errors.AppError {
	var fieldErrors []errors.FieldError

	validationErrors, ok := err.(validator.ValidationErrors)
	if !ok {
		return errors.NewBadRequestError("Invalid request payload", err.Error())
	}

	for _, fe := range validationErrors {
		fieldName := strings.ToLower(fe.Field())
		var message string

		switch fe.Tag() {
		case "required":
			message = fmt.Sprintf("%s is required", fieldName)
		case "email":
			message = fmt.Sprintf("%s must be a valid email address", fieldName)
		case "min":
			message = fmt.Sprintf("%s must be at least %s characters long", fieldName, fe.Param())
		case "max":
			message = fmt.Sprintf("%s must be at most %s characters long", fieldName, fe.Param())
		case "oneof":
			message = fmt.Sprintf("%s must be one of: %s", fieldName, fe.Param())
		default:
			message = fmt.Sprintf("%s failed on validation rule: %s", fieldName, fe.Tag())
		}

		fieldErrors = append(fieldErrors, errors.FieldError{
			Field:   fieldName,
			Tag:     fe.Tag(),
			Message: message,
		})
	}

	return errors.NewValidationError(fieldErrors)
}
