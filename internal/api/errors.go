package api

import (
	"errors"
	"net/http"
)

type AppError struct {
	Code    int    `json:"-"`
	Message string `json:"error"`
}

func (e *AppError) Error() string {
	return e.Message
}

var (
	ErrBadRequest          = &AppError{Code: http.StatusBadRequest, Message: "bad request"}
	ErrUnauthorized        = &AppError{Code: http.StatusUnauthorized, Message: "unauthorized"}
	ErrForbidden           = &AppError{Code: http.StatusForbidden, Message: "forbidden"}
	ErrNotFound            = &AppError{Code: http.StatusNotFound, Message: "not found"}
	ErrConflict            = &AppError{Code: http.StatusConflict, Message: "conflict"}
	ErrInternalServer      = &AppError{Code: http.StatusInternalServerError, Message: "internal server error"}
	ErrInvalidCredentials  = &AppError{Code: http.StatusUnauthorized, Message: "invalid email or password"}
	ErrEmailAlreadyExists  = &AppError{Code: http.StatusConflict, Message: "email already registered"}
	ErrInvalidToken        = &AppError{Code: http.StatusUnauthorized, Message: "invalid or expired token"}
	ErrOwnershipViolation  = &AppError{Code: http.StatusForbidden, Message: "access denied: ownership mismatch"}
	ErrValidation          = &AppError{Code: http.StatusBadRequest, Message: "validation error"}
)

func NewBadRequestError(msg string) *AppError {
	return &AppError{Code: http.StatusBadRequest, Message: msg}
}

func NewNotFoundError(msg string) *AppError {
	return &AppError{Code: http.StatusNotFound, Message: msg}
}

func NewConflictError(msg string) *AppError {
	return &AppError{Code: http.StatusConflict, Message: msg}
}

func NewValidationError(msg string) *AppError {
	return &AppError{Code: http.StatusBadRequest, Message: msg}
}

func HandleError(w http.ResponseWriter, err error) {
	var appErr *AppError
	if errors.As(err, &appErr) {
		JSONErrorMessage(w, appErr.Code, appErr.Message)
		return
	}
	JSONErrorMessage(w, http.StatusInternalServerError, "internal server error")
}
