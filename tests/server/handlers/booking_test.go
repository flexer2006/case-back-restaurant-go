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

func setupBookingTestApp(_ *testing.T) (*fiber.App, *MockBookingUseCase, context.Context) {
	app := fiber.New()
	bookingUseCase := new(MockBookingUseCase)
	handler := handlers.NewBookingHandler(bookingUseCase)

	testLogger := CreateTestLogger()

	ctx := logger.NewContext(context.Background(), testLogger)

	app.Use(func(c fiber.Ctx) error {
		c.Locals("ctx", ctx)
		return c.Next()
	})

	api := app.Group("/api/v1")
	api.Post("/bookings", handler.CreateBooking)
	api.Get("/bookings/:id", handler.GetBooking)
	api.Post("/bookings/:id/confirm", handler.ConfirmBooking)
	api.Post("/bookings/:id/reject", handler.RejectBooking)
	api.Post("/bookings/:id/cancel", handler.CancelBooking)
	api.Post("/bookings/:id/complete", handler.CompleteBooking)
	api.Post("/bookings/:id/alternative", handler.SuggestAlternativeTime)
	api.Post("/bookings/alternatives/:id/accept", handler.AcceptAlternative)
	api.Post("/bookings/alternatives/:id/reject", handler.RejectAlternative)

	return app, bookingUseCase, ctx
}

func TestCreateBooking_Success(t *testing.T) {
	app, bookingUseCase, _ := setupBookingTestApp(t)

	bookingUseCase.On("CreateBooking", mock.Anything, mock.MatchedBy(func(booking *domain.Booking) bool {
		return booking.RestaurantID == "restaurant1" &&
			booking.UserID == "user1" &&
			booking.GuestsCount == 2 &&
			booking.Time == "19:00" &&
			booking.Duration == 90
	})).Return("booking123", nil)

	bookingDate := time.Now().Add(24 * time.Hour).Round(time.Second)
	reqBody := handlers.CreateBookingRequest{
		RestaurantID: "restaurant1",
		UserID:       "user1",
		Date:         bookingDate,
		Time:         "19:00",
		Duration:     90,
		GuestsCount:  2,
		Comment:      "Special occasion",
	}
	reqJSON, _ := json.Marshal(reqBody)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/bookings", bytes.NewBuffer(reqJSON))
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req)
	require.NoError(t, err)
	assert.Equal(t, http.StatusCreated, resp.StatusCode)

	var respBody map[string]string
	err = json.NewDecoder(resp.Body).Decode(&respBody)
	require.NoError(t, err)
	assert.Equal(t, "booking123", respBody["id"])

	bookingUseCase.AssertExpectations(t)
}

func TestCreateBooking_InvalidParams(t *testing.T) {
	app, bookingUseCase, _ := setupBookingTestApp(t)

	bookingUseCase.On("CreateBooking", mock.Anything, mock.Anything).Return("", nil).Maybe()

	reqJSON := []byte(`{"restaurant_id": "restaurant1", "date": invalid-json}`)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/bookings", bytes.NewBuffer(reqJSON))
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req)
	require.NoError(t, err)
	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)

	var respBody map[string]string
	err = json.NewDecoder(resp.Body).Decode(&respBody)
	require.NoError(t, err)
	assert.Equal(t, common.ErrInvalidParams, respBody["error"])
}

func TestCreateBooking_InternalError(t *testing.T) {
	app, bookingUseCase, _ := setupBookingTestApp(t)

	bookingUseCase.On("CreateBooking", mock.Anything, mock.Anything).Return("", errors.New("database error"))

	bookingDate := time.Now().Add(24 * time.Hour)
	reqBody := handlers.CreateBookingRequest{
		RestaurantID: "restaurant1",
		UserID:       "user1",
		Date:         bookingDate,
		Time:         "19:00",
		Duration:     90,
		GuestsCount:  2,
	}
	reqJSON, _ := json.Marshal(reqBody)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/bookings", bytes.NewBuffer(reqJSON))
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req)
	require.NoError(t, err)
	assert.Equal(t, http.StatusInternalServerError, resp.StatusCode)

	var respBody map[string]string
	err = json.NewDecoder(resp.Body).Decode(&respBody)
	require.NoError(t, err)
	assert.Equal(t, common.ErrInternalServer, respBody["error"])

	bookingUseCase.AssertExpectations(t)
}

