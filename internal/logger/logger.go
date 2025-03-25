package logger

import (
	"context"
	"fmt"

	"github.com/flexer2006/case-back-restaurant-go/common"
	"github.com/flexer2006/case-back-restaurant-go/internal/logger/adapters/zap_adapter"
	"github.com/flexer2006/case-back-restaurant-go/internal/logger/ports"

	"go.uber.org/zap"
)

var defaultFactory ports.LoggerFactory = zap_adapter.NewZapLoggerFactory()
var defaultLogger ports.LoggerPort

func init() {
	var err error
	defaultLogger, err = defaultFactory.NewLogger()
	if err != nil {
		panic(fmt.Sprintf("%s: %v", common.ErrInitDefaultLogger, err))
	}
}

func NewLogger() (ports.LoggerPort, error) {
	return defaultFactory.NewLogger()
}

func SetLoggerFactory(factory ports.LoggerFactory) {
	defaultFactory = factory
	var err error
	defaultLogger, err = defaultFactory.NewLogger()
	if err != nil {
		panic(fmt.Sprintf("%s: %v", common.ErrInitDefaultLogger, err))
	}
}

func Info(ctx context.Context, msg string, fields ...zap.Field) {
	defaultLogger.Info(ctx, msg, fields...)
}

func Warn(ctx context.Context, msg string, fields ...zap.Field) {
	defaultLogger.Warn(ctx, msg, fields...)
}

func Error(ctx context.Context, msg string, fields ...zap.Field) {
	defaultLogger.Error(ctx, msg, fields...)
}

func Debug(ctx context.Context, msg string, fields ...zap.Field) {
	defaultLogger.Debug(ctx, msg, fields...)
}

func Fatal(ctx context.Context, msg string, fields ...zap.Field) {
	defaultLogger.Fatal(ctx, msg, fields...)
}
