package db_test

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/flexer2006/case-back-restaurant-go/common"
	"github.com/flexer2006/case-back-restaurant-go/configs"
	migratepkg "github.com/flexer2006/case-back-restaurant-go/db/postgres/migrate"
	"github.com/flexer2006/case-back-restaurant-go/internal/logger"
	"github.com/flexer2006/case-back-restaurant-go/internal/logger/ports"

	"github.com/golang-migrate/migrate/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"go.uber.org/zap"
)

type MockLogger struct {
	mock.Mock
	level ports.LogLevel
}

func (m *MockLogger) Info(ctx context.Context, msg string, fields ...zap.Field) {
	m.Called(ctx, msg, fields)
}

func (m *MockLogger) Warn(ctx context.Context, msg string, fields ...zap.Field) {
	m.Called(ctx, msg, fields)
}

func (m *MockLogger) Error(ctx context.Context, msg string, fields ...zap.Field) {
	m.Called(ctx, msg, fields)
}

func (m *MockLogger) Debug(ctx context.Context, msg string, fields ...zap.Field) {
	m.Called(ctx, msg, fields)
}

func (m *MockLogger) Fatal(ctx context.Context, msg string, fields ...zap.Field) {
	m.Called(ctx, msg, fields)
}

func (m *MockLogger) SetLevel(level ports.LogLevel) {
	m.level = level
}

func (m *MockLogger) GetLevel() ports.LogLevel {
	return m.level
}

func (m *MockLogger) With(fields ...zap.Field) ports.LoggerPort {
	return m
}

func (m *MockLogger) Sync() error {
	return nil
}

type MockLoggerFactory struct {
	Logger ports.LoggerPort
}

func (m *MockLoggerFactory) NewLogger() (ports.LoggerPort, error) {
	return m.Logger, nil
}

type MockMigrator struct {
	mock.Mock
}

func (m *MockMigrator) Up() error {
	args := m.Called()
	return args.Error(0)
}

func (m *MockMigrator) Down() error {
	args := m.Called()
	return args.Error(0)
}

func (m *MockMigrator) Version() (uint, bool, error) {
	args := m.Called()
	return args.Get(0).(uint), args.Bool(1), args.Error(2)
}

func (m *MockMigrator) Close() (source error, database error) {
	args := m.Called()
	return args.Error(0), args.Error(1)
}

func (m *MockMigrator) MigrateTo(version uint) error {
	args := m.Called(version)
	return args.Error(0)
}

type MockMigrationHandler struct {
	mock.Mock
}

func (m *MockMigrationHandler) Migrate(source, dsn string) (migratepkg.Migrator, error) {
	args := m.Called(source, dsn)
	return args.Get(0).(migratepkg.Migrator), args.Error(1)
}

func createDSN(cfg *configs.PostgresConfig) string {
	return fmt.Sprintf("postgres://%s:%s@%s:%d/%s?sslmode=%s",
		cfg.Username, cfg.Password, cfg.Host, cfg.Port, cfg.Database, cfg.SSLMode)
}

func TestMigrate(t *testing.T) {

	zapLogger, err := logger.NewLogger()
	assert.NoError(t, err)
	ctx := logger.NewContext(context.Background(), zapLogger)

	cfg := &configs.PostgresConfig{
		Host:     "localhost",
		Port:     5432,
		Username: "testuser",
		Password: "testpassword",
		Database: "testdb",
		SSLMode:  "disable",
	}

	mockMigrator := new(MockMigrator)
	mockHandler := new(MockMigrationHandler)

	mockMigrator.On("Up").Return(nil)
	mockMigrator.On("Version").Return(uint(2), false, nil)
	mockMigrator.On("Close").Return(nil, nil)

	dsn := createDSN(cfg)
	mockHandler.On("Migrate", migratepkg.DefaultMigrationsPath, dsn).Return(mockMigrator, nil)

	originalNewHandlerFunc := migratepkg.NewHandlerFunc
	migratepkg.NewHandlerFunc = func() migratepkg.MigrationHandler {
		return mockHandler
	}
	defer func() {
		migratepkg.NewHandlerFunc = originalNewHandlerFunc
	}()

	err = migratepkg.Migrate(ctx, cfg)

	assert.NoError(t, err)
	mockMigrator.AssertExpectations(t)
	mockHandler.AssertExpectations(t)
}

