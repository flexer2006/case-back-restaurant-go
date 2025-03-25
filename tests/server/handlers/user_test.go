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
	"github.com/flexer2006/case-back-restaurant-go/internal/logger/ports"
	"github.com/flexer2006/case-back-restaurant-go/internal/server/handlers"
	"github.com/flexer2006/case-back-restaurant-go/pkg/usecase"

	"github.com/gofiber/fiber/v3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

type MockLogger struct {
	mock.Mock
}

func (m *MockLogger) Info(ctx context.Context, msg string, fields ...zap.Field) {
	m.Called(ctx, msg, fields)
}

func (m *MockLogger) Warn(ctx context.Context, msg string, fields ...zap.Field) {
	m.Called(ctx, msg, fields)
}

func (m *MockLogger) Error(ctx context.Context, msg string, fields ...zap.Field) {
	m.Called(ctx, msg, fields)
}

func (m *MockLogger) Debug(ctx context.Context, msg string, fields ...zap.Field) {
	m.Called(ctx, msg, fields)
}

func (m *MockLogger) Fatal(ctx context.Context, msg string, fields ...zap.Field) {
	m.Called(ctx, msg, fields)
}

func (m *MockLogger) SetLevel(level ports.LogLevel) {
	m.Called(level)
}

func (m *MockLogger) GetLevel() ports.LogLevel {
	args := m.Called()
	return args.Get(0).(ports.LogLevel)
}

func (m *MockLogger) With(fields ...zap.Field) ports.LoggerPort {
	args := m.Called(fields)
	return args.Get(0).(ports.LoggerPort)
}

func (m *MockLogger) Sync() error {
	args := m.Called()
	return args.Error(0)
}

func CreateTestLogger() ports.LoggerPort {
	mockLogger := new(MockLogger)

	mockLogger.On("Info", mock.Anything, mock.Anything, mock.Anything).Return()
	mockLogger.On("Error", mock.Anything, mock.Anything, mock.Anything).Return()
	mockLogger.On("Warn", mock.Anything, mock.Anything, mock.Anything).Return()
	mockLogger.On("Debug", mock.Anything, mock.Anything, mock.Anything).Return()
	mockLogger.On("Fatal", mock.Anything, mock.Anything, mock.Anything).Return()
	mockLogger.On("GetLevel").Return(ports.InfoLevel)
	mockLogger.On("With", mock.Anything).Return(mockLogger)
	mockLogger.On("Sync").Return(nil)

	return mockLogger
}

type MockUserUseCase struct {
	mock.Mock
}

func (m *MockUserUseCase) GetUser(ctx context.Context, id string) (*domain.User, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.User), args.Error(1)
}

func (m *MockUserUseCase) GetUserByEmail(ctx context.Context, email string) (*domain.User, error) {
	args := m.Called(ctx, email)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.User), args.Error(1)
}

func (m *MockUserUseCase) CreateUser(ctx context.Context, user *domain.User) (string, error) {
	args := m.Called(ctx, user)
	return args.String(0), args.Error(1)
}

func (m *MockUserUseCase) UpdateUser(ctx context.Context, user *domain.User) error {
	args := m.Called(ctx, user)
	return args.Error(0)
}

type MockBookingUseCase struct {
	mock.Mock
}

func (m *MockBookingUseCase) GetBooking(ctx context.Context, id string) (*domain.Booking, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Booking), args.Error(1)
}

func (m *MockBookingUseCase) GetRestaurantBookings(ctx context.Context, restaurantID string) ([]*domain.Booking, error) {
	args := m.Called(ctx, restaurantID)
	return args.Get(0).([]*domain.Booking), args.Error(1)
}

func (m *MockBookingUseCase) GetUserBookings(ctx context.Context, userID string) ([]*domain.Booking, error) {
	args := m.Called(ctx, userID)
	return args.Get(0).([]*domain.Booking), args.Error(1)
}

func (m *MockBookingUseCase) CreateBooking(ctx context.Context, booking *domain.Booking) (string, error) {
	args := m.Called(ctx, booking)
	return args.String(0), args.Error(1)
}

