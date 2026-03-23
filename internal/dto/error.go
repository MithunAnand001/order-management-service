package dto

import "net/http"

type ErrorCode string

const (
	ErrCodeInternal     ErrorCode = "INTERNAL_ERROR"
	ErrCodeNotFound     ErrorCode = "NOT_FOUND"
	ErrCodeValidation   ErrorCode = "VALIDATION_ERROR"
	ErrCodeUnauthorized ErrorCode = "UNAUTHORIZED"
	ErrCodeConflict     ErrorCode = "CONFLICT"
	ErrCodeBadRequest   ErrorCode = "BAD_REQUEST"
)

type AppError struct {
	Code       ErrorCode
	Message    string
	HTTPStatus int
	Err        error
}

func (e *AppError) Error() string {
	if e.Err != nil {
		return e.Message + ": " + e.Err.Error()
	}
	return e.Message
}

func NewAppError(code ErrorCode, message string, httpStatus int, err error) *AppError {
	return &AppError{
		Code:       code,
		Message:    message,
		HTTPStatus: httpStatus,
		Err:        err,
	}
}

func NewInternalError(err error) *AppError {
	return NewAppError(ErrCodeInternal, "Internal server error", http.StatusInternalServerError, err)
}

func NewNotFoundError(message string) *AppError {
	return NewAppError(ErrCodeNotFound, message, http.StatusNotFound, nil)
}

func NewValidationError(message string, err error) *AppError {
	return NewAppError(ErrCodeValidation, message, http.StatusBadRequest, err)
}