func TestGetBooking_Success(t *testing.T) {
	app, bookingUseCase, _ := setupBookingTestApp(t)

	currentTime := time.Now()
	expectedBooking := &domain.Booking{
		ID:           "booking123",
		RestaurantID: "restaurant1",
		UserID:       "user1",
		Date:         currentTime,
		Time:         "19:00",
		Duration:     90,
		GuestsCount:  2,
		Status:       domain.BookingStatusConfirmed,
		CreatedAt:    currentTime,
		UpdatedAt:    currentTime,
		ConfirmedAt:  &currentTime,
	}

	bookingUseCase.On("GetBooking", mock.Anything, "booking123").Return(expectedBooking, nil)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/bookings/booking123", nil)
	resp, err := app.Test(req)
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var respBooking domain.Booking
	err = json.NewDecoder(resp.Body).Decode(&respBooking)
	require.NoError(t, err)
	assert.Equal(t, expectedBooking.ID, respBooking.ID)
	assert.Equal(t, expectedBooking.RestaurantID, respBooking.RestaurantID)
	assert.Equal(t, expectedBooking.UserID, respBooking.UserID)
	assert.Equal(t, expectedBooking.Status, respBooking.Status)

	bookingUseCase.AssertExpectations(t)
}

func TestGetBooking_NotFound(t *testing.T) {
	app, bookingUseCase, _ := setupBookingTestApp(t)

	bookingUseCase.On("GetBooking", mock.Anything, "nonexistent").Return(nil, nil)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/bookings/nonexistent", nil)
	resp, err := app.Test(req)
	require.NoError(t, err)
	assert.Equal(t, http.StatusNotFound, resp.StatusCode)

	var respBody map[string]string
	err = json.NewDecoder(resp.Body).Decode(&respBody)
	require.NoError(t, err)
	assert.Equal(t, common.ErrBookingNotFound, respBody["error"])

	bookingUseCase.AssertExpectations(t)
}

func TestGetBooking_InternalError(t *testing.T) {
	app, bookingUseCase, _ := setupBookingTestApp(t)

	bookingUseCase.On("GetBooking", mock.Anything, "booking123").Return(nil, errors.New("database error"))

	req := httptest.NewRequest(http.MethodGet, "/api/v1/bookings/booking123", nil)
	resp, err := app.Test(req)
	require.NoError(t, err)
	assert.Equal(t, http.StatusInternalServerError, resp.StatusCode)

	var respBody map[string]string
	err = json.NewDecoder(resp.Body).Decode(&respBody)
	require.NoError(t, err)
	assert.Equal(t, common.ErrInternalServer, respBody["error"])

	bookingUseCase.AssertExpectations(t)
}

func TestConfirmBooking_Success(t *testing.T) {
	app, bookingUseCase, _ := setupBookingTestApp(t)

	bookingUseCase.On("ConfirmBooking", mock.Anything, "booking123").Return(nil)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/bookings/booking123/confirm", nil)
	resp, err := app.Test(req)
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var respBody map[string]string
	err = json.NewDecoder(resp.Body).Decode(&respBody)
	require.NoError(t, err)
	assert.Equal(t, "success", respBody["status"])

	bookingUseCase.AssertExpectations(t)
}

func TestConfirmBooking_InternalError(t *testing.T) {
	app, bookingUseCase, _ := setupBookingTestApp(t)

	bookingUseCase.On("ConfirmBooking", mock.Anything, "booking123").Return(errors.New("database error"))

	req := httptest.NewRequest(http.MethodPost, "/api/v1/bookings/booking123/confirm", nil)
	resp, err := app.Test(req)
	require.NoError(t, err)
	assert.Equal(t, http.StatusInternalServerError, resp.StatusCode)

	var respBody map[string]string
	err = json.NewDecoder(resp.Body).Decode(&respBody)
	require.NoError(t, err)
	assert.Equal(t, common.ErrInternalServer, respBody["error"])

	bookingUseCase.AssertExpectations(t)
}

func TestRejectBooking_Success(t *testing.T) {
	app, bookingUseCase, _ := setupBookingTestApp(t)

	bookingUseCase.On("RejectBooking", mock.Anything, "booking123", "Fully booked").Return(nil)

	reqBody := handlers.RejectBookingRequest{
		Reason: "Fully booked",
	}
	reqJSON, _ := json.Marshal(reqBody)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/bookings/booking123/reject", bytes.NewBuffer(reqJSON))
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req)
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var respBody map[string]string
	err = json.NewDecoder(resp.Body).Decode(&respBody)
	require.NoError(t, err)
	assert.Equal(t, "success", respBody["status"])

	bookingUseCase.AssertExpectations(t)
}

func TestRejectBooking_InvalidParams(t *testing.T) {
	app, bookingUseCase, _ := setupBookingTestApp(t)

	bookingUseCase.On("RejectBooking", mock.Anything, mock.Anything, mock.Anything).Return(nil).Maybe()

	reqJSON := []byte(`{"reason": invalid-json}`)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/bookings/booking123/reject", bytes.NewBuffer(reqJSON))
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req)
	require.NoError(t, err)
	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)

	var respBody map[string]string
	err = json.NewDecoder(resp.Body).Decode(&respBody)
	require.NoError(t, err)
	assert.Equal(t, common.ErrInvalidParams, respBody["error"])
}

