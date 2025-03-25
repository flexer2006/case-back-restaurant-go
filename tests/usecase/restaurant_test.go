package usecase_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/flexer2006/case-back-restaurant-go/internal/domain"
	"github.com/flexer2006/case-back-restaurant-go/pkg/usecase"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func createTestRestaurant() *domain.Restaurant {
	return &domain.Restaurant{
		ID:           uuid.New().String(),
		Name:         "test restaurant",
		Address:      "test street, 123",
		Cuisine:      domain.Cuisine("italian"),
		Description:  "test restaurant description",
		ContactEmail: "test@restaurant.com",
		ContactPhone: "+7 (123) 456-78-90",
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}
}

type MockWorkingHoursRepository struct {
	mock.Mock
}

func (m *MockWorkingHoursRepository) GetByRestaurantID(ctx context.Context, restaurantID string) ([]*domain.WorkingHours, error) {
	args := m.Called(ctx, restaurantID)
	return args.Get(0).([]*domain.WorkingHours), args.Error(1)
}

func (m *MockWorkingHoursRepository) SetWorkingHours(ctx context.Context, hours *domain.WorkingHours) error {
	args := m.Called(ctx, hours)
	if hours.ID == "" {
		hours.ID = uuid.New().String()
	}
	return args.Error(0)
}

func (m *MockWorkingHoursRepository) DeleteWorkingHours(ctx context.Context, id string) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func TestRestaurantUseCase_GetRestaurant(t *testing.T) {

	ctx := newTestContext()
	mockRestaurantRepo := new(MockRestaurantRepository)
	mockWorkingHoursRepo := new(MockWorkingHoursRepository)

	useCase := usecase.NewRestaurantUseCase(mockRestaurantRepo, mockWorkingHoursRepo)

	restaurantID := uuid.New().String()
	expectedRestaurant := createTestRestaurant()
	expectedRestaurant.ID = restaurantID

	mockRestaurantRepo.On("GetByID", ctx, restaurantID).Return(expectedRestaurant, nil)

	result, err := useCase.GetRestaurant(ctx, restaurantID)

	assert.NoError(t, err)
	assert.Equal(t, expectedRestaurant, result)
	mockRestaurantRepo.AssertExpectations(t)
}

func TestRestaurantUseCase_GetRestaurantNotFound(t *testing.T) {

	ctx := newTestContext()
	mockRestaurantRepo := new(MockRestaurantRepository)
	mockWorkingHoursRepo := new(MockWorkingHoursRepository)

	useCase := usecase.NewRestaurantUseCase(mockRestaurantRepo, mockWorkingHoursRepo)

	restaurantID := uuid.New().String()
	expectedError := errors.New("restaurant not found")

	mockRestaurantRepo.On("GetByID", ctx, restaurantID).Return(nil, expectedError)

	result, err := useCase.GetRestaurant(ctx, restaurantID)

	assert.Error(t, err)
	assert.Equal(t, expectedError, err)
	assert.Nil(t, result)
	mockRestaurantRepo.AssertExpectations(t)
}

func TestRestaurantUseCase_ListRestaurants(t *testing.T) {

	ctx := newTestContext()
	mockRestaurantRepo := new(MockRestaurantRepository)
	mockWorkingHoursRepo := new(MockWorkingHoursRepository)

	useCase := usecase.NewRestaurantUseCase(mockRestaurantRepo, mockWorkingHoursRepo)

	offset, limit := 0, 10
	expectedRestaurants := []*domain.Restaurant{
		createTestRestaurant(),
		createTestRestaurant(),
	}

	mockRestaurantRepo.On("List", ctx, offset, limit).Return(expectedRestaurants, nil)

	result, err := useCase.ListRestaurants(ctx, offset, limit)

	assert.NoError(t, err)
	assert.Equal(t, expectedRestaurants, result)
	assert.Len(t, result, 2)
	mockRestaurantRepo.AssertExpectations(t)
}

