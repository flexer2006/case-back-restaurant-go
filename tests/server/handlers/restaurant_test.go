package handlers_test

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/flexer2006/case-back-restaurant-go/common"
	"github.com/flexer2006/case-back-restaurant-go/internal/domain"
	"github.com/flexer2006/case-back-restaurant-go/internal/logger"
	"github.com/flexer2006/case-back-restaurant-go/internal/server/handlers"

	"github.com/gofiber/fiber/v3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

type MockRestaurantUseCase struct {
	mock.Mock
}

func (m *MockRestaurantUseCase) GetRestaurant(ctx context.Context, id string) (*domain.Restaurant, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Restaurant), args.Error(1)
}

func (m *MockRestaurantUseCase) ListRestaurants(ctx context.Context, offset, limit int) ([]*domain.Restaurant, error) {
	args := m.Called(ctx, offset, limit)
	return args.Get(0).([]*domain.Restaurant), args.Error(1)
}

func (m *MockRestaurantUseCase) CreateRestaurant(ctx context.Context, restaurant *domain.Restaurant) (string, error) {
	args := m.Called(ctx, restaurant)
	return args.String(0), args.Error(1)
}

func (m *MockRestaurantUseCase) UpdateRestaurant(ctx context.Context, restaurant *domain.Restaurant) error {
	args := m.Called(ctx, restaurant)
	return args.Error(0)
}

func (m *MockRestaurantUseCase) DeleteRestaurant(ctx context.Context, id string) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockRestaurantUseCase) AddFact(ctx context.Context, restaurantID string, content string) (*domain.Fact, error) {
	args := m.Called(ctx, restaurantID, content)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Fact), args.Error(1)
}

func (m *MockRestaurantUseCase) GetFacts(ctx context.Context, restaurantID string) ([]domain.Fact, error) {
	args := m.Called(ctx, restaurantID)
	return args.Get(0).([]domain.Fact), args.Error(1)
}

func (m *MockRestaurantUseCase) GetRandomFacts(ctx context.Context, count int) ([]domain.Fact, error) {
	args := m.Called(ctx, count)
	return args.Get(0).([]domain.Fact), args.Error(1)
}

func (m *MockRestaurantUseCase) SetWorkingHours(ctx context.Context, restaurantID string, workingHours *domain.WorkingHours) error {
	args := m.Called(ctx, restaurantID, workingHours)
	return args.Error(0)
}

func (m *MockRestaurantUseCase) GetWorkingHours(ctx context.Context, restaurantID string) ([]*domain.WorkingHours, error) {
	args := m.Called(ctx, restaurantID)
	return args.Get(0).([]*domain.WorkingHours), args.Error(1)
}

type MockAvailabilityUseCase struct {
	mock.Mock
}

func (m *MockAvailabilityUseCase) GetAvailability(ctx context.Context, restaurantID string, date time.Time) ([]*domain.Availability, error) {
	args := m.Called(ctx, restaurantID, date)
	return args.Get(0).([]*domain.Availability), args.Error(1)
}

func (m *MockAvailabilityUseCase) SetAvailability(ctx context.Context, availability *domain.Availability) error {
	args := m.Called(ctx, availability)
	return args.Error(0)
}

func (m *MockAvailabilityUseCase) UpdateReservedSeats(ctx context.Context, availabilityID string, delta int) error {
	args := m.Called(ctx, availabilityID, delta)
	return args.Error(0)
}

func (m *MockAvailabilityUseCase) CheckAvailability(ctx context.Context, restaurantID string, date time.Time, timeSlot string, guestsCount int) (bool, error) {
	args := m.Called(ctx, restaurantID, date, timeSlot, guestsCount)
	return args.Bool(0), args.Error(1)
}