func TestRejectBooking_InternalError(t *testing.T) {
	app, bookingUseCase, _ := setupBookingTestApp(t)

	bookingUseCase.On("RejectBooking", mock.Anything, "booking123", "Fully booked").Return(errors.New("database error"))

	reqBody := handlers.RejectBookingRequest{
		Reason: "Fully booked",
	}
	reqJSON, _ := json.Marshal(reqBody)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/bookings/booking123/reject", bytes.NewBuffer(reqJSON))
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req)
	require.NoError(t, err)
	assert.Equal(t, http.StatusInternalServerError, resp.StatusCode)

	var respBody map[string]string
	err = json.NewDecoder(resp.Body).Decode(&respBody)
	require.NoError(t, err)
	assert.Equal(t, common.ErrInternalServer, respBody["error"])

	bookingUseCase.AssertExpectations(t)
}

func TestCancelBooking_Success(t *testing.T) {
	app, bookingUseCase, _ := setupBookingTestApp(t)

	bookingUseCase.On("CancelBooking", mock.Anything, "booking123").Return(nil)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/bookings/booking123/cancel", nil)
	resp, err := app.Test(req)
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var respBody map[string]string
	err = json.NewDecoder(resp.Body).Decode(&respBody)
	require.NoError(t, err)
	assert.Equal(t, "success", respBody["status"])

	bookingUseCase.AssertExpectations(t)
}

func TestCancelBooking_InternalError(t *testing.T) {
	app, bookingUseCase, _ := setupBookingTestApp(t)

	bookingUseCase.On("CancelBooking", mock.Anything, "booking123").Return(errors.New("database error"))

	req := httptest.NewRequest(http.MethodPost, "/api/v1/bookings/booking123/cancel", nil)
	resp, err := app.Test(req)
	require.NoError(t, err)
	assert.Equal(t, http.StatusInternalServerError, resp.StatusCode)

	var respBody map[string]string
	err = json.NewDecoder(resp.Body).Decode(&respBody)
	require.NoError(t, err)
	assert.Equal(t, common.ErrInternalServer, respBody["error"])

	bookingUseCase.AssertExpectations(t)
}

func TestCompleteBooking_Success(t *testing.T) {
	app, bookingUseCase, _ := setupBookingTestApp(t)

	bookingUseCase.On("CompleteBooking", mock.Anything, "booking123").Return(nil)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/bookings/booking123/complete", nil)
	resp, err := app.Test(req)
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var respBody map[string]string
	err = json.NewDecoder(resp.Body).Decode(&respBody)
	require.NoError(t, err)
	assert.Equal(t, "success", respBody["status"])

	bookingUseCase.AssertExpectations(t)
}

func TestCompleteBooking_InternalError(t *testing.T) {
	app, bookingUseCase, _ := setupBookingTestApp(t)

	bookingUseCase.On("CompleteBooking", mock.Anything, "booking123").Return(errors.New("database error"))

	req := httptest.NewRequest(http.MethodPost, "/api/v1/bookings/booking123/complete", nil)
	resp, err := app.Test(req)
	require.NoError(t, err)
	assert.Equal(t, http.StatusInternalServerError, resp.StatusCode)

	var respBody map[string]string
	err = json.NewDecoder(resp.Body).Decode(&respBody)
	require.NoError(t, err)
	assert.Equal(t, common.ErrInternalServer, respBody["error"])

	bookingUseCase.AssertExpectations(t)
}

func TestSuggestAlternativeTime_Success(t *testing.T) {
	app, bookingUseCase, _ := setupBookingTestApp(t)

	alternativeDate := time.Now().Add(48 * time.Hour).Round(time.Second)
	message := "How about the day after tomorrow?"

	bookingUseCase.On(
		"SuggestAlternativeTime",
		mock.Anything,
		"booking123",
		mock.MatchedBy(func(date time.Time) bool {
			return date.Equal(alternativeDate)
		}),
		"20:00",
		message,
	).Return("alternative123", nil)

	reqBody := handlers.SuggestAlternativeTimeRequest{
		Date:    alternativeDate,
		Time:    "20:00",
		Message: message,
	}
	reqJSON, _ := json.Marshal(reqBody)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/bookings/booking123/alternative", bytes.NewBuffer(reqJSON))
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req)
	require.NoError(t, err)
	assert.Equal(t, http.StatusCreated, resp.StatusCode)

	var respBody map[string]string
	err = json.NewDecoder(resp.Body).Decode(&respBody)
	require.NoError(t, err)
	assert.Equal(t, "alternative123", respBody["id"])

	bookingUseCase.AssertExpectations(t)
}