func TestRestaurantUseCase_CreateRestaurant(t *testing.T) {

	ctx := newTestContext()
	mockRestaurantRepo := new(MockRestaurantRepository)
	mockWorkingHoursRepo := new(MockWorkingHoursRepository)

	useCase := usecase.NewRestaurantUseCase(mockRestaurantRepo, mockWorkingHoursRepo)

	newRestaurant := &domain.Restaurant{
		Name:         "new restaurant",
		Address:      "new street, 456",
		Cuisine:      domain.Cuisine("french"),
		Description:  "new restaurant description",
		ContactEmail: "new@restaurant.com",
		ContactPhone: "+7 (987) 654-32-10",
	}

	expectedID := "test-restaurant-id"

	mockRestaurantRepo.On("Create", ctx, mock.AnythingOfType("*domain.Restaurant")).Run(func(args mock.Arguments) {
		restaurant := args.Get(1).(*domain.Restaurant)
		restaurant.ID = expectedID
	}).Return(nil)

	id, err := useCase.CreateRestaurant(ctx, newRestaurant)

	assert.NoError(t, err)
	assert.NotEmpty(t, id)
	assert.Equal(t, expectedID, id)
	assert.Equal(t, expectedID, newRestaurant.ID)
	assert.False(t, newRestaurant.CreatedAt.IsZero())
	assert.False(t, newRestaurant.UpdatedAt.IsZero())
	mockRestaurantRepo.AssertExpectations(t)
}

func TestRestaurantUseCase_UpdateRestaurant(t *testing.T) {

	ctx := newTestContext()
	mockRestaurantRepo := new(MockRestaurantRepository)
	mockWorkingHoursRepo := new(MockWorkingHoursRepository)

	useCase := usecase.NewRestaurantUseCase(mockRestaurantRepo, mockWorkingHoursRepo)

	restaurant := createTestRestaurant()
	oldUpdateTime := restaurant.UpdatedAt

	time.Sleep(1 * time.Millisecond)

	mockRestaurantRepo.On("Update", ctx, mock.AnythingOfType("*domain.Restaurant")).Return(nil)

	err := useCase.UpdateRestaurant(ctx, restaurant)

	assert.NoError(t, err)
	assert.True(t, restaurant.UpdatedAt.After(oldUpdateTime))
	mockRestaurantRepo.AssertExpectations(t)
}

func TestRestaurantUseCase_DeleteRestaurant(t *testing.T) {

	ctx := newTestContext()
	mockRestaurantRepo := new(MockRestaurantRepository)
	mockWorkingHoursRepo := new(MockWorkingHoursRepository)

	useCase := usecase.NewRestaurantUseCase(mockRestaurantRepo, mockWorkingHoursRepo)

	restaurantID := uuid.New().String()

	mockRestaurantRepo.On("Delete", ctx, restaurantID).Return(nil)

	err := useCase.DeleteRestaurant(ctx, restaurantID)

	assert.NoError(t, err)
	mockRestaurantRepo.AssertExpectations(t)
}

func TestRestaurantUseCase_AddFact(t *testing.T) {

	ctx := newTestContext()
	mockRestaurantRepo := new(MockRestaurantRepository)
	mockWorkingHoursRepo := new(MockWorkingHoursRepository)

	useCase := usecase.NewRestaurantUseCase(mockRestaurantRepo, mockWorkingHoursRepo)

	restaurantID := uuid.New().String()
	factContent := "interesting fact about the restaurant"
	expectedFact := &domain.Fact{
		ID:           uuid.New().String(),
		RestaurantID: restaurantID,
		Content:      factContent,
		CreatedAt:    time.Now(),
	}

	mockRestaurantRepo.On("AddFact", ctx, restaurantID, mock.AnythingOfType("domain.Fact")).Return(expectedFact, nil)

	fact, err := useCase.AddFact(ctx, restaurantID, factContent)

	assert.NoError(t, err)
	assert.Equal(t, expectedFact, fact)

	mockRestaurantRepo.AssertExpectations(t)
}

func TestRestaurantUseCase_GetFacts(t *testing.T) {

	ctx := newTestContext()
	mockRestaurantRepo := new(MockRestaurantRepository)
	mockWorkingHoursRepo := new(MockWorkingHoursRepository)

	useCase := usecase.NewRestaurantUseCase(mockRestaurantRepo, mockWorkingHoursRepo)

	restaurantID := uuid.New().String()
	expectedFacts := []domain.Fact{
		{
			ID:           uuid.New().String(),
			RestaurantID: restaurantID,
			Content:      "fact 1",
			CreatedAt:    time.Now(),
		},
		{
			ID:           uuid.New().String(),
			RestaurantID: restaurantID,
			Content:      "fact 2",
			CreatedAt:    time.Now(),
		},
	}

	mockRestaurantRepo.On("GetFacts", ctx, restaurantID).Return(expectedFacts, nil)

	result, err := useCase.GetFacts(ctx, restaurantID)

	assert.NoError(t, err)
	assert.Equal(t, expectedFacts, result)
	assert.Len(t, result, 2)
	mockRestaurantRepo.AssertExpectations(t)
}

