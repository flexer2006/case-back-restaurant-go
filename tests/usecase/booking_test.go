package usecase_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/flexer2006/case-back-restaurant-go/internal/domain"
	"github.com/flexer2006/case-back-restaurant-go/internal/logger"
	"github.com/flexer2006/case-back-restaurant-go/internal/logger/ports"
	"github.com/flexer2006/case-back-restaurant-go/pkg/usecase"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"go.uber.org/zap"
)

type MockLogger struct {
	mock.Mock
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
	m.Called(level)
}

func (m *MockLogger) GetLevel() ports.LogLevel {
	args := m.Called()
	return args.Get(0).(ports.LogLevel)
}

func (m *MockLogger) With(fields ...zap.Field) ports.LoggerPort {
	args := m.Called(fields)
	return args.Get(0).(ports.LoggerPort)
}

func (m *MockLogger) Sync() error {
	args := m.Called()
	return args.Error(0)
}

func newTestContext() context.Context {
	mockLogger := new(MockLogger)
	mockLogger.On("Info", mock.Anything, mock.Anything, mock.Anything).Return()
	mockLogger.On("Error", mock.Anything, mock.Anything, mock.Anything).Return()
	mockLogger.On("Warn", mock.Anything, mock.Anything, mock.Anything).Return()
	mockLogger.On("Debug", mock.Anything, mock.Anything, mock.Anything).Return()
	mockLogger.On("Fatal", mock.Anything, mock.Anything, mock.Anything).Return()
	mockLogger.On("With", mock.Anything).Return(mockLogger)
	mockLogger.On("GetLevel").Return(ports.InfoLevel)
	mockLogger.On("SetLevel", mock.Anything).Return()
	mockLogger.On("Sync").Return(nil)

	return logger.NewContext(context.Background(), mockLogger)
}

type MockBookingRepository struct {
	mock.Mock
}

func (m *MockBookingRepository) GetByID(ctx context.Context, id string) (*domain.Booking, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Booking), args.Error(1)
}

func (m *MockBookingRepository) GetByRestaurantID(ctx context.Context, restaurantID string) ([]*domain.Booking, error) {
	args := m.Called(ctx, restaurantID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*domain.Booking), args.Error(1)
}

func (m *MockBookingRepository) GetByUserID(ctx context.Context, userID string) ([]*domain.Booking, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*domain.Booking), args.Error(1)
}

func (m *MockBookingRepository) Create(ctx context.Context, booking *domain.Booking) error {
	args := m.Called(ctx, booking)
	return args.Error(0)
}

func (m *MockBookingRepository) UpdateStatus(ctx context.Context, id string, status domain.BookingStatus) error {
	args := m.Called(ctx, id, status)
	return args.Error(0)
}

func (m *MockBookingRepository) AddAlternative(ctx context.Context, alternative *domain.BookingAlternative) error {
	args := m.Called(ctx, alternative)
	return args.Error(0)
}

func (m *MockBookingRepository) AcceptAlternative(ctx context.Context, alternativeID string) error {
	args := m.Called(ctx, alternativeID)
	return args.Error(0)
}

func (m *MockBookingRepository) RejectAlternative(ctx context.Context, alternativeID string) error {
	args := m.Called(ctx, alternativeID)
	return args.Error(0)
}

func (m *MockBookingRepository) GetAlternativeByID(ctx context.Context, alternativeID string) (*domain.BookingAlternative, error) {
	args := m.Called(ctx, alternativeID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.BookingAlternative), args.Error(1)
}

type MockAvailabilityRepository struct {
	mock.Mock
}

func (m *MockAvailabilityRepository) GetByRestaurantAndDate(ctx context.Context, restaurantID string, date time.Time) ([]*domain.Availability, error) {
	args := m.Called(ctx, restaurantID, date)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*domain.Availability), args.Error(1)
}

func (m *MockAvailabilityRepository) SetAvailability(ctx context.Context, availability *domain.Availability) error {
	args := m.Called(ctx, availability)
	return args.Error(0)
}

func (m *MockAvailabilityRepository) UpdateReservedSeats(ctx context.Context, availabilityID string, delta int) error {
	args := m.Called(ctx, availabilityID, delta)
	return args.Error(0)
}

type MockNotificationService struct {
	mock.Mock
}

