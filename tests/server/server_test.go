package server_test

import (
	"context"
	"fmt"
	"net/http"
	"testing"
	"time"

	"github.com/flexer2006/case-back-restaurant-go/configs"
	"github.com/flexer2006/case-back-restaurant-go/internal/domain"
	"github.com/flexer2006/case-back-restaurant-go/internal/logger"
	"github.com/flexer2006/case-back-restaurant-go/internal/logger/ports"
	"github.com/flexer2006/case-back-restaurant-go/internal/server"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

type MockRestaurantUseCase struct {
	mock.Mock
}

type MockBookingUseCase struct {
	mock.Mock
}

type MockUserUseCase struct {
	mock.Mock
}

type MockFactsUseCase struct {
	mock.Mock
}

type MockAvailabilityUseCase struct {
	mock.Mock
}

type MockNotificationUseCase struct {
	mock.Mock
}

type MockLogger struct {
	mock.Mock
}

func (m *MockLogger) Info(ctx context.Context, msg string, _ ...zap.Field) {
	m.Called(ctx, msg)
}

func (m *MockLogger) Warn(ctx context.Context, msg string, _ ...zap.Field) {
	m.Called(ctx, msg)
}

func (m *MockLogger) Error(ctx context.Context, msg string, _ ...zap.Field) {
	m.Called(ctx, msg)
}

func (m *MockLogger) Debug(ctx context.Context, msg string, _ ...zap.Field) {
	m.Called(ctx, msg)
}

func (m *MockLogger) Fatal(ctx context.Context, msg string, _ ...zap.Field) {
	m.Called(ctx, msg)
}

func (m *MockLogger) SetLevel(level ports.LogLevel) {
	m.Called(level)
}

func (m *MockLogger) GetLevel() ports.LogLevel {
	args := m.Called()
	return args.Get(0).(ports.LogLevel)
}

func (m *MockLogger) With(fields ...zap.Field) ports.LoggerPort {
	args := m.Called()
	return args.Get(0).(ports.LoggerPort)
}

func (m *MockLogger) Sync() error {
	args := m.Called()
	return args.Error(0)
}

func createTestConfig() *configs.Config {
	return &configs.Config{
		Server: configs.ServerConfig{
			Host: "127.0.0.1",
			Port: 8080,
		},
		Shutdown: configs.ShutdownConfig{
			Timeout: 5 * time.Second,
		},
	}
}

func TestNewServer(t *testing.T) {
	ctx := context.Background()
	config := createTestConfig()

	restaurantUseCase := new(MockRestaurantUseCase)
	bookingUseCase := new(MockBookingUseCase)
	userUseCase := new(MockUserUseCase)
	factsUseCase := new(MockFactsUseCase)
	availabilityUseCase := new(MockAvailabilityUseCase)
	notificationUseCase := new(MockNotificationUseCase)

	s, err := server.NewServer(
		ctx,
		config,
		restaurantUseCase,
		bookingUseCase,
		userUseCase,
		factsUseCase,
		availabilityUseCase,
		notificationUseCase,
	)

	require.NoError(t, err)
	assert.NotNil(t, s)
}

func TestRegisterRoutes(t *testing.T) {
	ctx := context.Background()
	config := createTestConfig()

	restaurantUseCase := new(MockRestaurantUseCase)
	bookingUseCase := new(MockBookingUseCase)
	userUseCase := new(MockUserUseCase)
	factsUseCase := new(MockFactsUseCase)
	availabilityUseCase := new(MockAvailabilityUseCase)
	notificationUseCase := new(MockNotificationUseCase)

	s, err := server.NewServer(
		ctx,
		config,
		restaurantUseCase,
		bookingUseCase,
		userUseCase,
		factsUseCase,
		availabilityUseCase,
		notificationUseCase,
	)
	require.NoError(t, err)

	assert.NotPanics(t, func() {
		s.RegisterRoutes()
	})
}

func TestStartAndStopServer(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping server test in short mode")
	}

	ctx := context.Background()
	config := createTestConfig()
	config.Server.Port = 9090 + int(time.Now().Unix()%100)

	config.Shutdown.Timeout = 500 * time.Millisecond

	restaurantUseCase := new(MockRestaurantUseCase)
	bookingUseCase := new(MockBookingUseCase)
	userUseCase := new(MockUserUseCase)
	factsUseCase := new(MockFactsUseCase)
	availabilityUseCase := new(MockAvailabilityUseCase)
	notificationUseCase := new(MockNotificationUseCase)

	mockLogger := new(MockLogger)
	mockLogger.On("Info", mock.Anything, mock.Anything).Return()
	mockLogger.On("Error", mock.Anything, mock.Anything).Return().Maybe()
	mockLogger.On("With", mock.Anything).Return(mockLogger).Maybe()
	mockLogger.On("Sync").Return(nil).Maybe()

	ctx = logger.NewContext(ctx, mockLogger)

	s, err := server.NewServer(
		ctx,
		config,
		restaurantUseCase,
		bookingUseCase,
		userUseCase,
		factsUseCase,
		availabilityUseCase,
		notificationUseCase,
	)
	require.NoError(t, err)

	s.RegisterRoutes()

	go func() {
		t.Logf("starting server on port %d", config.Server.Port)
		err := s.Start(ctx)
		t.Logf("server shutdown with result: %v", err)
	}()

	time.Sleep(1 * time.Second)

	t.Logf("checking server availability")
	initialResp, err := http.Get(fmt.Sprintf("http://127.0.0.1:%d/", config.Server.Port))
	if err == nil {
		if closeErr := initialResp.Body.Close(); closeErr != nil {
			t.Logf("ошибка при закрытии тела ответа: %v", closeErr)
		}
		t.Logf("server is available, response status: %d", initialResp.StatusCode)
	} else {
		t.Logf("error checking server availability: %v", err)
		t.Fatal("server failed to start")
	}

	t.Logf("stopping server via stop method")

	stopCtx, stopCancel := context.WithTimeout(ctx, 2*time.Second)
	defer stopCancel()

	err = s.Stop(stopCtx)
	require.NoError(t, err, "error stopping server")

	time.Sleep(2 * time.Second)

	t.Logf("verifying server shutdown")
	client := &http.Client{
		Timeout: 100 * time.Millisecond,
	}
	_, err = client.Get(fmt.Sprintf("http://127.0.0.1:%d/", config.Server.Port))
	assert.Error(t, err, "server should be unavailable after shutdown")
	t.Logf("server is unavailable (as expected): %v", err)

	mockLogger.AssertCalled(t, "Info", mock.Anything, mock.Anything)
}

