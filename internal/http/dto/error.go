package dto

import "time"

// It follows RFC 7807 Problem Details for HTTP APIs principles.
type ErrorResponse struct {
	Success   bool         `json:"success"`
	Error     *ErrorDetail `json:"error"`
	RequestID string       `json:"request_id,omitempty"`
	Timestamp time.Time    `json:"timestamp"`
}

type ErrorDetail struct {
	Code    string                 `json:"code"`
	Message string                 `json:"message"`
	Details map[string]interface{} `json:"details,omitempty"`
}

const (
	ErrCodeBadRequest          = "bad_request"
	ErrCodeNotFound            = "not_found"
	ErrCodeConflict            = "conflict"
	ErrCodeInternalServerError = "internal_server_error"
	ErrCodeValidation          = "validation_error"
	ErrCodeDuplicateRequest    = "duplicate_request"
	ErrCodeRateLimitExceeded   = "rate_limit_exceeded"
	ErrCodeUnauthorized        = "unauthorized"
	ErrCodeForbidden           = "forbidden"
)

func NewErrorResponse(code, message string) *ErrorResponse {
	return &ErrorResponse{
		Success: false,
		Error: &ErrorDetail{
			Code:    code,
			Message: message,
		},
		Timestamp: time.Now(),
	}
}

func NewErrorResponseWithDetails(code, message string, details map[string]interface{}) *ErrorResponse {
	return &ErrorResponse{
		Success: false,
		Error: &ErrorDetail{
			Code:    code,
			Message: message,
			Details: details,
		},
		Timestamp: time.Now(),
	}
}

func (e *ErrorResponse) WithRequestID(requestID string) *ErrorResponse {
	e.RequestID = requestID
	return e
}

type ValidationError struct {
	Field   string `json:"field"`
	Message string `json:"message"`
}

// NewValidationErrorResponse creates a validation error response with field details.
func NewValidationErrorResponse(errors []ValidationError) *ErrorResponse {
	details := make(map[string]interface{})
	details["validation_errors"] = errors

	return &ErrorResponse{
		Success: false,
		Error: &ErrorDetail{
			Code:    ErrCodeValidation,
			Message: "validation failed",
			Details: details,
		},
		Timestamp: time.Now(),
	}
}
