package domain

import (
	"context"
	"time"
)

type NotificationType string

const (
	NotificationTypeNewBooking NotificationType = "new_booking"

	NotificationTypeBookingConfirmed NotificationType = "booking_confirmed"

	NotificationTypeBookingRejected NotificationType = "booking_rejected"

	NotificationTypeBookingCancelled NotificationType = "booking_cancelled"

	NotificationTypeAlternativeOffer NotificationType = "alternative_offer"

	NotificationTypeAlternativeAccepted NotificationType = "alternative_accepted"

	NotificationTypeAlternativeRejected NotificationType = "alternative_rejected"
)

type RecipientType string

const (
	RecipientTypeUser RecipientType = "user"

	RecipientTypeRestaurant RecipientType = "restaurant"
)

type Notification struct {
	ID            string           `json:"id"`
	RecipientType RecipientType    `json:"recipient_type"`
	RecipientID   string           `json:"recipient_id"`
	Type          NotificationType `json:"type"`
	Title         string           `json:"title"`
	Message       string           `json:"message"`
	IsRead        bool             `json:"is_read"`
	RelatedID     string           `json:"related_id"`
	CreatedAt     time.Time        `json:"created_at"`
}

type EmailSender interface {
	SendEmail(to, subject, body string) error
}

type NotificationService interface {
	NotifyRestaurant(ctx context.Context, restaurantID string, notificationType NotificationType,
		title, message string, relatedID string) error
	NotifyUser(ctx context.Context, userID string, notificationType NotificationType,
		title, message string, relatedID string) error
	GetUserNotifications(ctx context.Context, userID string) ([]Notification, error)
	MarkAsRead(ctx context.Context, notificationID string) error
}
