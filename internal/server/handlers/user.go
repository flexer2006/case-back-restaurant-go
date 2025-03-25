package handlers

import (
	"errors"

	"github.com/flexer2006/case-back-restaurant-go/common"
	"github.com/flexer2006/case-back-restaurant-go/internal/domain"
	"github.com/flexer2006/case-back-restaurant-go/pkg/usecase"

	"github.com/gofiber/fiber/v3"
	"go.uber.org/zap"
)

type UserHandler struct {
	userUseCase         usecase.UserUseCase
	bookingUseCase      usecase.BookingUseCase
	notificationUseCase usecase.NotificationUseCase
}

func NewUserHandler(
	userUseCase usecase.UserUseCase,
	bookingUseCase usecase.BookingUseCase,
	notificationUseCase usecase.NotificationUseCase,
) *UserHandler {
	return &UserHandler{
		userUseCase:         userUseCase,
		bookingUseCase:      bookingUseCase,
		notificationUseCase: notificationUseCase,
	}
}

type CreateUserRequest struct {
	Name  string `json:"name"  validate:"required"`
	Email string `json:"email" validate:"required,email"`
	Phone string `json:"phone" validate:"required"`
}

// CreateUser godoc
// @Summary Create user
// @Description Create a new user
// @Tags users
// @Accept json
// @Produce json
// @Param user body CreateUserRequest true "User data"
// @Success 201 {object} domain.User
// @Failure 400 {object} map[string]string "Invalid data"
// @Failure 409 {object} map[string]string "Email already exists"
// @Failure 500 {object} map[string]string
// @Router /users [post]
func (h *UserHandler) CreateUser(c fiber.Ctx) error {
	ctx, log, err := getContextAndLogger(c)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": common.ErrInternalServer,
		})
	}

	var request CreateUserRequest
	if err := c.Bind().Body(&request); err != nil {
		log.Error(ctx, common.ErrParseRequestBody, zap.Error(err))

		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": common.ErrInvalidParams,
		})
	}

	user := &domain.User{
		Name:  request.Name,
		Email: request.Email,
		Phone: request.Phone,
	}

	userID, err := h.userUseCase.CreateUser(ctx, user)
	if err != nil {
		if errors.Is(err, usecase.ErrEmailExists) {
			return c.Status(fiber.StatusConflict).JSON(fiber.Map{
				"error": common.ErrEmailAlreadyExistsMsg,
			})
		}

		log.Error(ctx, common.ErrCreateUserHandler, zap.Error(err))

		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": common.ErrInternalServer,
		})
	}

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"id": userID,
	})
}

// GetUser godoc
// @Summary Get user
// @Description Get user by ID
// @Tags users
// @Accept json
// @Produce json
// @Param id path string true "User ID"
// @Success 200 {object} domain.User
// @Failure 400 {object} map[string]string
// @Failure 404 {object} map[string]string "User not found"
// @Failure 500 {object} map[string]string
// @Router /users/{id} [get]
func (h *UserHandler) GetUser(c fiber.Ctx) error {
	ctx, log, err := getContextAndLogger(c)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": common.ErrInternalServer,
		})
	}

	id := c.Params("id")
	if id == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": common.ErrInvalidParams,
		})
	}

	user, err := h.userUseCase.GetUser(ctx, id)
	if err != nil {
		log.Error(ctx, common.ErrGetUserHandler, zap.String("id", id), zap.Error(err))

		if err.Error() == common.ErrUserNotFound {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"error": common.ErrUserNotFound,
			})
		}

		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": common.ErrInternalServer,
		})
	}

	if user == nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": common.ErrUserNotFound,
		})
	}

	return c.Status(fiber.StatusOK).JSON(user)
}

type UpdateUserRequest struct {
	Name  string `json:"name"  validate:"required"`
	Email string `json:"email" validate:"required,email"`
	Phone string `json:"phone" validate:"required"`
}

