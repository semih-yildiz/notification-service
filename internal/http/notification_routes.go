package http

import (
	"net/http"

	"github.com/labstack/echo/v4"

	"github.com/semih-yildiz/notification-service/internal/application/notification/command/create"
	"github.com/semih-yildiz/notification-service/internal/http/dto"
)

type NotificationHandler struct {
	createUsecase *create.UseCase
}

func NewNotificationHandler(createUsecase *create.UseCase) *NotificationHandler {
	return &NotificationHandler{
		createUsecase: createUsecase,
	}
}

func RegisterNotificationRoutes(g *echo.Group, handler *NotificationHandler) {
	g.POST("/notifications", handler.CreateNotification)
}

func (h *NotificationHandler) CreateNotification(c echo.Context) error {
	ctx := c.Request().Context()

	var item dto.NotificationItem
	if err := c.Bind(&item); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid json object"})
	}

	if err := item.Validate(); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": err.Error()})
	}

	cmd := &create.Command{
		Recipient:      item.Recipient,
		Channel:        item.Channel,
		Content:        item.Content,
		Priority:       item.Priority,
		IdempotencyKey: item.IdempotencyKey,
	}

	result, err := h.createUsecase.CreateNotification(ctx, cmd)
	if err != nil {
		return mapNotificationError(c, err)
	}

	return c.JSON(http.StatusCreated, result)
}
