package apperrors

import (
	"fmt"
)

func init() {
	var _ error = (*AppError)(nil)
}

type AppError struct {
	Type       AppErrorType
	Message    string
	FormatArgs []any
}

func (a AppError) Error() string {
	return fmt.Sprintf(a.Message, a.FormatArgs)
}
