package http

import (
	"net/http"

	"github.com/labstack/echo/v4"

	"github.com/semih-yildiz/notification-service/internal/domain/notification"
	"github.com/semih-yildiz/notification-service/internal/http/dto"
)

// mapNotificationError maps domain errors to standardized HTTP error responses.
func mapNotificationError(c echo.Context, err error) error {
	var errResp *dto.ErrorResponse
	var statusCode int

	switch err {
	case notification.ErrNotFound:
		errResp = dto.NewErrorResponse(dto.ErrCodeNotFound, "notification or batch not found")
		statusCode = http.StatusNotFound

	case notification.ErrInvalidChannel:
		errResp = dto.NewErrorResponse(dto.ErrCodeValidation, "invalid channel: must be sms, email, or push")
		statusCode = http.StatusBadRequest

	case notification.ErrInvalidPriority:
		errResp = dto.NewErrorResponse(dto.ErrCodeValidation, "invalid priority: must be high, normal, or low")
		statusCode = http.StatusBadRequest

	case notification.ErrInvalidContent:
		errResp = dto.NewErrorResponse(dto.ErrCodeValidation, "invalid content: check character limits and required fields")
		statusCode = http.StatusBadRequest

	case notification.ErrDuplicateRequest:
		errResp = dto.NewErrorResponse(dto.ErrCodeDuplicateRequest, "duplicate request: idempotency key already used")
		statusCode = http.StatusConflict

	case notification.ErrBatchTooLarge:
		errResp = dto.NewErrorResponseWithDetails(
			dto.ErrCodeValidation,
			"batch size exceeds maximum",
			map[string]interface{}{"max_size": 1000},
		)
		statusCode = http.StatusBadRequest

	case notification.ErrAlreadyTerminal:
		errResp = dto.NewErrorResponse(dto.ErrCodeConflict, "notification already in terminal state")
		statusCode = http.StatusConflict

	default:
		errResp = dto.NewErrorResponse(dto.ErrCodeInternalServerError, "internal server error")
		statusCode = http.StatusInternalServerError
	}

	// Add request ID from context if available
	if reqID := c.Response().Header().Get(echo.HeaderXRequestID); reqID != "" {
		errResp = errResp.WithRequestID(reqID)
	}

	return c.JSON(statusCode, errResp)
}
