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
	"github.com/jackc/pgx/v5"
	"go.uber.org/zap"
)

type BookingRepository struct {
	*Repository
}

func NewBookingRepository(repository *Repository) *BookingRepository {
	return &BookingRepository{
		Repository: repository,
	}
}

func (r *BookingRepository) GetByID(ctx context.Context, id string) (*domain.Booking, error) {
	logger, err := logger.FromContext(ctx)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", common.ErrGetBookingData, err)
	}

	const query = `
		SELECT id, restaurant_id, user_id, date, time, duration, guests_count, status, comment,
			   created_at, updated_at, confirmed_at, rejected_at, completed_at
		FROM bookings
		WHERE id = $1
	`

	executor, release, err := r.GetExecutor(ctx)
	if err != nil {
		logger.Error(ctx, common.ErrGetQueryExecutor, zap.Error(err))
		return nil, fmt.Errorf("%s: %w", common.ErrGetQueryExecutor, err)
	}
	defer release()

	var booking domain.Booking
	var confirmedAt, rejectedAt, completedAt *time.Time

	err = executor.QueryRow(ctx, query, id).Scan(
		&booking.ID,
		&booking.RestaurantID,
		&booking.UserID,
		&booking.Date,
		&booking.Time,
		&booking.Duration,
		&booking.GuestsCount,
		&booking.Status,
		&booking.Comment,
		&booking.CreatedAt,
		&booking.UpdatedAt,
		&confirmedAt,
		&rejectedAt,
		&completedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, fmt.Errorf("%s: %w", common.ErrBookingNotFound, err)
		}
		logger.Error(ctx, common.ErrGetBookingData, zap.Error(err))
		return nil, fmt.Errorf("%s: %w", common.ErrGetBookingData, err)
	}

	if confirmedAt != nil {
		booking.ConfirmedAt = confirmedAt
	}
	if rejectedAt != nil {
		booking.RejectedAt = rejectedAt
	}
	if completedAt != nil {
		booking.CompletedAt = completedAt
	}

	alternatives, err := r.getAlternatives(ctx, booking.ID, executor)
	if err != nil {
		logger.Error(ctx, common.ErrGetAlternativeOffers,
			zap.String("bookingID", booking.ID),
			zap.Error(err))
	} else {
		booking.Alternatives = alternatives
	}

	return &booking, nil
}

func (r *BookingRepository) scanBooking(rows pgx.Rows) (*domain.Booking, error) {
	var booking domain.Booking
	var confirmedAt, rejectedAt, completedAt *time.Time

	err := rows.Scan(
		&booking.ID,
		&booking.RestaurantID,
		&booking.UserID,
		&booking.Date,
		&booking.Time,
		&booking.Duration,
		&booking.GuestsCount,
		&booking.Status,
		&booking.Comment,
		&booking.CreatedAt,
		&booking.UpdatedAt,
		&confirmedAt,
		&rejectedAt,
		&completedAt,
	)
	if err != nil {
		return nil, err
	}

	if confirmedAt != nil {
		booking.ConfirmedAt = confirmedAt
	}
	if rejectedAt != nil {
		booking.RejectedAt = rejectedAt
	}
	if completedAt != nil {
		booking.CompletedAt = completedAt
	}

	return &booking, nil
}

func (r *BookingRepository) getBookingsByQuery(ctx context.Context, query string, args ...interface{}) ([]*domain.Booking, error) {
	log, _ := logger.FromContext(ctx)

	executor, release, err := r.GetExecutor(ctx)
	if err != nil {
		log.Error(ctx, common.ErrGetQueryExecutor, zap.Error(err))
		return nil, err
	}
	defer release()

	rows, err := executor.Query(ctx, query, args...)
	if err != nil {
		log.Error(ctx, common.ErrExecuteBookingsQuery, zap.Error(err))
		return nil, err
	}
	defer rows.Close()

	bookings := make([]*domain.Booking, 0)
	for rows.Next() {
		booking, err := r.scanBooking(rows)
		if err != nil {
			log.Error(ctx, common.ErrGetBookingData, zap.Error(err))
			return nil, err
		}
		bookings = append(bookings, booking)
	}

	if err = rows.Err(); err != nil {
		log.Error(ctx, common.ErrIterateBookings, zap.Error(err))
		return nil, err
	}

	return bookings, nil
}

