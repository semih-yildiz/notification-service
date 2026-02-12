package middleware

import (
	"context"
	"crypto/rand"
	"encoding/hex"

	"github.com/labstack/echo/v4"

	sharedctx "github.com/semih-yildiz/notification-service/internal/shared/context"
)

// CorrelationID returns middleware that adds correlation ID to request context.
func CorrelationID() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			id := c.Request().Header.Get("X-Correlation-ID")
			if id == "" {
				id = newCorrelationID()
			}
			c.Response().Header().Set("X-Correlation-ID", id)
			ctx := context.WithValue(c.Request().Context(), sharedctx.CorrelationIDKey(), id)
			c.SetRequest(c.Request().WithContext(ctx))
			return next(c)
		}
	}
}

func newCorrelationID() string {
	b := make([]byte, 8)
	_, _ = rand.Read(b)
	return hex.EncodeToString(b)
}
