package http

import (
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"

	httpmw "github.com/semih-yildiz/notification-service/internal/http/middleware"
)

// NewEcho creates Echo instance with middleware and routes.
func NewEcho(
	notificationHandler *NotificationHandler,
	basePath string,
) *echo.Echo {
	e := echo.New()
	e.HideBanner = true

	// Basic middleware
	e.Use(middleware.Recover())
	e.Use(middleware.RequestID())
	e.Use(httpmw.CorrelationID())

	// Simple health endpoint (no dependencies)
	e.GET("/health", func(c echo.Context) error {
		return c.JSON(http.StatusOK, map[string]string{"status": "ok"})
	})

	// API routes (base group)
	if notificationHandler != nil {
		g := e.Group(basePath)
		RegisterNotificationRoutes(g, notificationHandler)
	}

	return e
}
