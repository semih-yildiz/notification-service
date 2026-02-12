package http

import (
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/redis/go-redis/v9"

	"github.com/semih-yildiz/notification-service/internal/application/notification/port"
	"github.com/semih-yildiz/notification-service/internal/infrastructure/messaging/rabbitmq"
)

func RegisterHealthRoutes(e *echo.Echo, handler *HealthHandler) {
	e.GET("/health", handler.Health)
	e.GET("/metrics", handler.Metrics)
}

type HealthHandler struct {
	db              interface{ Ping() error }
	redis           *redis.Client
	metricsProvider port.MetricsProvider
	mqManagement    *rabbitmq.ManagementClient
}

func NewHealthHandler(
	db interface{ Ping() error },
	redisClient *redis.Client,
	metricsProvider port.MetricsProvider,
	mqManagement *rabbitmq.ManagementClient,
) *HealthHandler {
	return &HealthHandler{
		db:              db,
		redis:           redisClient,
		metricsProvider: metricsProvider,
		mqManagement:    mqManagement,
	}
}

func (h *HealthHandler) Health(c echo.Context) error {
	ctx := c.Request().Context()
	status := http.StatusOK

	if h.db != nil && h.db.Ping() != nil {
		status = http.StatusServiceUnavailable
	}

	if h.redis != nil && h.redis.Ping(ctx).Err() != nil {
		status = http.StatusServiceUnavailable
	}

	if status == http.StatusOK {
		return c.JSON(http.StatusOK, map[string]string{"status": "ok"})
	}

	return c.JSON(http.StatusServiceUnavailable, map[string]string{"status": "unhealthy"})
}

func (h *HealthHandler) Metrics(c echo.Context) error {
	ctx := c.Request().Context()

	if h.metricsProvider == nil {
		return c.String(http.StatusServiceUnavailable, "metrics provider not configured")
	}

	// Get DB metrics
	stats, err := h.metricsProvider.GetNotificationStats(ctx)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}

	response := map[string]interface{}{
		"notifications": map[string]int64{
			"pending": stats.Pending,
			"queued":  stats.Queued,
			"sent":    stats.Sent,
			"failed":  stats.Failed,
			"total":   stats.Total,
		},
		"success_rate": calculateRate(stats.Sent, stats.Total),
		"failure_rate": calculateRate(stats.Failed, stats.Total),
	}

	// Get RabbitMQ queue depths (if management client configured)
	if h.mqManagement != nil {
		depths, err := h.mqManagement.GetQueueDepths(ctx)
		if err == nil {
			queueMetrics := make(map[string]interface{})
			for _, d := range depths {
				queueMetrics[d.Queue] = map[string]int{
					"total":   d.Depth,
					"ready":   d.Ready,
					"unacked": d.Unacked,
				}
			}
			response["queues"] = queueMetrics
		}
	}

	return c.JSON(http.StatusOK, response)
}

func calculateRate(part, total int64) float64 {
	if total == 0 {
		return 0.0
	}
	return float64(part) / float64(total) * 100.0
}
