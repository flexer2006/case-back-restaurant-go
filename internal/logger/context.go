package logger

import (
	"context"
	"errors"

	"github.com/flexer2006/case-back-restaurant-go/common"
	"github.com/flexer2006/case-back-restaurant-go/internal/logger/ports"
)

type loggerKey struct{}

func NewContext(ctx context.Context, logger ports.LoggerPort) context.Context {
	return context.WithValue(ctx, loggerKey{}, logger)
}

func FromContext(ctx context.Context) (ports.LoggerPort, error) {
	logger, ok := ctx.Value(loggerKey{}).(ports.LoggerPort)
	if !ok {
		return nil, errors.New(common.ErrLoggerNotFound)
	}

	return logger, nil
}