func (r *BookingRepository) GetByRestaurantID(ctx context.Context, restaurantID string) ([]*domain.Booking, error) {
	const query = `
		SELECT id, restaurant_id, user_id, date, time, duration, guests_count, status, comment,
			   created_at, updated_at, confirmed_at, rejected_at, completed_at
		FROM bookings
		WHERE restaurant_id = $1
		ORDER BY date DESC, time DESC
	`

	log, _ := logger.FromContext(ctx)
	bookings, err := r.getBookingsByQuery(ctx, query, restaurantID)
	if err != nil {
		log.Error(ctx, common.ErrGetRestaurantBookings,
			zap.String("restaurantID", restaurantID),
			zap.Error(err))
	}
	return bookings, err
}

func (r *BookingRepository) GetByUserID(ctx context.Context, userID string) ([]*domain.Booking, error) {
	const query = `
		SELECT id, restaurant_id, user_id, date, time, duration, guests_count, status, comment,
			   created_at, updated_at, confirmed_at, rejected_at, completed_at
		FROM bookings
		WHERE user_id = $1
		ORDER BY date DESC, time DESC
	`

	log, _ := logger.FromContext(ctx)
	bookings, err := r.getBookingsByQuery(ctx, query, userID)
	if err != nil {
		log.Error(ctx, common.ErrGetUserBookings,
			zap.String("userID", userID),
			zap.Error(err))
	}
	return bookings, err
}

func (r *BookingRepository) Create(ctx context.Context, booking *domain.Booking) error {
	log, _ := logger.FromContext(ctx)

	if booking.ID == "" {
		booking.ID = uuid.New().String()
	}

	const query = `
		INSERT INTO bookings (id, restaurant_id, user_id, date, time, duration, guests_count, status, comment, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
	`

	executor, release, err := r.GetExecutor(ctx)
	if err != nil {
		log.Error(ctx, common.ErrGetQueryExecutor, zap.Error(err))
		return err
	}
	defer release()

	restaurantExists, err := r.checkRestaurantExists(ctx, booking.RestaurantID, executor)
	if err != nil {
		log.Error(ctx, common.ErrCheckRestaurantExistence,
			zap.String("restaurantID", booking.RestaurantID),
			zap.Error(err))
		return err
	}
	if !restaurantExists {
		return errors.New(common.ErrRestaurantNotFound)
	}

	userExists, err := r.checkUserExists(ctx, booking.UserID, executor)
	if err != nil {
		log.Error(ctx, common.ErrCheckUserExistence,
			zap.String("userID", booking.UserID),
			zap.Error(err))
		return err
	}
	if !userExists {
		return errors.New(common.ErrUserNotFound)
	}

	formattedDate := booking.Date.Format("2006-01-02")

	_, err = executor.Exec(ctx, query,
		booking.ID,
		booking.RestaurantID,
		booking.UserID,
		formattedDate,
		booking.Time,
		booking.Duration,
		booking.GuestsCount,
		booking.Status,
		booking.Comment,
		booking.CreatedAt,
		booking.UpdatedAt,
	)
	if err != nil {
		log.Error(ctx, common.ErrCreateBooking,
			zap.String("userID", booking.UserID),
			zap.String("restaurantID", booking.RestaurantID),
			zap.String("date", formattedDate),
			zap.Error(err))
		return err
	}

	return nil
}

func (r *BookingRepository) UpdateStatus(ctx context.Context, id string, status domain.BookingStatus) error {
	logger, err := logger.FromContext(ctx)
	if err != nil {
		return fmt.Errorf("%s: %w", common.ErrUpdateBookingStatus, err)
	}

	if !isValidStatus(status) {
		return fmt.Errorf("%s: %w", common.ErrInvalidBookingStatus, errors.New("неизвестный статус бронирования"))
	}

	return r.WithTransaction(ctx, func(tx pgx.Tx) error {
		const getQuery = `
			SELECT status FROM bookings
			WHERE id = $1 FOR UPDATE
		`
		var currentStatus string
		err := tx.QueryRow(ctx, getQuery, id).Scan(&currentStatus)
		if err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				return fmt.Errorf("%s: %w", common.ErrBookingNotFound, err)
			}
			logger.Error(ctx, common.ErrGetCurrentBookingStatus,
				zap.String("bookingID", id),
				zap.Error(err))
			return fmt.Errorf("%s: %w", common.ErrGetCurrentBookingStatus, err)
		}

		query := "UPDATE bookings SET status = $2, updated_at = $3"
		args := []interface{}{id, status, time.Now()}

		switch status {
		case domain.BookingStatusPending:

		case domain.BookingStatusConfirmed:
			query += ", confirmed_at = $4"
			args = append(args, time.Now())
		case domain.BookingStatusRejected:
			query += ", rejected_at = $4"
			args = append(args, time.Now())
		case domain.BookingStatusCancelled:

		case domain.BookingStatusCompleted:
			query += ", completed_at = $4"
			args = append(args, time.Now())
		}

		query += " WHERE id = $1"

		commandTag, err := tx.Exec(ctx, query, args...)
		if err != nil {
			logger.Error(ctx, common.ErrUpdateBookingStatus,
				zap.String("bookingID", id),
				zap.String("newStatus", string(status)),
				zap.Error(err))
			return fmt.Errorf("%s: %w", common.ErrUpdateBookingStatus, err)
		}

		if commandTag.RowsAffected() == 0 {
			return fmt.Errorf("%s: %w", common.ErrBookingNotFound, errors.New("booking record not affected"))
		}

		return nil
	})
}

