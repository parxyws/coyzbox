package validator

import (
	"errors"
	"fmt"

	"github.com/go-playground/validator/v10"
)

var Validate *validator.Validate

func init() {
	Validate = validator.New(validator.WithRequiredStructEnabled())
}

func TranslateValidationError(err error) map[string]string {
	errs := make(map[string]string)
	if err == nil {
		return errs
	}

	var validationErrors validator.ValidationErrors
	if errors.As(err, &validationErrors) {
		for _, fieldError := range validationErrors {
			var message string
			switch fieldError.Tag() {
			case "required":
				message = fmt.Sprintf("%s is required", fieldError.Field())
			case "alpha":
				message = fmt.Sprintf("%s can only contain letters", fieldError.Field())
			case "min":
				message = fmt.Sprintf("%s must be at least %s characters long", fieldError.Field(), fieldError.Param())
			case "email":
				message = fmt.Sprintf("%s must be a valid email address", fieldError.Field())
			default:
				message = fmt.Sprintf("%s is invalid", fieldError.Field())
			}
			errs[fieldError.Field()] = message
		}
	}

	return errs
}
