package postgres

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/flexer2006/case-back-restaurant-go/common"
	"github.com/flexer2006/case-back-restaurant-go/internal/domain"
	"github.com/flexer2006/case-back-restaurant-go/internal/logger"

	"github.com/google/uuid"
	"go.uber.org/zap"
)

var (
	ErrNotificationNotFound = errors.New(common.ErrNotificationNotFound)
)

type NotificationRepository struct {
	*Repository
}

func NewNotificationRepository(repository *Repository) *NotificationRepository {
	return &NotificationRepository{
		Repository: repository,
	}
}

func (r *NotificationRepository) scanNotification(rows interface{ Scan(dest ...any) error }) (domain.Notification, error) {
	var notification domain.Notification
	var isRead bool

	err := rows.Scan(
		&notification.ID,
		&notification.RecipientType,
		&notification.RecipientID,
		&notification.Type,
		&notification.Title,
		&notification.Message,
		&notification.RelatedID,
		&notification.CreatedAt,
		&isRead,
	)
	if err != nil {
		return notification, fmt.Errorf("%s: %w", common.ErrScanNotification, err)
	}

	notification.IsRead = isRead

	return notification, nil
}

func (r *NotificationRepository) Create(ctx context.Context, notification *domain.Notification) error {
	log, _ := logger.FromContext(ctx)

	if notification.ID == "" {
		notification.ID = uuid.New().String()
	}

	const query = `
		INSERT INTO notifications (id, recipient_type, recipient_id, type, title, message, is_read, related_id, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
	`

	if notification.CreatedAt.IsZero() {
		notification.CreatedAt = time.Now()
	}

	executor, release, err := r.GetExecutor(ctx)
	if err != nil {
		log.Error(ctx, common.ErrGetQueryExecutor, zap.Error(err))
		return err
	}
	defer release()

	_, err = executor.Exec(ctx, query,
		notification.ID,
		notification.RecipientType,
		notification.RecipientID,
		notification.Type,
		notification.Title,
		notification.Message,
		notification.IsRead,
		notification.RelatedID,
		notification.CreatedAt,
	)
	if err != nil {
		log.Error(ctx, common.ErrCreateNotification,
			zap.String("recipientID", notification.RecipientID),
			zap.String("type", string(notification.Type)),
			zap.Error(err))
		return err
	}

	return nil
}

func (r *NotificationRepository) getNotificationsByRecipient(ctx context.Context, recipientType domain.RecipientType, recipientID string) ([]domain.Notification, error) {
	log, _ := logger.FromContext(ctx)

	const query = `
		SELECT id, recipient_type, recipient_id, type, title, message, related_id, created_at, is_read
		FROM notifications
		WHERE recipient_type = $1 AND recipient_id = $2
		ORDER BY created_at DESC
	`

	executor, release, err := r.GetExecutor(ctx)
	if err != nil {
		log.Error(ctx, common.ErrGetQueryExecutor, zap.Error(err))
		return nil, err
	}
	defer release()

	rows, err := executor.Query(ctx, query, recipientType, recipientID)
	if err != nil {
		log.Error(ctx, common.ErrExecuteNotificationsQuery,
			zap.String("recipientType", string(recipientType)),
			zap.String("recipientID", recipientID),
			zap.Error(err))
		return nil, err
	}
	defer rows.Close()

	notifications := make([]domain.Notification, 0)
	for rows.Next() {
		notification, err := r.scanNotification(rows)
		if err != nil {
			log.Error(ctx, common.ErrScanNotification, zap.Error(err))
			return nil, err
		}
		notifications = append(notifications, notification)
	}

	if err = rows.Err(); err != nil {
		log.Error(ctx, common.ErrIterateNotifications, zap.Error(err))
		return nil, err
	}

	return notifications, nil
}

func (r *NotificationRepository) GetByUserID(ctx context.Context, userID string) ([]domain.Notification, error) {
	return r.getNotificationsByRecipient(ctx, domain.RecipientTypeUser, userID)
}

func (r *NotificationRepository) GetByRestaurantID(ctx context.Context, restaurantID string) ([]domain.Notification, error) {
	return r.getNotificationsByRecipient(ctx, domain.RecipientTypeRestaurant, restaurantID)
}