func (r *BookingRepository) AddAlternative(ctx context.Context, alternative *domain.BookingAlternative) error {
	log, _ := logger.FromContext(ctx)

	if alternative.ID == "" {
		alternative.ID = uuid.New().String()
	}

	const query = `
		INSERT INTO booking_alternatives (id, booking_id, date, time, message, created_at)
		VALUES ($1, $2, $3, $4, $5, $6)
	`

	executor, release, err := r.GetExecutor(ctx)
	if err != nil {
		log.Error(ctx, common.ErrGetQueryExecutor, zap.Error(err))
		return err
	}
	defer release()

	bookingExists, err := r.checkBookingExists(ctx, alternative.BookingID, executor)
	if err != nil {
		log.Error(ctx, common.ErrCheckBookingExistence,
			zap.String("bookingID", alternative.BookingID),
			zap.Error(err))
		return err
	}
	if !bookingExists {
		return fmt.Errorf("%s: %w", common.ErrBookingNotFound, err)
	}

	formattedDate := alternative.Date.Format("2006-01-02")

	_, err = executor.Exec(ctx, query,
		alternative.ID,
		alternative.BookingID,
		formattedDate,
		alternative.Time,
		alternative.Message,
		alternative.CreatedAt,
	)
	if err != nil {
		log.Error(ctx, common.ErrAddAlternativeOffer,
			zap.String("bookingID", alternative.BookingID),
			zap.String("date", formattedDate),
			zap.Error(err))
		return err
	}

	return nil
}

func (r *BookingRepository) AcceptAlternative(ctx context.Context, alternativeID string) error {
	log, _ := logger.FromContext(ctx)

	return r.WithTransaction(ctx, func(tx pgx.Tx) error {
		const getAltQuery = `
			SELECT booking_id, date, time FROM booking_alternatives
			WHERE id = $1 AND accepted_at IS NULL AND rejected_at IS NULL
			FOR UPDATE
		`
		var bookingID string
		var date time.Time
		var timeSlot string
		err := tx.QueryRow(ctx, getAltQuery, alternativeID).Scan(&bookingID, &date, &timeSlot)
		if err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				return fmt.Errorf("%s: %w", common.ErrAlternativeNotFound, err)
			}
			log.Error(ctx, common.ErrGetAlternativeInfo,
				zap.String("alternativeID", alternativeID),
				zap.Error(err))
			return err
		}

		now := time.Now()
		const updateAltQuery = `
			UPDATE booking_alternatives
			SET accepted_at = $2
			WHERE id = $1
		`
		_, err = tx.Exec(ctx, updateAltQuery, alternativeID, now)
		if err != nil {
			log.Error(ctx, common.ErrUpdateAlternativeOffer,
				zap.String("alternativeID", alternativeID),
				zap.Error(err))
			return err
		}

		const updateBookingQuery = `
			UPDATE bookings
			SET date = $2, time = $3, updated_at = $4, status = $5, confirmed_at = $6
			WHERE id = $1
		`
		_, err = tx.Exec(ctx, updateBookingQuery, bookingID, date, timeSlot, now, domain.BookingStatusConfirmed, now)
		if err != nil {
			log.Error(ctx, common.ErrUpdateBooking,
				zap.String("bookingID", bookingID),
				zap.Error(err))
			return err
		}

		return nil
	})
}

func (r *BookingRepository) RejectAlternative(ctx context.Context, alternativeID string) error {
	log, _ := logger.FromContext(ctx)
	log.Info(ctx, "rejecting booking alternative", zap.String("alternativeID", alternativeID))

	const query = `
		UPDATE booking_alternatives
		SET rejected_at = NOW()
		WHERE id = $1 AND rejected_at IS NULL AND accepted_at IS NULL
	`

	executor, release, err := r.GetExecutor(ctx)
	if err != nil {
		log.Error(ctx, common.ErrGetQueryExecutor, zap.Error(err))
		return err
	}
	defer release()

	commandTag, err := executor.Exec(ctx, query, alternativeID)
	if err != nil {
		log.Error(ctx, common.ErrUpdateAlternativeOffer,
			zap.String("alternativeID", alternativeID),
			zap.Error(err))
		return err
	}

	if commandTag.RowsAffected() == 0 {
		return fmt.Errorf("%s: %w", common.ErrAlternativeNotFound, err)
	}

	return nil
}

