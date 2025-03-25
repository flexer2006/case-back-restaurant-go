package logger_test

import (
	"context"
	"testing"

	"github.com/flexer2006/case-back-restaurant-go/internal/logger"
	"github.com/flexer2006/case-back-restaurant-go/internal/logger/adapters/zap_adapter"
	"github.com/flexer2006/case-back-restaurant-go/internal/logger/ports"
	"github.com/flexer2006/case-back-restaurant-go/internal/logger/utils"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"go.uber.org/zap/zaptest/observer"
)

func TestNewLogger(t *testing.T) {
	l, err := logger.NewLogger()
	require.NoError(t, err)
	require.NotNil(t, l)
}

func TestSetLoggerFactory(t *testing.T) {
	core, logs := observer.New(zapcore.InfoLevel)
	zapLogger := zap.New(core)

	factory := &testLoggerFactory{
		logger: &testLogger{
			zapLogger: zapLogger,
			logs:      logs,
		},
	}

	logger.SetLoggerFactory(factory)

	ctx := context.Background()
	logger.Info(ctx, "test message")

	require.Equal(t, 1, logs.Len())
	require.Equal(t, "test message", logs.All()[0].Message)
}

func TestLogLevels(t *testing.T) {
	core, logs := observer.New(zapcore.DebugLevel)
	zapLogger := zap.New(core)

	factory := &testLoggerFactory{
		logger: &testLogger{
			zapLogger: zapLogger,
			logs:      logs,
			level:     ports.DebugLevel,
		},
	}
	logger.SetLoggerFactory(factory)

	ctx := context.Background()

	logger.Debug(ctx, "debug message")
	require.Equal(t, 1, logs.Len())
	require.Equal(t, "debug message", logs.All()[0].Message)
	require.Equal(t, zapcore.DebugLevel, logs.All()[0].Level)
	logs.TakeAll()

	logger.Info(ctx, "info message")
	require.Equal(t, 1, logs.Len())
	require.Equal(t, "info message", logs.All()[0].Message)
	require.Equal(t, zapcore.InfoLevel, logs.All()[0].Level)
	logs.TakeAll()

	logger.Warn(ctx, "warn message")
	require.Equal(t, 1, logs.Len())
	require.Equal(t, "warn message", logs.All()[0].Message)
	require.Equal(t, zapcore.WarnLevel, logs.All()[0].Level)
	logs.TakeAll()

	logger.Error(ctx, "error message")
	require.Equal(t, 1, logs.Len())
	require.Equal(t, "error message", logs.All()[0].Message)
	require.Equal(t, zapcore.ErrorLevel, logs.All()[0].Level)
	logs.TakeAll()

	logger.Fatal(ctx, "fatal message")
	require.Equal(t, 1, logs.Len())
	require.Equal(t, "FATAL: fatal message", logs.All()[0].Message)
	require.Equal(t, zapcore.ErrorLevel, logs.All()[0].Level)
	logs.TakeAll()
}

func TestLoggerWithFields(t *testing.T) {
	strValue := "value"
	intValue := 123
	boolValue := true
	floatValue := 123.456
	errorValue := assert.AnError

	fields := logger.Fields(
		"string", strValue,
		"int", intValue,
		"bool", boolValue,
		"float", floatValue,
		"error", errorValue,
	)

	require.Len(t, fields, 5)

	fieldKeys := make(map[string]bool)
	for _, field := range fields {
		fieldKeys[field.Key] = true
	}

	expectedKeys := []string{"string", "int", "bool", "float", "error"}
	for _, key := range expectedKeys {
		require.True(t, fieldKeys[key], "missing key field %s", key)
	}

	core, _ := observer.New(zapcore.InfoLevel)
	zapLogger := zap.New(core)
	testFactory := &testLoggerFactory{
		logger: &testLogger{
			zapLogger: zapLogger,
			level:     ports.InfoLevel,
		},
	}

	logger.SetLoggerFactory(testFactory)

	loggerWithFields := testFactory.logger.With(fields...)
	require.NotNil(t, loggerWithFields)
}

func TestLoggerContext(t *testing.T) {
	l, err := logger.NewLogger()
	require.NoError(t, err)

	ctx := logger.NewContext(context.Background(), l)

	loggerFromCtx, err := logger.FromContext(ctx)
	require.NoError(t, err)
	require.NotNil(t, loggerFromCtx)

	require.Equal(t, l.GetLevel(), loggerFromCtx.GetLevel())

	_, err = logger.FromContext(context.Background())
	require.Error(t, err)
}