func (m *MockNotificationService) NotifyRestaurant(ctx context.Context, restaurantID string, notificationType domain.NotificationType, title, message string, relatedID string) error {
	args := m.Called(ctx, restaurantID, notificationType, title, message, relatedID)
	return args.Error(0)
}

func (m *MockNotificationService) NotifyUser(ctx context.Context, userID string, notificationType domain.NotificationType, title, message string, relatedID string) error {
	args := m.Called(ctx, userID, notificationType, title, message, relatedID)
	return args.Error(0)
}

func (m *MockNotificationService) GetUserNotifications(ctx context.Context, userID string) ([]domain.Notification, error) {
	args := m.Called(ctx, userID)
	return args.Get(0).([]domain.Notification), args.Error(1)
}

func (m *MockNotificationService) MarkAsRead(ctx context.Context, notificationID string) error {
	args := m.Called(ctx, notificationID)
	return args.Error(0)
}

func TestGetBooking(t *testing.T) {
	bookingRepo := new(MockBookingRepository)
	availabilityRepo := new(MockAvailabilityRepository)
	notificationSvc := new(MockNotificationService)

	booking := &domain.Booking{
		ID:           "booking-123",
		RestaurantID: "restaurant-456",
		UserID:       "user-789",
		Date:         time.Now().Add(24 * time.Hour),
		Time:         "19:00",
		GuestsCount:  4,
		Status:       domain.BookingStatusPending,
	}

	bookingRepo.On("GetByID", mock.Anything, "booking-123").Return(booking, nil)
	bookingRepo.On("GetByID", mock.Anything, "non-existent").Return(nil, errors.New("booking not found"))

	uc := usecase.NewBookingUseCase(bookingRepo, availabilityRepo, notificationSvc)

	t.Run("successful booking retrieval", func(t *testing.T) {
		ctx := newTestContext()
		result, err := uc.GetBooking(ctx, "booking-123")

		assert.NoError(t, err)
		assert.Equal(t, booking, result)
	})

	t.Run("booking not found", func(t *testing.T) {
		ctx := newTestContext()
		result, err := uc.GetBooking(ctx, "non-existent")

		assert.Error(t, err)
		assert.Nil(t, result)
	})
}

func TestGetRestaurantBookings(t *testing.T) {
	bookingRepo := new(MockBookingRepository)
	availabilityRepo := new(MockAvailabilityRepository)
	notificationSvc := new(MockNotificationService)

	bookings := []*domain.Booking{
		{
			ID:           "booking-123",
			RestaurantID: "restaurant-456",
			UserID:       "user-789",
			Date:         time.Now().Add(24 * time.Hour),
			Time:         "19:00",
			GuestsCount:  4,
			Status:       domain.BookingStatusPending,
		},
		{
			ID:           "booking-124",
			RestaurantID: "restaurant-456",
			UserID:       "user-790",
			Date:         time.Now().Add(48 * time.Hour),
			Time:         "20:00",
			GuestsCount:  2,
			Status:       domain.BookingStatusConfirmed,
		},
	}

	bookingRepo.On("GetByRestaurantID", mock.Anything, "restaurant-456").Return(bookings, nil)
	bookingRepo.On("GetByRestaurantID", mock.Anything, "non-existent").Return(nil, errors.New("restaurant not found"))

	uc := usecase.NewBookingUseCase(bookingRepo, availabilityRepo, notificationSvc)

	t.Run("successful restaurant bookings retrieval", func(t *testing.T) {
		ctx := newTestContext()
		result, err := uc.GetRestaurantBookings(ctx, "restaurant-456")

		assert.NoError(t, err)
		assert.Equal(t, bookings, result)
	})

	t.Run("restaurant not found", func(t *testing.T) {
		ctx := newTestContext()
		result, err := uc.GetRestaurantBookings(ctx, "non-existent")

		assert.Error(t, err)
		assert.Nil(t, result)
	})
}

