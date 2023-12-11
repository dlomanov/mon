package apperrors

import (
	"fmt"
)

func init() {
	var _ error = (*AppError)(nil)
	var _ fmt.Stringer = (*Code)(nil)
}

type AppError struct {
	Code       Code
	Message    string
	FormatArgs []any
}

func (a AppError) Error() string {
	if a.Code == "" {
		return ""
	}

	if a.Message == "" {
		return string(a.Code)
	}

	msg := fmt.Sprintf(a.Message, a.FormatArgs)
	return fmt.Sprintf("%s: %s", a.Code, msg)
}

type Code string

func (c Code) String() string {
	return string(c)
}

func (c Code) New(message string, args ...any) error {
	return AppError{
		Code:       c,
		Message:    message,
		FormatArgs: args,
	}
}
