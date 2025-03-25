package repository

import (
	"context"
	"time"

	"github.com/flexer2006/case-back-restaurant-go/internal/domain"
)

type RestaurantRepository interface {
	GetByID(ctx context.Context, id string) (*domain.Restaurant, error)
	List(ctx context.Context, offset, limit int) ([]*domain.Restaurant, error)
	Create(ctx context.Context, restaurant *domain.Restaurant) error
	Update(ctx context.Context, restaurant *domain.Restaurant) error
	Delete(ctx context.Context, id string) error

	AddFact(ctx context.Context, restaurantID string, fact domain.Fact) (*domain.Fact, error)
	GetFacts(ctx context.Context, restaurantID string) ([]domain.Fact, error)
	GetRandomFacts(ctx context.Context, count int) ([]domain.Fact, error)
}

type WorkingHoursRepository interface {
	GetByRestaurantID(ctx context.Context, restaurantID string) ([]*domain.WorkingHours, error)
	SetWorkingHours(ctx context.Context, hours *domain.WorkingHours) error
	DeleteWorkingHours(ctx context.Context, id string) error
}

type AvailabilityRepository interface {
	GetByRestaurantAndDate(ctx context.Context, restaurantID string, date time.Time) ([]*domain.Availability, error)
	SetAvailability(ctx context.Context, availability *domain.Availability) error
	UpdateReservedSeats(ctx context.Context, availabilityID string, delta int) error
}

type BookingRepository interface {
	GetByID(ctx context.Context, id string) (*domain.Booking, error)
	GetByRestaurantID(ctx context.Context, restaurantID string) ([]*domain.Booking, error)
	GetByUserID(ctx context.Context, userID string) ([]*domain.Booking, error)
	Create(ctx context.Context, booking *domain.Booking) error
	UpdateStatus(ctx context.Context, id string, status domain.BookingStatus) error
	AddAlternative(ctx context.Context, alternative *domain.BookingAlternative) error
	GetAlternativeByID(ctx context.Context, alternativeID string) (*domain.BookingAlternative, error)
	AcceptAlternative(ctx context.Context, alternativeID string) error
	RejectAlternative(ctx context.Context, alternativeID string) error
}

type UserRepository interface {
	GetByID(ctx context.Context, id string) (*domain.User, error)
	GetByEmail(ctx context.Context, email string) (*domain.User, error)
	Create(ctx context.Context, user *domain.User) error
	Update(ctx context.Context, user *domain.User) error
}