func TestSuggestAlternativeTime_InvalidParams(t *testing.T) {
	app, bookingUseCase, _ := setupBookingTestApp(t)

	bookingUseCase.On("SuggestAlternativeTime", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return("", nil).Maybe()

	reqJSON := []byte(`{"date": invalid-json, "time": "20:00"}`)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/bookings/booking123/alternative", bytes.NewBuffer(reqJSON))
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req)
	require.NoError(t, err)
	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)

	var respBody map[string]string
	err = json.NewDecoder(resp.Body).Decode(&respBody)
	require.NoError(t, err)
	assert.Equal(t, common.ErrInvalidParams, respBody["error"])
}

func TestSuggestAlternativeTime_InternalError(t *testing.T) {
	app, bookingUseCase, _ := setupBookingTestApp(t)

	alternativeDate := time.Now().Add(48 * time.Hour)

	bookingUseCase.On(
		"SuggestAlternativeTime",
		mock.Anything,
		"booking123",
		mock.Anything,
		"20:00",
		"Alternative time suggestion",
	).Return("", errors.New("database error"))

	reqBody := handlers.SuggestAlternativeTimeRequest{
		Date:    alternativeDate,
		Time:    "20:00",
		Message: "Alternative time suggestion",
	}
	reqJSON, _ := json.Marshal(reqBody)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/bookings/booking123/alternative", bytes.NewBuffer(reqJSON))
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req)
	require.NoError(t, err)
	assert.Equal(t, http.StatusInternalServerError, resp.StatusCode)

	var respBody map[string]string
	err = json.NewDecoder(resp.Body).Decode(&respBody)
	require.NoError(t, err)
	assert.Equal(t, common.ErrInternalServer, respBody["error"])

	bookingUseCase.AssertExpectations(t)
}

func TestAcceptAlternative_Success(t *testing.T) {
	app, bookingUseCase, _ := setupBookingTestApp(t)

	bookingUseCase.On("AcceptAlternative", mock.Anything, "alternative123").Return(nil)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/bookings/alternatives/alternative123/accept", nil)
	resp, err := app.Test(req)
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var respBody map[string]string
	err = json.NewDecoder(resp.Body).Decode(&respBody)
	require.NoError(t, err)
	assert.Equal(t, "success", respBody["status"])

	bookingUseCase.AssertExpectations(t)
}

func TestAcceptAlternative_InternalError(t *testing.T) {
	app, bookingUseCase, _ := setupBookingTestApp(t)

	bookingUseCase.On("AcceptAlternative", mock.Anything, "alternative123").Return(errors.New("database error"))

	req := httptest.NewRequest(http.MethodPost, "/api/v1/bookings/alternatives/alternative123/accept", nil)
	resp, err := app.Test(req)
	require.NoError(t, err)
	assert.Equal(t, http.StatusInternalServerError, resp.StatusCode)

	var respBody map[string]string
	err = json.NewDecoder(resp.Body).Decode(&respBody)
	require.NoError(t, err)
	assert.Equal(t, common.ErrInternalServer, respBody["error"])

	bookingUseCase.AssertExpectations(t)
}

func TestRejectAlternative_Success(t *testing.T) {
	app, bookingUseCase, _ := setupBookingTestApp(t)

	bookingUseCase.On("RejectAlternative", mock.Anything, "alternative123").Return(nil)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/bookings/alternatives/alternative123/reject", nil)
	resp, err := app.Test(req)
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var respBody map[string]string
	err = json.NewDecoder(resp.Body).Decode(&respBody)
	require.NoError(t, err)
	assert.Equal(t, "success", respBody["status"])

	bookingUseCase.AssertExpectations(t)
}

func TestRejectAlternative_InternalError(t *testing.T) {
	app, bookingUseCase, _ := setupBookingTestApp(t)

	bookingUseCase.On("RejectAlternative", mock.Anything, "alternative123").Return(errors.New("database error"))

	req := httptest.NewRequest(http.MethodPost, "/api/v1/bookings/alternatives/alternative123/reject", nil)
	resp, err := app.Test(req)
	require.NoError(t, err)
	assert.Equal(t, http.StatusInternalServerError, resp.StatusCode)

	var respBody map[string]string
	err = json.NewDecoder(resp.Body).Decode(&respBody)
	require.NoError(t, err)
	assert.Equal(t, common.ErrInternalServer, respBody["error"])

	bookingUseCase.AssertExpectations(t)
}
