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
	ErrWorkingHoursNotFound = errors.New(common.ErrWorkingHoursNotFound)
)

type WorkingHoursRepository struct {
	*Repository
}

func NewWorkingHoursRepository(repository *Repository) *WorkingHoursRepository {
	return &WorkingHoursRepository{
		Repository: repository,
	}
}

func (r *WorkingHoursRepository) GetByRestaurantID(ctx context.Context, restaurantID string) ([]*domain.WorkingHours, error) {
	logger, err := logger.FromContext(ctx)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", common.ErrExecuteWorkingHoursQuery, err)
	}

	const query = `
		SELECT id, restaurant_id, week_day, open_time, close_time, valid_from, valid_to
		FROM working_hours
		WHERE restaurant_id = $1 AND (valid_to IS NULL OR valid_to > CURRENT_DATE)
		ORDER BY week_day, open_time
	`

	executor, release, err := r.GetExecutor(ctx)
	if err != nil {
		logger.Error(ctx, common.ErrGetQueryExecutor, zap.Error(err))
		return nil, fmt.Errorf("%s: %w", common.ErrGetQueryExecutor, err)
	}
	defer release()

	rows, err := executor.Query(ctx, query, restaurantID)
	if err != nil {
		logger.Error(ctx, common.ErrExecuteWorkingHoursQuery, zap.Error(err))
		return nil, fmt.Errorf("%s: %w", common.ErrExecuteWorkingHoursQuery, err)
	}
	defer rows.Close()

	hours := make([]*domain.WorkingHours, 0)
	for rows.Next() {
		var h domain.WorkingHours
		var validTo *time.Time

		err = rows.Scan(
			&h.ID,
			&h.RestaurantID,
			&h.WeekDay,
			&h.OpenTime,
			&h.CloseTime,
			&h.ValidFrom,
			&validTo,
		)
		if err != nil {
			logger.Error(ctx, common.ErrScanWorkingHours, zap.Error(err))
			return nil, fmt.Errorf("%s: %w", common.ErrScanWorkingHours, err)
		}

		if validTo != nil {
			h.ValidTo = *validTo
		}

		hours = append(hours, &h)
	}

	if err = rows.Err(); err != nil {
		logger.Error(ctx, common.ErrIterateWorkingHours, zap.Error(err))
		return nil, fmt.Errorf("%s: %w", common.ErrIterateWorkingHours, err)
	}

	return hours, nil
}

func (r *WorkingHoursRepository) SetWorkingHours(ctx context.Context, hours *domain.WorkingHours) error {
	log, _ := logger.FromContext(ctx)

	if hours.ID == "" {
		hours.ID = uuid.New().String()
	}

	const checkQuery = `
		SELECT id FROM working_hours
		WHERE restaurant_id = $1 AND week_day = $2 AND
		      (valid_to IS NULL OR valid_to > $3) AND valid_from <= $3
	`

	executor, release, err := r.GetExecutor(ctx)
	if err != nil {
		log.Error(ctx, common.ErrGetQueryExecutor, zap.Error(err))
		return err
	}
	defer release()

	var existingID string
	now := time.Now()
	err = executor.QueryRow(ctx, checkQuery, hours.RestaurantID, hours.WeekDay, now).Scan(&existingID)
	if err != nil && !errors.Is(err, pgx.ErrNoRows) {
		log.Error(ctx, common.ErrCheckWorkingHoursExistence,
			zap.String("restaurantID", hours.RestaurantID),
			zap.Int("weekDay", int(hours.WeekDay)),
			zap.Error(err))
		return err
	}

	exist, err := r.checkRestaurantExists(ctx, hours.RestaurantID, executor)
	if err != nil {
		log.Error(ctx, common.ErrCheckRestaurantExistence,
			zap.String("restaurantID", hours.RestaurantID),
			zap.Error(err))
		return err
	}
	if !exist {
		return errors.New(common.ErrRestaurantNotFound)
	}

	return r.WithTransaction(ctx, func(tx pgx.Tx) error {
		if existingID != "" {
			const updateQuery = `
				UPDATE working_hours
				SET valid_to = $2
				WHERE id = $1
			`
			_, err := tx.Exec(ctx, updateQuery, existingID, now)
			if err != nil {
				log.Error(ctx, common.ErrTerminateWorkingHours,
					zap.String("id", existingID),
					zap.Error(err))
				return err
			}
		}

		const insertQuery = `
			INSERT INTO working_hours (id, restaurant_id, week_day, open_time, close_time, is_closed, valid_from, valid_to)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		`
		var validTo *time.Time
		if !hours.ValidTo.IsZero() {
			validTo = &hours.ValidTo
		}

		_, err := tx.Exec(ctx, insertQuery,
			hours.ID,
			hours.RestaurantID,
			hours.WeekDay,
			hours.OpenTime,
			hours.CloseTime,
			hours.IsClosed,
			hours.ValidFrom,
			validTo,
		)
		if err != nil {
			log.Error(ctx, common.ErrInsertWorkingHours,
				zap.String("restaurantID", hours.RestaurantID),
				zap.Int("weekDay", int(hours.WeekDay)),
				zap.Error(err))
			return err
		}

		return nil
	})
}

func (r *WorkingHoursRepository) DeleteWorkingHours(ctx context.Context, id string) error {
	log, _ := logger.FromContext(ctx)

	const query = `
		DELETE FROM working_hours
		WHERE id = $1
	`

	executor, release, err := r.GetExecutor(ctx)
	if err != nil {
		log.Error(ctx, common.ErrGetQueryExecutor, zap.Error(err))
		return err
	}
	defer release()

	commandTag, err := executor.Exec(ctx, query, id)
	if err != nil {
		log.Error(ctx, common.ErrDeleteWorkingHours,
			zap.String("id", id),
			zap.Error(err))
		return err
	}

	if commandTag.RowsAffected() == 0 {
		return ErrWorkingHoursNotFound
	}

	return nil
}

func (r *WorkingHoursRepository) checkRestaurantExists(ctx context.Context, id string, executor DBExecutor) (bool, error) {
	const query = `
		SELECT EXISTS(SELECT 1 FROM restaurants WHERE id = $1)
	`

	var exists bool
	err := executor.QueryRow(ctx, query, id).Scan(&exists)
	if err != nil {
		return false, errors.New(common.ErrCheckRestaurantExistence)
	}

	return exists, nil
}
