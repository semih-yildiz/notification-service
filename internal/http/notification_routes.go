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
	g.POST("/notifications/batches", handler.CreateNotificationBatches)
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

// CreateNotificationBatches creates multiple notifications (batch).
func (h *NotificationHandler) CreateNotificationBatches(c echo.Context) error {
	ctx := c.Request().Context()

	var items []dto.NotificationItem
	if err := c.Bind(&items); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid json array"})
	}

	if len(items) == 0 || len(items) > 1000 {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "batch must contain 1-1000 items"})
	}

	for i := range items {
		if err := items[i].Validate(); err != nil {
			return c.JSON(http.StatusBadRequest, map[string]string{"error": err.Error()})
		}
	}

	batchItems := make([]create.BatchItem, len(items))
	for i, item := range items {
		batchItems[i] = create.BatchItem{
			Recipient: item.Recipient,
			Channel:   item.Channel,
			Content:   item.Content,
			Priority:  item.Priority,
		}
	}

	var batchIdempotencyKey *string
	if len(items) > 0 && items[0].IdempotencyKey != nil {
		batchIdempotencyKey = items[0].IdempotencyKey
	}

	cmd := &create.BatchCommand{
		Items:          batchItems,
		IdempotencyKey: batchIdempotencyKey,
	}

	result, err := h.createUsecase.CreateNotificationBatches(ctx, cmd)
	if err != nil {
		return mapNotificationError(c, err)
	}

	return c.JSON(http.StatusCreated, result)
}
