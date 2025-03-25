package usecase

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/flexer2006/case-back-restaurant-go/internal/domain"
	"github.com/flexer2006/case-back-restaurant-go/internal/logger"
	"github.com/flexer2006/case-back-restaurant-go/internal/repository"

	"go.uber.org/zap"
)

var (
	ErrNoAvailability       = errors.New("no availability for this time")
	ErrInvalidBookingStatus = errors.New("invalid booking status")
)

type BookingUseCase interface {
	GetBooking(ctx context.Context, id string) (*domain.Booking, error)

	GetRestaurantBookings(ctx context.Context, restaurantID string) ([]*domain.Booking, error)

	GetUserBookings(ctx context.Context, userID string) ([]*domain.Booking, error)

	CreateBooking(ctx context.Context, booking *domain.Booking) (string, error)

	ConfirmBooking(ctx context.Context, id string) error

	RejectBooking(ctx context.Context, id string, reason string) error

	CancelBooking(ctx context.Context, id string) error

	CompleteBooking(ctx context.Context, id string) error

	SuggestAlternativeTime(ctx context.Context, bookingID string, date time.Time, time string, message string) (string, error)

	AcceptAlternative(ctx context.Context, alternativeID string) error

	RejectAlternative(ctx context.Context, alternativeID string) error
}

type bookingUseCase struct {
	bookingRepo      repository.BookingRepository
	availabilityRepo repository.AvailabilityRepository
	notificationSvc  domain.NotificationService
}

func NewBookingUseCase(
	bookingRepo repository.BookingRepository,
	availabilityRepo repository.AvailabilityRepository,
	notificationSvc domain.NotificationService,
) BookingUseCase {
	return &bookingUseCase{
		bookingRepo:      bookingRepo,
		availabilityRepo: availabilityRepo,
		notificationSvc:  notificationSvc,
	}
}

func (u *bookingUseCase) GetBooking(ctx context.Context, id string) (*domain.Booking, error) {
	return u.bookingRepo.GetByID(ctx, id)
}

func (u *bookingUseCase) GetRestaurantBookings(ctx context.Context, restaurantID string) ([]*domain.Booking, error) {
	return u.bookingRepo.GetByRestaurantID(ctx, restaurantID)
}

func (u *bookingUseCase) GetUserBookings(ctx context.Context, userID string) ([]*domain.Booking, error) {
	return u.bookingRepo.GetByUserID(ctx, userID)
}

func (u *bookingUseCase) CreateBooking(ctx context.Context, booking *domain.Booking) (string, error) {
	log, _ := logger.FromContext(ctx)
	log.Info(ctx, "creating new booking",
		zap.String("restaurantID", booking.RestaurantID),
		zap.Time("date", booking.Date),
		zap.String("time", booking.Time),
		zap.Int("guests", booking.GuestsCount))

	availabilities, err := u.availabilityRepo.GetByRestaurantAndDate(ctx, booking.RestaurantID, booking.Date)
	if err != nil {
		log.Error(ctx, "failed to get availability",
			zap.String("restaurantID", booking.RestaurantID),
			zap.Time("date", booking.Date),
			zap.Error(err))
		return "", err
	}

	var availabilityID string
	var availableSeats int

	for _, avail := range availabilities {
		if avail.TimeSlot == booking.Time {
			availabilityID = avail.ID
			availableSeats = avail.AvailableSeats()
			break
		}
	}

	if availabilityID == "" || availableSeats < booking.GuestsCount {
		log.Warn(ctx, "no availability for booking",
			zap.String("restaurantID", booking.RestaurantID),
			zap.Time("date", booking.Date),
			zap.String("time", booking.Time),
			zap.Int("requestedSeats", booking.GuestsCount),
			zap.Int("availableSeats", availableSeats))
		return "", ErrNoAvailability
	}

	now := time.Now()
	booking.Status = domain.BookingStatusPending
	booking.CreatedAt = now
	booking.UpdatedAt = now

	if err := u.bookingRepo.Create(ctx, booking); err != nil {
		log.Error(ctx, "failed to create booking", zap.Error(err))
		return "", err
	}

	if err := u.availabilityRepo.UpdateReservedSeats(ctx, availabilityID, booking.GuestsCount); err != nil {
		deleteErr := u.bookingRepo.UpdateStatus(ctx, booking.ID, domain.BookingStatusCancelled)
		if deleteErr != nil {
			log.Error(ctx, "failed to cancel booking after unsuccessful availability update",
				zap.String("bookingID", booking.ID),
				zap.Error(deleteErr))
			fmt.Printf("failed to cancel booking %s after unsuccessful availability update: %v\n",
				booking.ID, deleteErr)
		}
		log.Error(ctx, "failed to update seats availability",
			zap.String("availabilityID", availabilityID),
			zap.Int("guestsCount", booking.GuestsCount),
			zap.Error(err))
		return "", fmt.Errorf("failed to update seats availability: %w", err)
	}

	err = u.notificationSvc.NotifyRestaurant(
		ctx,
		booking.RestaurantID,
		domain.NotificationTypeNewBooking,
		"New booking",
		"You have a new booking on "+booking.Date.Format("02.01.2006")+" at "+booking.Time,
		booking.ID,
	)
	if err != nil {
		log.Error(ctx, "failed to send notification to restaurant",
			zap.String("restaurantID", booking.RestaurantID),
			zap.String("bookingID", booking.ID),
			zap.Error(err))
	}

	log.Info(ctx, "booking successfully created",
		zap.String("bookingID", booking.ID),
		zap.String("restaurantID", booking.RestaurantID),
		zap.Time("date", booking.Date),
		zap.String("time", booking.Time))

	return booking.ID, nil
}

