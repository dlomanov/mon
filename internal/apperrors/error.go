package apperrors

import (
	"fmt"
)

type AppError struct {
	Type       AppErrorType
	Message    string
	FormatArgs []any
}

func (a AppError) Error() string {
	return fmt.Sprintf(a.Message, a.FormatArgs)
}
