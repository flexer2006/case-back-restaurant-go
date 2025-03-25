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

type RestaurantRepository struct {
	*Repository
}

func NewRestaurantRepository(repository *Repository) *RestaurantRepository {
	return &RestaurantRepository{
		Repository: repository,
	}
}

func (r *RestaurantRepository) GetByID(ctx context.Context, id string) (*domain.Restaurant, error) {
	logger, err := logger.FromContext(ctx)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", common.ErrGetRestaurant, err)
	}

	const query = `
		SELECT id, name, address, cuisine, description, created_at, updated_at, contact_email, contact_phone
		FROM restaurants
		WHERE id = $1
	`

	executor, release, err := r.GetExecutor(ctx)
	if err != nil {
		logger.Error(ctx, common.ErrGetQueryExecutor, zap.Error(err))
		return nil, fmt.Errorf("%s: %w", common.ErrGetQueryExecutor, err)
	}
	defer release()

	row := executor.QueryRow(ctx, query, id)
	var restaurant domain.Restaurant
	err = row.Scan(
		&restaurant.ID,
		&restaurant.Name,
		&restaurant.Address,
		&restaurant.Cuisine,
		&restaurant.Description,
		&restaurant.CreatedAt,
		&restaurant.UpdatedAt,
		&restaurant.ContactEmail,
		&restaurant.ContactPhone,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, errors.New(common.ErrRestaurantNotFound)
		}
		logger.Error(ctx, common.ErrScanRestaurant,
			zap.String("restaurantID", id),
			zap.Error(err))
		return nil, fmt.Errorf("%s: %w", common.ErrScanRestaurant, err)
	}

	facts, err := r.GetFacts(ctx, id)
	if err != nil {
		logger.Warn(ctx, common.ErrGetRestaurantFacts,
			zap.String("restaurantID", id),
			zap.Error(err))
	}
	restaurant.Facts = facts

	return &restaurant, nil
}

func (r *RestaurantRepository) List(ctx context.Context, offset, limit int) ([]*domain.Restaurant, error) {
	log, _ := logger.FromContext(ctx)

	const query = `
		SELECT id, name, address, cuisine, description, created_at, updated_at, contact_email, contact_phone
		FROM restaurants
		ORDER BY name
		LIMIT $1 OFFSET $2
	`

	executor, release, err := r.GetExecutor(ctx)
	if err != nil {
		log.Error(ctx, common.ErrGetQueryExecutor, zap.Error(err))
		return nil, err
	}
	defer release()

	rows, err := executor.Query(ctx, query, limit, offset)
	if err != nil {
		log.Error(ctx, common.ErrExecuteRestaurantsQuery, zap.Error(err))
		return nil, err
	}
	defer rows.Close()

	restaurants := make([]*domain.Restaurant, 0, limit)
	for rows.Next() {
		var restaurant domain.Restaurant
		err = rows.Scan(
			&restaurant.ID,
			&restaurant.Name,
			&restaurant.Address,
			&restaurant.Cuisine,
			&restaurant.Description,
			&restaurant.CreatedAt,
			&restaurant.UpdatedAt,
			&restaurant.ContactEmail,
			&restaurant.ContactPhone,
		)
		if err != nil {
			log.Error(ctx, common.ErrScanRestaurant, zap.Error(err))
			return nil, err
		}
		restaurants = append(restaurants, &restaurant)
	}

	if err = rows.Err(); err != nil {
		log.Error(ctx, common.ErrIterateRestaurants, zap.Error(err))
		return nil, err
	}

	return restaurants, nil
}

func (r *RestaurantRepository) Create(ctx context.Context, restaurant *domain.Restaurant) error {
	log, _ := logger.FromContext(ctx)

	const query = `
		INSERT INTO restaurants (id, name, address, cuisine, description, created_at, updated_at, contact_email, contact_phone)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
	`

	if restaurant.ID == "" {
		restaurant.ID = uuid.New().String()
	}

	now := time.Now()
	restaurant.CreatedAt = now
	restaurant.UpdatedAt = now

	executor, release, err := r.GetExecutor(ctx)
	if err != nil {
		log.Error(ctx, common.ErrGetQueryExecutor, zap.Error(err))
		return err
	}
	defer release()

	_, err = executor.Exec(ctx, query,
		restaurant.ID,
		restaurant.Name,
		restaurant.Address,
		restaurant.Cuisine,
		restaurant.Description,
		restaurant.CreatedAt,
		restaurant.UpdatedAt,
		restaurant.ContactEmail,
		restaurant.ContactPhone,
	)
	if err != nil {
		log.Error(ctx, common.ErrCreateRestaurant,
			zap.String("name", restaurant.Name),
			zap.Error(err))
		return err
	}

	return nil
}

func (r *RestaurantRepository) Update(ctx context.Context, restaurant *domain.Restaurant) error {
	log, _ := logger.FromContext(ctx)

	const query = `
		UPDATE restaurants
		SET name = $2, address = $3, cuisine = $4, description = $5, updated_at = $6, contact_email = $7, contact_phone = $8
		WHERE id = $1
	`

	restaurant.UpdatedAt = time.Now()

	executor, release, err := r.GetExecutor(ctx)
	if err != nil {
		log.Error(ctx, common.ErrGetQueryExecutor, zap.Error(err))
		return err
	}
	defer release()

	commandTag, err := executor.Exec(ctx, query,
		restaurant.ID,
		restaurant.Name,
		restaurant.Address,
		restaurant.Cuisine,
		restaurant.Description,
		restaurant.UpdatedAt,
		restaurant.ContactEmail,
		restaurant.ContactPhone,
	)
	if err != nil {
		log.Error(ctx, common.ErrUpdateRestaurant,
			zap.String("restaurantID", restaurant.ID),
			zap.Error(err))
		return err
	}

	if commandTag.RowsAffected() == 0 {
		return errors.New(common.ErrRestaurantNotFound)
	}

	return nil
}

