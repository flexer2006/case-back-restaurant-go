package postgres

import (
	"context"
	"fmt"

	"github.com/flexer2006/case-back-restaurant-go/common"
	"github.com/jackc/pgx/v5/pgxpool"
)

type PgxConnWrapper struct {
	Conn *pgxpool.Conn
}

func (c *PgxConnWrapper) Release() {
	c.Conn.Release()
}

func (c *PgxConnWrapper) GetConnection() *pgxpool.Conn {
	return c.Conn
}

type PgxPoolAdapter struct {
	pool *pgxpool.Pool
}

func NewPgxPoolAdapter(pool *pgxpool.Pool) *PgxPoolAdapter {
	return &PgxPoolAdapter{pool: pool}
}

func (a *PgxPoolAdapter) Close() {
	a.pool.Close()
}

func (a *PgxPoolAdapter) Ping(ctx context.Context) error {
	err := a.pool.Ping(ctx)
	if err != nil {
		return fmt.Errorf("%s: %w", common.ErrPingPostgresPool, err)
	}

	return nil
}

func (a *PgxPoolAdapter) Acquire(ctx context.Context) (Conn, error) {
	conn, err := a.pool.Acquire(ctx)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", common.ErrAcquireConnection, err)
	}

	return &PgxConnWrapper{Conn: conn}, nil
}

func (a *PgxPoolAdapter) GetInternalPool() *pgxpool.Pool {
	return a.pool
}