func (r *BookingRepository) GetAlternativeByID(ctx context.Context, alternativeID string) (*domain.BookingAlternative, error) {
	log, _ := logger.FromContext(ctx)
	log.Info(ctx, "getting booking alternative by ID", zap.String("alternativeID", alternativeID))

	const query = `
		SELECT id, booking_id, date, time, message, created_at, accepted_at, rejected_at
		FROM booking_alternatives
		WHERE id = $1
	`

	executor, release, err := r.GetExecutor(ctx)
	if err != nil {
		log.Error(ctx, common.ErrGetQueryExecutor, zap.Error(err))
		return nil, err
	}
	defer release()

	row := executor.QueryRow(ctx, query, alternativeID)

	var alt domain.BookingAlternative
	var acceptedAt, rejectedAt *time.Time

	err = row.Scan(
		&alt.ID,
		&alt.BookingID,
		&alt.Date,
		&alt.Time,
		&alt.Message,
		&alt.CreatedAt,
		&acceptedAt,
		&rejectedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, fmt.Errorf("%s: alternative offer not found", common.ErrAlternativeNotFound)
		}
		log.Error(ctx, common.ErrScanAlternativeOffer, zap.Error(err))
		return nil, fmt.Errorf("%s: %w", common.ErrGetAlternativeInfo, err)
	}

	if acceptedAt != nil {
		alt.AcceptedAt = acceptedAt
	}
	if rejectedAt != nil {
		alt.RejectedAt = rejectedAt
	}

	return &alt, nil
}

func (r *BookingRepository) getAlternatives(ctx context.Context, bookingID string, executor DBExecutor) ([]domain.BookingAlternative, error) {
	const query = `
		SELECT id, booking_id, date, time, message, created_at, accepted_at, rejected_at
		FROM booking_alternatives
		WHERE booking_id = $1
		ORDER BY created_at DESC
	`

	rows, err := executor.Query(ctx, query, bookingID)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", common.ErrQueryAlternativeOffers, err)
	}
	defer rows.Close()

	alternatives := make([]domain.BookingAlternative, 0)
	for rows.Next() {
		var alt domain.BookingAlternative
		var acceptedAt, rejectedAt *time.Time

		err = rows.Scan(
			&alt.ID,
			&alt.BookingID,
			&alt.Date,
			&alt.Time,
			&alt.Message,
			&alt.CreatedAt,
			&acceptedAt,
			&rejectedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("%s: %w", common.ErrScanAlternativeOffer, err)
		}

		if acceptedAt != nil {
			alt.AcceptedAt = acceptedAt
		}
		if rejectedAt != nil {
			alt.RejectedAt = rejectedAt
		}

		alternatives = append(alternatives, alt)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("%s: %w", common.ErrIterateAlternativeOffers, err)
	}

	return alternatives, nil
}

func (r *BookingRepository) checkRestaurantExists(ctx context.Context, id string, executor DBExecutor) (bool, error) {
	const query = `
		SELECT EXISTS(SELECT 1 FROM restaurants WHERE id = $1)
	`

	var exists bool
	err := executor.QueryRow(ctx, query, id).Scan(&exists)
	if err != nil {
		return false, err
	}

	return exists, nil
}

func (r *BookingRepository) checkUserExists(ctx context.Context, id string, executor DBExecutor) (bool, error) {
	const query = `
		SELECT EXISTS(SELECT 1 FROM users WHERE id = $1)
	`

	var exists bool
	err := executor.QueryRow(ctx, query, id).Scan(&exists)
	if err != nil {
		return false, errors.New(common.ErrCheckUserExistence)
	}

	return exists, nil
}

func (r *BookingRepository) checkBookingExists(ctx context.Context, id string, executor DBExecutor) (bool, error) {
	const query = `
		SELECT EXISTS(SELECT 1 FROM bookings WHERE id = $1)
	`

	var exists bool
	err := executor.QueryRow(ctx, query, id).Scan(&exists)
	if err != nil {
		return false, errors.New(common.ErrCheckBookingExistence)
	}

	return exists, nil
}

func isValidStatus(status domain.BookingStatus) bool {
	validStatuses := []domain.BookingStatus{
		domain.BookingStatusPending,
		domain.BookingStatusConfirmed,
		domain.BookingStatusRejected,
		domain.BookingStatusCancelled,
		domain.BookingStatusCompleted,
	}

	for _, s := range validStatuses {
		if s == status {
			return true
		}
	}

	return false
}