func (m *MockBookingUseCase) ConfirmBooking(ctx context.Context, id string) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockBookingUseCase) RejectBooking(ctx context.Context, id string, reason string) error {
	args := m.Called(ctx, id, reason)
	return args.Error(0)
}

func (m *MockBookingUseCase) CancelBooking(ctx context.Context, id string) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockBookingUseCase) CompleteBooking(ctx context.Context, id string) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockBookingUseCase) SuggestAlternativeTime(ctx context.Context, bookingID string, date time.Time, time string, message string) (string, error) {
	args := m.Called(ctx, bookingID, date, time, message)
	return args.String(0), args.Error(1)
}

func (m *MockBookingUseCase) AcceptAlternative(ctx context.Context, alternativeID string) error {
	args := m.Called(ctx, alternativeID)
	return args.Error(0)
}

func (m *MockBookingUseCase) RejectAlternative(ctx context.Context, alternativeID string) error {
	args := m.Called(ctx, alternativeID)
	return args.Error(0)
}

type MockNotificationUseCase struct {
	mock.Mock
}

func (m *MockNotificationUseCase) NotifyRestaurant(ctx context.Context, restaurantID string, notificationType domain.NotificationType, title, message, relatedID string) error {
	args := m.Called(ctx, restaurantID, notificationType, title, message, relatedID)
	return args.Error(0)
}

func (m *MockNotificationUseCase) NotifyUser(ctx context.Context, userID string, notificationType domain.NotificationType, title, message, relatedID string) error {
	args := m.Called(ctx, userID, notificationType, title, message, relatedID)
	return args.Error(0)
}

func (m *MockNotificationUseCase) GetUserNotifications(ctx context.Context, userID string) ([]domain.Notification, error) {
	args := m.Called(ctx, userID)
	return args.Get(0).([]domain.Notification), args.Error(1)
}

func (m *MockNotificationUseCase) MarkAsRead(ctx context.Context, notificationID string) error {
	args := m.Called(ctx, notificationID)
	return args.Error(0)
}

func setupTestApp(_ *testing.T) (*fiber.App, *MockUserUseCase, *MockBookingUseCase, *MockNotificationUseCase, context.Context) {
	app := fiber.New()
	userUseCase := new(MockUserUseCase)
	bookingUseCase := new(MockBookingUseCase)
	notificationUseCase := new(MockNotificationUseCase)
	handler := handlers.NewUserHandler(userUseCase, bookingUseCase, notificationUseCase)

	testLogger := CreateTestLogger()

	ctx := logger.NewContext(context.Background(), testLogger)

	app.Use(func(c fiber.Ctx) error {
		c.Locals("ctx", ctx)
		return c.Next()
	})

	api := app.Group("/api/v1")
	api.Post("/users", handler.CreateUser)
	api.Get("/users/:id", handler.GetUser)
	api.Put("/users/:id", handler.UpdateUser)
	api.Get("/users/:id/bookings", handler.GetUserBookings)
	api.Get("/users/:id/notifications", handler.GetUserNotifications)

	return app, userUseCase, bookingUseCase, notificationUseCase, ctx
}

func TestCreateUser_Success(t *testing.T) {
	app, userUseCase, _, _, _ := setupTestApp(t)

	userUseCase.On("CreateUser", mock.Anything, mock.MatchedBy(func(user *domain.User) bool {
		return user.Name == "Test User" && user.Email == "test@example.com" && user.Phone == "+71234567890"
	})).Return("user123", nil)

	reqBody := handlers.CreateUserRequest{
		Name:  "Test User",
		Email: "test@example.com",
		Phone: "+71234567890",
	}
	reqJSON, _ := json.Marshal(reqBody)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/users", bytes.NewBuffer(reqJSON))
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req)
	require.NoError(t, err)
	assert.Equal(t, http.StatusCreated, resp.StatusCode)

	var respBody map[string]string
	err = json.NewDecoder(resp.Body).Decode(&respBody)
	require.NoError(t, err)
	assert.Equal(t, "user123", respBody["id"])

	userUseCase.AssertExpectations(t)
}

