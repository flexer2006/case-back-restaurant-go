package usecase

import (
	"context"
	"time"

	"github.com/flexer2006/case-back-restaurant-go/internal/domain"
	"github.com/flexer2006/case-back-restaurant-go/internal/logger"
	"github.com/flexer2006/case-back-restaurant-go/internal/repository"

	"go.uber.org/zap"
)

type AvailabilityUseCase interface {
	GetAvailability(ctx context.Context, restaurantID string, date time.Time) ([]*domain.Availability, error)

	SetAvailability(ctx context.Context, availability *domain.Availability) error

	UpdateReservedSeats(ctx context.Context, availabilityID string, delta int) error

	CheckAvailability(ctx context.Context, restaurantID string, date time.Time, timeSlot string, guestsCount int) (bool, error)
}

type availabilityUseCase struct {
	availabilityRepo repository.AvailabilityRepository
	restaurantRepo   repository.RestaurantRepository
	workingHoursRepo repository.WorkingHoursRepository
}

func NewAvailabilityUseCase(
	availabilityRepo repository.AvailabilityRepository,
	restaurantRepo repository.RestaurantRepository,
	workingHoursRepo repository.WorkingHoursRepository,
) AvailabilityUseCase {
	return &availabilityUseCase{
		availabilityRepo: availabilityRepo,
		restaurantRepo:   restaurantRepo,
		workingHoursRepo: workingHoursRepo,
	}
}

func (u *availabilityUseCase) GetAvailability(ctx context.Context, restaurantID string, date time.Time) ([]*domain.Availability, error) {
	return u.availabilityRepo.GetByRestaurantAndDate(ctx, restaurantID, date)
}

func (u *availabilityUseCase) SetAvailability(ctx context.Context, availability *domain.Availability) error {
	log, _ := logger.FromContext(ctx)
	log.Info(ctx, "setting restaurant availability",
		zap.String("restaurantID", availability.RestaurantID),
		zap.Time("date", availability.Date),
		zap.String("timeSlot", availability.TimeSlot),
		zap.Int("capacity", availability.Capacity),
		zap.Int("reserved", availability.Reserved))

	availability.UpdatedAt = time.Now()

	if err := u.availabilityRepo.SetAvailability(ctx, availability); err != nil {
		log.Error(ctx, "failed to set restaurant availability",
			zap.String("restaurantID", availability.RestaurantID),
			zap.Time("date", availability.Date),
			zap.Error(err))
		return err
	}

	log.Info(ctx, "restaurant availability successfully set",
		zap.String("availabilityID", availability.ID),
		zap.String("restaurantID", availability.RestaurantID),
		zap.Time("date", availability.Date))
	return nil
}

func (u *availabilityUseCase) UpdateReservedSeats(ctx context.Context, availabilityID string, delta int) error {
	log, _ := logger.FromContext(ctx)
	log.Info(ctx, "updating reserved seats count",
		zap.String("availabilityID", availabilityID),
		zap.Int("delta", delta))

	if err := u.availabilityRepo.UpdateReservedSeats(ctx, availabilityID, delta); err != nil {
		log.Error(ctx, "failed to update reserved seats count",
			zap.String("availabilityID", availabilityID),
			zap.Int("delta", delta),
			zap.Error(err))
		return err
	}

	log.Info(ctx, "reserved seats count successfully updated",
		zap.String("availabilityID", availabilityID),
		zap.Int("delta", delta))
	return nil
}

func (u *availabilityUseCase) CheckAvailability(ctx context.Context, restaurantID string, date time.Time, timeSlot string, guestsCount int) (bool, error) {
	log, _ := logger.FromContext(ctx)
	log.Info(ctx, "checking restaurant availability",
		zap.String("restaurantID", restaurantID),
		zap.Time("date", date),
		zap.String("timeSlot", timeSlot),
		zap.Int("guestsCount", guestsCount))

	availabilities, err := u.availabilityRepo.GetByRestaurantAndDate(ctx, restaurantID, date)
	if err != nil {
		log.Error(ctx, "failed to get restaurant availability",
			zap.String("restaurantID", restaurantID),
			zap.Time("date", date),
			zap.Error(err))
		return false, err
	}

	for _, avail := range availabilities {
		if avail.TimeSlot == timeSlot {
			isAvailable := avail.AvailableSeats() >= guestsCount
			log.Info(ctx, "availability check result",
				zap.String("restaurantID", restaurantID),
				zap.Time("date", date),
				zap.String("timeSlot", timeSlot),
				zap.Int("guestsCount", guestsCount),
				zap.Bool("isAvailable", isAvailable),
				zap.Int("availableSeats", avail.AvailableSeats()))
			return isAvailable, nil
		}
	}

	log.Warn(ctx, "time slot not found",
		zap.String("restaurantID", restaurantID),
		zap.Time("date", date),
		zap.String("timeSlot", timeSlot))
	return false, nil
}