func TestGetUserBookings(t *testing.T) {
	bookingRepo := new(MockBookingRepository)
	availabilityRepo := new(MockAvailabilityRepository)
	notificationSvc := new(MockNotificationService)

	bookings := []*domain.Booking{
		{
			ID:           "booking-123",
			RestaurantID: "restaurant-456",
			UserID:       "user-789",
			Date:         time.Now().Add(24 * time.Hour),
			Time:         "19:00",
			GuestsCount:  4,
			Status:       domain.BookingStatusPending,
		},
		{
			ID:           "booking-124",
			RestaurantID: "restaurant-457",
			UserID:       "user-789",
			Date:         time.Now().Add(48 * time.Hour),
			Time:         "20:00",
			GuestsCount:  2,
			Status:       domain.BookingStatusConfirmed,
		},
	}

	bookingRepo.On("GetByUserID", mock.Anything, "user-789").Return(bookings, nil)
	bookingRepo.On("GetByUserID", mock.Anything, "non-existent").Return(nil, errors.New("user not found"))

	uc := usecase.NewBookingUseCase(bookingRepo, availabilityRepo, notificationSvc)

	t.Run("successful user bookings retrieval", func(t *testing.T) {
		ctx := newTestContext()
		result, err := uc.GetUserBookings(ctx, "user-789")

		assert.NoError(t, err)
		assert.Equal(t, bookings, result)
	})

	t.Run("user not found", func(t *testing.T) {
		ctx := newTestContext()
		result, err := uc.GetUserBookings(ctx, "non-existent")

		assert.Error(t, err)
		assert.Nil(t, result)
	})
}

func TestCreateBooking(t *testing.T) {
	bookingRepo := new(MockBookingRepository)
	availabilityRepo := new(MockAvailabilityRepository)
	notificationSvc := new(MockNotificationService)

	bookingDate := time.Now().Add(24 * time.Hour)

	booking := &domain.Booking{
		RestaurantID: "restaurant-456",
		UserID:       "user-789",
		Date:         bookingDate,
		Time:         "19:00",
		GuestsCount:  4,
		Status:       domain.BookingStatusPending,
	}

	availabilities := []*domain.Availability{
		{
			ID:           "avail-123",
			RestaurantID: "restaurant-456",
			Date:         bookingDate,
			TimeSlot:     "19:00",
			Capacity:     20,
			Reserved:     10,
		},
		{
			ID:           "avail-124",
			RestaurantID: "restaurant-456",
			Date:         bookingDate,
			TimeSlot:     "20:00",
			Capacity:     20,
			Reserved:     15,
		},
	}

	bookingRepo.On("Create", mock.Anything, mock.MatchedBy(func(b *domain.Booking) bool {
		b.ID = "booking-new-123"
		return b.RestaurantID == booking.RestaurantID && b.Time == booking.Time
	})).Return(nil)

	availabilityRepo.On("GetByRestaurantAndDate", mock.Anything, "restaurant-456", bookingDate).Return(availabilities, nil)
	availabilityRepo.On("UpdateReservedSeats", mock.Anything, "avail-123", 4).Return(nil)

	notificationSvc.On("NotifyRestaurant", mock.Anything, "restaurant-456", domain.NotificationTypeNewBooking, mock.Anything, mock.Anything, mock.Anything).Return(nil)

	uc := usecase.NewBookingUseCase(bookingRepo, availabilityRepo, notificationSvc)

	t.Run("successful booking creation", func(t *testing.T) {
		ctx := newTestContext()
		bookingID, err := uc.CreateBooking(ctx, booking)

		assert.NoError(t, err)
		assert.NotEmpty(t, bookingID)
	})

	t.Run("no availability for booking", func(t *testing.T) {
		ctx := newTestContext()
		noAvailBooking := &domain.Booking{
			RestaurantID: "restaurant-456",
			UserID:       "user-789",
			Date:         bookingDate,
			Time:         "21:00",
			GuestsCount:  4,
		}

		bookingID, err := uc.CreateBooking(ctx, noAvailBooking)

		assert.Error(t, err)
		assert.Equal(t, usecase.ErrNoAvailability, err)
		assert.Empty(t, bookingID)
	})

	t.Run("not enough seats", func(t *testing.T) {
		ctx := newTestContext()
		largeBooking := &domain.Booking{
			RestaurantID: "restaurant-456",
			UserID:       "user-789",
			Date:         bookingDate,
			Time:         "20:00",
			GuestsCount:  10,
		}

		bookingID, err := uc.CreateBooking(ctx, largeBooking)

		assert.Error(t, err)
		assert.Equal(t, usecase.ErrNoAvailability, err)
		assert.Empty(t, bookingID)
	})
}

