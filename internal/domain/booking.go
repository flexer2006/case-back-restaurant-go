package domain

import (
	"time"
)

type BookingStatus string

const (
	BookingStatusPending BookingStatus = "pending"

	BookingStatusConfirmed BookingStatus = "confirmed"

	BookingStatusRejected BookingStatus = "rejected"

	BookingStatusCancelled BookingStatus = "cancelled"

	BookingStatusCompleted BookingStatus = "completed"
)

type BookingAlternative struct {
	ID         string     `json:"id"`
	BookingID  string     `json:"booking_id"`
	Date       time.Time  `json:"date"`
	Time       string     `json:"time"`
	Message    string     `json:"message"`
	CreatedAt  time.Time  `json:"created_at"`
	AcceptedAt *time.Time `json:"accepted_at,omitempty"`
	RejectedAt *time.Time `json:"rejected_at,omitempty"`
}

type Booking struct {
	ID           string               `json:"id"`
	RestaurantID string               `json:"restaurant_id"`
	UserID       string               `json:"user_id"`
	Date         time.Time            `json:"date"`
	Time         string               `json:"time"`
	Duration     int                  `json:"duration"`
	GuestsCount  int                  `json:"guests_count"`
	Status       BookingStatus        `json:"status"`
	Comment      string               `json:"comment"`
	CreatedAt    time.Time            `json:"created_at"`
	UpdatedAt    time.Time            `json:"updated_at"`
	ConfirmedAt  *time.Time           `json:"confirmed_at,omitempty"`
	RejectedAt   *time.Time           `json:"rejected_at,omitempty"`
	CompletedAt  *time.Time           `json:"completed_at,omitempty"`
	Alternatives []BookingAlternative `json:"alternatives,omitempty"`
}