func TestCreateUser_EmailExists(t *testing.T) {
	app, userUseCase, _, _, _ := setupTestApp(t)

	userUseCase.On("CreateUser", mock.Anything, mock.MatchedBy(func(user *domain.User) bool {
		return user.Email == "existing@example.com"
	})).Return("", usecase.ErrEmailExists)

	reqBody := handlers.CreateUserRequest{
		Name:  "Test User",
		Email: "existing@example.com",
		Phone: "+71234567890",
	}
	reqJSON, _ := json.Marshal(reqBody)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/users", bytes.NewBuffer(reqJSON))
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req)
	require.NoError(t, err)
	assert.Equal(t, http.StatusConflict, resp.StatusCode)

	var respBody map[string]string
	err = json.NewDecoder(resp.Body).Decode(&respBody)
	require.NoError(t, err)
	assert.Equal(t, common.ErrEmailAlreadyExistsMsg, respBody["error"])

	userUseCase.AssertExpectations(t)
}

func TestCreateUser_InvalidParams(t *testing.T) {
	app, userUseCase, _, _, _ := setupTestApp(t)

	userUseCase.On("CreateUser", mock.Anything, mock.Anything).Return("", nil).Maybe()

	reqJSON := []byte(`{"name": "Test User", "email": invalid-json, "phone": "+71234567890"}`)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/users", bytes.NewBuffer(reqJSON))
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req)
	require.NoError(t, err)
	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)

	var respBody map[string]string
	err = json.NewDecoder(resp.Body).Decode(&respBody)
	require.NoError(t, err)
	assert.Equal(t, common.ErrInvalidParams, respBody["error"])
}

func TestCreateUser_InternalError(t *testing.T) {
	app, userUseCase, _, _, _ := setupTestApp(t)

	userUseCase.On("CreateUser", mock.Anything, mock.Anything).Return("", errors.New("database error"))

	reqBody := handlers.CreateUserRequest{
		Name:  "Test User",
		Email: "test@example.com",
		Phone: "+71234567890",
	}
	reqJSON, _ := json.Marshal(reqBody)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/users", bytes.NewBuffer(reqJSON))
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req)
	require.NoError(t, err)
	assert.Equal(t, http.StatusInternalServerError, resp.StatusCode)

	var respBody map[string]string
	err = json.NewDecoder(resp.Body).Decode(&respBody)
	require.NoError(t, err)
	assert.Equal(t, common.ErrInternalServer, respBody["error"])

	userUseCase.AssertExpectations(t)
}

func TestGetUser_Success(t *testing.T) {
	app, userUseCase, _, _, _ := setupTestApp(t)

	expectedUser := &domain.User{
		ID:        "user123",
		Name:      "Test User",
		Email:     "test@example.com",
		Phone:     "+71234567890",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	userUseCase.On("GetUser", mock.Anything, "user123").Return(expectedUser, nil)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/users/user123", nil)
	resp, err := app.Test(req)
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var respUser domain.User
	err = json.NewDecoder(resp.Body).Decode(&respUser)
	require.NoError(t, err)
	assert.Equal(t, expectedUser.ID, respUser.ID)
	assert.Equal(t, expectedUser.Name, respUser.Name)
	assert.Equal(t, expectedUser.Email, respUser.Email)

	userUseCase.AssertExpectations(t)
}

func TestGetUser_NotFound(t *testing.T) {
	app, userUseCase, _, _, _ := setupTestApp(t)

	userUseCase.On("GetUser", mock.Anything, "nonexistent").Return(nil, nil)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/users/nonexistent", nil)
	resp, err := app.Test(req)
	require.NoError(t, err)
	assert.Equal(t, http.StatusNotFound, resp.StatusCode)

	var respBody map[string]string
	err = json.NewDecoder(resp.Body).Decode(&respBody)
	require.NoError(t, err)
	assert.Equal(t, common.ErrUserNotFound, respBody["error"])

	userUseCase.AssertExpectations(t)
}

func TestGetUser_InternalError(t *testing.T) {
	app, userUseCase, _, _, _ := setupTestApp(t)

	userUseCase.On("GetUser", mock.Anything, "user123").Return(nil, errors.New("database error"))

	req := httptest.NewRequest(http.MethodGet, "/api/v1/users/user123", nil)
	resp, err := app.Test(req)
	require.NoError(t, err)
	assert.Equal(t, http.StatusInternalServerError, resp.StatusCode)

	var respBody map[string]string
	err = json.NewDecoder(resp.Body).Decode(&respBody)
	require.NoError(t, err)
	assert.Equal(t, common.ErrInternalServer, respBody["error"])

	userUseCase.AssertExpectations(t)
}

func TestUpdateUser_Success(t *testing.T) {
	app, userUseCase, _, _, _ := setupTestApp(t)

	userUseCase.On("UpdateUser", mock.Anything, mock.MatchedBy(func(user *domain.User) bool {
		return user.ID == "user123" &&
			user.Name == "Updated User" &&
			user.Email == "updated@example.com" &&
			user.Phone == "+71234567891"
	})).Return(nil)

	reqBody := handlers.UpdateUserRequest{
		Name:  "Updated User",
		Email: "updated@example.com",
		Phone: "+71234567891",
	}
	reqJSON, _ := json.Marshal(reqBody)

	req := httptest.NewRequest(http.MethodPut, "/api/v1/users/user123", bytes.NewBuffer(reqJSON))
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req)
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var respBody map[string]string
	err = json.NewDecoder(resp.Body).Decode(&respBody)
	require.NoError(t, err)
	assert.Equal(t, common.MsgSuccess, respBody["status"])

	userUseCase.AssertExpectations(t)
}

