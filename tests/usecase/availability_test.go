package usecase_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/flexer2006/case-back-restaurant-go/internal/domain"
	"github.com/flexer2006/case-back-restaurant-go/internal/logger"
	"github.com/flexer2006/case-back-restaurant-go/pkg/usecase"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type mockAvailabilityRepository struct {
	mock.Mock
}

func (m *mockAvailabilityRepository) GetByRestaurantAndDate(ctx context.Context, restaurantID string, date time.Time) ([]*domain.Availability, error) {
	args := m.Called(ctx, restaurantID, date)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*domain.Availability), args.Error(1)
}

func (m *mockAvailabilityRepository) SetAvailability(ctx context.Context, availability *domain.Availability) error {
	args := m.Called(ctx, availability)
	return args.Error(0)
}

func (m *mockAvailabilityRepository) UpdateReservedSeats(ctx context.Context, availabilityID string, delta int) error {
	args := m.Called(ctx, availabilityID, delta)
	return args.Error(0)
}

type mockRestaurantRepository struct {
	mock.Mock
}

func (m *mockRestaurantRepository) GetByID(ctx context.Context, id string) (*domain.Restaurant, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Restaurant), args.Error(1)
}

func (m *mockRestaurantRepository) List(ctx context.Context, offset, limit int) ([]*domain.Restaurant, error) {
	args := m.Called(ctx, offset, limit)
	return args.Get(0).([]*domain.Restaurant), args.Error(1)
}

func (m *mockRestaurantRepository) Create(ctx context.Context, restaurant *domain.Restaurant) error {
	args := m.Called(ctx, restaurant)
	return args.Error(0)
}

func (m *mockRestaurantRepository) Update(ctx context.Context, restaurant *domain.Restaurant) error {
	args := m.Called(ctx, restaurant)
	return args.Error(0)
}

func (m *mockRestaurantRepository) Delete(ctx context.Context, id string) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *mockRestaurantRepository) AddFact(ctx context.Context, restaurantID string, fact domain.Fact) (*domain.Fact, error) {
	args := m.Called(ctx, restaurantID, fact)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Fact), args.Error(1)
}

func (m *mockRestaurantRepository) GetFacts(ctx context.Context, restaurantID string) ([]domain.Fact, error) {
	args := m.Called(ctx, restaurantID)
	return args.Get(0).([]domain.Fact), args.Error(1)
}

func (m *mockRestaurantRepository) GetRandomFacts(ctx context.Context, count int) ([]domain.Fact, error) {
	args := m.Called(ctx, count)
	return args.Get(0).([]domain.Fact), args.Error(1)
}

type mockWorkingHoursRepository struct {
	mock.Mock
}

func (m *mockWorkingHoursRepository) GetByRestaurantID(ctx context.Context, restaurantID string) ([]*domain.WorkingHours, error) {
	args := m.Called(ctx, restaurantID)
	return args.Get(0).([]*domain.WorkingHours), args.Error(1)
}

func (m *mockWorkingHoursRepository) SetWorkingHours(ctx context.Context, hours *domain.WorkingHours) error {
	args := m.Called(ctx, hours)
	return args.Error(0)
}

func (m *mockWorkingHoursRepository) DeleteWorkingHours(ctx context.Context, id string) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func setupTestContext() context.Context {
	loggerInstance, _ := logger.NewLogger()
	ctx := context.Background()
	return logger.NewContext(ctx, loggerInstance)
}

