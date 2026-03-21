package httpx

import "net/http"

type AppError struct {
	Status  int         `json:"-"`
	Code    string      `json:"code"`
	Message string      `json:"message"`
	Details interface{} `json:"details,omitempty"`
}

func (e *AppError) Error() string {
	return e.Message
}

func NewError(status int, code, message string) *AppError {
	return &AppError{
		Status:  status,
		Code:    code,
		Message: message,
	}
}

func BadRequest(code, message string) *AppError {
	return NewError(http.StatusBadRequest, code, message)
}

func Unauthorized(code, message string) *AppError {
	return NewError(http.StatusUnauthorized, code, message)
}

func Forbidden(code, message string) *AppError {
	return NewError(http.StatusForbidden, code, message)
}

func NotFound(code, message string) *AppError {
	return NewError(http.StatusNotFound, code, message)
}

func Conflict(code, message string) *AppError {
	return NewError(http.StatusConflict, code, message)
}

func Internal(code string) *AppError {
	return NewError(http.StatusInternalServerError, code, "internal server error")
}