func TestServerWithConfig(t *testing.T) {
	ctx := context.Background()
	config := createTestConfig()

	restaurantUseCase := new(MockRestaurantUseCase)
	bookingUseCase := new(MockBookingUseCase)
	userUseCase := new(MockUserUseCase)
	factsUseCase := new(MockFactsUseCase)
	availabilityUseCase := new(MockAvailabilityUseCase)
	notificationUseCase := new(MockNotificationUseCase)

	s1, err := server.NewServer(
		ctx,
		config,
		restaurantUseCase,
		bookingUseCase,
		userUseCase,
		factsUseCase,
		availabilityUseCase,
		notificationUseCase,
	)
	require.NoError(t, err)
	assert.NotNil(t, s1)

	config.Server.Port = 9999
	s2, err := server.NewServer(
		ctx,
		config,
		restaurantUseCase,
		bookingUseCase,
		userUseCase,
		factsUseCase,
		availabilityUseCase,
		notificationUseCase,
	)
	require.NoError(t, err)
	assert.NotNil(t, s2)
}

func TestStopServerWithContext(t *testing.T) {
	ctx := context.Background()
	config := createTestConfig()

	restaurantUseCase := new(MockRestaurantUseCase)
	bookingUseCase := new(MockBookingUseCase)
	userUseCase := new(MockUserUseCase)
	factsUseCase := new(MockFactsUseCase)
	availabilityUseCase := new(MockAvailabilityUseCase)
	notificationUseCase := new(MockNotificationUseCase)

	mockLogger := new(MockLogger)
	mockLogger.On("Info", mock.Anything, mock.Anything).Return()
	mockLogger.On("With").Return(mockLogger).Maybe()

	ctx = logger.NewContext(ctx, mockLogger)

	s, err := server.NewServer(
		ctx,
		config,
		restaurantUseCase,
		bookingUseCase,
		userUseCase,
		factsUseCase,
		availabilityUseCase,
		notificationUseCase,
	)
	require.NoError(t, err)

	err = s.Stop(ctx)
	assert.NoError(t, err)

	timeoutCtx, cancel := context.WithTimeout(ctx, 100*time.Millisecond)
	defer cancel()
	_ = s.Stop(timeoutCtx)
}

