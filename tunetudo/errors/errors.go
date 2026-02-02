package errors

import (
	"errors"
	"tunetudo/logger"
)

// AppError represents application-level errors with safe user messages
type AppError struct {
	Code       string // Error code for client reference
	Message    string // Safe message for end user (no sensitive info)
	StatusCode int    // HTTP status code
	Internal   error  // Internal error (not exposed to client, logged separately)
}

func (e *AppError) Error() string {
	// Return only safe message, not internal error
	return e.Message
}

// Common error codes
const (
	ErrCodeValidation    = "VALIDATION_ERROR"
	ErrCodeAuth          = "AUTH_ERROR"
	ErrCodeNotFound      = "NOT_FOUND"
	ErrCodeUnauthorized  = "UNAUTHORIZED"
	ErrCodeForbidden     = "FORBIDDEN"
	ErrCodeInternal      = "INTERNAL_ERROR"
	ErrCodeBadRequest    = "BAD_REQUEST"
	ErrCodeConflict      = "CONFLICT"
	ErrCodeRateLimit     = "RATE_LIMIT"
)

// NewAppError creates a new application error
func NewAppError(code, message string, statusCode int, internal error) *AppError {
	return &AppError{
		Code:       code,
		Message:    message,
		StatusCode: statusCode,
		Internal:   internal,
	}
}

// "Do not reveal sensitive information from the error to the user"

func ValidationError(message string, internal error) *AppError {
	// Log internal error details but don't expose to user
	if internal != nil {
		logger.Error(logger.CategoryValidation, "Validation failed", internal)
	}
	return NewAppError(ErrCodeValidation, message, 400, internal)
}

func AuthError(message string, internal error) *AppError {
	// Log authentication errors without exposing details
	if internal != nil {
		logger.Error(logger.CategoryAuth, "Authentication error", internal)
	}
	// Always return generic message: "Just tell them 'authorization failed'"
	return NewAppError(ErrCodeAuth, "Authorization failed", 401, internal)
}

func NotFoundError(message string) *AppError {
	return NewAppError(ErrCodeNotFound, message, 404, nil)
}

func UnauthorizedError(message string) *AppError {
	return NewAppError(ErrCodeUnauthorized, "Authorization required", 401, nil)
}

func ForbiddenError(message string) *AppError {
	return NewAppError(ErrCodeForbidden, "Access forbidden", 403, nil)
}

func InternalError(internal error) *AppError {
	// Log detailed error internally
	if internal != nil {
		logger.Error(logger.CategoryAPI, "Internal error occurred", internal)
	}
	// Return generic message to user
	return NewAppError(
		ErrCodeInternal,
		"An internal error occurred. Please try again later.",
		500,
		internal,
	)
}

func BadRequestError(message string) *AppError {
	return NewAppError(ErrCodeBadRequest, message, 400, nil)
}

func ConflictError(message string) *AppError {
	return NewAppError(ErrCodeConflict, message, 409, nil)
}

func RateLimitError() *AppError {
	return NewAppError(
		ErrCodeRateLimit,
		"Too many requests. Please try again later.",
		429,
		nil,
	)
}

// WrapError wraps a standard error into an AppError
func WrapError(err error) *AppError {
	if err == nil {
		return nil
	}

	var appErr *AppError
	if errors.As(err, &appErr) {
		return appErr
	}

	return InternalError(err)
}

// IsAppError checks if an error is an AppError
func IsAppError(err error) bool {
	var appErr *AppError
	return errors.As(err, &appErr)
}

// GetAppError extracts AppError from error
func GetAppError(err error) *AppError {
	var appErr *AppError
	if errors.As(err, &appErr) {
		return appErr
	}
	return nil
}