func (u *bookingUseCase) ConfirmBooking(ctx context.Context, id string) error {
	log, _ := logger.FromContext(ctx)
	log.Info(ctx, "confirming booking", zap.String("bookingID", id))

	booking, err := u.bookingRepo.GetByID(ctx, id)
	if err != nil {
		log.Error(ctx, "failed to get booking", zap.String("bookingID", id), zap.Error(err))
		return err
	}

	if booking.Status != domain.BookingStatusPending {
		log.Warn(ctx, "invalid booking status for confirmation",
			zap.String("bookingID", id),
			zap.String("currentStatus", string(booking.Status)))
		return ErrInvalidBookingStatus
	}

	now := time.Now()
	booking.Status = domain.BookingStatusConfirmed
	booking.UpdatedAt = now
	booking.ConfirmedAt = &now

	if err := u.bookingRepo.UpdateStatus(ctx, id, domain.BookingStatusConfirmed); err != nil {
		log.Error(ctx, "failed to update booking status",
			zap.String("bookingID", id),
			zap.Error(err))
		return err
	}

	err = u.notificationSvc.NotifyUser(
		ctx,
		booking.UserID,
		domain.NotificationTypeBookingConfirmed,
		"Booking confirmed",
		"Your booking on "+booking.Date.Format("02.01.2006")+" at "+booking.Time+" has been confirmed by the restaurant.",
		booking.ID,
	)
	if err != nil {
		log.Error(ctx, "failed to send notification to user",
			zap.String("userID", booking.UserID),
			zap.String("bookingID", id),
			zap.Error(err))
	}

	log.Info(ctx, "booking successfully confirmed",
		zap.String("bookingID", id),
		zap.String("restaurantID", booking.RestaurantID),
		zap.String("userID", booking.UserID))

	return nil
}

func (u *bookingUseCase) RejectBooking(ctx context.Context, id string, reason string) error {
	log, _ := logger.FromContext(ctx)
	log.Info(ctx, "rejecting booking",
		zap.String("bookingID", id),
		zap.String("reason", reason))

	booking, err := u.bookingRepo.GetByID(ctx, id)
	if err != nil {
		log.Error(ctx, "failed to get booking",
			zap.String("bookingID", id),
			zap.Error(err))
		return err
	}

	if booking.Status != domain.BookingStatusPending {
		log.Warn(ctx, "invalid booking status for rejection",
			zap.String("bookingID", id),
			zap.String("currentStatus", string(booking.Status)))
		return ErrInvalidBookingStatus
	}

	now := time.Now()
	booking.Status = domain.BookingStatusRejected
	booking.UpdatedAt = now
	booking.RejectedAt = &now

	if err := u.bookingRepo.UpdateStatus(ctx, id, domain.BookingStatusRejected); err != nil {
		log.Error(ctx, "failed to update booking status",
			zap.String("bookingID", id),
			zap.Error(err))
		return err
	}

	message := "Your booking on " + booking.Date.Format("02.01.2006") + " at " + booking.Time + " has been rejected by the restaurant."
	if reason != "" {
		message += " Reason: " + reason
	}

	err = u.notificationSvc.NotifyUser(
		ctx,
		booking.UserID,
		domain.NotificationTypeBookingRejected,
		"Booking rejected",
		message,
		booking.ID,
	)
	if err != nil {
		log.Error(ctx, "failed to send notification to user",
			zap.String("userID", booking.UserID),
			zap.String("bookingID", id),
			zap.Error(err))
	}

	log.Info(ctx, "booking successfully rejected",
		zap.String("bookingID", id),
		zap.String("restaurantID", booking.RestaurantID),
		zap.String("userID", booking.UserID),
		zap.String("reason", reason))

	return nil
}