func (m *MockRestaurantUseCase) GetRestaurant(ctx context.Context, id string) (*domain.Restaurant, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Restaurant), args.Error(1)
}

func (m *MockRestaurantUseCase) ListRestaurants(ctx context.Context, offset, limit int) ([]*domain.Restaurant, error) {
	args := m.Called(ctx, offset, limit)
	return args.Get(0).([]*domain.Restaurant), args.Error(1)
}

func (m *MockRestaurantUseCase) CreateRestaurant(ctx context.Context, restaurant *domain.Restaurant) (string, error) {
	args := m.Called(ctx, restaurant)
	return args.String(0), args.Error(1)
}

func (m *MockRestaurantUseCase) UpdateRestaurant(ctx context.Context, restaurant *domain.Restaurant) error {
	args := m.Called(ctx, restaurant)
	return args.Error(0)
}

func (m *MockRestaurantUseCase) DeleteRestaurant(ctx context.Context, id string) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockRestaurantUseCase) AddFact(ctx context.Context, restaurantID string, content string) (*domain.Fact, error) {
	args := m.Called(ctx, restaurantID, content)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Fact), args.Error(1)
}

func (m *MockRestaurantUseCase) GetFacts(ctx context.Context, restaurantID string) ([]domain.Fact, error) {
	args := m.Called(ctx, restaurantID)
	return args.Get(0).([]domain.Fact), args.Error(1)
}

func (m *MockRestaurantUseCase) GetRandomFacts(ctx context.Context, count int) ([]domain.Fact, error) {
	args := m.Called(ctx, count)
	return args.Get(0).([]domain.Fact), args.Error(1)
}

func (m *MockRestaurantUseCase) SetWorkingHours(ctx context.Context, restaurantID string, workingHours *domain.WorkingHours) error {
	args := m.Called(ctx, restaurantID, workingHours)
	return args.Error(0)
}

func (m *MockRestaurantUseCase) GetWorkingHours(ctx context.Context, restaurantID string) ([]*domain.WorkingHours, error) {
	args := m.Called(ctx, restaurantID)
	return args.Get(0).([]*domain.WorkingHours), args.Error(1)
}

func (m *MockBookingUseCase) GetBooking(ctx context.Context, id string) (*domain.Booking, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Booking), args.Error(1)
}

func (m *MockBookingUseCase) GetRestaurantBookings(ctx context.Context, restaurantID string) ([]*domain.Booking, error) {
	args := m.Called(ctx, restaurantID)
	return args.Get(0).([]*domain.Booking), args.Error(1)
}

func (m *MockBookingUseCase) GetUserBookings(ctx context.Context, userID string) ([]*domain.Booking, error) {
	args := m.Called(ctx, userID)
	return args.Get(0).([]*domain.Booking), args.Error(1)
}

func (m *MockBookingUseCase) CreateBooking(ctx context.Context, booking *domain.Booking) (string, error) {
	args := m.Called(ctx, booking)
	return args.String(0), args.Error(1)
}

