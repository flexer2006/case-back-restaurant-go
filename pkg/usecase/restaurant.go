package usecase

import (
	"context"
	"time"

	"github.com/flexer2006/case-back-restaurant-go/internal/domain"
	"github.com/flexer2006/case-back-restaurant-go/internal/logger"
	"github.com/flexer2006/case-back-restaurant-go/internal/repository"

	"go.uber.org/zap"
)

type RestaurantUseCase interface {
	GetRestaurant(ctx context.Context, id string) (*domain.Restaurant, error)

	ListRestaurants(ctx context.Context, offset, limit int) ([]*domain.Restaurant, error)

	CreateRestaurant(ctx context.Context, restaurant *domain.Restaurant) (string, error)

	UpdateRestaurant(ctx context.Context, restaurant *domain.Restaurant) error

	DeleteRestaurant(ctx context.Context, id string) error

	AddFact(ctx context.Context, restaurantID string, content string) (*domain.Fact, error)

	GetFacts(ctx context.Context, restaurantID string) ([]domain.Fact, error)

	GetRandomFacts(ctx context.Context, count int) ([]domain.Fact, error)

	SetWorkingHours(ctx context.Context, restaurantID string, workingHours *domain.WorkingHours) error

	GetWorkingHours(ctx context.Context, restaurantID string) ([]*domain.WorkingHours, error)
}

type restaurantUseCase struct {
	restaurantRepo   repository.RestaurantRepository
	workingHoursRepo repository.WorkingHoursRepository
}

func NewRestaurantUseCase(
	restaurantRepo repository.RestaurantRepository,
	workingHoursRepo repository.WorkingHoursRepository,
) RestaurantUseCase {
	return &restaurantUseCase{
		restaurantRepo:   restaurantRepo,
		workingHoursRepo: workingHoursRepo,
	}
}

func (u *restaurantUseCase) GetRestaurant(ctx context.Context, id string) (*domain.Restaurant, error) {
	return u.restaurantRepo.GetByID(ctx, id)
}

func (u *restaurantUseCase) ListRestaurants(ctx context.Context, offset, limit int) ([]*domain.Restaurant, error) {
	return u.restaurantRepo.List(ctx, offset, limit)
}

func (u *restaurantUseCase) CreateRestaurant(ctx context.Context, restaurant *domain.Restaurant) (string, error) {
	log, _ := logger.FromContext(ctx)
	log.Info(ctx, "creating new restaurant",
		zap.String("name", restaurant.Name),
		zap.String("address", restaurant.Address),
		zap.String("cuisine", string(restaurant.Cuisine)))

	now := time.Now()
	restaurant.CreatedAt = now
	restaurant.UpdatedAt = now
	if err := u.restaurantRepo.Create(ctx, restaurant); err != nil {
		log.Error(ctx, "failed to create restaurant",
			zap.String("name", restaurant.Name),
			zap.Error(err))
		return "", err
	}

	log.Info(ctx, "restaurant successfully created", zap.String("restaurantID", restaurant.ID))
	return restaurant.ID, nil
}

func (u *restaurantUseCase) UpdateRestaurant(ctx context.Context, restaurant *domain.Restaurant) error {
	log, _ := logger.FromContext(ctx)
	log.Info(ctx, "updating restaurant",
		zap.String("restaurantID", restaurant.ID),
		zap.String("name", restaurant.Name))

	restaurant.UpdatedAt = time.Now()

	if err := u.restaurantRepo.Update(ctx, restaurant); err != nil {
		log.Error(ctx, "failed to update restaurant",
			zap.String("restaurantID", restaurant.ID),
			zap.Error(err))
		return err
	}

	log.Info(ctx, "restaurant successfully updated", zap.String("restaurantID", restaurant.ID))
	return nil
}

func (u *restaurantUseCase) DeleteRestaurant(ctx context.Context, id string) error {
	log, _ := logger.FromContext(ctx)
	log.Info(ctx, "deleting restaurant", zap.String("restaurantID", id))

	if err := u.restaurantRepo.Delete(ctx, id); err != nil {
		log.Error(ctx, "failed to delete restaurant",
			zap.String("restaurantID", id),
			zap.Error(err))
		return err
	}

	log.Info(ctx, "restaurant successfully deleted", zap.String("restaurantID", id))
	return nil
}

func (u *restaurantUseCase) AddFact(ctx context.Context, restaurantID string, content string) (*domain.Fact, error) {
	log, _ := logger.FromContext(ctx)
	log.Info(ctx, "adding restaurant fact",
		zap.String("restaurantID", restaurantID))

	fact := domain.Fact{
		RestaurantID: restaurantID,
		Content:      content,
		CreatedAt:    time.Now(),
	}

	createdFact, err := u.restaurantRepo.AddFact(ctx, restaurantID, fact)
	if err != nil {
		log.Error(ctx, "failed to add restaurant fact",
			zap.String("restaurantID", restaurantID),
			zap.Error(err))
		return nil, err
	}

	log.Info(ctx, "restaurant fact successfully added",
		zap.String("factID", createdFact.ID),
		zap.String("restaurantID", restaurantID))
	return createdFact, nil
}

func (u *restaurantUseCase) GetFacts(ctx context.Context, restaurantID string) ([]domain.Fact, error) {
	return u.restaurantRepo.GetFacts(ctx, restaurantID)
}

func (u *restaurantUseCase) GetRandomFacts(ctx context.Context, count int) ([]domain.Fact, error) {
	return u.restaurantRepo.GetRandomFacts(ctx, count)
}

func (u *restaurantUseCase) SetWorkingHours(ctx context.Context, restaurantID string, workingHours *domain.WorkingHours) error {
	log, _ := logger.FromContext(ctx)
	log.Info(ctx, "setting restaurant working hours",
		zap.String("restaurantID", restaurantID),
		zap.Int("weekDay", int(workingHours.WeekDay)),
		zap.String("openTime", workingHours.OpenTime),
		zap.String("closeTime", workingHours.CloseTime))

	workingHours.RestaurantID = restaurantID
	if err := u.workingHoursRepo.SetWorkingHours(ctx, workingHours); err != nil {
		log.Error(ctx, "failed to set restaurant working hours",
			zap.String("restaurantID", restaurantID),
			zap.Error(err))
		return err
	}

	log.Info(ctx, "restaurant working hours successfully set",
		zap.String("restaurantID", restaurantID),
		zap.Int("weekDay", int(workingHours.WeekDay)))
	return nil
}

func (u *restaurantUseCase) GetWorkingHours(ctx context.Context, restaurantID string) ([]*domain.WorkingHours, error) {
	return u.workingHoursRepo.GetByRestaurantID(ctx, restaurantID)
}
