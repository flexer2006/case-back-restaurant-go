package usecase_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/flexer2006/case-back-restaurant-go/internal/domain"
	"github.com/flexer2006/case-back-restaurant-go/pkg/usecase"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

var _ = newTestContext

type MockRestaurantRepository struct {
	mock.Mock
}

func (m *MockRestaurantRepository) GetByID(ctx context.Context, id string) (*domain.Restaurant, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Restaurant), args.Error(1)
}

func (m *MockRestaurantRepository) List(ctx context.Context, offset, limit int) ([]*domain.Restaurant, error) {
	args := m.Called(ctx, offset, limit)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*domain.Restaurant), args.Error(1)
}

func (m *MockRestaurantRepository) Create(ctx context.Context, restaurant *domain.Restaurant) error {
	args := m.Called(ctx, restaurant)
	return args.Error(0)
}

func (m *MockRestaurantRepository) Update(ctx context.Context, restaurant *domain.Restaurant) error {
	args := m.Called(ctx, restaurant)
	return args.Error(0)
}

func (m *MockRestaurantRepository) Delete(ctx context.Context, id string) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockRestaurantRepository) AddFact(ctx context.Context, restaurantID string, fact domain.Fact) (*domain.Fact, error) {
	args := m.Called(ctx, restaurantID, fact)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Fact), args.Error(1)
}

func (m *MockRestaurantRepository) GetFacts(ctx context.Context, restaurantID string) ([]domain.Fact, error) {
	args := m.Called(ctx, restaurantID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]domain.Fact), args.Error(1)
}

func (m *MockRestaurantRepository) GetRandomFacts(ctx context.Context, count int) ([]domain.Fact, error) {
	args := m.Called(ctx, count)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]domain.Fact), args.Error(1)
}

func TestGetRandomFacts(t *testing.T) {
	testCases := []struct {
		name          string
		count         int
		expectedCount int
		mockFacts     []domain.Fact
		mockError     error
	}{
		{
			name:          "successful random facts retrieval",
			count:         3,
			expectedCount: 3,
			mockFacts: []domain.Fact{
				{ID: "1", RestaurantID: "r1", Content: "Fact 1", CreatedAt: time.Now()},
				{ID: "2", RestaurantID: "r2", Content: "Fact 2", CreatedAt: time.Now()},
				{ID: "3", RestaurantID: "r3", Content: "Fact 3", CreatedAt: time.Now()},
			},
			mockError: nil,
		},
		{
			name:          "negative fact count should return 3 facts",
			count:         -5,
			expectedCount: 3,
			mockFacts: []domain.Fact{
				{ID: "1", RestaurantID: "r1", Content: "Fact 1", CreatedAt: time.Now()},
				{ID: "2", RestaurantID: "r2", Content: "Fact 2", CreatedAt: time.Now()},
				{ID: "3", RestaurantID: "r3", Content: "Fact 3", CreatedAt: time.Now()},
			},
			mockError: nil,
		},
		{
			name:          "too large fact count should return maximum of 10",
			count:         15,
			expectedCount: 10,
			mockFacts: []domain.Fact{
				{ID: "1", RestaurantID: "r1", Content: "Fact 1", CreatedAt: time.Now()},
				{ID: "2", RestaurantID: "r2", Content: "Fact 2", CreatedAt: time.Now()},
				{ID: "3", RestaurantID: "r3", Content: "Fact 3", CreatedAt: time.Now()},
				{ID: "4", RestaurantID: "r4", Content: "Fact 4", CreatedAt: time.Now()},
				{ID: "5", RestaurantID: "r5", Content: "Fact 5", CreatedAt: time.Now()},
				{ID: "6", RestaurantID: "r6", Content: "Fact 6", CreatedAt: time.Now()},
				{ID: "7", RestaurantID: "r7", Content: "Fact 7", CreatedAt: time.Now()},
				{ID: "8", RestaurantID: "r8", Content: "Fact 8", CreatedAt: time.Now()},
				{ID: "9", RestaurantID: "r9", Content: "Fact 9", CreatedAt: time.Now()},
				{ID: "10", RestaurantID: "r10", Content: "Fact 10", CreatedAt: time.Now()},
			},
			mockError: nil,
		},
		{
			name:          "repository error",
			count:         3,
			expectedCount: 3,
			mockFacts:     nil,
			mockError:     errors.New("database error"),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctx := newTestContext()
			mockRepo := new(MockRestaurantRepository)

			expectedRepoCount := tc.expectedCount
			mockRepo.On("GetRandomFacts", ctx, expectedRepoCount).Return(tc.mockFacts, tc.mockError)

			factsUC := usecase.NewFactsUseCase(mockRepo)

			facts, err := factsUC.GetRandomFacts(ctx, tc.count)

			if tc.mockError != nil {
				assert.Error(t, err)
				assert.Equal(t, tc.mockError, err)
				assert.Nil(t, facts)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tc.mockFacts, facts)
				assert.Len(t, facts, len(tc.mockFacts))
			}

			mockRepo.AssertExpectations(t)
		})
	}
}

func TestGetRestaurantFacts(t *testing.T) {
	testCases := []struct {
		name         string
		restaurantID string
		mockFacts    []domain.Fact
		mockError    error
	}{
		{
			name:         "successful restaurant facts retrieval",
			restaurantID: "resto123",
			mockFacts: []domain.Fact{
				{ID: "1", RestaurantID: "resto123", Content: "Fact 1", CreatedAt: time.Now()},
				{ID: "2", RestaurantID: "resto123", Content: "Fact 2", CreatedAt: time.Now()},
				{ID: "3", RestaurantID: "resto123", Content: "Fact 3", CreatedAt: time.Now()},
			},
			mockError: nil,
		},
		{
			name:         "restaurant without facts",
			restaurantID: "resto456",
			mockFacts:    []domain.Fact{},
			mockError:    nil,
		},
		{
			name:         "repository error",
			restaurantID: "resto789",
			mockFacts:    nil,
			mockError:    errors.New("restaurant not found"),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctx := newTestContext()
			mockRepo := new(MockRestaurantRepository)

			mockRepo.On("GetFacts", ctx, tc.restaurantID).Return(tc.mockFacts, tc.mockError)

			factsUC := usecase.NewFactsUseCase(mockRepo)

			facts, err := factsUC.GetRestaurantFacts(ctx, tc.restaurantID)

			if tc.mockError != nil {
				assert.Error(t, err)
				assert.Equal(t, tc.mockError, err)
				assert.Nil(t, facts)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tc.mockFacts, facts)
				assert.Len(t, facts, len(tc.mockFacts))
			}

			mockRepo.AssertExpectations(t)
		})
	}
}
