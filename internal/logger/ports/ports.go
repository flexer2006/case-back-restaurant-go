package ports

import (
	"context"

	"go.uber.org/zap"
)

type LogLevel string

const (
	DebugLevel LogLevel = "debug"
	InfoLevel  LogLevel = "info"
	WarnLevel  LogLevel = "warn"
	ErrorLevel LogLevel = "error"
	FatalLevel LogLevel = "fatal"
)

type LoggerPort interface {
	Info(ctx context.Context, msg string, fields ...zap.Field)
	Warn(ctx context.Context, msg string, fields ...zap.Field)
	Error(ctx context.Context, msg string, fields ...zap.Field)
	Debug(ctx context.Context, msg string, fields ...zap.Field)
	Fatal(ctx context.Context, msg string, fields ...zap.Field)
	SetLevel(level LogLevel)
	GetLevel() LogLevel
	With(fields ...zap.Field) LoggerPort
	Sync() error
}

type LoggerFactory interface {
	NewLogger() (LoggerPort, error)
}
