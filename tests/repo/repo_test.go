package repo_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/flexer2006/case-back-restaurant-go/common"
	"github.com/flexer2006/case-back-restaurant-go/internal/domain"
	"github.com/flexer2006/case-back-restaurant-go/internal/repository/postgres"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type MockCommandTag struct {
	mock.Mock
}

func (m *MockCommandTag) RowsAffected() int64 {
	args := m.Called()
	return args.Get(0).(int64)
}

func (m *MockCommandTag) String() string {
	args := m.Called()
	return args.String(0)
}

func (m *MockCommandTag) Insert() bool {
	args := m.Called()
	return args.Bool(0)
}

func (m *MockCommandTag) Update() bool {
	args := m.Called()
	return args.Bool(0)
}

func (m *MockCommandTag) Delete() bool {
	args := m.Called()
	return args.Bool(0)
}

func (m *MockCommandTag) Select() bool {
	args := m.Called()
	return args.Bool(0)
}

type MockDBExecutor struct {
	mock.Mock
}

func (m *MockDBExecutor) Exec(ctx context.Context, sql string, arguments ...interface{}) (pgconn.CommandTag, error) {
	args := m.Called(ctx, sql, arguments)
	if args.Get(0) == nil {
		return pgconn.CommandTag{}, args.Error(1)
	}
	return args.Get(0).(pgconn.CommandTag), args.Error(1)
}

func (m *MockDBExecutor) Query(ctx context.Context, sql string, args ...interface{}) (pgx.Rows, error) {
	mockArgs := m.Called(ctx, sql, args)
	if mockArgs.Get(0) == nil {
		return nil, mockArgs.Error(1)
	}
	return mockArgs.Get(0).(pgx.Rows), mockArgs.Error(1)
}

func (m *MockDBExecutor) QueryRow(ctx context.Context, sql string, args ...interface{}) pgx.Row {
	mockArgs := m.Called(ctx, sql, args)
	return mockArgs.Get(0).(pgx.Row)
}

type MockTx struct {
	MockDBExecutor
}

func (m *MockTx) Begin(ctx context.Context) (pgx.Tx, error) {
	args := m.Called(ctx)
	return args.Get(0).(pgx.Tx), args.Error(1)
}

func (m *MockTx) Commit(ctx context.Context) error {
	args := m.Called(ctx)
	return args.Error(0)
}

func (m *MockTx) Rollback(ctx context.Context) error {
	args := m.Called(ctx)
	return args.Error(0)
}

func (m *MockTx) CopyFrom(ctx context.Context, tableName pgx.Identifier, columnNames []string, rowSrc pgx.CopyFromSource) (int64, error) {
	args := m.Called(ctx, tableName, columnNames, rowSrc)
	return args.Get(0).(int64), args.Error(1)
}

func (m *MockTx) SendBatch(ctx context.Context, b *pgx.Batch) pgx.BatchResults {
	args := m.Called(ctx, b)
	return args.Get(0).(pgx.BatchResults)
}

func (m *MockTx) LargeObjects() pgx.LargeObjects {
	args := m.Called()
	return args.Get(0).(pgx.LargeObjects)
}

func (m *MockTx) Prepare(ctx context.Context, name, sql string) (*pgconn.StatementDescription, error) {
	args := m.Called(ctx, name, sql)
	return args.Get(0).(*pgconn.StatementDescription), args.Error(1)
}

func (m *MockTx) Conn() *pgx.Conn {
	args := m.Called()
	return args.Get(0).(*pgx.Conn)
}

type MockRow struct {
	mock.Mock
}

func (m *MockRow) Scan(dest ...interface{}) error {
	args := m.Called(dest)
	return args.Error(0)
}

type MockRows struct {
	mock.Mock
}

func (m *MockRows) Close() {
	m.Called()
}

func (m *MockRows) Err() error {
	args := m.Called()
	return args.Error(0)
}

func (m *MockRows) CommandTag() pgconn.CommandTag {
	args := m.Called()
	return args.Get(0).(pgconn.CommandTag)
}

func (m *MockRows) FieldDescriptions() []pgconn.FieldDescription {
	args := m.Called()
	return args.Get(0).([]pgconn.FieldDescription)
}

