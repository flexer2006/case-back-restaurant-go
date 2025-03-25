// Package postgres provides database connection and interaction functionality
// using PostgreSQL as the backend database system.
package postgres

import (
	"context"
	"errors"
	"fmt"

	"github.com/flexer2006/case-back-restaurant-go/common"
	"github.com/flexer2006/case-back-restaurant-go/configs"
	migrate2 "github.com/flexer2006/case-back-restaurant-go/db/postgres/migrate"
	"github.com/flexer2006/case-back-restaurant-go/internal/logger"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/zap"
)

type DB struct {
	pool Pool
}

func New(ctx context.Context, cfg *configs.PostgresConfig) (Database, error) {
	log, err := logger.FromContext(ctx)
	if err != nil {
		log, logErr := logger.NewLogger()
		if logErr != nil {
			return nil, fmt.Errorf("%s: %w", common.ErrInitLogger, logErr)
		}

		ctx = logger.NewContext(ctx, log)
	}

	log.Info(ctx, common.MsgConnectingToPostgres,
		zap.String("host", cfg.Host),
		zap.Int("port", cfg.Port),
		zap.String("database", cfg.Database),
		zap.Int("maxConnections", cfg.MaxConnections),
		zap.Int("minConnections", cfg.MinConnections))

	poolCfg, err := pgxpool.ParseConfig(fmt.Sprintf(
		"host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		cfg.Host, cfg.Port, cfg.Username, cfg.Password, cfg.Database, cfg.SSLMode,
	))
	if err != nil {
		log.Error(ctx, common.ErrParsePoolConfig, zap.Error(err))

		return nil, fmt.Errorf("%s: %w", common.ErrParsePoolConfig, err)
	}

	poolCfg.MaxConns = int32(cfg.MaxConnections)
	poolCfg.MinConns = int32(cfg.MinConnections)

	pool, err := pgxpool.NewWithConfig(ctx, poolCfg)
	if err != nil {
		log.Error(ctx, common.ErrCreateConnectionPool, zap.Error(err))

		return nil, fmt.Errorf("%s: %w", common.ErrCreateConnectionPool, err)
	}

	pgxAdapter := NewPgxPoolAdapter(pool)

	if err := pgxAdapter.Ping(ctx); err != nil {
		log.Error(ctx, common.ErrPingPostgresPool, zap.Error(err))

		return nil, fmt.Errorf("%s: %w", common.ErrPingPostgresPool, err)
	}

	log.Info(ctx, common.MsgPostgresConnected)

	if err := migrate2.Migrate(ctx, cfg); err != nil && !errors.Is(err, migrate.ErrNoChange) {
		log.Error(ctx, common.ErrApplyDBMigrations, zap.Error(err))
		pgxAdapter.Close()

		return nil, fmt.Errorf("%s: %w", common.ErrApplyDBMigrations, err)
	}

	log.Info(ctx, common.MsgDBMigrationsApplied)

	return &DB{pool: pgxAdapter}, nil
}

func (db *DB) Close(ctx context.Context) error {
	log, err := logger.FromContext(ctx)
	if err != nil {
		log, logErr := logger.NewLogger()
		if logErr != nil {
			return fmt.Errorf("%s: %w", common.ErrInitLogger, logErr)
		}

		ctx = logger.NewContext(ctx, log)
	}

	log.Info(ctx, common.MsgClosingPostgresPool)
	db.pool.Close()

	return nil
}

func (db *DB) GetPool() Pool {
	return db.pool
}

func (db *DB) Ping(ctx context.Context) error {
	log, err := logger.FromContext(ctx)
	if err != nil {
		log, logErr := logger.NewLogger()
		if logErr != nil {
			return fmt.Errorf("%s: %w", common.ErrInitLogger, logErr)
		}

		ctx = logger.NewContext(ctx, log)
	}

	if err := db.pool.Ping(ctx); err != nil {
		log.Error(ctx, common.ErrPingPostgresPool, zap.Error(err))

		return fmt.Errorf("%s: %w", common.ErrPingPostgresPool, err)
	}

	return nil
}
