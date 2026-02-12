package http

import (
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

// NewEcho creates Echo instance with basic middleware.
func NewEcho() *echo.Echo {
	e := echo.New()
	e.HideBanner = true

	// Basic middleware
	e.Use(middleware.Recover())
	e.Use(middleware.RequestID())

	// Simple health endpoint (no dependencies)
	e.GET("/health", func(c echo.Context) error {
		return c.JSON(http.StatusOK, map[string]string{"status": "ok"})
	})

	return e
}