func TestGetAvailability(t *testing.T) {
	availabilityRepo := new(mockAvailabilityRepository)
	restaurantRepo := new(mockRestaurantRepository)
	workingHoursRepo := new(mockWorkingHoursRepository)
	ctx := setupTestContext()

	useCase := usecase.NewAvailabilityUseCase(availabilityRepo, restaurantRepo, workingHoursRepo)

	restaurantID := "rest123"
	date := time.Date(2023, 10, 15, 0, 0, 0, 0, time.UTC)

	t.Run("successful availability retrieval", func(t *testing.T) {
		expected := []*domain.Availability{
			{
				ID:           "avail1",
				RestaurantID: restaurantID,
				Date:         date,
				TimeSlot:     "18:00",
				Capacity:     50,
				Reserved:     20,
			},
			{
				ID:           "avail2",
				RestaurantID: restaurantID,
				Date:         date,
				TimeSlot:     "19:00",
				Capacity:     50,
				Reserved:     30,
			},
		}

		availabilityRepo.On("GetByRestaurantAndDate", ctx, restaurantID, date).Return(expected, nil).Once()

		result, err := useCase.GetAvailability(ctx, restaurantID, date)

		assert.NoError(t, err)
		assert.Equal(t, expected, result)
		availabilityRepo.AssertExpectations(t)
	})

	t.Run("error when retrieving availability", func(t *testing.T) {
		expectedErr := errors.New("database error")
		availabilityRepo.On("GetByRestaurantAndDate", ctx, restaurantID, date).Return(nil, expectedErr).Once()

		result, err := useCase.GetAvailability(ctx, restaurantID, date)

		assert.Error(t, err)
		assert.Equal(t, expectedErr, err)
		assert.Nil(t, result)
		availabilityRepo.AssertExpectations(t)
	})
}

func TestSetAvailability(t *testing.T) {
	availabilityRepo := new(mockAvailabilityRepository)
	restaurantRepo := new(mockRestaurantRepository)
	workingHoursRepo := new(mockWorkingHoursRepository)
	ctx := setupTestContext()

	useCase := usecase.NewAvailabilityUseCase(availabilityRepo, restaurantRepo, workingHoursRepo)

	t.Run("successful availability setting", func(t *testing.T) {
		availability := &domain.Availability{
			ID:           "avail1",
			RestaurantID: "rest123",
			Date:         time.Date(2023, 10, 15, 0, 0, 0, 0, time.UTC),
			TimeSlot:     "18:00",
			Capacity:     50,
			Reserved:     0,
		}

		availabilityRepo.On("SetAvailability", mock.Anything, mock.MatchedBy(func(a *domain.Availability) bool {
			return a.ID == availability.ID && !a.UpdatedAt.IsZero()
		})).Return(nil).Once()

		err := useCase.SetAvailability(ctx, availability)

		assert.NoError(t, err)
		assert.False(t, availability.UpdatedAt.IsZero(), "updatedAt should be set")
		availabilityRepo.AssertExpectations(t)
	})

	t.Run("error when setting availability", func(t *testing.T) {
		availability := &domain.Availability{
			ID:           "avail1",
			RestaurantID: "rest123",
			Date:         time.Date(2023, 10, 15, 0, 0, 0, 0, time.UTC),
			TimeSlot:     "18:00",
			Capacity:     50,
			Reserved:     0,
		}

		expectedErr := errors.New("database error")
		availabilityRepo.On("SetAvailability", mock.Anything, mock.MatchedBy(func(a *domain.Availability) bool {
			return a.ID == availability.ID && !a.UpdatedAt.IsZero()
		})).Return(expectedErr).Once()

		err := useCase.SetAvailability(ctx, availability)

		assert.Error(t, err)
		assert.Equal(t, expectedErr, err)
		availabilityRepo.AssertExpectations(t)
	})
}

func TestUpdateReservedSeats(t *testing.T) {
	availabilityRepo := new(mockAvailabilityRepository)
	restaurantRepo := new(mockRestaurantRepository)
	workingHoursRepo := new(mockWorkingHoursRepository)
	ctx := setupTestContext()

	useCase := usecase.NewAvailabilityUseCase(availabilityRepo, restaurantRepo, workingHoursRepo)
	availabilityID := "avail1"

	t.Run("successful reserved seats update (increase)", func(t *testing.T) {
		delta := 3
		availabilityRepo.On("UpdateReservedSeats", ctx, availabilityID, delta).Return(nil).Once()

		err := useCase.UpdateReservedSeats(ctx, availabilityID, delta)

		assert.NoError(t, err)
		availabilityRepo.AssertExpectations(t)
	})

	t.Run("successful reserved seats update (decrease)", func(t *testing.T) {
		delta := -2
		availabilityRepo.On("UpdateReservedSeats", ctx, availabilityID, delta).Return(nil).Once()

		err := useCase.UpdateReservedSeats(ctx, availabilityID, delta)

		assert.NoError(t, err)
		availabilityRepo.AssertExpectations(t)
	})

	t.Run("error when updating reserved seats", func(t *testing.T) {
		delta := 3
		expectedErr := errors.New("database error")
		availabilityRepo.On("UpdateReservedSeats", ctx, availabilityID, delta).Return(expectedErr).Once()

		err := useCase.UpdateReservedSeats(ctx, availabilityID, delta)

		assert.Error(t, err)
		assert.Equal(t, expectedErr, err)
		availabilityRepo.AssertExpectations(t)
	})
}

