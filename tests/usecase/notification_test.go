package usecase_test

import (
	"errors"
	"testing"

	"github.com/flexer2006/case-back-restaurant-go/internal/domain"

	"github.com/flexer2006/case-back-restaurant-go/pkg/usecase"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type MockEmailService struct {
	mock.Mock
}

func (m *MockEmailService) SendEmail(to, subject, body string) error {
	args := m.Called(to, subject, body)
	return args.Error(0)
}

func TestNotifyRestaurant_Success(t *testing.T) {
	mockEmailService := new(MockEmailService)
	mockNotifier := new(MockNotificationService)

	notificationUseCase := usecase.NewNotificationUseCase(mockEmailService, mockNotifier)

	ctx := newTestContext()
	restaurantID := "rest123"
	notificationType := domain.NotificationTypeNewBooking
	title := "new booking"
	message := "you have a new booking"
	relatedID := "booking123"

	mockNotifier.On("NotifyRestaurant", ctx, restaurantID, notificationType, title, message, relatedID).Return(nil)
	mockEmailService.On("SendEmail", restaurantID+"@example.com", title, message).Return(nil)

	err := notificationUseCase.NotifyRestaurant(ctx, restaurantID, notificationType, title, message, relatedID)

	assert.NoError(t, err)
	mockNotifier.AssertExpectations(t)
	mockEmailService.AssertExpectations(t)
}

func TestNotifyRestaurant_Error(t *testing.T) {
	mockEmailService := new(MockEmailService)
	mockNotifier := new(MockNotificationService)

	notificationUseCase := usecase.NewNotificationUseCase(mockEmailService, mockNotifier)

	ctx := newTestContext()
	restaurantID := "rest123"
	notificationType := domain.NotificationTypeNewBooking
	title := "new booking"
	message := "you have a new booking"
	relatedID := "booking123"

	expectedErr := errors.New("notification service error")
	mockNotifier.On("NotifyRestaurant", ctx, restaurantID, notificationType, title, message, relatedID).Return(expectedErr)

	err := notificationUseCase.NotifyRestaurant(ctx, restaurantID, notificationType, title, message, relatedID)

	assert.Error(t, err)
	assert.Equal(t, expectedErr, err)
	mockNotifier.AssertExpectations(t)
	mockEmailService.AssertExpectations(t)
}

func TestNotifyUser_Success(t *testing.T) {
	mockEmailService := new(MockEmailService)
	mockNotifier := new(MockNotificationService)

	notificationUseCase := usecase.NewNotificationUseCase(mockEmailService, mockNotifier)

	ctx := newTestContext()
	userID := "user123"
	userEmail := userID + "@example.com" // Добавляем суффикс email как в реализации
	notificationType := domain.NotificationTypeBookingConfirmed
	title := "booking confirmed"
	message := "your booking has been confirmed"
	relatedID := "booking123"

	mockNotifier.On("NotifyUser", ctx, userID, notificationType, title, message, relatedID).Return(nil)
	mockEmailService.On("SendEmail", userEmail, title, message).Return(nil) // Используем правильный email

	err := notificationUseCase.NotifyUser(ctx, userID, notificationType, title, message, relatedID)

	assert.NoError(t, err)
	mockNotifier.AssertExpectations(t)
	mockEmailService.AssertExpectations(t)
}

func TestNotifyUser_Error(t *testing.T) {
	mockEmailService := new(MockEmailService)
	mockNotifier := new(MockNotificationService)

	notificationUseCase := usecase.NewNotificationUseCase(mockEmailService, mockNotifier)

	ctx := newTestContext()
	userID := "user123"
	notificationType := domain.NotificationTypeBookingConfirmed
	title := "booking confirmed"
	message := "your booking has been confirmed"
	relatedID := "booking123"

	expectedErr := errors.New("notification service error")
	mockNotifier.On("NotifyUser", ctx, userID, notificationType, title, message, relatedID).Return(expectedErr)

	err := notificationUseCase.NotifyUser(ctx, userID, notificationType, title, message, relatedID)

	assert.Error(t, err)
	assert.Equal(t, expectedErr, err)
	mockNotifier.AssertExpectations(t)
	mockEmailService.AssertExpectations(t)
}

func TestGetUserNotifications_Success(t *testing.T) {
	mockEmailService := new(MockEmailService)
	mockNotifier := new(MockNotificationService)

	notificationUseCase := usecase.NewNotificationUseCase(mockEmailService, mockNotifier)

	ctx := newTestContext()
	userID := "user123"

	expectedNotifications := []domain.Notification{
		{
			ID:            "notif1",
			RecipientType: domain.RecipientTypeUser,
			RecipientID:   userID,
			Type:          domain.NotificationTypeBookingConfirmed,
			Title:         "booking confirmed",
			Message:       "your booking has been confirmed",
			IsRead:        false,
			RelatedID:     "booking123",
		},
		{
			ID:            "notif2",
			RecipientType: domain.RecipientTypeUser,
			RecipientID:   userID,
			Type:          domain.NotificationTypeBookingRejected,
			Title:         "booking rejected",
			Message:       "your booking has been rejected",
			IsRead:        true,
			RelatedID:     "booking456",
		},
	}

	mockNotifier.On("GetUserNotifications", ctx, userID).Return(expectedNotifications, nil)

	notifications, err := notificationUseCase.GetUserNotifications(ctx, userID)

	assert.NoError(t, err)
	assert.Equal(t, expectedNotifications, notifications)
	assert.Len(t, notifications, 2)
	mockNotifier.AssertExpectations(t)
}

func TestGetUserNotifications_Error(t *testing.T) {
	mockEmailService := new(MockEmailService)
	mockNotifier := new(MockNotificationService)

	notificationUseCase := usecase.NewNotificationUseCase(mockEmailService, mockNotifier)

	ctx := newTestContext()
	userID := "user123"

	expectedErr := errors.New("notification service error")

	mockNotifier.On("GetUserNotifications", ctx, userID).Return([]domain.Notification(nil), expectedErr)

	notifications, err := notificationUseCase.GetUserNotifications(ctx, userID)

	assert.Error(t, err)
	assert.Equal(t, expectedErr, err)
	assert.Nil(t, notifications)
	mockNotifier.AssertExpectations(t)
}

func TestMarkAsRead_Success(t *testing.T) {
	mockEmailService := new(MockEmailService)
	mockNotifier := new(MockNotificationService)

	notificationUseCase := usecase.NewNotificationUseCase(mockEmailService, mockNotifier)

	ctx := newTestContext()
	notificationID := "notif123"

	mockNotifier.On("MarkAsRead", ctx, notificationID).Return(nil)

	err := notificationUseCase.MarkAsRead(ctx, notificationID)

	assert.NoError(t, err)
	mockNotifier.AssertExpectations(t)
}

func TestMarkAsRead_Error(t *testing.T) {
	mockEmailService := new(MockEmailService)
	mockNotifier := new(MockNotificationService)

	notificationUseCase := usecase.NewNotificationUseCase(mockEmailService, mockNotifier)

	ctx := newTestContext()
	notificationID := "notif123"

	expectedErr := errors.New("notification service error")
	mockNotifier.On("MarkAsRead", ctx, notificationID).Return(expectedErr)

	err := notificationUseCase.MarkAsRead(ctx, notificationID)

	assert.Error(t, err)
	assert.Equal(t, expectedErr, err)
	mockNotifier.AssertExpectations(t)
}
