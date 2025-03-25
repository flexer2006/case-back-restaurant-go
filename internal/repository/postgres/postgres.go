// Package postgres provides an implementation of repositories for working with PostgreSQL database,
// including adapters for working with various entities of the restaurant reservation system.
package postgres

import (
	"context"
	"errors"
	"fmt"

	"github.com/flexer2006/case-back-restaurant-go/common"
	"github.com/flexer2006/case-back-restaurant-go/db/postgres"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

type DBExecutor interface {
	Exec(ctx context.Context, sql string, arguments ...interface{}) (pgconn.CommandTag, error)
	Query(ctx context.Context, sql string, args ...interface{}) (pgx.Rows, error)
	QueryRow(ctx context.Context, sql string, args ...interface{}) pgx.Row
}

type Repository struct {
	pool postgres.Pool
}

func NewRepository(pool postgres.Pool) *Repository {
	return &Repository{
		pool: pool,
	}
}

func (r *Repository) GetExecutor(ctx context.Context) (DBExecutor, func(), error) {
	conn, err := r.pool.Acquire(ctx)
	if err != nil {
		return nil, nil, fmt.Errorf("%s: %w", common.ErrAcquireConnection, err)
	}

	release := func() {
		conn.Release()
	}

	pgxConnWrapper, ok := conn.(*postgres.PgxConnWrapper)
	if !ok {
		pgxConnAdapter, ok := conn.(*PgxConnAdapter)
		if !ok {
			return nil, release, fmt.Errorf(common.ErrUnknownConnectionType)
		}
		return pgxConnAdapter.conn, release, nil
	}

	return pgxConnWrapper.Conn, release, nil
}

func (r *Repository) GetPool() postgres.Pool {
	return r.pool
}

func (r *Repository) WithTransaction(ctx context.Context, fn func(tx pgx.Tx) error) error {
	_, release, err := r.GetExecutor(ctx)
	if err != nil {
		return fmt.Errorf("%s: %w", common.ErrGetQueryExecutor, err)
	}
	defer release()

	conn, err := r.pool.Acquire(ctx)
	if err != nil {
		return fmt.Errorf("%s: %w", common.ErrAcquireConnection, err)
	}
	defer conn.Release()

	pgxConn, ok := conn.(*postgres.PgxConnWrapper)
	if !ok {
		return fmt.Errorf("%s: %w", common.ErrUnknownConnectionType, errors.New("expected PgxConnWrapper"))
	}

	tx, err := pgxConn.Conn.Begin(ctx)
	if err != nil {
		return fmt.Errorf("%s: %w", common.ErrBeginTransaction, err)
	}

	if err := fn(tx); err != nil {
		rbErr := tx.Rollback(ctx)
		if rbErr != nil {
			return fmt.Errorf("%s: %v, original error: %w", common.ErrRollbackTransaction, rbErr, err)
		}
		return err
	}

	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("%s: %w", common.ErrCommitTransaction, err)
	}

	return nil
}