func setupRestaurantTestApp(_ *testing.T) (*fiber.App, *MockRestaurantUseCase, *MockBookingUseCase, *MockAvailabilityUseCase, context.Context) {
	app := fiber.New()
	restaurantUseCase := new(MockRestaurantUseCase)
	bookingUseCase := new(MockBookingUseCase)
	availabilityUseCase := new(MockAvailabilityUseCase)
	handler := handlers.NewRestaurantHandler(restaurantUseCase, bookingUseCase, availabilityUseCase)

	testLogger := CreateTestLogger()

	ctx := logger.NewContext(context.Background(), testLogger)

	app.Use(func(c fiber.Ctx) error {
		c.Locals("ctx", ctx)
		return c.Next()
	})

	api := app.Group("/api/v1")
	api.Get("/restaurants", handler.ListRestaurants)
	api.Post("/restaurants", handler.CreateRestaurant)
	api.Get("/restaurants/:id", handler.GetRestaurant)
	api.Put("/restaurants/:id", handler.UpdateRestaurant)
	api.Delete("/restaurants/:id", handler.DeleteRestaurant)
	api.Get("/restaurants/:id/facts", handler.GetFacts)
	api.Post("/restaurants/:id/facts", handler.AddFact)
	api.Get("/restaurants/:id/working-hours", handler.GetWorkingHours)
	api.Post("/restaurants/:id/working-hours", handler.SetWorkingHours)
	api.Get("/restaurants/:id/availability", handler.GetAvailability)
	api.Post("/restaurants/:id/availability", handler.SetAvailability)
	api.Get("/restaurants/:id/bookings", handler.GetRestaurantBookings)

	return app, restaurantUseCase, bookingUseCase, availabilityUseCase, ctx
}

func TestListRestaurants_Success(t *testing.T) {
	app, restaurantUseCase, _, _, _ := setupRestaurantTestApp(t)

	currentTime := time.Now()
	restaurants := []*domain.Restaurant{
		{
			ID:           "restaurant1",
			Name:         "Restaurant 1",
			Address:      "123 Main St",
			Cuisine:      "Italian",
			Description:  "Italian cuisine restaurant",
			ContactEmail: "restaurant1@example.com",
			ContactPhone: "+11234567890",
			CreatedAt:    currentTime,
			UpdatedAt:    currentTime,
		},
		{
			ID:           "restaurant2",
			Name:         "Restaurant 2",
			Address:      "456 Broadway",
			Cuisine:      "Mexican",
			Description:  "Mexican cuisine restaurant",
			ContactEmail: "restaurant2@example.com",
			ContactPhone: "+12345678901",
			CreatedAt:    currentTime,
			UpdatedAt:    currentTime,
		},
	}

	restaurantUseCase.On("ListRestaurants", mock.Anything, 0, 20).Return(restaurants, nil)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/restaurants", nil)
	resp, err := app.Test(req)
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var respRestaurants []*domain.Restaurant
	err = json.NewDecoder(resp.Body).Decode(&respRestaurants)
	require.NoError(t, err)
	assert.Len(t, respRestaurants, 2)
	assert.Equal(t, restaurants[0].ID, respRestaurants[0].ID)
	assert.Equal(t, restaurants[1].ID, respRestaurants[1].ID)

	restaurantUseCase.AssertExpectations(t)
}

func TestListRestaurants_WithPagination(t *testing.T) {
	app, restaurantUseCase, _, _, _ := setupRestaurantTestApp(t)

	currentTime := time.Now()
	restaurants := []*domain.Restaurant{
		{
			ID:           "restaurant3",
			Name:         "Restaurant 3",
			Address:      "789 Market St",
			Cuisine:      "Japanese",
			ContactEmail: "restaurant3@example.com",
			ContactPhone: "+13456789012",
			CreatedAt:    currentTime,
			UpdatedAt:    currentTime,
		},
	}

	restaurantUseCase.On("ListRestaurants", mock.Anything, 10, 5).Return(restaurants, nil)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/restaurants?offset=10&limit=5", nil)
	resp, err := app.Test(req)
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var respRestaurants []*domain.Restaurant
	err = json.NewDecoder(resp.Body).Decode(&respRestaurants)
	require.NoError(t, err)
	assert.Len(t, respRestaurants, 1)
	assert.Equal(t, restaurants[0].ID, respRestaurants[0].ID)

	restaurantUseCase.AssertExpectations(t)
}

func TestListRestaurants_InternalError(t *testing.T) {
	app, restaurantUseCase, _, _, _ := setupRestaurantTestApp(t)

	restaurantUseCase.On("ListRestaurants", mock.Anything, 0, 20).Return([]*domain.Restaurant{}, errors.New("database error"))

	req := httptest.NewRequest(http.MethodGet, "/api/v1/restaurants", nil)
	resp, err := app.Test(req)
	require.NoError(t, err)
	assert.Equal(t, http.StatusInternalServerError, resp.StatusCode)

	var respBody map[string]string
	err = json.NewDecoder(resp.Body).Decode(&respBody)
	require.NoError(t, err)
	assert.Equal(t, common.ErrInternalServer, respBody["error"])

	restaurantUseCase.AssertExpectations(t)
}