func TestMigrateError(t *testing.T) {

	zapLogger, err := logger.NewLogger()
	assert.NoError(t, err)
	ctx := logger.NewContext(context.Background(), zapLogger)

	cfg := &configs.PostgresConfig{
		Host:     "localhost",
		Port:     5432,
		Username: "testuser",
		Password: "testpassword",
		Database: "testdb",
		SSLMode:  "disable",
	}

	mockMigrator := new(MockMigrator)
	mockHandler := new(MockMigrationHandler)

	migrateErr := errors.New("migration error")
	mockMigrator.On("Up").Return(migrateErr)
	mockMigrator.On("Close").Return(nil, nil)

	dsn := createDSN(cfg)
	mockHandler.On("Migrate", migratepkg.DefaultMigrationsPath, dsn).Return(mockMigrator, nil)

	originalNewHandlerFunc := migratepkg.NewHandlerFunc
	migratepkg.NewHandlerFunc = func() migratepkg.MigrationHandler {
		return mockHandler
	}
	defer func() {
		migratepkg.NewHandlerFunc = originalNewHandlerFunc
	}()

	err = migratepkg.Migrate(ctx, cfg)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), common.ErrMigrateApply)
	mockMigrator.AssertExpectations(t)
	mockHandler.AssertExpectations(t)
}

func TestMigrateNoChange(t *testing.T) {

	zapLogger, err := logger.NewLogger()
	assert.NoError(t, err)
	ctx := logger.NewContext(context.Background(), zapLogger)

	cfg := &configs.PostgresConfig{
		Host:     "localhost",
		Port:     5432,
		Username: "testuser",
		Password: "testpassword",
		Database: "testdb",
		SSLMode:  "disable",
	}

	mockMigrator := new(MockMigrator)
	mockHandler := new(MockMigrationHandler)

	mockMigrator.On("Up").Return(migrate.ErrNoChange)
	mockMigrator.On("Version").Return(uint(2), false, nil)
	mockMigrator.On("Close").Return(nil, nil)

	dsn := createDSN(cfg)
	mockHandler.On("Migrate", migratepkg.DefaultMigrationsPath, dsn).Return(mockMigrator, nil)

	originalNewHandlerFunc := migratepkg.NewHandlerFunc
	migratepkg.NewHandlerFunc = func() migratepkg.MigrationHandler {
		return mockHandler
	}
	defer func() {
		migratepkg.NewHandlerFunc = originalNewHandlerFunc
	}()

	err = migratepkg.Migrate(ctx, cfg)
	assert.NoError(t, err)
	mockMigrator.AssertExpectations(t)
	mockHandler.AssertExpectations(t)
}

func TestMigrateTo(t *testing.T) {

	zapLogger, err := logger.NewLogger()
	assert.NoError(t, err)
	ctx := logger.NewContext(context.Background(), zapLogger)

	cfg := &configs.PostgresConfig{
		Host:     "localhost",
		Port:     5432,
		Username: "testuser",
		Password: "testpassword",
		Database: "testdb",
		SSLMode:  "disable",
	}

	targetVersion := uint(3)

	mockMigrator := new(MockMigrator)
	mockHandler := new(MockMigrationHandler)

	mockMigrator.On("MigrateTo", targetVersion).Return(nil)
	mockMigrator.On("Version").Return(targetVersion, false, nil)
	mockMigrator.On("Close").Return(nil, nil)

	dsn := createDSN(cfg)
	mockHandler.On("Migrate", migratepkg.DefaultMigrationsPath, dsn).Return(mockMigrator, nil)

	originalNewHandlerFunc := migratepkg.NewHandlerFunc
	migratepkg.NewHandlerFunc = func() migratepkg.MigrationHandler {
		return mockHandler
	}
	defer func() {
		migratepkg.NewHandlerFunc = originalNewHandlerFunc
	}()

	err = migratepkg.MigrateTo(ctx, cfg, targetVersion, "")

	assert.NoError(t, err)
	mockMigrator.AssertExpectations(t)
	mockHandler.AssertExpectations(t)
}

