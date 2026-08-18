package apperrors

import (
	"errors"
	"fmt"
)

// Sentinel errors representing common failure types in our pipeline
var (
	ErrNotFound     = errors.New("resource not found")
	ErrInvalidInput = errors.New("invalid record or input data")
	ErrJobCancelled = errors.New("pipeline job cancelled by user")
	ErrJobFailed    = errors.New("pipeline job execution failed")
	ErrDatabaseOp   = errors.New("database operation failed")
)

// AppError represents a structured error with HTTP status code support
type AppError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Err     error  `json:"-"`
}

func (e *AppError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("%s: %v", e.Message, e.Err)
	}
	return e.Message
}

func (e *AppError) Unwrap() error {
	return e.Err
}

// constructors for basic API errors
func NewBadRequest(msg string, err error) *AppError {
	return &AppError{Code: 400, Message: msg, Err: err}
}

func NewNotFound(msg string, err error) *AppError {
	return &AppError{Code: 404, Message: msg, Err: err}
}

func NewInternalServerError(msg string, err error) *AppError {
	return &AppError{Code: 500, Message: msg, Err: err}
}