func TestGetRestaurant_Success(t *testing.T) {
	app, restaurantUseCase, _, _, _ := setupRestaurantTestApp(t)

	currentTime := time.Now()
	expectedRestaurant := &domain.Restaurant{
		ID:           "restaurant1",
		Name:         "Restaurant 1",
		Address:      "123 Main St",
		Cuisine:      "Italian",
		Description:  "Italian cuisine restaurant",
		ContactEmail: "restaurant1@example.com",
		ContactPhone: "+11234567890",
		CreatedAt:    currentTime,
		UpdatedAt:    currentTime,
	}

	restaurantUseCase.On("GetRestaurant", mock.Anything, "restaurant1").Return(expectedRestaurant, nil)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/restaurants/restaurant1", nil)
	resp, err := app.Test(req)
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var respRestaurant domain.Restaurant
	err = json.NewDecoder(resp.Body).Decode(&respRestaurant)
	require.NoError(t, err)
	assert.Equal(t, expectedRestaurant.ID, respRestaurant.ID)
	assert.Equal(t, expectedRestaurant.Name, respRestaurant.Name)
	assert.Equal(t, expectedRestaurant.Address, respRestaurant.Address)

	restaurantUseCase.AssertExpectations(t)
}

func TestGetRestaurant_NotFound(t *testing.T) {
	app, restaurantUseCase, _, _, _ := setupRestaurantTestApp(t)

	restaurantUseCase.On("GetRestaurant", mock.Anything, "nonexistent").Return(nil, nil)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/restaurants/nonexistent", nil)
	resp, err := app.Test(req)
	require.NoError(t, err)
	assert.Equal(t, http.StatusNotFound, resp.StatusCode)

	var respBody map[string]string
	err = json.NewDecoder(resp.Body).Decode(&respBody)
	require.NoError(t, err)
	assert.Equal(t, common.ErrRestaurantNotFound, respBody["error"])

	restaurantUseCase.AssertExpectations(t)
}

func TestGetRestaurant_InternalError(t *testing.T) {
	app, restaurantUseCase, _, _, _ := setupRestaurantTestApp(t)

	restaurantUseCase.On("GetRestaurant", mock.Anything, "restaurant1").Return(nil, errors.New("database error"))

	req := httptest.NewRequest(http.MethodGet, "/api/v1/restaurants/restaurant1", nil)
	resp, err := app.Test(req)
	require.NoError(t, err)
	assert.Equal(t, http.StatusInternalServerError, resp.StatusCode)

	var respBody map[string]string
	err = json.NewDecoder(resp.Body).Decode(&respBody)
	require.NoError(t, err)
	assert.Equal(t, common.ErrInternalServer, respBody["error"])

	restaurantUseCase.AssertExpectations(t)
}

func TestCreateRestaurant_Success(t *testing.T) {
	app, restaurantUseCase, _, _, _ := setupRestaurantTestApp(t)

	restaurantUseCase.On("CreateRestaurant", mock.Anything, mock.MatchedBy(func(restaurant *domain.Restaurant) bool {
		return restaurant.Name == "Test Restaurant" &&
			restaurant.Address == "123 Test St" &&
			string(restaurant.Cuisine) == "Italian" &&
			restaurant.ContactEmail == "test@example.com" &&
			restaurant.ContactPhone == "+71234567890"
	})).Return("restaurant123", nil)

	fact1 := &domain.Fact{
		ID:           "fact1",
		RestaurantID: "restaurant123",
		Content:      "Amazing pizza",
		CreatedAt:    time.Now(),
	}

	fact2 := &domain.Fact{
		ID:           "fact2",
		RestaurantID: "restaurant123",
		Content:      "Fresh ingredients",
		CreatedAt:    time.Now(),
	}

	restaurantUseCase.On("AddFact", mock.Anything, "restaurant123", "Amazing pizza").Return(fact1, nil)
	restaurantUseCase.On("AddFact", mock.Anything, "restaurant123", "Fresh ingredients").Return(fact2, nil)

	reqBody := handlers.CreateRestaurantRequest{
		Name:         "Test Restaurant",
		Address:      "123 Test St",
		Cuisine:      "Italian",
		Description:  "Test description",
		ContactEmail: "test@example.com",
		ContactPhone: "+71234567890",
		Facts:        []string{"Amazing pizza", "Fresh ingredients"},
	}
	reqJSON, _ := json.Marshal(reqBody)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/restaurants", bytes.NewBuffer(reqJSON))
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req)
	require.NoError(t, err)
	assert.Equal(t, http.StatusCreated, resp.StatusCode)

	var respBody map[string]string
	err = json.NewDecoder(resp.Body).Decode(&respBody)
	require.NoError(t, err)
	assert.Equal(t, "restaurant123", respBody["id"])

	restaurantUseCase.AssertExpectations(t)
}