func TestCheckAvailability(t *testing.T) {
	availabilityRepo := new(mockAvailabilityRepository)
	restaurantRepo := new(mockRestaurantRepository)
	workingHoursRepo := new(mockWorkingHoursRepository)
	ctx := setupTestContext()

	useCase := usecase.NewAvailabilityUseCase(availabilityRepo, restaurantRepo, workingHoursRepo)

	restaurantID := "rest123"
	date := time.Date(2023, 10, 15, 0, 0, 0, 0, time.UTC)
	timeSlot := "18:00"

	t.Run("enough seats available", func(t *testing.T) {
		guestsCount := 5
		availabilities := []*domain.Availability{
			{
				ID:           "avail1",
				RestaurantID: restaurantID,
				Date:         date,
				TimeSlot:     timeSlot,
				Capacity:     50,
				Reserved:     20,
			},
		}

		availabilityRepo.On("GetByRestaurantAndDate", ctx, restaurantID, date).Return(availabilities, nil).Once()

		isAvailable, err := useCase.CheckAvailability(ctx, restaurantID, date, timeSlot, guestsCount)

		assert.NoError(t, err)
		assert.True(t, isAvailable)
		availabilityRepo.AssertExpectations(t)
	})

	t.Run("not enough seats", func(t *testing.T) {
		guestsCount := 35
		availabilities := []*domain.Availability{
			{
				ID:           "avail1",
				RestaurantID: restaurantID,
				Date:         date,
				TimeSlot:     timeSlot,
				Capacity:     50,
				Reserved:     20,
			},
		}

		availabilityRepo.On("GetByRestaurantAndDate", ctx, restaurantID, date).Return(availabilities, nil).Once()

		isAvailable, err := useCase.CheckAvailability(ctx, restaurantID, date, timeSlot, guestsCount)

		assert.NoError(t, err)
		assert.False(t, isAvailable)
		availabilityRepo.AssertExpectations(t)
	})

	t.Run("time slot not found", func(t *testing.T) {
		guestsCount := 5
		nonExistentTimeSlot := "20:00"
		availabilities := []*domain.Availability{
			{
				ID:           "avail1",
				RestaurantID: restaurantID,
				Date:         date,
				TimeSlot:     timeSlot,
				Capacity:     50,
				Reserved:     20,
			},
		}

		availabilityRepo.On("GetByRestaurantAndDate", ctx, restaurantID, date).Return(availabilities, nil).Once()

		isAvailable, err := useCase.CheckAvailability(ctx, restaurantID, date, nonExistentTimeSlot, guestsCount)

		assert.NoError(t, err)
		assert.False(t, isAvailable)
		availabilityRepo.AssertExpectations(t)
	})

	t.Run("error when getting availability data", func(t *testing.T) {
		guestsCount := 5
		expectedErr := errors.New("database error")
		availabilityRepo.On("GetByRestaurantAndDate", ctx, restaurantID, date).Return(nil, expectedErr).Once()

		isAvailable, err := useCase.CheckAvailability(ctx, restaurantID, date, timeSlot, guestsCount)

		assert.Error(t, err)
		assert.Equal(t, expectedErr, err)
		assert.False(t, isAvailable)
		availabilityRepo.AssertExpectations(t)
	})
}
