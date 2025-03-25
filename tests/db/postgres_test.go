package db_test

import (
	"context"
	"errors"
	"testing"

	"github.com/flexer2006/case-back-restaurant-go/db/postgres"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type MockPool struct {
	mock.Mock
}

func (m *MockPool) Close() {
	m.Called()
}

func (m *MockPool) Ping(ctx context.Context) error {
	args := m.Called(ctx)
	return args.Error(0)
}

func (m *MockPool) Acquire(ctx context.Context) (postgres.Conn, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(postgres.Conn), args.Error(1)
}

type MockConn struct {
	mock.Mock
}

func (m *MockConn) Release() {
	m.Called()
}

type testDatabase struct {
	pool *MockPool
}

func (db *testDatabase) Close(_ context.Context) error {
	db.pool.Close()
	return nil
}

func (db *testDatabase) Ping(ctx context.Context) error {
	return db.pool.Ping(ctx)
}

func (db *testDatabase) GetPool() postgres.Pool {
	return db.pool
}

func TestDatabase_Close(t *testing.T) {
	mockPool := new(MockPool)
	mockPool.On("Close").Return()

	db := &testDatabase{pool: mockPool}

	ctx := context.Background()
	err := db.Close(ctx)

	assert.NoError(t, err)
	mockPool.AssertCalled(t, "Close")
}

func TestDatabase_Ping(t *testing.T) {
	t.Run("успешный пинг", func(t *testing.T) {
		mockPool := new(MockPool)
		mockPool.On("Ping", mock.Anything).Return(nil)

		db := &testDatabase{pool: mockPool}

		ctx := context.Background()
		err := db.Ping(ctx)

		assert.NoError(t, err)
		mockPool.AssertExpectations(t)
	})

	t.Run("ошибка пинга", func(t *testing.T) {
		mockPool := new(MockPool)
		mockPool.On("Ping", mock.Anything).Return(errors.New("ping error"))

		db := &testDatabase{pool: mockPool}

		ctx := context.Background()
		err := db.Ping(ctx)

		assert.Error(t, err)
		assert.Equal(t, "ping error", err.Error())
		mockPool.AssertExpectations(t)
	})
}

func TestDatabase_GetPool(t *testing.T) {
	mockPool := new(MockPool)

	db := &testDatabase{pool: mockPool}

	pool := db.GetPool()

	assert.Equal(t, mockPool, pool)
}