func TestCreateRestaurant_InvalidParams(t *testing.T) {
	app, restaurantUseCase, _, _, _ := setupRestaurantTestApp(t)

	restaurantUseCase.On("CreateRestaurant", mock.Anything, mock.Anything).Return("", nil).Maybe()

	reqJSON := []byte(`{"name": "Test Restaurant", "address": invalid-json}`)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/restaurants", bytes.NewBuffer(reqJSON))
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req)
	require.NoError(t, err)
	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)

	var respBody map[string]string
	err = json.NewDecoder(resp.Body).Decode(&respBody)
	require.NoError(t, err)
	assert.Equal(t, common.ErrInvalidParams, respBody["error"])
}

func TestCreateRestaurant_InternalError(t *testing.T) {
	app, restaurantUseCase, _, _, _ := setupRestaurantTestApp(t)

	restaurantUseCase.On("CreateRestaurant", mock.Anything, mock.Anything).Return("", errors.New("database error"))

	reqBody := handlers.CreateRestaurantRequest{
		Name:         "Test Restaurant",
		Address:      "123 Test St",
		Cuisine:      "Italian",
		Description:  "Test description",
		ContactEmail: "test@example.com",
		ContactPhone: "+71234567890",
	}
	reqJSON, _ := json.Marshal(reqBody)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/restaurants", bytes.NewBuffer(reqJSON))
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req)
	require.NoError(t, err)
	assert.Equal(t, http.StatusInternalServerError, resp.StatusCode)

	var respBody map[string]string
	err = json.NewDecoder(resp.Body).Decode(&respBody)
	require.NoError(t, err)
	assert.Equal(t, common.ErrInternalServer, respBody["error"])

	restaurantUseCase.AssertExpectations(t)
}

func TestUpdateRestaurant_Success(t *testing.T) {
	app, restaurantUseCase, _, _, _ := setupRestaurantTestApp(t)

	currentTime := time.Now()
	existingRestaurant := &domain.Restaurant{
		ID:           "restaurant1",
		Name:         "Old Restaurant Name",
		Address:      "Old Address",
		Cuisine:      "Italian",
		Description:  "Old description",
		ContactEmail: "old@example.com",
		ContactPhone: "+10987654321",
		CreatedAt:    currentTime,
		UpdatedAt:    currentTime,
	}

	restaurantUseCase.On("GetRestaurant", mock.Anything, "restaurant1").Return(existingRestaurant, nil)
	restaurantUseCase.On("UpdateRestaurant", mock.Anything, mock.MatchedBy(func(restaurant *domain.Restaurant) bool {
		return restaurant.ID == "restaurant1" &&
			restaurant.Name == "Updated Restaurant" &&
			restaurant.Address == "456 New St" &&
			string(restaurant.Cuisine) == "Mexican" &&
			restaurant.ContactEmail == "updated@example.com"
	})).Return(nil)

	reqBody := handlers.UpdateRestaurantRequest{
		Name:         "Updated Restaurant",
		Address:      "456 New St",
		Cuisine:      "Mexican",
		Description:  "Updated description",
		ContactEmail: "updated@example.com",
		ContactPhone: "+70987654321",
	}
	reqJSON, _ := json.Marshal(reqBody)

	req := httptest.NewRequest(http.MethodPut, "/api/v1/restaurants/restaurant1", bytes.NewBuffer(reqJSON))
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req)
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var respBody map[string]string
	err = json.NewDecoder(resp.Body).Decode(&respBody)
	require.NoError(t, err)
	assert.Equal(t, "success", respBody["status"])

	restaurantUseCase.AssertExpectations(t)
}