func (u *bookingUseCase) CancelBooking(ctx context.Context, id string) error {
	log, _ := logger.FromContext(ctx)
	log.Info(ctx, "canceling booking", zap.String("bookingID", id))

	booking, err := u.bookingRepo.GetByID(ctx, id)
	if err != nil {
		log.Error(ctx, "failed to get booking",
			zap.String("bookingID", id),
			zap.Error(err))
		return err
	}

	if booking.Status == domain.BookingStatusCompleted || booking.Status == domain.BookingStatusRejected {
		log.Warn(ctx, "invalid booking status for cancellation",
			zap.String("bookingID", id),
			zap.String("currentStatus", string(booking.Status)))
		return ErrInvalidBookingStatus
	}

	booking.Status = domain.BookingStatusCancelled
	booking.UpdatedAt = time.Now()

	if err := u.bookingRepo.UpdateStatus(ctx, id, domain.BookingStatusCancelled); err != nil {
		log.Error(ctx, "failed to update booking status",
			zap.String("bookingID", id),
			zap.Error(err))
		return err
	}

	err = u.notificationSvc.NotifyRestaurant(
		ctx,
		booking.RestaurantID,
		domain.NotificationTypeBookingCancelled,
		"Booking cancelled",
		"Booking on "+booking.Date.Format("02.01.2006")+" at "+booking.Time+" has been cancelled by the user.",
		booking.ID,
	)
	if err != nil {
		log.Error(ctx, "failed to send notification to restaurant",
			zap.String("restaurantID", booking.RestaurantID),
			zap.String("bookingID", id),
			zap.Error(err))
	}

	log.Info(ctx, "booking successfully cancelled",
		zap.String("bookingID", id),
		zap.String("restaurantID", booking.RestaurantID),
		zap.String("userID", booking.UserID))

	return nil
}

func (u *bookingUseCase) CompleteBooking(ctx context.Context, id string) error {
	log, _ := logger.FromContext(ctx)
	log.Info(ctx, "completing booking", zap.String("bookingID", id))

	booking, err := u.bookingRepo.GetByID(ctx, id)
	if err != nil {
		log.Error(ctx, "failed to get booking", zap.String("bookingID", id), zap.Error(err))
		return err
	}

	if booking.Status != domain.BookingStatusConfirmed {
		log.Warn(ctx, "invalid booking status for completion",
			zap.String("bookingID", id),
			zap.String("currentStatus", string(booking.Status)))
		return ErrInvalidBookingStatus
	}

	now := time.Now()
	booking.Status = domain.BookingStatusCompleted
	booking.UpdatedAt = now
	booking.CompletedAt = &now

	if err := u.bookingRepo.UpdateStatus(ctx, id, domain.BookingStatusCompleted); err != nil {
		log.Error(ctx, "failed to update booking status",
			zap.String("bookingID", id),
			zap.Error(err))
		return err
	}

	log.Info(ctx, "booking successfully completed",
		zap.String("bookingID", id),
		zap.String("restaurantID", booking.RestaurantID),
		zap.String("userID", booking.UserID))

	return nil
}

func (u *bookingUseCase) SuggestAlternativeTime(ctx context.Context, bookingID string, date time.Time, timeSlot string, message string) (string, error) {
	log, _ := logger.FromContext(ctx)
	log.Info(ctx, "suggesting alternative booking time",
		zap.String("bookingID", bookingID),
		zap.Time("alternativeDate", date),
		zap.String("alternativeTime", timeSlot))

	booking, err := u.bookingRepo.GetByID(ctx, bookingID)
	if err != nil {
		log.Error(ctx, "failed to get booking",
			zap.String("bookingID", bookingID),
			zap.Error(err))
		return "", err
	}

	if booking.Status != domain.BookingStatusPending {
		log.Warn(ctx, "invalid booking status for suggesting alternative time",
			zap.String("bookingID", bookingID),
			zap.String("currentStatus", string(booking.Status)))
		return "", ErrInvalidBookingStatus
	}

	alternative := &domain.BookingAlternative{
		BookingID: bookingID,
		Date:      date,
		Time:      timeSlot,
		Message:   message,
		CreatedAt: time.Now(),
	}

	if err := u.bookingRepo.AddAlternative(ctx, alternative); err != nil {
		log.Error(ctx, "failed to add booking alternative",
			zap.String("bookingID", bookingID),
			zap.Error(err))
		return "", err
	}

	err = u.notificationSvc.NotifyUser(
		ctx,
		booking.UserID,
		domain.NotificationTypeAlternativeOffer,
		"Alternative time offered",
		"Restaurant offers alternative time for your booking: "+date.Format("02.01.2006")+" at "+timeSlot,
		bookingID,
	)
	if err != nil {
		log.Error(ctx, "failed to send notification to user",
			zap.String("userID", booking.UserID),
			zap.String("bookingID", bookingID),
			zap.Error(err))
	}

	log.Info(ctx, "alternative booking time successfully suggested",
		zap.String("alternativeID", alternative.ID),
		zap.String("bookingID", bookingID),
		zap.String("userID", booking.UserID))

	return alternative.ID, nil
}

