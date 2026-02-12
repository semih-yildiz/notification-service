package process

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"

	"github.com/semih-yildiz/notification-service/internal/application/notification/port"
	"github.com/semih-yildiz/notification-service/internal/domain/notification"
)

const maxDeliveryAttempts = 5
const initialBackoff = time.Second

// UseCase processes a notification: rate limit, deliver with retry, update status.
type UseCase struct {
	notifRepo   port.NotificationRepository
	attemptRepo port.DeliveryAttemptRepository
	rateLimit   port.RateLimiter
	delivery    port.DeliveryClient
	log         port.Logger
}

// NewUseCase returns a new process use case.
func NewUseCase(
	notifRepo port.NotificationRepository,
	attemptRepo port.DeliveryAttemptRepository,
	rateLimit port.RateLimiter,
	delivery port.DeliveryClient,
	log port.Logger,
) *UseCase {
	return &UseCase{
		notifRepo:   notifRepo,
		attemptRepo: attemptRepo,
		rateLimit:   rateLimit,
		delivery:    delivery,
		log:         log,
	}
}

// Execute processes one notification.
func (u *UseCase) Execute(ctx context.Context, cmd *Command) error {
	u.log.Info(ctx, "processing notification", port.F("notification_id", cmd.NotificationID))

	n, err := u.notifRepo.GetByID(ctx, cmd.NotificationID)
	if err != nil {
		u.log.Error(ctx, "failed to get notification", port.F("error", err), port.F("notification_id", cmd.NotificationID))
		return err
	}

	if n.Status.Terminal() {
		u.log.Info(ctx, "notification already in terminal state", port.F("notification_id", cmd.NotificationID), port.F("status", n.Status))
		return nil
	}

	allowed, err := u.rateLimit.Allow(ctx, n.Channel)
	if err != nil || !allowed {
		if !allowed {
			u.log.Warn(ctx, "rate limit exceeded", port.F("notification_id", cmd.NotificationID), port.F("channel", n.Channel))
			return fmt.Errorf("rate limit exceeded for channel %s", n.Channel)
		}
		u.log.Error(ctx, "rate limiter error", port.F("error", err), port.F("notification_id", cmd.NotificationID))
		return err
	}

	req := &port.DeliveryRequest{
		To:      n.Recipient,
		Channel: n.Channel.String(),
		Content: n.Content,
	}

	var lastErr error
	var lastCode int
	backoff := initialBackoff

	for attempt := 1; attempt <= maxDeliveryAttempts; attempt++ {
		u.log.Info(ctx, "delivery attempt", port.F("notification_id", cmd.NotificationID), port.F("attempt", attempt), port.F("channel", n.Channel))

		resp, code, err := u.delivery.Deliver(ctx, req)

		da := &notification.DeliveryAttempt{
			ID:             uuid.New().String(),
			NotificationID: cmd.NotificationID,
			AttemptNumber:  attempt,
			StatusCode:     code,
			CreatedAt:      time.Now(),
		}

		if err != nil {
			lastErr = err
			lastCode = code
			msg := err.Error()
			da.Success = false
			da.ErrorMessage = &msg
			if err := u.attemptRepo.Create(ctx, da); err != nil {
				u.log.Error(ctx, "failed to record delivery attempt", port.F("error", err), port.F("notification_id", cmd.NotificationID), port.F("attempt", attempt))
			}
			u.log.Warn(ctx, "delivery attempt failed", port.F("notification_id", cmd.NotificationID), port.F("attempt", attempt), port.F("status_code", code), port.F("error", err))

			if attempt < maxDeliveryAttempts {
				u.log.Info(ctx, "retrying after backoff", port.F("notification_id", cmd.NotificationID), port.F("backoff_ms", backoff.Milliseconds()))
				time.Sleep(backoff)
				backoff *= 2
			}
			continue
		}

		da.Success = true
		if resp != nil {
			da.ResponseBody = resp.MessageID
		}
		if err := u.attemptRepo.Create(ctx, da); err != nil {
			u.log.Error(ctx, "failed to record successful delivery attempt", port.F("error", err), port.F("notification_id", cmd.NotificationID), port.F("attempt", attempt))
		}

		now := time.Now()
		if err := u.notifRepo.UpdateStatus(ctx, cmd.NotificationID, notification.StatusSent, &now, nil); err != nil {
			u.log.Error(ctx, "failed to update status to sent", port.F("error", err), port.F("notification_id", cmd.NotificationID))
			// Don't fail the delivery since it was successful
		}
		u.log.Info(ctx, "notification delivered successfully", port.F("notification_id", cmd.NotificationID), port.F("attempt", attempt), port.F("message_id", da.ResponseBody))
		return nil
	}

	reason := fmt.Sprintf("failed after %d attempts: %v", maxDeliveryAttempts, lastErr)
	if lastCode > 0 {
		reason = fmt.Sprintf("status %d: %s", lastCode, reason)
	}
	if err := u.notifRepo.UpdateStatus(ctx, cmd.NotificationID, notification.StatusFailed, nil, &reason); err != nil {
		u.log.Error(ctx, "failed to update status to failed", port.F("error", err), port.F("notification_id", cmd.NotificationID))
		// Continue anyway since we want to return the delivery error
	}
	u.log.Error(ctx, "notification delivery failed permanently", port.F("notification_id", cmd.NotificationID), port.F("attempts", maxDeliveryAttempts), port.F("last_error", lastErr), port.F("last_code", lastCode))

	return lastErr
}