func TestConfirmBooking(t *testing.T) {
	bookingRepo := new(MockBookingRepository)
	availabilityRepo := new(MockAvailabilityRepository)
	notificationSvc := new(MockNotificationService)

	pendingBooking := &domain.Booking{
		ID:           "booking-123",
		RestaurantID: "restaurant-456",
		UserID:       "user-789",
		Date:         time.Now().Add(24 * time.Hour),
		Time:         "19:00",
		GuestsCount:  4,
		Status:       domain.BookingStatusPending,
	}

	confirmedBooking := &domain.Booking{
		ID:           "booking-124",
		RestaurantID: "restaurant-456",
		UserID:       "user-789",
		Date:         time.Now().Add(24 * time.Hour),
		Time:         "20:00",
		GuestsCount:  2,
		Status:       domain.BookingStatusConfirmed,
	}

	bookingRepo.On("GetByID", mock.Anything, "booking-123").Return(pendingBooking, nil)
	bookingRepo.On("GetByID", mock.Anything, "booking-124").Return(confirmedBooking, nil)
	bookingRepo.On("GetByID", mock.Anything, "non-existent").Return(nil, errors.New("booking not found"))

	bookingRepo.On("UpdateStatus", mock.Anything, "booking-123", domain.BookingStatusConfirmed).Return(nil)

	notificationSvc.On("NotifyUser", mock.Anything, "user-789", domain.NotificationTypeBookingConfirmed, mock.Anything, mock.Anything, mock.Anything).Return(nil)

	uc := usecase.NewBookingUseCase(bookingRepo, availabilityRepo, notificationSvc)

	t.Run("successful booking confirmation", func(t *testing.T) {
		ctx := newTestContext()
		err := uc.ConfirmBooking(ctx, "booking-123")

		assert.NoError(t, err)
	})

	t.Run("booking already confirmed", func(t *testing.T) {
		ctx := newTestContext()
		err := uc.ConfirmBooking(ctx, "booking-124")

		assert.Error(t, err)
		assert.Equal(t, usecase.ErrInvalidBookingStatus, err)
	})

	t.Run("booking not found", func(t *testing.T) {
		ctx := newTestContext()
		err := uc.ConfirmBooking(ctx, "non-existent")

		assert.Error(t, err)
	})
}

func TestRejectBooking(t *testing.T) {
	bookingRepo := new(MockBookingRepository)
	availabilityRepo := new(MockAvailabilityRepository)
	notificationSvc := new(MockNotificationService)

	pendingBooking := &domain.Booking{
		ID:           "booking-123",
		RestaurantID: "restaurant-456",
		UserID:       "user-789",
		Date:         time.Now().Add(24 * time.Hour),
		Time:         "19:00",
		GuestsCount:  4,
		Status:       domain.BookingStatusPending,
	}

	confirmedBooking := &domain.Booking{
		ID:           "booking-124",
		RestaurantID: "restaurant-456",
		UserID:       "user-789",
		Date:         time.Now().Add(24 * time.Hour),
		Time:         "20:00",
		GuestsCount:  2,
		Status:       domain.BookingStatusConfirmed,
	}

	bookingRepo.On("GetByID", mock.Anything, "booking-123").Return(pendingBooking, nil)
	bookingRepo.On("GetByID", mock.Anything, "booking-124").Return(confirmedBooking, nil)
	bookingRepo.On("GetByID", mock.Anything, "non-existent").Return(nil, errors.New("booking not found"))

	bookingRepo.On("UpdateStatus", mock.Anything, "booking-123", domain.BookingStatusRejected).Return(nil)

	notificationSvc.On("NotifyUser", mock.Anything, "user-789", domain.NotificationTypeBookingRejected, mock.Anything, mock.Anything, mock.Anything).Return(nil)

	uc := usecase.NewBookingUseCase(bookingRepo, availabilityRepo, notificationSvc)

	t.Run("successful booking rejection", func(t *testing.T) {
		ctx := newTestContext()
		err := uc.RejectBooking(ctx, "booking-123", "no available tables")

		assert.NoError(t, err)
	})

	t.Run("booking already confirmed", func(t *testing.T) {
		ctx := newTestContext()
		err := uc.RejectBooking(ctx, "booking-124", "")

		assert.Error(t, err)
		assert.Equal(t, usecase.ErrInvalidBookingStatus, err)
	})

	t.Run("booking not found", func(t *testing.T) {
		ctx := newTestContext()
		err := uc.RejectBooking(ctx, "non-existent", "")

		assert.Error(t, err)
	})
}