func (m *MockBookingUseCase) ConfirmBooking(ctx context.Context, id string) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockBookingUseCase) RejectBooking(ctx context.Context, id string, reason string) error {
	args := m.Called(ctx, id, reason)
	return args.Error(0)
}

func (m *MockBookingUseCase) CancelBooking(ctx context.Context, id string) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockBookingUseCase) CompleteBooking(ctx context.Context, id string) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockBookingUseCase) SuggestAlternativeTime(ctx context.Context, bookingID string, date time.Time, time string, message string) (string, error) {
	args := m.Called(ctx, bookingID, date, time, message)
	return args.String(0), args.Error(1)
}

func (m *MockBookingUseCase) AcceptAlternative(ctx context.Context, alternativeID string) error {
	args := m.Called(ctx, alternativeID)
	return args.Error(0)
}

func (m *MockBookingUseCase) RejectAlternative(ctx context.Context, alternativeID string) error {
	args := m.Called(ctx, alternativeID)
	return args.Error(0)
}

func (m *MockUserUseCase) GetUser(ctx context.Context, id string) (*domain.User, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.User), args.Error(1)
}

func (m *MockUserUseCase) GetUserByEmail(ctx context.Context, email string) (*domain.User, error) {
	args := m.Called(ctx, email)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.User), args.Error(1)
}

func (m *MockUserUseCase) CreateUser(ctx context.Context, user *domain.User) (string, error) {
	args := m.Called(ctx, user)
	return args.String(0), args.Error(1)
}

func (m *MockUserUseCase) UpdateUser(ctx context.Context, user *domain.User) error {
	args := m.Called(ctx, user)
	return args.Error(0)
}

func (m *MockFactsUseCase) GetRandomFacts(ctx context.Context, count int) ([]domain.Fact, error) {
	args := m.Called(ctx, count)
	return args.Get(0).([]domain.Fact), args.Error(1)
}

func (m *MockFactsUseCase) GetRestaurantFacts(ctx context.Context, restaurantID string) ([]domain.Fact, error) {
	args := m.Called(ctx, restaurantID)
	return args.Get(0).([]domain.Fact), args.Error(1)
}

func (m *MockAvailabilityUseCase) GetAvailability(ctx context.Context, restaurantID string, date time.Time) ([]*domain.Availability, error) {
	args := m.Called(ctx, restaurantID, date)
	return args.Get(0).([]*domain.Availability), args.Error(1)
}

func (m *MockAvailabilityUseCase) SetAvailability(ctx context.Context, availability *domain.Availability) error {
	args := m.Called(ctx, availability)
	return args.Error(0)
}

func (m *MockAvailabilityUseCase) UpdateReservedSeats(ctx context.Context, availabilityID string, delta int) error {
	args := m.Called(ctx, availabilityID, delta)
	return args.Error(0)
}

func (m *MockAvailabilityUseCase) CheckAvailability(ctx context.Context, restaurantID string, date time.Time, timeSlot string, guestsCount int) (bool, error) {
	args := m.Called(ctx, restaurantID, date, timeSlot, guestsCount)
	return args.Bool(0), args.Error(1)
}

func (m *MockNotificationUseCase) NotifyRestaurant(ctx context.Context, restaurantID string, notificationType domain.NotificationType, title, message, relatedID string) error {
	args := m.Called(ctx, restaurantID, notificationType, title, message, relatedID)
	return args.Error(0)
}

func (m *MockNotificationUseCase) NotifyUser(ctx context.Context, userID string, notificationType domain.NotificationType, title, message, relatedID string) error {
	args := m.Called(ctx, userID, notificationType, title, message, relatedID)
	return args.Error(0)
}

func (m *MockNotificationUseCase) GetUserNotifications(ctx context.Context, userID string) ([]domain.Notification, error) {
	args := m.Called(ctx, userID)
	return args.Get(0).([]domain.Notification), args.Error(1)
}

func (m *MockNotificationUseCase) MarkAsRead(ctx context.Context, notificationID string) error {
	args := m.Called(ctx, notificationID)
	return args.Error(0)
}