func TestRequestID(t *testing.T) {
	requestID := "test-request-id"
	ctx := context.WithValue(context.Background(), utils.RequestID, requestID)

	id, ok := utils.GetRequestID(ctx)
	require.True(t, ok)
	require.Equal(t, requestID, id)

	fields := []zap.Field{zap.String("test", "value")}
	fieldsWithRequestID := utils.AddRequestID(ctx, fields)

	require.Len(t, fieldsWithRequestID, 2)
	require.Equal(t, "test", fieldsWithRequestID[0].Key)
	require.Equal(t, "request_id", fieldsWithRequestID[1].Key)
	require.Equal(t, requestID, fieldsWithRequestID[1].String)
}

func TestZapLoggerAdapter(t *testing.T) {
	encoderConfig := zapcore.EncoderConfig{
		MessageKey:     "message",
		LevelKey:       "level",
		TimeKey:        "time",
		NameKey:        "logger",
		CallerKey:      "caller",
		StacktraceKey:  "stacktrace",
		LineEnding:     zapcore.DefaultLineEnding,
		EncodeLevel:    zapcore.LowercaseLevelEncoder,
		EncodeTime:     zapcore.ISO8601TimeEncoder,
		EncodeDuration: zapcore.SecondsDurationEncoder,
		EncodeCaller:   zapcore.ShortCallerEncoder,
	}

	config := zap.Config{
		Level:            zap.NewAtomicLevelAt(zapcore.InfoLevel),
		Development:      false,
		Encoding:         "json",
		EncoderConfig:    encoderConfig,
		OutputPaths:      []string{"stdout"},
		ErrorOutputPaths: []string{"stderr"},
	}

	factory := zap_adapter.NewZapLoggerFactoryWithConfig(config)
	l, err := factory.NewLogger()
	require.NoError(t, err)
	require.NotNil(t, l)

	l.SetLevel(ports.DebugLevel)
	require.Equal(t, ports.DebugLevel, l.GetLevel())

	l.SetLevel(ports.InfoLevel)
	require.Equal(t, ports.InfoLevel, l.GetLevel())

	l.SetLevel(ports.WarnLevel)
	require.Equal(t, ports.WarnLevel, l.GetLevel())

	l.SetLevel(ports.ErrorLevel)
	require.Equal(t, ports.ErrorLevel, l.GetLevel())

	l.SetLevel(ports.FatalLevel)
	require.Equal(t, ports.FatalLevel, l.GetLevel())

	newLogger := l.With(zap.String("key", "value"))
	require.NotNil(t, newLogger)
	require.Equal(t, l.GetLevel(), newLogger.GetLevel())
}

type testLoggerFactory struct {
	logger ports.LoggerPort
}

func (f *testLoggerFactory) NewLogger() (ports.LoggerPort, error) {
	return f.logger, nil
}

type testLogger struct {
	zapLogger *zap.Logger
	logs      *observer.ObservedLogs
	level     ports.LogLevel
}

func (l *testLogger) Info(ctx context.Context, msg string, fields ...zap.Field) {
	l.zapLogger.Info(msg, utils.AddRequestID(ctx, fields)...)
}

func (l *testLogger) Warn(ctx context.Context, msg string, fields ...zap.Field) {
	l.zapLogger.Warn(msg, utils.AddRequestID(ctx, fields)...)
}

func (l *testLogger) Error(ctx context.Context, msg string, fields ...zap.Field) {
	l.zapLogger.Error(msg, utils.AddRequestID(ctx, fields)...)
}

func (l *testLogger) Debug(ctx context.Context, msg string, fields ...zap.Field) {
	l.zapLogger.Debug(msg, utils.AddRequestID(ctx, fields)...)
}

func (l *testLogger) Fatal(ctx context.Context, msg string, fields ...zap.Field) {
	l.zapLogger.Error("FATAL: "+msg, utils.AddRequestID(ctx, fields)...)
}

func (l *testLogger) SetLevel(level ports.LogLevel) {
	l.level = level
}

func (l *testLogger) GetLevel() ports.LogLevel {
	return l.level
}

func (l *testLogger) With(fields ...zap.Field) ports.LoggerPort {
	return &testLogger{
		zapLogger: l.zapLogger.With(fields...),
		logs:      l.logs,
		level:     l.level,
	}
}

func (l *testLogger) Sync() error {
	return nil
}