func TestMigrateDown(t *testing.T) {

	zapLogger, err := logger.NewLogger()
	assert.NoError(t, err)
	ctx := logger.NewContext(context.Background(), zapLogger)

	cfg := &configs.PostgresConfig{
		Host:     "localhost",
		Port:     5432,
		Username: "testuser",
		Password: "testpassword",
		Database: "testdb",
		SSLMode:  "disable",
	}

	mockMigrator := new(MockMigrator)
	mockHandler := new(MockMigrationHandler)

	mockMigrator.On("Down").Return(nil)
	mockMigrator.On("Close").Return(nil, nil)

	dsn := createDSN(cfg)
	mockHandler.On("Migrate", migratepkg.DefaultMigrationsPath, dsn).Return(mockMigrator, nil)

	originalNewHandlerFunc := migratepkg.NewHandlerFunc
	migratepkg.NewHandlerFunc = func() migratepkg.MigrationHandler {
		return mockHandler
	}
	defer func() {
		migratepkg.NewHandlerFunc = originalNewHandlerFunc
	}()

	MigrateDown := func(ctx context.Context, cfg *configs.PostgresConfig, migrationsPath string) error {
		if migrationsPath == "" {
			migrationsPath = migratepkg.DefaultMigrationsPath
		}

		log, err := logger.FromContext(ctx)
		if err != nil {
			return fmt.Errorf("%s: %w", common.ErrLoggerNotFound, err)
		}

		handler := migratepkg.NewHandler()
		dsn := createDSN(cfg)

		m, err := handler.Migrate(migrationsPath, dsn)
		if err != nil {
			return fmt.Errorf("%s: %w", common.ErrMigrateInstanceCreation, err)
		}

		var finalErr error

		defer func() {
			sourceErr, dbErr := m.Close()
			if finalErr == nil && (sourceErr != nil || dbErr != nil) {
				if sourceErr != nil {
					finalErr = fmt.Errorf("%s: %w", common.ErrCloseMigrationSource, sourceErr)
				} else {
					finalErr = fmt.Errorf("%s: %w", common.ErrCloseDBConnection, dbErr)
				}
			}
		}()

		if err := m.Down(); err != nil && !errors.Is(err, migrate.ErrNoChange) {
			log.Error(ctx, common.ErrMigrateDown, zap.Error(err))
			finalErr = fmt.Errorf("%s: %w", common.ErrMigrateDown, err)
		}

		return finalErr
	}

	err = MigrateDown(ctx, cfg, "")

	assert.NoError(t, err)
	mockMigrator.AssertExpectations(t)
	mockHandler.AssertExpectations(t)
}

func TestCloseError(t *testing.T) {

	zapLogger, err := logger.NewLogger()
	assert.NoError(t, err)
	ctx := logger.NewContext(context.Background(), zapLogger)

	cfg := &configs.PostgresConfig{
		Host:     "localhost",
		Port:     5432,
		Username: "testuser",
		Password: "testpassword",
		Database: "testdb",
		SSLMode:  "disable",
	}

	mockMigrator := new(MockMigrator)
	mockHandler := new(MockMigrationHandler)

	mockMigrator.On("Up").Return(nil)
	mockMigrator.On("Version").Return(uint(2), false, nil)
	mockMigrator.On("Close").Return(errors.New("source error"), errors.New("database error"))

	dsn := createDSN(cfg)
	mockHandler.On("Migrate", migratepkg.DefaultMigrationsPath, dsn).Return(mockMigrator, nil)

	originalNewHandlerFunc := migratepkg.NewHandlerFunc
	migratepkg.NewHandlerFunc = func() migratepkg.MigrationHandler {
		return mockHandler
	}
	defer func() {
		migratepkg.NewHandlerFunc = originalNewHandlerFunc
	}()

	err = migratepkg.Migrate(ctx, cfg)

	assert.NoError(t, err)
	mockMigrator.AssertExpectations(t)
	mockHandler.AssertExpectations(t)
}