func TestUpdateRestaurant_NotFound(t *testing.T) {
	app, restaurantUseCase, _, _, _ := setupRestaurantTestApp(t)

	restaurantUseCase.On("GetRestaurant", mock.Anything, "nonexistent").Return(nil, nil)

	reqBody := handlers.UpdateRestaurantRequest{
		Name:         "Updated Restaurant",
		Address:      "456 New St",
		Cuisine:      "Mexican",
		Description:  "Updated description",
		ContactEmail: "updated@example.com",
		ContactPhone: "+70987654321",
	}
	reqJSON, _ := json.Marshal(reqBody)

	req := httptest.NewRequest(http.MethodPut, "/api/v1/restaurants/nonexistent", bytes.NewBuffer(reqJSON))
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req)
	require.NoError(t, err)
	assert.Equal(t, http.StatusNotFound, resp.StatusCode)

	var respBody map[string]string
	err = json.NewDecoder(resp.Body).Decode(&respBody)
	require.NoError(t, err)
	assert.Equal(t, common.ErrRestaurantNotFound, respBody["error"])

	restaurantUseCase.AssertExpectations(t)
}

func TestDeleteRestaurant_Success(t *testing.T) {
	app, restaurantUseCase, _, _, _ := setupRestaurantTestApp(t)

	restaurantUseCase.On("DeleteRestaurant", mock.Anything, "restaurant1").Return(nil)

	req := httptest.NewRequest(http.MethodDelete, "/api/v1/restaurants/restaurant1", nil)
	resp, err := app.Test(req)
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var respBody map[string]string
	err = json.NewDecoder(resp.Body).Decode(&respBody)
	require.NoError(t, err)
	assert.Equal(t, "success", respBody["status"])

	restaurantUseCase.AssertExpectations(t)
}

func TestDeleteRestaurant_InternalError(t *testing.T) {
	app, restaurantUseCase, _, _, _ := setupRestaurantTestApp(t)

	restaurantUseCase.On("DeleteRestaurant", mock.Anything, "restaurant1").Return(errors.New("database error"))

	req := httptest.NewRequest(http.MethodDelete, "/api/v1/restaurants/restaurant1", nil)
	resp, err := app.Test(req)
	require.NoError(t, err)
	assert.Equal(t, http.StatusInternalServerError, resp.StatusCode)

	var respBody map[string]string
	err = json.NewDecoder(resp.Body).Decode(&respBody)
	require.NoError(t, err)
	assert.Equal(t, common.ErrInternalServer, respBody["error"])

	restaurantUseCase.AssertExpectations(t)
}

func TestAddFact_Success(t *testing.T) {
	app, restaurantUseCase, _, _, _ := setupRestaurantTestApp(t)

	// Проверяем существование ресторана
	restaurantUseCase.On("GetRestaurant", mock.Anything, "restaurant1").Return(&domain.Restaurant{
		ID: "restaurant1",
	}, nil)

	// Создаем объект факта, который будет возвращен
	createdFact := &domain.Fact{
		ID:           "fact123",
		RestaurantID: "restaurant1",
		Content:      "New interesting fact",
		CreatedAt:    time.Now(),
	}

	restaurantUseCase.On("AddFact", mock.Anything, "restaurant1", "New interesting fact").Return(createdFact, nil)

	reqBody := handlers.AddFactRequest{
		Content: "New interesting fact",
	}
	reqJSON, _ := json.Marshal(reqBody)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/restaurants/restaurant1/facts", bytes.NewBuffer(reqJSON))
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req)
	require.NoError(t, err)
	assert.Equal(t, http.StatusCreated, resp.StatusCode)

	var respBody domain.Fact
	err = json.NewDecoder(resp.Body).Decode(&respBody)
	require.NoError(t, err)
	assert.Equal(t, "fact123", respBody.ID)
	assert.Equal(t, "restaurant1", respBody.RestaurantID)
	assert.Equal(t, "New interesting fact", respBody.Content)

	restaurantUseCase.AssertExpectations(t)
}