func (m *MockRows) Next() bool {
	args := m.Called()
	return args.Bool(0)
}

func (m *MockRows) Scan(dest ...interface{}) error {
	args := m.Called(dest)
	return args.Error(0)
}

func (m *MockRows) Values() ([]interface{}, error) {
	args := m.Called()
	return args.Get(0).([]interface{}), args.Error(1)
}

func (m *MockRows) RawValues() [][]byte {
	args := m.Called()
	return args.Get(0).([][]byte)
}

type MockRepository struct {
	mock.Mock
}

func (m *MockRepository) GetExecutor(ctx context.Context) (postgres.DBExecutor, func(), error) {
	args := m.Called(ctx)
	return args.Get(0).(postgres.DBExecutor), args.Get(1).(func()), args.Error(2)
}

func (m *MockRepository) WithTransaction(ctx context.Context, fn func(tx pgx.Tx) error) error {
	args := m.Called(ctx, fn)
	return args.Error(0)
}

type MockRestaurantRepository struct {
	mock.Mock
}

func (m *MockRestaurantRepository) GetByID(ctx context.Context, id string) (*domain.Restaurant, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Restaurant), args.Error(1)
}

type MockBookingRepository struct {
	mock.Mock
}

func (m *MockBookingRepository) Create(ctx context.Context, booking *domain.Booking) error {
	args := m.Called(ctx, booking)
	return args.Error(0)
}

func (m *MockBookingRepository) UpdateStatus(ctx context.Context, id string, status domain.BookingStatus) error {
	args := m.Called(ctx, id, status)
	return args.Error(0)
}

type MockAvailabilityRepository struct {
	mock.Mock
}

func (m *MockAvailabilityRepository) SetAvailability(ctx context.Context, availability *domain.Availability) error {
	args := m.Called(ctx, availability)
	return args.Error(0)
}

type MockWorkingHoursRepository struct {
	mock.Mock
}

func (m *MockWorkingHoursRepository) SetWorkingHours(ctx context.Context, hours *domain.WorkingHours) error {
	args := m.Called(ctx, hours)
	return args.Error(0)
}

type MockNotificationRepository struct {
	mock.Mock
}

func (m *MockNotificationRepository) Create(ctx context.Context, notification *domain.Notification) error {
	args := m.Called(ctx, notification)
	return args.Error(0)
}

type MockUserRepository struct {
	mock.Mock
}

func (m *MockUserRepository) GetByID(ctx context.Context, id string) (*domain.User, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.User), args.Error(1)
}

func TestRestaurantRepository_GetByID(t *testing.T) {
	restaurantID := uuid.New().String()
	now := time.Now()
	expectedRestaurant := &domain.Restaurant{
		ID:           restaurantID,
		Name:         "Test Restaurant",
		Address:      "Test Address",
		Cuisine:      "Italian",
		Description:  "Test Description",
		CreatedAt:    now,
		UpdatedAt:    now,
		ContactEmail: "test@example.com",
		ContactPhone: "+1234567890",
		Facts:        []domain.Fact{},
	}

	mockRepo := new(MockRestaurantRepository)
	mockRepo.On("GetByID", mock.Anything, restaurantID).Return(expectedRestaurant, nil)

	result, err := mockRepo.GetByID(context.Background(), restaurantID)

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, expectedRestaurant.ID, result.ID)
	assert.Equal(t, expectedRestaurant.Name, result.Name)
	mockRepo.AssertExpectations(t)
}

func TestBookingRepository_Create(t *testing.T) {
	bookingID := uuid.New().String()
	restaurantID := uuid.New().String()
	userID := uuid.New().String()
	now := time.Now()
	date := time.Now().AddDate(0, 0, 1)

	booking := &domain.Booking{
		ID:           bookingID,
		RestaurantID: restaurantID,
		UserID:       userID,
		Date:         date,
		Time:         "18:00",
		Duration:     120,
		GuestsCount:  4,
		Status:       domain.BookingStatusPending,
		Comment:      "Test comment",
		CreatedAt:    now,
		UpdatedAt:    now,
	}

	mockRepo := new(MockBookingRepository)
	mockRepo.On("Create", mock.Anything, booking).Return(nil)

	err := mockRepo.Create(context.Background(), booking)

	assert.NoError(t, err)
	mockRepo.AssertExpectations(t)
}

