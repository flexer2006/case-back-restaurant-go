package utils

import (
	"context"

	"go.uber.org/zap"
)

type requestIDKey struct{}

var RequestID = requestIDKey{}

func GetRequestID(ctx context.Context) (string, bool) {
	id, ok := ctx.Value(RequestID).(string)
	return id, ok
}

func AddRequestID(ctx context.Context, fields []zap.Field) []zap.Field {
	if id, ok := GetRequestID(ctx); ok {
		fields = append(fields, zap.String("request_id", id))
	}
	return fields
}
