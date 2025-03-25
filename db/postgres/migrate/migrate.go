// Package migrate provides tools for managing PostgreSQL database migrations.
package migrate

import (
	"context"
	"errors"
	"fmt"
	"net"
	"strconv"

	"github.com/golang-migrate/migrate/v4"
	"go.uber.org/zap"

	"github.com/flexer2006/case-back-restaurant-go/common"
	"github.com/flexer2006/case-back-restaurant-go/configs"
	"github.com/flexer2006/case-back-restaurant-go/internal/logger"
)

func createDSN(cfg *configs.PostgresConfig) string {
	hostPort := net.JoinHostPort(cfg.Host, strconv.Itoa(cfg.Port))

	return fmt.Sprintf("postgres://%s:%s@%s/%s?sslmode=%s",
		cfg.Username, cfg.Password, hostPort, cfg.Database, cfg.SSLMode)
}

const DefaultMigrationsPath = "file://db/migrations"

func Migrate(ctx context.Context, cfg *configs.PostgresConfig, opts ...string) error {
	migrationsPath := DefaultMigrationsPath
	if len(opts) > 0 && opts[0] != "" {
		migrationsPath = opts[0]
	}

	return migrateWithPath(ctx, cfg, migrationsPath)
}

func MigrateTo(ctx context.Context, cfg *configs.PostgresConfig, version uint, migrationsPath string) error {
	if migrationsPath == "" {
		migrationsPath = DefaultMigrationsPath
	}

	log, err := logger.FromContext(ctx)
	if err != nil {
		log, logErr := logger.NewLogger()
		if logErr != nil {
			return fmt.Errorf("%s: %w", common.ErrLoggerCreation, logErr)
		}

		ctx = logger.NewContext(ctx, log)
	}

	dsn := createDSN(cfg)

	log.Info(ctx, common.MsgDBMigrationStarted,
		zap.String("source", migrationsPath),
		zap.String("host", cfg.Host),
		zap.Int("port", cfg.Port),
		zap.Uint("target_version", version))

	handler := NewHandler()
	m, err := handler.Migrate(migrationsPath, dsn)
	if err != nil {
		log.Error(ctx, common.ErrMigrateInstanceCreation, zap.Error(err))

		return fmt.Errorf("%s: %w", common.ErrMigrateInstanceCreation, err)
	}

	defer func() {
		srcErr, dbErr := m.Close()
		if srcErr != nil {
			log.Error(ctx, common.ErrCloseMigrationSource, zap.Error(srcErr))
		}

		if dbErr != nil {
			log.Error(ctx, common.ErrCloseDBConnection, zap.Error(dbErr))
		}
	}()

	if err := m.MigrateTo(version); err != nil && !errors.Is(err, migrate.ErrNoChange) {
		log.Error(ctx, common.ErrMigrateToVersion, zap.Uint("version", version), zap.Error(err))

		return fmt.Errorf("%s to version %d: %w", common.ErrMigrateToVersion, version, err)
	}

	currentVersion, dirty, err := m.Version()
	if err != nil && !errors.Is(err, migrate.ErrNilVersion) {
		log.Error(ctx, common.ErrMigrateVersion, zap.Error(err))

		return fmt.Errorf("%s: %w", common.ErrMigrateVersion, err)
	}

	if dirty {
		log.Error(ctx, common.ErrMigrateDirtyState, zap.Uint("version", currentVersion))

		return fmt.Errorf("%s: version %d", common.ErrMigrateDirtyState, currentVersion)
	}

	log.Info(ctx, common.MsgDBMigrationCompleted,
		zap.Uint("version", currentVersion),
		zap.Bool("dirty", dirty))

	return nil
}

func migrateWithPath(ctx context.Context, cfg *configs.PostgresConfig, migrationsPath string) error {
	log, err := logger.FromContext(ctx)
	if err != nil {
		log, logErr := logger.NewLogger()
		if logErr != nil {
			return fmt.Errorf("%s: %w", common.ErrLoggerCreation, logErr)
		}

		ctx = logger.NewContext(ctx, log)
	}

	dsn := createDSN(cfg)

	log.Info(ctx, common.MsgDBMigrationStarted,
		zap.String("source", migrationsPath),
		zap.String("host", cfg.Host),
		zap.Int("port", cfg.Port))

	handler := NewHandler()
	m, err := handler.Migrate(migrationsPath, dsn)
	if err != nil {
		log.Error(ctx, common.ErrMigrateInstanceCreation, zap.Error(err))

		return fmt.Errorf("%s: %w", common.ErrMigrateInstanceCreation, err)
	}

	defer func() {
		srcErr, dbErr := m.Close()
		if srcErr != nil {
			log.Error(ctx, common.ErrCloseMigrationSource, zap.Error(srcErr))
		}

		if dbErr != nil {
			log.Error(ctx, common.ErrCloseDBConnection, zap.Error(dbErr))
		}
	}()

	if err := m.Up(); err != nil && !errors.Is(err, migrate.ErrNoChange) {
		log.Error(ctx, common.ErrMigrateApply, zap.Error(err))

		return fmt.Errorf("%s: %w", common.ErrMigrateApply, err)
	}

	version, dirty, err := m.Version()
	if err != nil && !errors.Is(err, migrate.ErrNilVersion) {
		log.Error(ctx, common.ErrMigrateVersion, zap.Error(err))

		return fmt.Errorf("%s: %w", common.ErrMigrateVersion, err)
	}

	if dirty {
		log.Error(ctx, common.ErrMigrateDirtyState, zap.Uint("version", version))

		return fmt.Errorf("%s: version %d", common.ErrMigrateDirtyState, version)
	}

	log.Info(ctx, common.MsgDBMigrationCompleted,
		zap.Uint("version", version),
		zap.Bool("dirty", dirty))

	return nil
}