func TestGetFacts_Success(t *testing.T) {
	app, restaurantUseCase, _, _, _ := setupRestaurantTestApp(t)

	currentTime := time.Now()
	facts := []domain.Fact{
		{
			ID:           "fact1",
			RestaurantID: "restaurant1",
			Content:      "Interesting fact 1",
			CreatedAt:    currentTime,
		},
		{
			ID:           "fact2",
			RestaurantID: "restaurant1",
			Content:      "Interesting fact 2",
			CreatedAt:    currentTime,
		},
	}

	restaurantUseCase.On("GetFacts", mock.Anything, "restaurant1").Return(facts, nil)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/restaurants/restaurant1/facts", nil)
	resp, err := app.Test(req)
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var respFacts []domain.Fact
	err = json.NewDecoder(resp.Body).Decode(&respFacts)
	require.NoError(t, err)
	assert.Len(t, respFacts, 2)
	assert.Equal(t, facts[0].ID, respFacts[0].ID)
	assert.Equal(t, facts[1].ID, respFacts[1].ID)

	restaurantUseCase.AssertExpectations(t)
}

func TestSetWorkingHours_Success(t *testing.T) {
	app, restaurantUseCase, _, _, _ := setupRestaurantTestApp(t)

	restaurantUseCase.On("SetWorkingHours", mock.Anything, "restaurant1", mock.MatchedBy(func(wh *domain.WorkingHours) bool {
		return wh.RestaurantID == "restaurant1" &&
			wh.WeekDay == domain.Monday &&
			wh.OpenTime == "09:00" &&
			wh.CloseTime == "22:00"
	})).Return(nil)

	validFrom := time.Now()
	validTo := validFrom.AddDate(0, 6, 0) // +6 месяцев

	reqBody := handlers.SetWorkingHoursRequest{
		WeekDay:   domain.Monday,
		OpenTime:  "09:00",
		CloseTime: "22:00",
		ValidFrom: validFrom,
		ValidTo:   validTo,
	}
	reqJSON, _ := json.Marshal(reqBody)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/restaurants/restaurant1/working-hours", bytes.NewBuffer(reqJSON))
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req)
	require.NoError(t, err)
	assert.Equal(t, http.StatusCreated, resp.StatusCode)

	var respBody map[string]string
	err = json.NewDecoder(resp.Body).Decode(&respBody)
	require.NoError(t, err)
	assert.Equal(t, "success", respBody["status"])

	restaurantUseCase.AssertExpectations(t)
}

func TestGetWorkingHours_Success(t *testing.T) {
	app, restaurantUseCase, _, _, _ := setupRestaurantTestApp(t)

	validFrom := time.Now()
	validTo := validFrom.AddDate(0, 6, 0)

	workingHours := []*domain.WorkingHours{
		{
			ID:           "wh1",
			RestaurantID: "restaurant1",
			WeekDay:      domain.Monday,
			OpenTime:     "09:00",
			CloseTime:    "22:00",
			ValidFrom:    validFrom,
			ValidTo:      validTo,
		},
		{
			ID:           "wh2",
			RestaurantID: "restaurant1",
			WeekDay:      domain.Tuesday,
			OpenTime:     "09:00",
			CloseTime:    "22:00",
			ValidFrom:    validFrom,
			ValidTo:      validTo,
		},
	}

	restaurantUseCase.On("GetWorkingHours", mock.Anything, "restaurant1").Return(workingHours, nil)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/restaurants/restaurant1/working-hours", nil)
	resp, err := app.Test(req)
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var respWorkingHours []*domain.WorkingHours
	err = json.NewDecoder(resp.Body).Decode(&respWorkingHours)
	require.NoError(t, err)
	assert.Len(t, respWorkingHours, 2)
	assert.Equal(t, workingHours[0].ID, respWorkingHours[0].ID)
	assert.Equal(t, workingHours[1].ID, respWorkingHours[1].ID)

	restaurantUseCase.AssertExpectations(t)
}

func TestSetAvailability_Success(t *testing.T) {
	app, _, _, availabilityUseCase, _ := setupRestaurantTestApp(t)

	availabilityUseCase.On("SetAvailability", mock.Anything, mock.MatchedBy(func(a *domain.Availability) bool {
		return a.RestaurantID == "restaurant1" &&
			a.TimeSlot == "19:00" &&
			a.Capacity == 20
	})).Return(nil)

	date := time.Now().AddDate(0, 0, 1)

	reqBody := handlers.SetAvailabilityRequest{
		Date:     date,
		TimeSlot: "19:00",
		Capacity: 20,
	}
	reqJSON, _ := json.Marshal(reqBody)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/restaurants/restaurant1/availability", bytes.NewBuffer(reqJSON))
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req)
	require.NoError(t, err)
	assert.Equal(t, http.StatusCreated, resp.StatusCode)

	var respBody map[string]string
	err = json.NewDecoder(resp.Body).Decode(&respBody)
	require.NoError(t, err)
	assert.Equal(t, "success", respBody["status"])

	availabilityUseCase.AssertExpectations(t)
}