// UpdateUser godoc
// @Summary Update user
// @Description Update an existing user
// @Tags users
// @Accept json
// @Produce json
// @Param id path string true "User ID"
// @Param user body UpdateUserRequest true "User data"
// @Success 200 {object} domain.User
// @Failure 400 {object} map[string]string "Invalid data"
// @Failure 404 {object} map[string]string "User not found"
// @Failure 409 {object} map[string]string "Email already exists"
// @Failure 500 {object} map[string]string
// @Router /users/{id} [put]
func (h *UserHandler) UpdateUser(c fiber.Ctx) error {
	ctx, log, err := getContextAndLogger(c)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": common.ErrInternalServer,
		})
	}

	id := c.Params("id")
	if id == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": common.ErrInvalidParams,
		})
	}

	var request UpdateUserRequest
	if err := c.Bind().Body(&request); err != nil {
		log.Error(ctx, common.ErrParseRequestBody, zap.Error(err))

		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": common.ErrInvalidParams,
		})
	}

	user := &domain.User{
		ID:    id,
		Name:  request.Name,
		Email: request.Email,
		Phone: request.Phone,
	}

	if err := h.userUseCase.UpdateUser(ctx, user); err != nil {
		if errors.Is(err, usecase.ErrUserNotFound) {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"error": common.ErrUserNotFound,
			})
		}

		if errors.Is(err, usecase.ErrEmailExists) {
			return c.Status(fiber.StatusConflict).JSON(fiber.Map{
				"error": common.ErrEmailAlreadyExistsMsg,
			})
		}

		log.Error(ctx, common.ErrUpdateUserHandler, zap.String("id", id), zap.Error(err))

		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": common.ErrInternalServer,
		})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"status": common.MsgSuccess,
	})
}

// GetUserBookings godoc
// @Summary Get user bookings
// @Description Get all bookings of a user
// @Tags users,bookings
// @Accept json
// @Produce json
// @Param id path string true "User ID"
// @Success 200 {array} domain.Booking
// @Failure 400 {object} map[string]string
// @Failure 404 {object} map[string]string "User not found"
// @Failure 500 {object} map[string]string
// @Router /users/{id}/bookings [get]
func (h *UserHandler) GetUserBookings(c fiber.Ctx) error {
	ctx, log, err := getContextAndLogger(c)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": common.ErrInternalServer,
		})
	}

	id := c.Params("id")
	if id == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": common.ErrInvalidParams,
		})
	}

	bookings, err := h.bookingUseCase.GetUserBookings(ctx, id)
	if err != nil {
		log.Error(ctx, common.ErrGetUserBookings, zap.String("userID", id), zap.Error(err))

		if err.Error() == common.ErrUserNotFound {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"error": common.ErrUserNotFound,
			})
		}

		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": common.ErrInternalServer,
		})
	}

	return c.Status(fiber.StatusOK).JSON(bookings)
}

// GetUserNotifications godoc
// @Summary Get user notifications
// @Description Get all notifications of a user
// @Tags users,notifications
// @Accept json
// @Produce json
// @Param id path string true "User ID"
// @Success 200 {array} domain.Notification
// @Failure 400 {object} map[string]string
// @Failure 404 {object} map[string]string "User not found"
// @Failure 500 {object} map[string]string
// @Router /users/{id}/notifications [get]
func (h *UserHandler) GetUserNotifications(c fiber.Ctx) error {
	ctx, log, err := getContextAndLogger(c)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": common.ErrInternalServer,
		})
	}

	id := c.Params("id")
	if id == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": common.ErrInvalidParams,
		})
	}

	notifications, err := h.notificationUseCase.GetUserNotifications(ctx, id)
	if err != nil {
		log.Error(ctx, common.ErrGetUserNotifications, zap.String("userID", id), zap.Error(err))

		if err.Error() == common.ErrUserNotFound {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"error": common.ErrUserNotFound,
			})
		}

		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": common.ErrInternalServer,
		})
	}

	return c.Status(fiber.StatusOK).JSON(notifications)
}
