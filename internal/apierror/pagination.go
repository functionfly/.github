package apierror

import (
	"net/http"
)

type PaginationErrorCode ErrorCode

const (
	ErrCodeInvalidOffset PaginationErrorCode = "INVALID_OFFSET"
	ErrCodeInvalidLimit  PaginationErrorCode = "INVALID_LIMIT"
	ErrCodeInvalidCursor PaginationErrorCode = "INVALID_CURSOR"
)

func (e PaginationErrorCode) ToErrorCode() ErrorCode {
	return ErrorCode(e)
}

func NewInvalidOffset(message string) *APIError {
	return &APIError{Status: http.StatusBadRequest, Code: ErrCodeInvalidOffset.ToErrorCode(), Message: message}
}

func NewInvalidLimit(message string) *APIError {
	return &APIError{Status: http.StatusBadRequest, Code: ErrCodeInvalidLimit.ToErrorCode(), Message: message}
}

func NewInvalidCursor(message string) *APIError {
	return &APIError{Status: http.StatusBadRequest, Code: ErrCodeInvalidCursor.ToErrorCode(), Message: message}
}
