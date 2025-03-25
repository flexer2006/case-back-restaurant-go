package postgres

import (
	"context"
	"fmt"

	"github.com/flexer2006/case-back-restaurant-go/common"
	"github.com/flexer2006/case-back-restaurant-go/internal/domain"
	"github.com/flexer2006/case-back-restaurant-go/internal/logger"

	"go.uber.org/zap"
)

type NotificationService struct {
	repo *NotificationRepository
}

func NewNotificationService(repo *NotificationRepository) domain.NotificationService {
	return &NotificationService{
		repo: repo,
	}
}

func (s *NotificationService) NotifyRestaurant(ctx context.Context, restaurantID string, notificationType domain.NotificationType, title, message string, relatedID string) error {
	logger1, err := logger.FromContext(ctx)
	if err != nil {
		return fmt.Errorf("%s: %w", common.ErrGetLoggerFromContext, err)
	}

	err = s.repo.NotifyRestaurant(ctx, restaurantID, notificationType, title, message, relatedID)
	if err != nil {
		logger1.Error(ctx, common.MsgNotifyRestaurant,
			zap.String("restaurantID", restaurantID),
			zap.Error(err))
		return fmt.Errorf("%s: %w", common.ErrCreateRestaurantNotification, err)
	}

	return nil
}

func (s *NotificationService) NotifyUser(ctx context.Context, userID string, notificationType domain.NotificationType, title, message string, relatedID string) error {
	logger1, err := logger.FromContext(ctx)
	if err != nil {
		return fmt.Errorf("%s: %w", common.ErrGetLoggerFromContext, err)
	}

	err = s.repo.NotifyUser(ctx, userID, notificationType, title, message, relatedID)
	if err != nil {
		logger1.Error(ctx, common.MsgNotifyUser,
			zap.String("userID", userID),
			zap.Error(err))
		return fmt.Errorf("%s: %w", common.ErrCreateUserNotification, err)
	}

	return nil
}

func (s *NotificationService) GetUserNotifications(ctx context.Context, userID string) ([]domain.Notification, error) {
	return s.repo.GetByUserID(ctx, userID)
}

func (s *NotificationService) MarkAsRead(ctx context.Context, notificationID string) error {
	return s.repo.MarkAsRead(ctx, notificationID)
}

type MockEmailService struct{}

func NewMockEmailService() *MockEmailService {
	return &MockEmailService{}
}

func (s *MockEmailService) SendEmail(to, subject, body string) error {
	fmt.Printf("[MOCK EMAIL] To: %s, Subject: %s, Body: %s\n", to, subject, body)
	return nil
}
