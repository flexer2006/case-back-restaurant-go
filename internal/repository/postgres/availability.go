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

var (
	ErrAvailabilityNotFound = errors.New(common.ErrAvailabilityNotFound)
	ErrInsufficientCapacity = errors.New(common.ErrInsufficientCapacity)
)

type AvailabilityRepository struct {
	*Repository
}

func NewAvailabilityRepository(repository *Repository) *AvailabilityRepository {
	return &AvailabilityRepository{
		Repository: repository,
	}
}

func (r *AvailabilityRepository) GetByRestaurantAndDate(ctx context.Context, restaurantID string, date time.Time) ([]*domain.Availability, error) {
	logger, err := logger.FromContext(ctx)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", common.ErrExecuteAvailabilityQuery, err)
	}

	const query = `
		SELECT id, restaurant_id, date, time_slot, capacity, reserved
		FROM availability
		WHERE restaurant_id = $1 AND date = $2
		ORDER BY time_slot
	`

	executor, release, err := r.GetExecutor(ctx)
	if err != nil {
		logger.Error(ctx, common.ErrGetQueryExecutor, zap.Error(err))
		return nil, fmt.Errorf("%s: %w", common.ErrGetQueryExecutor, err)
	}
	defer release()

	formattedDate := date.Format("2006-01-02")
	rows, err := executor.Query(ctx, query, restaurantID, formattedDate)
	if err != nil {
		logger.Error(ctx, common.ErrExecuteAvailabilityQuery, zap.Error(err))
		return nil, fmt.Errorf("%s: %w", common.ErrExecuteAvailabilityQuery, err)
	}
	defer rows.Close()

	availabilities := make([]*domain.Availability, 0)
	for rows.Next() {
		var a domain.Availability
		err = rows.Scan(
			&a.ID,
			&a.RestaurantID,
			&a.Date,
			&a.TimeSlot,
			&a.Capacity,
			&a.Reserved,
		)
		if err != nil {
			logger.Error(ctx, common.ErrScanAvailability, zap.Error(err))
			return nil, fmt.Errorf("%s: %w", common.ErrScanAvailability, err)
		}

		availabilities = append(availabilities, &a)
	}

	if err = rows.Err(); err != nil {
		logger.Error(ctx, common.ErrIterateAvailability, zap.Error(err))
		return nil, fmt.Errorf("%s: %w", common.ErrIterateAvailability, err)
	}

	return availabilities, nil
}

func (r *AvailabilityRepository) SetAvailability(ctx context.Context, availability *domain.Availability) error {
	log, _ := logger.FromContext(ctx)

	if availability.ID == "" {
		availability.ID = uuid.New().String()
	}

	executor, release, err := r.GetExecutor(ctx)
	if err != nil {
		log.Error(ctx, common.ErrGetQueryExecutor, zap.Error(err))
		return err
	}
	defer release()

	exist, err := r.checkRestaurantExists(ctx, availability.RestaurantID, executor)
	if err != nil {
		log.Error(ctx, common.ErrCheckRestaurantExistence,
			zap.String("restaurantID", availability.RestaurantID),
			zap.Error(err))
		return err
	}
	if !exist {
		return errors.New(common.ErrRestaurantNotFound)
	}

	const checkQuery = `
		SELECT id, reserved FROM availability
		WHERE restaurant_id = $1 AND date = $2 AND time_slot = $3
	`

	formattedDate := availability.Date.Format("2006-01-02")
	var existingID string
	var reserved int
	err = executor.QueryRow(ctx, checkQuery, availability.RestaurantID, formattedDate, availability.TimeSlot).Scan(&existingID, &reserved)

	availability.UpdatedAt = time.Now()

	if err == nil {
		const updateQuery = `
			UPDATE availability
			SET capacity = $2, updated_at = $3
			WHERE id = $1
		`

		_, err = executor.Exec(ctx, updateQuery, existingID, availability.Capacity, availability.UpdatedAt)
		if err != nil {
			log.Error(ctx, common.ErrUpdateAvailability,
				zap.String("id", existingID),
				zap.Int("capacity", availability.Capacity),
				zap.Error(err))
			return err
		}

		availability.ID = existingID
		availability.Reserved = reserved

		return nil
	}

	if !errors.Is(err, pgx.ErrNoRows) {
		log.Error(ctx, common.ErrCheckAvailabilityExistence,
			zap.String("restaurantID", availability.RestaurantID),
			zap.String("date", formattedDate),
			zap.String("timeSlot", availability.TimeSlot),
			zap.Error(err))
		return err
	}

	const insertQuery = `
		INSERT INTO availability (id, restaurant_id, date, time_slot, capacity, reserved, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
	`

	_, err = executor.Exec(ctx, insertQuery,
		availability.ID,
		availability.RestaurantID,
		formattedDate,
		availability.TimeSlot,
		availability.Capacity,
		0,
		availability.UpdatedAt,
	)
	if err != nil {
		log.Error(ctx, common.ErrInsertAvailability,
			zap.String("restaurantID", availability.RestaurantID),
			zap.String("date", formattedDate),
			zap.String("timeSlot", availability.TimeSlot),
			zap.Error(err))
		return err
	}

	availability.Reserved = 0
	return nil
}

func (r *AvailabilityRepository) UpdateReservedSeats(ctx context.Context, availabilityID string, delta int) error {
	log, _ := logger.FromContext(ctx)

	return r.WithTransaction(ctx, func(tx pgx.Tx) error {
		const getQuery = `
			SELECT capacity, reserved FROM availability
			WHERE id = $1 FOR UPDATE
		`
		var capacity, reserved int
		err := tx.QueryRow(ctx, getQuery, availabilityID).Scan(&capacity, &reserved)
		if err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				return fmt.Errorf("%s: %w", common.ErrAvailabilityNotFound, err)
			}
			log.Error(ctx, common.ErrGetCurrentAvailability,
				zap.String("availabilityID", availabilityID),
				zap.Error(err))
			return err
		}

		newReserved := reserved + delta
		if newReserved > capacity {
			return fmt.Errorf("%s: %w", common.ErrInsufficientCapacity, err)
		}
		if newReserved < 0 {
			newReserved = 0
		}

		const updateQuery = `
			UPDATE availability
			SET reserved = $2, updated_at = $3
			WHERE id = $1
		`
		now := time.Now()
		_, err = tx.Exec(ctx, updateQuery, availabilityID, newReserved, now)
		if err != nil {
			log.Error(ctx, common.ErrUpdateReservedSeats,
				zap.String("availabilityID", availabilityID),
				zap.Int("newReserved", newReserved),
				zap.Error(err))
			return err
		}

		return nil
	})
}

func (r *AvailabilityRepository) checkRestaurantExists(ctx context.Context, id string, executor DBExecutor) (bool, error) {
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
