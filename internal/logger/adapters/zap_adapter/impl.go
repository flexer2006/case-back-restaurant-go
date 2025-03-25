package zap_adapter

import (
	"context"
	"github.com/flexer2006/case-back-restaurant-go/internal/logger/ports"
	"github.com/flexer2006/case-back-restaurant-go/internal/logger/utils"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type ZapLogger struct {
	l     *zap.Logger
	level ports.LogLevel
}

type ZapLoggerFactory struct {
	config zap.Config
}

func NewZapLoggerFactory() *ZapLoggerFactory {
	return &ZapLoggerFactory{
		config: zap.NewProductionConfig(),
	}
}

func NewZapLoggerFactoryWithConfig(config zap.Config) *ZapLoggerFactory {
	return &ZapLoggerFactory{
		config: config,
	}
}

func (f *ZapLoggerFactory) NewLogger() (ports.LoggerPort, error) {
	zapLogger, err := f.config.Build()
	if err != nil {
		return nil, err
	}
	return &ZapLogger{
		l:     zapLogger,
		level: ports.InfoLevel,
	}, nil
}

func (l *ZapLogger) Info(ctx context.Context, msg string, fields ...zap.Field) {
	l.l.Info(msg, utils.AddRequestID(ctx, fields)...)
}

func (l *ZapLogger) Warn(ctx context.Context, msg string, fields ...zap.Field) {
	l.l.Warn(msg, utils.AddRequestID(ctx, fields)...)
}

func (l *ZapLogger) Error(ctx context.Context, msg string, fields ...zap.Field) {
	l.l.Error(msg, utils.AddRequestID(ctx, fields)...)
}

func (l *ZapLogger) Debug(ctx context.Context, msg string, fields ...zap.Field) {
	l.l.Debug(msg, utils.AddRequestID(ctx, fields)...)
}

func (l *ZapLogger) Fatal(ctx context.Context, msg string, fields ...zap.Field) {
	l.l.Fatal(msg, utils.AddRequestID(ctx, fields)...)
}

func (l *ZapLogger) Sync() error {
	return l.l.Sync()
}

func (l *ZapLogger) SetLevel(level ports.LogLevel) {
	var zapLevel zapcore.Level

	switch level {
	case ports.DebugLevel:
		zapLevel = zap.DebugLevel
	case ports.InfoLevel:
		zapLevel = zap.InfoLevel
	case ports.WarnLevel:
		zapLevel = zap.WarnLevel
	case ports.ErrorLevel:
		zapLevel = zap.ErrorLevel
	case ports.FatalLevel:
		zapLevel = zap.FatalLevel
	default:
		zapLevel = zap.InfoLevel
		level = ports.InfoLevel
	}

	if atom, ok := l.l.Core().(zapcore.LevelEnabler); ok {
		if atomLevel, ok := atom.(*zap.AtomicLevel); ok {
			atomLevel.SetLevel(zapLevel)
		}
	}

	l.level = level
}

func (l *ZapLogger) GetLevel() ports.LogLevel {
	return l.level
}

func (l *ZapLogger) With(fields ...zap.Field) ports.LoggerPort {
	newZapLogger := l.l.With(fields...)
	return &ZapLogger{
		l:     newZapLogger,
		level: l.level,
	}
}