func TestUpdateUser_NotFound(t *testing.T) {
	app, userUseCase, _, _, _ := setupTestApp(t)

	userUseCase.On("UpdateUser", mock.Anything, mock.Anything).Return(usecase.ErrUserNotFound)

	reqBody := handlers.UpdateUserRequest{
		Name:  "Updated User",
		Email: "updated@example.com",
		Phone: "+71234567891",
	}
	reqJSON, _ := json.Marshal(reqBody)

	req := httptest.NewRequest(http.MethodPut, "/api/v1/users/nonexistent", bytes.NewBuffer(reqJSON))
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req)
	require.NoError(t, err)
	assert.Equal(t, http.StatusNotFound, resp.StatusCode)

	var respBody map[string]string
	err = json.NewDecoder(resp.Body).Decode(&respBody)
	require.NoError(t, err)
	assert.Equal(t, common.ErrUserNotFound, respBody["error"])

	userUseCase.AssertExpectations(t)
}

func TestUpdateUser_EmailExists(t *testing.T) {
	app, userUseCase, _, _, _ := setupTestApp(t)

	userUseCase.On("UpdateUser", mock.Anything, mock.Anything).Return(usecase.ErrEmailExists)

	reqBody := handlers.UpdateUserRequest{
		Name:  "Updated User",
		Email: "existing@example.com",
		Phone: "+71234567891",
	}
	reqJSON, _ := json.Marshal(reqBody)

	req := httptest.NewRequest(http.MethodPut, "/api/v1/users/user123", bytes.NewBuffer(reqJSON))
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req)
	require.NoError(t, err)
	assert.Equal(t, http.StatusConflict, resp.StatusCode)

	var respBody map[string]string
	err = json.NewDecoder(resp.Body).Decode(&respBody)
	require.NoError(t, err)
	assert.Equal(t, common.ErrEmailAlreadyExistsMsg, respBody["error"])

	userUseCase.AssertExpectations(t)
}

func TestUpdateUser_InternalError(t *testing.T) {
	app, userUseCase, _, _, _ := setupTestApp(t)

	userUseCase.On("UpdateUser", mock.Anything, mock.Anything).Return(errors.New("database error"))

	reqBody := handlers.UpdateUserRequest{
		Name:  "Updated User",
		Email: "updated@example.com",
		Phone: "+71234567891",
	}
	reqJSON, _ := json.Marshal(reqBody)

	req := httptest.NewRequest(http.MethodPut, "/api/v1/users/user123", bytes.NewBuffer(reqJSON))
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req)
	require.NoError(t, err)
	assert.Equal(t, http.StatusInternalServerError, resp.StatusCode)

	var respBody map[string]string
	err = json.NewDecoder(resp.Body).Decode(&respBody)
	require.NoError(t, err)
	assert.Equal(t, common.ErrInternalServer, respBody["error"])

	userUseCase.AssertExpectations(t)
}