func TestRestaurantUseCase_GetRandomFacts(t *testing.T) {

	ctx := newTestContext()
	mockRestaurantRepo := new(MockRestaurantRepository)
	mockWorkingHoursRepo := new(MockWorkingHoursRepository)

	useCase := usecase.NewRestaurantUseCase(mockRestaurantRepo, mockWorkingHoursRepo)

	count := 3
	expectedFacts := []domain.Fact{
		{
			ID:           uuid.New().String(),
			RestaurantID: uuid.New().String(),
			Content:      "random fact 1",
			CreatedAt:    time.Now(),
		},
		{
			ID:           uuid.New().String(),
			RestaurantID: uuid.New().String(),
			Content:      "random fact 2",
			CreatedAt:    time.Now(),
		},
		{
			ID:           uuid.New().String(),
			RestaurantID: uuid.New().String(),
			Content:      "random fact 3",
			CreatedAt:    time.Now(),
		},
	}

	mockRestaurantRepo.On("GetRandomFacts", ctx, count).Return(expectedFacts, nil)

	result, err := useCase.GetRandomFacts(ctx, count)

	assert.NoError(t, err)
	assert.Equal(t, expectedFacts, result)
	assert.Len(t, result, count)
	mockRestaurantRepo.AssertExpectations(t)
}

func TestRestaurantUseCase_SetWorkingHours(t *testing.T) {

	ctx := newTestContext()
	mockRestaurantRepo := new(MockRestaurantRepository)
	mockWorkingHoursRepo := new(MockWorkingHoursRepository)

	useCase := usecase.NewRestaurantUseCase(mockRestaurantRepo, mockWorkingHoursRepo)

	restaurantID := uuid.New().String()
	workingHours := &domain.WorkingHours{
		WeekDay:   domain.Monday,
		OpenTime:  "09:00",
		CloseTime: "21:00",
		ValidFrom: time.Now(),
		ValidTo:   time.Now().AddDate(0, 3, 0),
	}

	mockWorkingHoursRepo.On("SetWorkingHours", ctx, mock.AnythingOfType("*domain.WorkingHours")).Return(nil)

	err := useCase.SetWorkingHours(ctx, restaurantID, workingHours)

	assert.NoError(t, err)
	assert.Equal(t, restaurantID, workingHours.RestaurantID)
	mockWorkingHoursRepo.AssertExpectations(t)
}

func TestRestaurantUseCase_GetWorkingHours(t *testing.T) {

	ctx := newTestContext()
	mockRestaurantRepo := new(MockRestaurantRepository)
	mockWorkingHoursRepo := new(MockWorkingHoursRepository)

	useCase := usecase.NewRestaurantUseCase(mockRestaurantRepo, mockWorkingHoursRepo)

	restaurantID := uuid.New().String()
	expectedWorkingHours := []*domain.WorkingHours{
		{
			ID:           uuid.New().String(),
			RestaurantID: restaurantID,
			WeekDay:      domain.Monday,
			OpenTime:     "09:00",
			CloseTime:    "21:00",
			ValidFrom:    time.Now(),
			ValidTo:      time.Now().AddDate(0, 3, 0),
		},
		{
			ID:           uuid.New().String(),
			RestaurantID: restaurantID,
			WeekDay:      domain.Tuesday,
			OpenTime:     "10:00",
			CloseTime:    "22:00",
			ValidFrom:    time.Now(),
			ValidTo:      time.Now().AddDate(0, 3, 0),
		},
	}

	mockWorkingHoursRepo.On("GetByRestaurantID", ctx, restaurantID).Return(expectedWorkingHours, nil)

	result, err := useCase.GetWorkingHours(ctx, restaurantID)

	assert.NoError(t, err)
	assert.Equal(t, expectedWorkingHours, result)
	assert.Len(t, result, 2)
	mockWorkingHoursRepo.AssertExpectations(t)
}
