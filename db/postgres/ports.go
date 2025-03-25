package postgres

import "context"

type Database interface {
	Close(ctx context.Context) error
	Ping(ctx context.Context) error
	GetPool() Pool
}

type Pool interface {
	Close()
	Ping(ctx context.Context) error
	Acquire(ctx context.Context) (Conn, error)
}

type Conn interface {
	Release()
}