func TestCancelBooking(t *testing.T) {
	bookingRepo := new(MockBookingRepository)
	availabilityRepo := new(MockAvailabilityRepository)
	notificationSvc := new(MockNotificationService)

	pendingBooking := &domain.Booking{
		ID:           "booking-123",
		RestaurantID: "restaurant-456",
		UserID:       "user-789",
		Date:         time.Now().Add(24 * time.Hour),
		Time:         "19:00",
		GuestsCount:  4,
		Status:       domain.BookingStatusPending,
	}

	completedBooking := &domain.Booking{
		ID:           "booking-124",
		RestaurantID: "restaurant-456",
		UserID:       "user-789",
		Date:         time.Now().Add(24 * time.Hour),
		Time:         "20:00",
		GuestsCount:  2,
		Status:       domain.BookingStatusCompleted,
	}

	bookingRepo.On("GetByID", mock.Anything, "booking-123").Return(pendingBooking, nil)
	bookingRepo.On("GetByID", mock.Anything, "booking-124").Return(completedBooking, nil)
	bookingRepo.On("GetByID", mock.Anything, "non-existent").Return(nil, errors.New("booking not found"))

	bookingRepo.On("UpdateStatus", mock.Anything, "booking-123", domain.BookingStatusCancelled).Return(nil)

	notificationSvc.On("NotifyRestaurant", mock.Anything, "restaurant-456", domain.NotificationTypeBookingCancelled, mock.Anything, mock.Anything, mock.Anything).Return(nil)

	uc := usecase.NewBookingUseCase(bookingRepo, availabilityRepo, notificationSvc)

	t.Run("successful booking cancellation", func(t *testing.T) {
		ctx := newTestContext()
		err := uc.CancelBooking(ctx, "booking-123")

		assert.NoError(t, err)
	})

	t.Run("booking already completed", func(t *testing.T) {
		ctx := newTestContext()
		err := uc.CancelBooking(ctx, "booking-124")

		assert.Error(t, err)
		assert.Equal(t, usecase.ErrInvalidBookingStatus, err)
	})

	t.Run("booking not found", func(t *testing.T) {
		ctx := newTestContext()
		err := uc.CancelBooking(ctx, "non-existent")

		assert.Error(t, err)
	})
}

func TestCompleteBooking(t *testing.T) {
	bookingRepo := new(MockBookingRepository)
	availabilityRepo := new(MockAvailabilityRepository)
	notificationSvc := new(MockNotificationService)

	confirmedBooking := &domain.Booking{
		ID:           "booking-123",
		RestaurantID: "restaurant-456",
		UserID:       "user-789",
		Date:         time.Now().Add(24 * time.Hour),
		Time:         "19:00",
		GuestsCount:  4,
		Status:       domain.BookingStatusConfirmed,
	}

	pendingBooking := &domain.Booking{
		ID:           "booking-124",
		RestaurantID: "restaurant-456",
		UserID:       "user-789",
		Date:         time.Now().Add(24 * time.Hour),
		Time:         "20:00",
		GuestsCount:  2,
		Status:       domain.BookingStatusPending,
	}

	bookingRepo.On("GetByID", mock.Anything, "booking-123").Return(confirmedBooking, nil)
	bookingRepo.On("GetByID", mock.Anything, "booking-124").Return(pendingBooking, nil)
	bookingRepo.On("GetByID", mock.Anything, "non-existent").Return(nil, errors.New("booking not found"))

	bookingRepo.On("UpdateStatus", mock.Anything, "booking-123", domain.BookingStatusCompleted).Return(nil)

	uc := usecase.NewBookingUseCase(bookingRepo, availabilityRepo, notificationSvc)

	t.Run("successful booking completion", func(t *testing.T) {
		ctx := newTestContext()
		err := uc.CompleteBooking(ctx, "booking-123")

		assert.NoError(t, err)
	})

	t.Run("booking in pending status", func(t *testing.T) {
		ctx := newTestContext()
		err := uc.CompleteBooking(ctx, "booking-124")

		assert.Error(t, err)
		assert.Equal(t, usecase.ErrInvalidBookingStatus, err)
	})

	t.Run("booking not found", func(t *testing.T) {
		ctx := newTestContext()
		err := uc.CompleteBooking(ctx, "non-existent")

		assert.Error(t, err)
	})
}

