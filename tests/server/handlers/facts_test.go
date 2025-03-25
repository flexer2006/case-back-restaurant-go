package handlers_test

import (
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

type MockFactsUseCase struct {
	mock.Mock
}

func (m *MockFactsUseCase) GetRandomFacts(ctx context.Context, count int) ([]domain.Fact, error) {
	args := m.Called(ctx, count)
	return args.Get(0).([]domain.Fact), args.Error(1)
}

func (m *MockFactsUseCase) GetRestaurantFacts(ctx context.Context, restaurantID string) ([]domain.Fact, error) {
	args := m.Called(ctx, restaurantID)
	return args.Get(0).([]domain.Fact), args.Error(1)
}

func setupFactsTestApp(_ *testing.T) (*fiber.App, *MockFactsUseCase, context.Context) {
	app := fiber.New()
	factsUseCase := new(MockFactsUseCase)
	handler := handlers.NewFactsHandler(factsUseCase)

	testLogger := CreateTestLogger()

	ctx := logger.NewContext(context.Background(), testLogger)

	app.Use(func(c fiber.Ctx) error {
		c.Locals("ctx", ctx)
		return c.Next()
	})

	api := app.Group("/api/v1")
	api.Get("/facts/random", handler.GetRandomFacts)

	return app, factsUseCase, ctx
}

func TestGetRandomFacts_Default(t *testing.T) {
	app, factsUseCase, _ := setupFactsTestApp(t)

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
			RestaurantID: "restaurant2",
			Content:      "Interesting fact 2",
			CreatedAt:    currentTime,
		},
		{
			ID:           "fact3",
			RestaurantID: "restaurant3",
			Content:      "Interesting fact 3",
			CreatedAt:    currentTime,
		},
	}

	factsUseCase.On("GetRandomFacts", mock.Anything, 3).Return(facts, nil)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/facts/random", nil)
	resp, err := app.Test(req)
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var respFacts []domain.Fact
	err = json.NewDecoder(resp.Body).Decode(&respFacts)
	require.NoError(t, err)
	assert.Len(t, respFacts, 3)
	assert.Equal(t, facts[0].ID, respFacts[0].ID)
	assert.Equal(t, facts[1].ID, respFacts[1].ID)
	assert.Equal(t, facts[2].ID, respFacts[2].ID)

	factsUseCase.AssertExpectations(t)
}

func TestGetRandomFacts_WithCount(t *testing.T) {
	app, factsUseCase, _ := setupFactsTestApp(t)

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
			RestaurantID: "restaurant2",
			Content:      "Interesting fact 2",
			CreatedAt:    currentTime,
		},
		{
			ID:           "fact3",
			RestaurantID: "restaurant3",
			Content:      "Interesting fact 3",
			CreatedAt:    currentTime,
		},
		{
			ID:           "fact4",
			RestaurantID: "restaurant4",
			Content:      "Interesting fact 4",
			CreatedAt:    currentTime,
		},
		{
			ID:           "fact5",
			RestaurantID: "restaurant5",
			Content:      "Interesting fact 5",
			CreatedAt:    currentTime,
		},
	}

	factsUseCase.On("GetRandomFacts", mock.Anything, 5).Return(facts, nil)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/facts/random?count=5", nil)
	resp, err := app.Test(req)
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var respFacts []domain.Fact
	err = json.NewDecoder(resp.Body).Decode(&respFacts)
	require.NoError(t, err)
	assert.Len(t, respFacts, 5)
	for i := 0; i < 5; i++ {
		assert.Equal(t, facts[i].ID, respFacts[i].ID)
	}

	factsUseCase.AssertExpectations(t)
}

func TestGetRandomFacts_InvalidCount(t *testing.T) {
	app, factsUseCase, _ := setupFactsTestApp(t)

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
			RestaurantID: "restaurant2",
			Content:      "Interesting fact 2",
			CreatedAt:    currentTime,
		},
		{
			ID:           "fact3",
			RestaurantID: "restaurant3",
			Content:      "Interesting fact 3",
			CreatedAt:    currentTime,
		},
	}

	factsUseCase.On("GetRandomFacts", mock.Anything, 3).Return(facts, nil)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/facts/random?count=-5", nil)
	resp, err := app.Test(req)
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var respFacts []domain.Fact
	err = json.NewDecoder(resp.Body).Decode(&respFacts)
	require.NoError(t, err)
	assert.Len(t, respFacts, 3)

	req = httptest.NewRequest(http.MethodGet, "/api/v1/facts/random?count=abc", nil)
	resp, err = app.Test(req)
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	err = json.NewDecoder(resp.Body).Decode(&respFacts)
	require.NoError(t, err)
	assert.Len(t, respFacts, 3)

	req = httptest.NewRequest(http.MethodGet, "/api/v1/facts/random?count=20", nil)
	resp, err = app.Test(req)
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	err = json.NewDecoder(resp.Body).Decode(&respFacts)
	require.NoError(t, err)
	assert.Len(t, respFacts, 3)

	factsUseCase.AssertExpectations(t)
}

func TestGetRandomFacts_InternalError(t *testing.T) {
	app, factsUseCase, _ := setupFactsTestApp(t)

	factsUseCase.On("GetRandomFacts", mock.Anything, 3).Return([]domain.Fact{}, errors.New("database error"))

	req := httptest.NewRequest(http.MethodGet, "/api/v1/facts/random", nil)
	resp, err := app.Test(req)
	require.NoError(t, err)
	assert.Equal(t, http.StatusInternalServerError, resp.StatusCode)

	var respBody map[string]string
	err = json.NewDecoder(resp.Body).Decode(&respBody)
	require.NoError(t, err)
	assert.Equal(t, common.ErrInternalError, respBody["error"])

	factsUseCase.AssertExpectations(t)
}
