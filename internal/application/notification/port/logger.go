package port

import "context"

type Logger interface {
	Info(ctx context.Context, msg string, fields ...Field)
	Error(ctx context.Context, msg string, fields ...Field)
	Warn(ctx context.Context, msg string, fields ...Field)
}

type Field struct {
	Key   string
	Value interface{}
}

func F(key string, value interface{}) Field {
	return Field{Key: key, Value: value}
}