func TestSuggestAlternativeTime(t *testing.T) {
	bookingRepo := new(MockBookingRepository)
	availabilityRepo := new(MockAvailabilityRepository)
	notificationSvc := new(MockNotificationService)

	pendingBooking := &domain.Booking{
		ID:           "booking-123",
		RestaurantID: "restaurant-456",
		UserID:       "user-789",
		Date:         time.Now().Add(24 * time.Hour),
		Time:         "19:00",
		GuestsCount:  4,
		Status:       domain.BookingStatusPending,
	}

	confirmedBooking := &domain.Booking{
		ID:           "booking-124",
		RestaurantID: "restaurant-456",
		UserID:       "user-789",
		Date:         time.Now().Add(24 * time.Hour),
		Time:         "20:00",
		GuestsCount:  2,
		Status:       domain.BookingStatusConfirmed,
	}

	alternativeDate := time.Now().Add(25 * time.Hour)
	alternativeTime := "20:00"
	message := "unfortunately, all tables at 19:00 are occupied, but we can offer a time at 20:00"

	bookingRepo.On("GetByID", mock.Anything, "booking-123").Return(pendingBooking, nil)
	bookingRepo.On("GetByID", mock.Anything, "booking-124").Return(confirmedBooking, nil)
	bookingRepo.On("GetByID", mock.Anything, "non-existent").Return(nil, errors.New("booking not found"))

	bookingRepo.On("AddAlternative", mock.Anything, mock.MatchedBy(func(alt *domain.BookingAlternative) bool {
		alt.ID = "alt-new-123"
		return alt.BookingID == "booking-123" && alt.Time == alternativeTime
	})).Return(nil)

	notificationSvc.On("NotifyUser", mock.Anything, "user-789", domain.NotificationTypeAlternativeOffer, mock.Anything, mock.Anything, mock.Anything).Return(nil)

	uc := usecase.NewBookingUseCase(bookingRepo, availabilityRepo, notificationSvc)

	t.Run("successful alternative time suggestion", func(t *testing.T) {
		ctx := newTestContext()
		alternativeID, err := uc.SuggestAlternativeTime(ctx, "booking-123", alternativeDate, alternativeTime, message)

		assert.NoError(t, err)
		assert.NotEmpty(t, alternativeID)
	})

	t.Run("booking already confirmed", func(t *testing.T) {
		ctx := newTestContext()
		alternativeID, err := uc.SuggestAlternativeTime(ctx, "booking-124", alternativeDate, alternativeTime, message)

		assert.Error(t, err)
		assert.Equal(t, usecase.ErrInvalidBookingStatus, err)
		assert.Empty(t, alternativeID)
	})

	t.Run("booking not found", func(t *testing.T) {
		ctx := newTestContext()
		alternativeID, err := uc.SuggestAlternativeTime(ctx, "non-existent", alternativeDate, alternativeTime, message)

		assert.Error(t, err)
		assert.Empty(t, alternativeID)
	})
}

