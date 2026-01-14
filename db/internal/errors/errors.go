package errors

import "fmt"

type ErrorCode int

const (
	ErrGeneral ErrorCode = iota
	ErrSyntax
	ErrTypeMismatch
	ErrConstraintViolation
)

type DBError struct {
	Code    ErrorCode
	Message string
	Hint    string
	Cause   error
}

func (e *DBError) Error() string {
	if e.Hint != "" {
		return fmt.Sprintf("%s\nHint: %s", e.Message, e.Hint)
	}
	return e.Message
}

func (e *DBError) Unwrap() error {
	return e.Cause
}

func New(code ErrorCode, msg, hint string) *DBError {
	return &DBError{Code: code, Message: msg, Hint: hint}
}

func Wrap(cause error, code ErrorCode, msg, hint string) *DBError {
	return &DBError{Code: code, Message: msg, Hint: hint, Cause: cause}
}
