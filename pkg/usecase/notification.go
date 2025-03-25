package usecase

import (
	"context"

	"github.com/flexer2006/case-back-restaurant-go/internal/domain"
	"github.com/flexer2006/case-back-restaurant-go/internal/logger"

	"go.uber.org/zap"
)

type NotificationUseCase interface {
	NotifyRestaurant(ctx context.Context, restaurantID string, notificationType domain.NotificationType, title, message, relatedID string) error

	NotifyUser(ctx context.Context, userID string, notificationType domain.NotificationType, title, message, relatedID string) error

	GetUserNotifications(ctx context.Context, userID string) ([]domain.Notification, error)

	MarkAsRead(ctx context.Context, notificationID string) error
}

type notificationUseCase struct {
	emailService EmailService
	notifier     domain.NotificationService
}

type EmailService interface {
	SendEmail(to, subject, body string) error
}

func NewNotificationUseCase(
	emailService EmailService,
	notifier domain.NotificationService,
) NotificationUseCase {
	return &notificationUseCase{
		emailService: emailService,
		notifier:     notifier,
	}
}

func (u *notificationUseCase) NotifyRestaurant(ctx context.Context, restaurantID string, notificationType domain.NotificationType, title, message, relatedID string) error {
	log, _ := logger.FromContext(ctx)
	log.Info(ctx, "sending notification to restaurant",
		zap.String("restaurantID", restaurantID),
		zap.String("type", string(notificationType)),
		zap.String("title", title),
		zap.String("relatedID", relatedID))

	err := u.notifier.NotifyRestaurant(ctx, restaurantID, notificationType, title, message, relatedID)
	if err != nil {
		log.Error(ctx, "failed to send notification to restaurant",
			zap.String("restaurantID", restaurantID),
			zap.Error(err))
		return err
	}

	restaurantEmail := u.getRestaurantEmail(restaurantID)

	if err := u.emailService.SendEmail(restaurantEmail, title, message); err != nil {
		log.Error(ctx, "failed to send email to restaurant",
			zap.String("restaurantID", restaurantID),
			zap.Error(err))

	}

	log.Info(ctx, "notification to restaurant successfully sent",
		zap.String("restaurantID", restaurantID),
		zap.String("type", string(notificationType)))

	return nil
}

func (u *notificationUseCase) NotifyUser(ctx context.Context, userID string, notificationType domain.NotificationType, title, message, relatedID string) error {
	log, _ := logger.FromContext(ctx)
	log.Info(ctx, "sending notification to user",
		zap.String("userID", userID),
		zap.String("type", string(notificationType)),
		zap.String("title", title),
		zap.String("relatedID", relatedID))

	err := u.notifier.NotifyUser(ctx, userID, notificationType, title, message, relatedID)
	if err != nil {
		log.Error(ctx, "failed to send notification to user",
			zap.String("userID", userID),
			zap.Error(err))
		return err
	}

	userEmail := u.getUserEmail(userID)

	if err := u.emailService.SendEmail(userEmail, title, message); err != nil {
		log.Error(ctx, "failed to send email to user",
			zap.String("userID", userID),
			zap.Error(err))
	}

	log.Info(ctx, "notification to user successfully sent",
		zap.String("userID", userID),
		zap.String("type", string(notificationType)))

	return nil
}

func (u *notificationUseCase) getUserEmail(userID string) string {

	return userID + "@example.com"
}

func (u *notificationUseCase) getRestaurantEmail(restaurantID string) string {

	return restaurantID + "@example.com"
}

func (u *notificationUseCase) GetUserNotifications(ctx context.Context, userID string) ([]domain.Notification, error) {
	log, _ := logger.FromContext(ctx)
	log.Info(ctx, "getting user notifications",
		zap.String("userID", userID))

	notifications, err := u.notifier.GetUserNotifications(ctx, userID)
	if err != nil {
		log.Error(ctx, "failed to get user notifications",
			zap.String("userID", userID),
			zap.Error(err))
		return nil, err
	}

	log.Info(ctx, "user notifications successfully retrieved",
		zap.String("userID", userID),
		zap.Int("count", len(notifications)))
	return notifications, nil
}

func (u *notificationUseCase) MarkAsRead(ctx context.Context, notificationID string) error {
	log, _ := logger.FromContext(ctx)
	log.Info(ctx, "marking notification as read",
		zap.String("notificationID", notificationID))

	if err := u.notifier.MarkAsRead(ctx, notificationID); err != nil {
		log.Error(ctx, "failed to mark notification as read",
			zap.String("notificationID", notificationID),
			zap.Error(err))
		return err
	}

	log.Info(ctx, "notification successfully marked as read",
		zap.String("notificationID", notificationID))
	return nil
}