func TestGetAvailability_Success(t *testing.T) {
	app, _, _, availabilityUseCase, _ := setupRestaurantTestApp(t)

	date := time.Now().AddDate(0, 0, 1)
	dateStr := date.Format("2006-01-02")

	availabilities := []*domain.Availability{
		{
			ID:           "a1",
			RestaurantID: "restaurant1",
			Date:         date,
			TimeSlot:     "18:00",
			Capacity:     20,
			Reserved:     5,
		},
		{
			ID:           "a2",
			RestaurantID: "restaurant1",
			Date:         date,
			TimeSlot:     "19:00",
			Capacity:     20,
			Reserved:     10,
		},
	}

	availabilityUseCase.On("GetAvailability", mock.Anything, "restaurant1", mock.MatchedBy(func(d time.Time) bool {
		return d.Format("2006-01-02") == dateStr
	})).Return(availabilities, nil)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/restaurants/restaurant1/availability?date="+dateStr, nil)
	resp, err := app.Test(req)
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var respAvailabilities []*domain.Availability
	err = json.NewDecoder(resp.Body).Decode(&respAvailabilities)
	require.NoError(t, err)
	assert.Len(t, respAvailabilities, 2)
	assert.Equal(t, availabilities[0].ID, respAvailabilities[0].ID)
	assert.Equal(t, availabilities[1].ID, respAvailabilities[1].ID)

	availabilityUseCase.AssertExpectations(t)
}

func TestGetRestaurantBookings_Success(t *testing.T) {
	app, _, bookingUseCase, _, _ := setupRestaurantTestApp(t)

	currentTime := time.Now()
	bookings := []*domain.Booking{
		{
			ID:           "booking1",
			RestaurantID: "restaurant1",
			UserID:       "user1",
			Date:         currentTime,
			Time:         "18:00",
			Duration:     90,
			GuestsCount:  2,
			Status:       domain.BookingStatusConfirmed,
			CreatedAt:    currentTime,
			UpdatedAt:    currentTime,
			ConfirmedAt:  &currentTime,
		},
		{
			ID:           "booking2",
			RestaurantID: "restaurant1",
			UserID:       "user2",
			Date:         currentTime.Add(24 * time.Hour),
			Time:         "19:00",
			Duration:     120,
			GuestsCount:  4,
			Status:       domain.BookingStatusPending,
			CreatedAt:    currentTime,
			UpdatedAt:    currentTime,
		},
	}

	bookingUseCase.On("GetRestaurantBookings", mock.Anything, "restaurant1").Return(bookings, nil)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/restaurants/restaurant1/bookings", nil)
	resp, err := app.Test(req)
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var respBookings []*domain.Booking
	err = json.NewDecoder(resp.Body).Decode(&respBookings)
	require.NoError(t, err)
	assert.Len(t, respBookings, 2)
	assert.Equal(t, bookings[0].ID, respBookings[0].ID)
	assert.Equal(t, bookings[1].ID, respBookings[1].ID)

	bookingUseCase.AssertExpectations(t)
}

func TestGetRestaurantBookings_InternalError(t *testing.T) {
	app, _, bookingUseCase, _, _ := setupRestaurantTestApp(t)

	bookingUseCase.On("GetRestaurantBookings", mock.Anything, "restaurant1").Return([]*domain.Booking{}, errors.New("database error"))

	req := httptest.NewRequest(http.MethodGet, "/api/v1/restaurants/restaurant1/bookings", nil)
	resp, err := app.Test(req)
	require.NoError(t, err)
	assert.Equal(t, http.StatusInternalServerError, resp.StatusCode)

	var respBody map[string]string
	err = json.NewDecoder(resp.Body).Decode(&respBody)
	require.NoError(t, err)
	assert.Equal(t, common.ErrInternalServer, respBody["error"])

	bookingUseCase.AssertExpectations(t)
}
