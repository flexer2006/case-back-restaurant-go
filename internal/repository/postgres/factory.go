package postgres

import (
	"context"
	"fmt"

	"github.com/flexer2006/case-back-restaurant-go/common"
	"github.com/flexer2006/case-back-restaurant-go/db/postgres"

	"github.com/jackc/pgx/v5/pgxpool"
)

type PgxConnAdapter struct {
	conn *pgxpool.Conn
}

func (a *PgxConnAdapter) Release() {
	a.conn.Release()
}

type PoolAdapter struct {
	pool *pgxpool.Pool
}

func (a *PoolAdapter) Acquire(ctx context.Context) (postgres.Conn, error) {
	conn, err := a.pool.Acquire(ctx)
	if err != nil {
		return nil, err
	}
	return &postgres.PgxConnWrapper{Conn: conn}, nil
}

func (a *PoolAdapter) Close() {
	a.pool.Close()
}

func (a *PoolAdapter) Ping(ctx context.Context) error {
	err := a.pool.Ping(ctx)
	if err != nil {
		return fmt.Errorf("%s: %w", common.ErrPingPostgresPool, err)
	}
	return nil
}

type RepositoryFactory struct {
	db postgres.Database
}

func NewRepositoryFactory(db postgres.Database) *RepositoryFactory {
	return &RepositoryFactory{
		db: db,
	}
}

func (f *RepositoryFactory) Restaurant() *RestaurantRepository {
	return NewRestaurantRepository(NewRepository(f.db.GetPool()))
}

func (f *RepositoryFactory) WorkingHours() *WorkingHoursRepository {
	return NewWorkingHoursRepository(NewRepository(f.db.GetPool()))
}

func (f *RepositoryFactory) Availability() *AvailabilityRepository {
	return NewAvailabilityRepository(NewRepository(f.db.GetPool()))
}

func (f *RepositoryFactory) Booking() *BookingRepository {
	return NewBookingRepository(NewRepository(f.db.GetPool()))
}

func (f *RepositoryFactory) User() *UserRepository {
	return NewUserRepository(NewRepository(f.db.GetPool()))
}

func (f *RepositoryFactory) Notification() *NotificationRepository {
	return NewNotificationRepository(NewRepository(f.db.GetPool()))
}

type PostgresFactory struct {
	pool *pgxpool.Pool
}

func NewPostgresFactory(pool *pgxpool.Pool) *PostgresFactory {
	return &PostgresFactory{pool: pool}
}

func (a *PostgresFactory) GetConnection(ctx context.Context) (postgres.Conn, func(), error) {
	conn, err := a.pool.Acquire(ctx)
	if err != nil {
		return nil, nil, fmt.Errorf("%s: %w", common.ErrGetQueryExecutor, err)
	}

	return &postgres.PgxConnWrapper{Conn: conn}, func() { conn.Release() }, nil
}

func (a *PostgresFactory) Ping(ctx context.Context) error {
	err := a.pool.Ping(ctx)
	if err != nil {
		return fmt.Errorf("%s: %w", common.ErrPingPostgresPool, err)
	}
	return nil
}