func TestAcceptAlternative(t *testing.T) {
	bookingRepo := new(MockBookingRepository)
	availabilityRepo := new(MockAvailabilityRepository)
	notificationSvc := new(MockNotificationService)

	// Создаем тестовые данные
	alternativeID := "alt-123"
	bookingID := "booking-123"
	restaurantID := "rest-123"
	alternativeDate := time.Now().AddDate(0, 0, 1)
	alternativeTime := "18:00"

	alternative := &domain.BookingAlternative{
		ID:        alternativeID,
		BookingID: bookingID,
		Date:      alternativeDate,
		Time:      alternativeTime,
		Message:   "New proposed time",
		CreatedAt: time.Now(),
	}

	booking := &domain.Booking{
		ID:           bookingID,
		RestaurantID: restaurantID,
		UserID:       "user-123",
		Status:       domain.BookingStatusPending,
	}

	// Настраиваем моки
	bookingRepo.On("GetAlternativeByID", mock.Anything, alternativeID).Return(alternative, nil)
	bookingRepo.On("GetAlternativeByID", mock.Anything, "non-existent").Return(nil, errors.New("alternative not found"))
	bookingRepo.On("GetByID", mock.Anything, bookingID).Return(booking, nil)
	bookingRepo.On("AcceptAlternative", mock.Anything, alternativeID).Return(nil)

	notificationSvc.On("NotifyRestaurant", mock.Anything, restaurantID, domain.NotificationTypeAlternativeAccepted, mock.Anything, mock.Anything, bookingID).Return(nil)

	uc := usecase.NewBookingUseCase(bookingRepo, availabilityRepo, notificationSvc)

	t.Run("successful alternative time acceptance", func(t *testing.T) {
		ctx := newTestContext()
		err := uc.AcceptAlternative(ctx, alternativeID)

		assert.NoError(t, err)
		bookingRepo.AssertCalled(t, "GetAlternativeByID", mock.Anything, alternativeID)
		bookingRepo.AssertCalled(t, "GetByID", mock.Anything, bookingID)
		bookingRepo.AssertCalled(t, "AcceptAlternative", mock.Anything, alternativeID)
		notificationSvc.AssertCalled(t, "NotifyRestaurant", mock.Anything, restaurantID, domain.NotificationTypeAlternativeAccepted, mock.Anything, mock.Anything, bookingID)
	})

	t.Run("alternative not found", func(t *testing.T) {
		ctx := newTestContext()
		err := uc.AcceptAlternative(ctx, "non-existent")

		assert.Error(t, err)
		bookingRepo.AssertCalled(t, "GetAlternativeByID", mock.Anything, "non-existent")
		bookingRepo.AssertNotCalled(t, "AcceptAlternative", mock.Anything, "non-existent")
	})
}

func TestRejectAlternative(t *testing.T) {
	bookingRepo := new(MockBookingRepository)
	availabilityRepo := new(MockAvailabilityRepository)
	notificationSvc := new(MockNotificationService)

	// Создаем тестовые данные
	alternativeID := "alt-123"
	bookingID := "booking-123"
	restaurantID := "rest-123"
	alternativeDate := time.Now().AddDate(0, 0, 1)
	alternativeTime := "18:00"

	alternative := &domain.BookingAlternative{
		ID:        alternativeID,
		BookingID: bookingID,
		Date:      alternativeDate,
		Time:      alternativeTime,
		Message:   "New proposed time",
		CreatedAt: time.Now(),
	}

	booking := &domain.Booking{
		ID:           bookingID,
		RestaurantID: restaurantID,
		UserID:       "user-123",
		Status:       domain.BookingStatusPending,
	}

	// Настраиваем моки
	bookingRepo.On("GetAlternativeByID", mock.Anything, alternativeID).Return(alternative, nil)
	bookingRepo.On("GetAlternativeByID", mock.Anything, "non-existent").Return(nil, errors.New("alternative not found"))
	bookingRepo.On("GetByID", mock.Anything, bookingID).Return(booking, nil)
	bookingRepo.On("RejectAlternative", mock.Anything, alternativeID).Return(nil)

	notificationSvc.On("NotifyRestaurant", mock.Anything, restaurantID, domain.NotificationTypeAlternativeRejected, mock.Anything, mock.Anything, bookingID).Return(nil)

	uc := usecase.NewBookingUseCase(bookingRepo, availabilityRepo, notificationSvc)

	t.Run("successful alternative time rejection", func(t *testing.T) {
		ctx := newTestContext()
		err := uc.RejectAlternative(ctx, alternativeID)

		assert.NoError(t, err)
		bookingRepo.AssertCalled(t, "GetAlternativeByID", mock.Anything, alternativeID)
		bookingRepo.AssertCalled(t, "GetByID", mock.Anything, bookingID)
		bookingRepo.AssertCalled(t, "RejectAlternative", mock.Anything, alternativeID)
		notificationSvc.AssertCalled(t, "NotifyRestaurant", mock.Anything, restaurantID, domain.NotificationTypeAlternativeRejected, mock.Anything, mock.Anything, bookingID)
	})

	t.Run("alternative not found", func(t *testing.T) {
		ctx := newTestContext()
		err := uc.RejectAlternative(ctx, "non-existent")

		assert.Error(t, err)
		bookingRepo.AssertCalled(t, "GetAlternativeByID", mock.Anything, "non-existent")
		bookingRepo.AssertNotCalled(t, "RejectAlternative", mock.Anything, "non-existent")
	})
}
