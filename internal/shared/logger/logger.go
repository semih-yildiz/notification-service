package logger

import (
	"context"
	"log/slog"
	"os"

	"github.com/semih-yildiz/notification-service/internal/application/notification/port"
	sharedctx "github.com/semih-yildiz/notification-service/internal/shared/context"
)

var _ port.Logger = (*SlogLogger)(nil)

type SlogLogger struct {
	log *slog.Logger
}

// New returns a structured JSON logger.
func New() *SlogLogger {
	return &SlogLogger{
		log: slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
			Level: slog.LevelInfo,
		})),
	}
}

func (l *SlogLogger) Info(ctx context.Context, msg string, fields ...port.Field) {
	attrs := toAttrs(ctx, fields)
	l.log.InfoContext(ctx, msg, attrs...)
}

func (l *SlogLogger) Error(ctx context.Context, msg string, fields ...port.Field) {
	attrs := toAttrs(ctx, fields)
	l.log.ErrorContext(ctx, msg, attrs...)
}

func (l *SlogLogger) Warn(ctx context.Context, msg string, fields ...port.Field) {
	attrs := toAttrs(ctx, fields)
	l.log.WarnContext(ctx, msg, attrs...)
}

func toAttrs(ctx context.Context, fields []port.Field) []any {
	attrs := make([]any, 0, len(fields)*2+2)
	// Add correlation ID from context if available
	if cid := ctx.Value(sharedctx.CorrelationIDKey()); cid != nil {
		attrs = append(attrs, "correlation_id", cid)
	}
	for _, f := range fields {
		attrs = append(attrs, f.Key, f.Value)
	}
	return attrs
}