func (r *RestaurantRepository) Delete(ctx context.Context, id string) error {
	log, _ := logger.FromContext(ctx)

	const query = `
		DELETE FROM restaurants
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
		log.Error(ctx, common.ErrDeleteRestaurant,
			zap.String("restaurantID", id),
			zap.Error(err))
		return err
	}

	if commandTag.RowsAffected() == 0 {
		return errors.New(common.ErrRestaurantNotFound)
	}

	return nil
}

func (r *RestaurantRepository) AddFact(ctx context.Context, restaurantID string, fact domain.Fact) (*domain.Fact, error) {
	log, _ := logger.FromContext(ctx)

	const query = `
		INSERT INTO facts (id, restaurant_id, content, created_at)
		VALUES ($1, $2, $3, $4)
	`

	if fact.ID == "" {
		fact.ID = uuid.New().String()
	}

	if fact.CreatedAt.IsZero() {
		fact.CreatedAt = time.Now()
	}

	executor, release, err := r.GetExecutor(ctx)
	if err != nil {
		log.Error(ctx, common.ErrGetQueryExecutor, zap.Error(err))
		return nil, err
	}
	defer release()

	exist, err := r.checkRestaurantExists(ctx, restaurantID, executor)
	if err != nil {
		log.Error(ctx, common.ErrCheckRestaurantExistence,
			zap.String("restaurantID", restaurantID),
			zap.Error(err))
		return nil, err
	}
	if !exist {
		return nil, errors.New(common.ErrRestaurantNotFound)
	}

	_, err = executor.Exec(ctx, query,
		fact.ID,
		restaurantID,
		fact.Content,
		fact.CreatedAt,
	)
	if err != nil {
		log.Error(ctx, common.ErrAddRestaurantFact,
			zap.String("restaurantID", restaurantID),
			zap.Error(err))
		return nil, err
	}

	return &fact, nil
}

func (r *RestaurantRepository) GetFacts(ctx context.Context, restaurantID string) ([]domain.Fact, error) {
	log, _ := logger.FromContext(ctx)

	const query = `
		SELECT id, restaurant_id, content, created_at
		FROM facts
		WHERE restaurant_id = $1
		ORDER BY created_at DESC
	`

	executor, release, err := r.GetExecutor(ctx)
	if err != nil {
		log.Error(ctx, common.ErrGetQueryExecutor, zap.Error(err))
		return nil, err
	}
	defer release()

	rows, err := executor.Query(ctx, query, restaurantID)
	if err != nil {
		log.Error(ctx, common.ErrExecuteFactsQuery,
			zap.String("restaurantID", restaurantID),
			zap.Error(err))
		return nil, err
	}
	defer rows.Close()

	facts := make([]domain.Fact, 0)
	for rows.Next() {
		var fact domain.Fact
		err = rows.Scan(
			&fact.ID,
			&fact.RestaurantID,
			&fact.Content,
			&fact.CreatedAt,
		)
		if err != nil {
			log.Error(ctx, common.ErrScanFact, zap.Error(err))
			return nil, err
		}
		facts = append(facts, fact)
	}

	if err = rows.Err(); err != nil {
		log.Error(ctx, common.ErrIterateFacts, zap.Error(err))
		return nil, err
	}

	return facts, nil
}

func (r *RestaurantRepository) GetRandomFacts(ctx context.Context, count int) ([]domain.Fact, error) {
	log, _ := logger.FromContext(ctx)

	const query = `
		SELECT f.id, f.restaurant_id, f.content, f.created_at
		FROM facts f
		JOIN restaurants r ON f.restaurant_id = r.id
		ORDER BY RANDOM()
		LIMIT $1
	`

	executor, release, err := r.GetExecutor(ctx)
	if err != nil {
		log.Error(ctx, common.ErrGetQueryExecutor, zap.Error(err))
		return nil, err
	}
	defer release()

	rows, err := executor.Query(ctx, query, count)
	if err != nil {
		log.Error(ctx, common.ErrExecuteRandomFactsQuery,
			zap.Int("count", count),
			zap.Error(err))
		return nil, err
	}
	defer rows.Close()

	facts := make([]domain.Fact, 0, count)
	for rows.Next() {
		var fact domain.Fact
		err = rows.Scan(
			&fact.ID,
			&fact.RestaurantID,
			&fact.Content,
			&fact.CreatedAt,
		)
		if err != nil {
			log.Error(ctx, common.ErrScanFact, zap.Error(err))
			return nil, err
		}
		facts = append(facts, fact)
	}

	if err = rows.Err(); err != nil {
		log.Error(ctx, common.ErrIterateRandomFacts, zap.Error(err))
		return nil, err
	}

	return facts, nil
}

func (r *RestaurantRepository) checkRestaurantExists(ctx context.Context, id string, executor DBExecutor) (bool, error) {
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
