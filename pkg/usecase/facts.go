package usecase

import (
	"context"

	"github.com/flexer2006/case-back-restaurant-go/internal/domain"
	"github.com/flexer2006/case-back-restaurant-go/internal/logger"
	"github.com/flexer2006/case-back-restaurant-go/internal/repository"

	"go.uber.org/zap"
)

type FactsUseCase interface {
	GetRandomFacts(ctx context.Context, count int) ([]domain.Fact, error)

	GetRestaurantFacts(ctx context.Context, restaurantID string) ([]domain.Fact, error)
}

type factsUseCase struct {
	restaurantRepo repository.RestaurantRepository
}

func NewFactsUseCase(restaurantRepo repository.RestaurantRepository) FactsUseCase {
	return &factsUseCase{
		restaurantRepo: restaurantRepo,
	}
}

func (u *factsUseCase) GetRandomFacts(ctx context.Context, count int) ([]domain.Fact, error) {
	log, _ := logger.FromContext(ctx)
	log.Info(ctx, "getting random facts", zap.Int("count", count))

	if count <= 0 {
		count = 3
	}
	if count > 10 {
		count = 10
	}

	facts, err := u.restaurantRepo.GetRandomFacts(ctx, count)
	if err != nil {
		log.Error(ctx, "failed to get random facts", zap.Error(err))
		return nil, err
	}

	log.Info(ctx, "retrieved random facts", zap.Int("count", len(facts)))
	return facts, nil
}

func (u *factsUseCase) GetRestaurantFacts(ctx context.Context, restaurantID string) ([]domain.Fact, error) {
	return u.restaurantRepo.GetFacts(ctx, restaurantID)
}