func (r *NotificationRepository) MarkAsRead(ctx context.Context, notificationID string) error {
	log, _ := logger.FromContext(ctx)

	const query = `
		UPDATE notifications
		SET is_read = true
		WHERE id = $1 AND is_read = false
	`

	executor, release, err := r.GetExecutor(ctx)
	if err != nil {
		log.Error(ctx, common.ErrGetQueryExecutor, zap.Error(err))
		return err
	}
	defer release()

	exists, err := r.checkNotificationExists(ctx, notificationID, executor)
	if err != nil {
		log.Error(ctx, common.ErrCheckNotificationExistence,
			zap.String("notificationID", notificationID),
			zap.Error(err))
		return err
	}
	if !exists {
		return ErrNotificationNotFound
	}

	_, err = executor.Exec(ctx, query, notificationID)
	if err != nil {
		log.Error(ctx, common.ErrMarkNotificationAsRead,
			zap.String("notificationID", notificationID),
			zap.Error(err))
		return err
	}

	return nil
}

func (r *NotificationRepository) NotifyUser(ctx context.Context, userID string, notificationType domain.NotificationType, title, message, relatedID string) error {
	notification := domain.Notification{
		ID:            uuid.New().String(),
		RecipientType: domain.RecipientTypeUser,
		RecipientID:   userID,
		Type:          notificationType,
		Title:         title,
		Message:       message,
		IsRead:        false,
		RelatedID:     relatedID,
		CreatedAt:     time.Now(),
	}

	return r.Create(ctx, &notification)
}

func (r *NotificationRepository) NotifyRestaurant(ctx context.Context, restaurantID string, notificationType domain.NotificationType, title, message, relatedID string) error {
	notification := domain.Notification{
		ID:            uuid.New().String(),
		RecipientType: domain.RecipientTypeRestaurant,
		RecipientID:   restaurantID,
		Type:          notificationType,
		Title:         title,
		Message:       message,
		IsRead:        false,
		RelatedID:     relatedID,
		CreatedAt:     time.Now(),
	}

	return r.Create(ctx, &notification)
}

func (r *NotificationRepository) checkNotificationExists(ctx context.Context, id string, executor DBExecutor) (bool, error) {
	const query = `
		SELECT EXISTS(SELECT 1 FROM notifications WHERE id = $1)
	`

	var exists bool
	err := executor.QueryRow(ctx, query, id).Scan(&exists)
	if err != nil {
		return false, err
	}

	return exists, nil
}

func (r *NotificationRepository) GetByID(ctx context.Context, id string) (*domain.Notification, error) {
	logger, err := logger.FromContext(ctx)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", common.ErrScanNotification, err)
	}

	const query = `
		SELECT id, recipient_type, recipient_id, type, title, message, related_id, created_at, read_at
		FROM notifications
		WHERE id = $1
	`

	executor, release, err := r.GetExecutor(ctx)
	if err != nil {
		logger.Error(ctx, common.ErrGetQueryExecutor, zap.Error(err))
		return nil, fmt.Errorf("%s: %w", common.ErrGetQueryExecutor, err)
	}
	defer release()

	rows, err := executor.Query(ctx, query, id)
	if err != nil {
		logger.Error(ctx, common.ErrExecuteNotificationsQuery, zap.Error(err))
		return nil, fmt.Errorf("%s: %w", common.ErrExecuteNotificationsQuery, err)
	}
	defer rows.Close()

	if !rows.Next() {
		if err := rows.Err(); err != nil {
			logger.Error(ctx, common.ErrIterateNotifications, zap.Error(err))
			return nil, fmt.Errorf("%s: %w", common.ErrIterateNotifications, err)
		}
		return nil, fmt.Errorf("%s: %w", common.ErrNotificationNotFound, errors.New("notification not found"))
	}

	notification, err := r.scanNotification(rows)
	if err != nil {
		logger.Error(ctx, common.ErrScanNotification, zap.Error(err))
		return nil, fmt.Errorf("%s: %w", common.ErrScanNotification, err)
	}

	return &notification, nil
}
