package http

import (
	"net/http"
	"strconv"
	"time"

	"github.com/labstack/echo/v4"

	"github.com/semih-yildiz/notification-service/internal/application/notification/command/cancel"
	"github.com/semih-yildiz/notification-service/internal/application/notification/command/create"
	"github.com/semih-yildiz/notification-service/internal/application/notification/query/get"
	"github.com/semih-yildiz/notification-service/internal/application/notification/query/list"
	"github.com/semih-yildiz/notification-service/internal/domain/notification"
	"github.com/semih-yildiz/notification-service/internal/http/dto"
)

type NotificationHandler struct {
	createUsecase *create.UseCase
	cancelUsecase *cancel.UseCase
	getUsecase    *get.UseCase
	listUsecase   *list.UseCase
}

func NewNotificationHandler(
	createUsecase *create.UseCase,
	cancelUsecase *cancel.UseCase,
	getUsecase *get.UseCase,
	listUsecase *list.UseCase,
) *NotificationHandler {
	return &NotificationHandler{
		createUsecase: createUsecase,
		cancelUsecase: cancelUsecase,
		getUsecase:    getUsecase,
		listUsecase:   listUsecase,
	}
}

func RegisterNotificationRoutes(g *echo.Group, handler *NotificationHandler) {
	g.POST("/notifications", handler.CreateNotification)
	g.POST("/notifications/batches", handler.CreateNotificationBatches)
	g.GET("/notifications/:id", handler.GetByID)
	g.GET("/notifications", handler.List)
	g.POST("/notifications/:id/cancel", handler.Cancel)
	g.GET("/batches/:id/notifications", handler.GetBatch)
	g.POST("/batches/:id/cancel", handler.CancelBatch)
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

// GetByID handles GET /notifications/:id
func (h *NotificationHandler) GetByID(c echo.Context) error {
	ctx := c.Request().Context()
	id := c.Param("id")

	result, err := h.getUsecase.Notification(ctx, &get.ByID{ID: id})
	if err != nil {
		return mapNotificationError(c, err)
	}

	return c.JSON(http.StatusOK, result)
}

// List handles GET /notifications
func (h *NotificationHandler) List(c echo.Context) error {
	ctx := c.Request().Context()

	var fromTime, toTime *time.Time
	if s := c.QueryParam("from"); s != "" {
		if t, err := time.Parse(time.RFC3339, s); err == nil {
			fromTime = &t
		}
	}
	if s := c.QueryParam("to"); s != "" {
		if t, err := time.Parse(time.RFC3339, s); err == nil {
			toTime = &t
		}
	}

	var batchID *string
	if s := c.QueryParam("batch_id"); s != "" {
		batchID = &s
	}

	var status *notification.Status
	var ch *notification.Channel
	if s := c.QueryParam("status"); s != "" {
		st := notification.Status(s)
		if st.Valid() {
			status = &st
		}
	}
	if s := c.QueryParam("channel"); s != "" {
		chVal := notification.Channel(s)
		if chVal.Valid() {
			ch = &chVal
		}
	}

	limit := 100
	if l := c.QueryParam("limit"); l != "" {
		if n, err := strconv.Atoi(l); err == nil && n > 0 && n <= 1000 {
			limit = n
		}
	}

	offset := 0
	if o := c.QueryParam("offset"); o != "" {
		if n, err := strconv.Atoi(o); err == nil && n >= 0 {
			offset = n
		}
	}

	query := &list.Query{
		Status:   status,
		Channel:  ch,
		FromTime: fromTime,
		ToTime:   toTime,
		BatchID:  batchID,
		Limit:    limit,
		Offset:   offset,
	}

	result, err := h.listUsecase.ListByQuery(ctx, query)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "internal error"})
	}

	return c.JSON(http.StatusOK, result)
}

// Cancel handles POST /notifications/:id/cancel
func (h *NotificationHandler) Cancel(c echo.Context) error {
	ctx := c.Request().Context()
	id := c.Param("id")

	cmd := &cancel.Command{NotificationID: id}
	if err := h.cancelUsecase.CancelPendingNotification(ctx, cmd); err != nil {
		return mapNotificationError(c, err)
	}

	return c.NoContent(http.StatusNoContent)
}

// GetBatch handles GET /batches/:id/notifications
func (h *NotificationHandler) GetBatch(c echo.Context) error {
	ctx := c.Request().Context()
	batchID := c.Param("id")

	batch, notifications, err := h.getUsecase.Batch(ctx, &get.BatchByID{BatchID: batchID})
	if err != nil {
		return mapNotificationError(c, err)
	}

	response := dto.BatchWithNotificationsResponse{
		Batch:         batch,
		Notifications: notifications,
	}

	return c.JSON(http.StatusOK, response)
}

// CancelBatch handles POST /batches/:id/cancel
func (h *NotificationHandler) CancelBatch(c echo.Context) error {
	ctx := c.Request().Context()
	batchID := c.Param("id")

	cmd := &cancel.BatchCommand{BatchID: batchID}
	count, err := h.cancelUsecase.CancelPendingNotificationBatch(ctx, cmd)
	if err != nil {
		return mapNotificationError(c, err)
	}

	response := dto.CancelBatchResponse{Cancelled: count}
	return c.JSON(http.StatusOK, response)
}
