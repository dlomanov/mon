package handlers

import "fmt"

func init() {
	var _ error = (*ValidationError)(nil)
}

type ValidationError struct {
	description string
}

func (v ValidationError) Error() string {
	if v.description == "" {
		return "validation error"
	}
	return fmt.Sprintf("validation error: %s", v.description)
}

func newValidationError(description string) error {
	return ValidationError{description: description}
}