func TestGetUserBookings_Success(t *testing.T) {
	app, _, bookingUseCase, _, _ := setupTestApp(t)

	currentTime := time.Now()
	bookings := []*domain.Booking{
		{
			ID:           "booking1",
			RestaurantID: "restaurant1",
			UserID:       "user123",
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
			RestaurantID: "restaurant2",
			UserID:       "user123",
			Date:         currentTime.Add(24 * time.Hour),
			Time:         "19:00",
			Duration:     120,
			GuestsCount:  4,
			Status:       domain.BookingStatusPending,
			CreatedAt:    currentTime,
			UpdatedAt:    currentTime,
		},
	}

	bookingUseCase.On("GetUserBookings", mock.Anything, "user123").Return(bookings, nil)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/users/user123/bookings", nil)
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

func TestGetUserBookings_InternalError(t *testing.T) {
	app, _, bookingUseCase, _, _ := setupTestApp(t)

	bookingUseCase.On("GetUserBookings", mock.Anything, "user123").Return([]*domain.Booking{}, errors.New("database error"))

	req := httptest.NewRequest(http.MethodGet, "/api/v1/users/user123/bookings", nil)
	resp, err := app.Test(req)
	require.NoError(t, err)
	assert.Equal(t, http.StatusInternalServerError, resp.StatusCode)

	var respBody map[string]string
	err = json.NewDecoder(resp.Body).Decode(&respBody)
	require.NoError(t, err)
	assert.Equal(t, common.ErrInternalServer, respBody["error"])

	bookingUseCase.AssertExpectations(t)
}

func TestGetUserNotifications_Success(t *testing.T) {
	app, _, _, notificationUseCase, _ := setupTestApp(t)

	currentTime := time.Now()
	notifications := []domain.Notification{
		{
			ID:            "notification1",
			RecipientType: domain.RecipientTypeUser,
			RecipientID:   "user123",
			Type:          domain.NotificationTypeBookingConfirmed,
			Title:         "Booking confirmed",
			Message:       "Your booking has been confirmed",
			IsRead:        false,
			RelatedID:     "booking1",
			CreatedAt:     currentTime,
		},
		{
			ID:            "notification2",
			RecipientType: domain.RecipientTypeUser,
			RecipientID:   "user123",
			Type:          domain.NotificationTypeBookingRejected,
			Title:         "Booking rejected",
			Message:       "Your booking has been rejected",
			IsRead:        true,
			RelatedID:     "booking2",
			CreatedAt:     currentTime.Add(-time.Hour),
		},
	}

	notificationUseCase.On("GetUserNotifications", mock.Anything, "user123").Return(notifications, nil)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/users/user123/notifications", nil)
	resp, err := app.Test(req)
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var respNotifications []domain.Notification
	err = json.NewDecoder(resp.Body).Decode(&respNotifications)
	require.NoError(t, err)
	assert.Len(t, respNotifications, 2)
	assert.Equal(t, notifications[0].ID, respNotifications[0].ID)
	assert.Equal(t, notifications[1].ID, respNotifications[1].ID)

	notificationUseCase.AssertExpectations(t)
}

func TestGetUserNotifications_InternalError(t *testing.T) {
	app, _, _, notificationUseCase, _ := setupTestApp(t)

	notificationUseCase.On("GetUserNotifications", mock.Anything, "user123").Return([]domain.Notification{}, errors.New("database error"))

	req := httptest.NewRequest(http.MethodGet, "/api/v1/users/user123/notifications", nil)
	resp, err := app.Test(req)
	require.NoError(t, err)
	assert.Equal(t, http.StatusInternalServerError, resp.StatusCode)

	var respBody map[string]string
	err = json.NewDecoder(resp.Body).Decode(&respBody)
	require.NoError(t, err)
	assert.Equal(t, common.ErrInternalServer, respBody["error"])

	notificationUseCase.AssertExpectations(t)
}