func (u *bookingUseCase) AcceptAlternative(ctx context.Context, alternativeID string) error {
	log, _ := logger.FromContext(ctx)
	log.Info(ctx, "accepting alternative booking offer", zap.String("alternativeID", alternativeID))

	alternative, err := u.bookingRepo.GetAlternativeByID(ctx, alternativeID)
	if err != nil {
		log.Error(ctx, "failed to get alternative offer",
			zap.String("alternativeID", alternativeID),
			zap.Error(err))
		return err
	}

	booking, err := u.bookingRepo.GetByID(ctx, alternative.BookingID)
	if err != nil {
		log.Error(ctx, "failed to get booking for alternative offer",
			zap.String("bookingID", alternative.BookingID),
			zap.String("alternativeID", alternativeID),
			zap.Error(err))
		return err
	}

	if err := u.bookingRepo.AcceptAlternative(ctx, alternativeID); err != nil {
		log.Error(ctx, "failed to accept alternative offer",
			zap.String("alternativeID", alternativeID),
			zap.Error(err))
		return err
	}

	err = u.notificationSvc.NotifyRestaurant(
		ctx,
		booking.RestaurantID,
		domain.NotificationTypeAlternativeAccepted,
		"Alternative booking accepted",
		"User has accepted your alternative booking offer for "+alternative.Date.Format("02.01.2006")+" at "+alternative.Time,
		booking.ID,
	)
	if err != nil {
		log.Error(ctx, "failed to send notification to restaurant",
			zap.String("restaurantID", booking.RestaurantID),
			zap.String("bookingID", booking.ID),
			zap.Error(err))

	}

	log.Info(ctx, "alternative booking offer successfully accepted",
		zap.String("alternativeID", alternativeID),
		zap.String("bookingID", booking.ID),
		zap.String("restaurantID", booking.RestaurantID))

	return nil
}

func (u *bookingUseCase) RejectAlternative(ctx context.Context, alternativeID string) error {
	log, _ := logger.FromContext(ctx)
	log.Info(ctx, "rejecting alternative booking offer", zap.String("alternativeID", alternativeID))

	alternative, err := u.bookingRepo.GetAlternativeByID(ctx, alternativeID)
	if err != nil {
		log.Error(ctx, "failed to get alternative offer",
			zap.String("alternativeID", alternativeID),
			zap.Error(err))
		return err
	}

	booking, err := u.bookingRepo.GetByID(ctx, alternative.BookingID)
	if err != nil {
		log.Error(ctx, "failed to get booking for alternative offer",
			zap.String("bookingID", alternative.BookingID),
			zap.String("alternativeID", alternativeID),
			zap.Error(err))
		return err
	}

	if err := u.bookingRepo.RejectAlternative(ctx, alternativeID); err != nil {
		log.Error(ctx, "failed to reject alternative offer",
			zap.String("alternativeID", alternativeID),
			zap.Error(err))
		return err
	}

	err = u.notificationSvc.NotifyRestaurant(
		ctx,
		booking.RestaurantID,
		domain.NotificationTypeAlternativeRejected,
		"Alternative booking rejected",
		"User has rejected your alternative booking offer for "+alternative.Date.Format("02.01.2006")+" at "+alternative.Time,
		booking.ID,
	)
	if err != nil {
		log.Error(ctx, "failed to send notification to restaurant",
			zap.String("restaurantID", booking.RestaurantID),
			zap.String("bookingID", booking.ID),
			zap.Error(err))

	}

	log.Info(ctx, "alternative booking offer successfully rejected",
		zap.String("alternativeID", alternativeID),
		zap.String("bookingID", booking.ID),
		zap.String("restaurantID", booking.RestaurantID))

	return nil
}