func TestAvailabilityRepository_SetAvailability(t *testing.T) {
	availabilityID := uuid.New().String()
	restaurantID := uuid.New().String()
	now := time.Now()
	date := time.Now().AddDate(0, 0, 1)

	availability := &domain.Availability{
		ID:           availabilityID,
		RestaurantID: restaurantID,
		Date:         date,
		TimeSlot:     "18:00",
		Capacity:     20,
		Reserved:     0,
		UpdatedAt:    now,
	}

	mockRepo := new(MockAvailabilityRepository)
	mockRepo.On("SetAvailability", mock.Anything, availability).Return(nil)

	err := mockRepo.SetAvailability(context.Background(), availability)

	assert.NoError(t, err)
	mockRepo.AssertExpectations(t)
}

func TestWorkingHoursRepository_SetWorkingHours(t *testing.T) {
	workingHoursID := uuid.New().String()
	restaurantID := uuid.New().String()
	now := time.Now()

	workingHours := &domain.WorkingHours{
		ID:           workingHoursID,
		RestaurantID: restaurantID,
		WeekDay:      domain.Monday,
		OpenTime:     "09:00",
		CloseTime:    "22:00",
		IsClosed:     false,
		ValidFrom:    now,
	}

	mockRepo := new(MockWorkingHoursRepository)
	mockRepo.On("SetWorkingHours", mock.Anything, workingHours).Return(nil)

	err := mockRepo.SetWorkingHours(context.Background(), workingHours)

	assert.NoError(t, err)
	mockRepo.AssertExpectations(t)
}

func TestNotificationRepository_Create(t *testing.T) {
	notificationID := uuid.New().String()
	userID := uuid.New().String()
	now := time.Now()

	notification := &domain.Notification{
		ID:            notificationID,
		RecipientType: domain.RecipientTypeUser,
		RecipientID:   userID,
		Type:          domain.NotificationTypeNewBooking,
		Title:         "New Booking",
		Message:       "You have a new booking request",
		IsRead:        false,
		RelatedID:     uuid.New().String(),
		CreatedAt:     now,
	}

	mockRepo := new(MockNotificationRepository)
	mockRepo.On("Create", mock.Anything, notification).Return(nil)

	err := mockRepo.Create(context.Background(), notification)

	assert.NoError(t, err)
	mockRepo.AssertExpectations(t)
}

func TestUserRepository_GetByID(t *testing.T) {
	userID := uuid.New().String()
	now := time.Now()
	expectedUser := &domain.User{
		ID:        userID,
		Name:      "Test User",
		Email:     "test@example.com",
		Phone:     "+1234567890",
		CreatedAt: now,
		UpdatedAt: now,
	}

	mockRepo := new(MockUserRepository)
	mockRepo.On("GetByID", mock.Anything, userID).Return(expectedUser, nil)

	result, err := mockRepo.GetByID(context.Background(), userID)

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, expectedUser.ID, result.ID)
	assert.Equal(t, expectedUser.Name, result.Name)
	assert.Equal(t, expectedUser.Email, result.Email)
	assert.Equal(t, expectedUser.Phone, result.Phone)
	mockRepo.AssertExpectations(t)
}

func TestBookingRepository_UpdateStatus(t *testing.T) {
	bookingID := uuid.New().String()
	newStatus := domain.BookingStatusConfirmed

	mockRepo := new(MockBookingRepository)
	mockRepo.On("UpdateStatus", mock.Anything, bookingID, newStatus).Return(nil)

	err := mockRepo.UpdateStatus(context.Background(), bookingID, newStatus)

	assert.NoError(t, err)
	mockRepo.AssertExpectations(t)
}

func TestRestaurantRepository_GetByID_NotFound(t *testing.T) {
	restaurantID := uuid.New().String()

	mockRepo := new(MockRestaurantRepository)
	mockRepo.On("GetByID", mock.Anything, restaurantID).Return(nil, errors.New(common.ErrRestaurantNotFound))

	result, err := mockRepo.GetByID(context.Background(), restaurantID)

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Equal(t, common.ErrRestaurantNotFound, err.Error())
	mockRepo.AssertExpectations(t)
}
