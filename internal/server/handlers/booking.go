// Package handlers contain HTTP handlers for the REST API.
package handlers

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/flexer2006/case-back-restaurant-go/common"
	"github.com/flexer2006/case-back-restaurant-go/internal/domain"
	"github.com/flexer2006/case-back-restaurant-go/internal/logger"
	"github.com/flexer2006/case-back-restaurant-go/internal/logger/ports"
	"github.com/flexer2006/case-back-restaurant-go/pkg/usecase"

	"github.com/gofiber/fiber/v3"
	"go.uber.org/zap"
)

var (
	ErrContextNotFound = errors.New("context not found")
	ErrLoggerNotFound  = errors.New("logger not found")
)

type BookingHandler struct {
	bookingUseCase usecase.BookingUseCase
}

func NewBookingHandler(bookingUseCase usecase.BookingUseCase) *BookingHandler {
	return &BookingHandler{
		bookingUseCase: bookingUseCase,
	}
}

type CreateBookingRequest struct {
	RestaurantID string    `json:"restaurant_id" validate:"required"`
	UserID       string    `json:"user_id" validate:"required"`
	Date         time.Time `json:"date" validate:"required"`
	Time         string    `json:"time" validate:"required"`
	Duration     int       `json:"duration" validate:"required,min=30"`
	GuestsCount  int       `json:"guests_count" validate:"required,min=1"`
	Comment      string    `json:"comment"`
}

func getContextAndLogger(c fiber.Ctx) (context.Context, ports.LoggerPort, error) {
	ctxValue := c.Locals("ctx")
	ctx, ok := ctxValue.(context.Context)
	if !ok {
		return nil, nil, ErrContextNotFound
	}

	log, err := logger.FromContext(ctx)
	if err != nil {
		return ctx, nil, fmt.Errorf("%w: %s", ErrLoggerNotFound, err.Error())
	}

	return ctx, log, nil
}

// CreateBooking godoc
// @Summary Create booking
// @Description Create a new booking for a restaurant
// @Tags bookings
// @Accept json
// @Produce json
// @Param booking body CreateBookingRequest true "Booking data"
// @Success 201 {object} domain.Booking
// @Failure 400 {object} map[string]string
// @Failure 404 {object} map[string]string "Restaurant or user not found"
// @Failure 422 {object} map[string]string "Not enough seats at the specified time"
// @Failure 500 {object} map[string]string
// @Router /bookings [post]
func (h *BookingHandler) CreateBooking(c fiber.Ctx) error {
	ctx, log, err := getContextAndLogger(c)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": common.ErrInternalServer,
		})
	}

	var request CreateBookingRequest
	if err := c.Bind().Body(&request); err != nil {
		log.Error(ctx, common.ErrParseRequestBody, zap.Error(err))

		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": common.ErrInvalidParams,
		})
	}

	booking := &domain.Booking{
		RestaurantID: request.RestaurantID,
		UserID:       request.UserID,
		Date:         request.Date,
		Time:         request.Time,
		Duration:     request.Duration,
		GuestsCount:  request.GuestsCount,
		Comment:      request.Comment,
		Status:       domain.BookingStatusPending,
	}

	bookingID, err := h.bookingUseCase.CreateBooking(ctx, booking)
	if err != nil {
		log.Error(ctx, common.ErrCreateBooking, zap.Error(err))

		if err.Error() == common.ErrRestaurantNotFound {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"error": common.ErrRestaurantNotFound,
			})
		}

		if err.Error() == common.ErrUserNotFound {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"error": common.ErrUserNotFound,
			})
		}

		if err.Error() == common.ErrInsufficientCapacity {
			return c.Status(fiber.StatusUnprocessableEntity).JSON(fiber.Map{
				"error": common.ErrInsufficientCapacity,
			})
		}

		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": common.ErrInternalServer,
		})
	}

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"id": bookingID,
	})
}

// GetBooking godoc
// @Summary Get booking
// @Description Get detailed information about a booking by ID
// @Tags bookings
// @Accept json
// @Produce json
// @Param id path string true "Booking ID"
// @Success 200 {object} domain.Booking
// @Failure 400 {object} map[string]string
// @Failure 404 {object} map[string]string "Booking not found"
// @Failure 500 {object} map[string]string
// @Router /bookings/{id} [get]
func (h *BookingHandler) GetBooking(c fiber.Ctx) error {
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

	booking, err := h.bookingUseCase.GetBooking(ctx, id)
	if err != nil {
		log.Error(ctx, common.ErrGetBookingByID, zap.String("id", id), zap.Error(err))

		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": common.ErrInternalServer,
		})
	}

	if booking == nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": common.ErrBookingNotFound,
		})
	}

	return c.Status(fiber.StatusOK).JSON(booking)
}

// ConfirmBooking godoc
// @Summary Confirm booking
// @Description Confirm a booking by the restaurant
// @Tags bookings
// @Accept json
// @Produce json
// @Param id path string true "Booking ID"
// @Success 200 {object} domain.Booking
// @Failure 400 {object} map[string]string
// @Failure 404 {object} map[string]string "Booking not found"
// @Failure 422 {object} map[string]string "Cannot confirm booking in current status"
// @Failure 500 {object} map[string]string
// @Router /bookings/{id}/confirm [post]
func (h *BookingHandler) ConfirmBooking(c fiber.Ctx) error {
	return h.handleBookingStatusChange(c, h.bookingUseCase.ConfirmBooking, common.ErrConfirmBookingByID)
}

type RejectBookingRequest struct {
	Reason string `json:"reason" validate:"required"`
}

// RejectBooking godoc
// @Summary Reject booking
// @Description Reject a booking by the restaurant with a reason
// @Tags bookings
// @Accept json
// @Produce json
// @Param id path string true "Booking ID"
// @Param reason body RejectBookingRequest true "Rejection reason"
// @Success 200 {object} domain.Booking
// @Failure 400 {object} map[string]string
// @Failure 404 {object} map[string]string "Booking not found"
// @Failure 422 {object} map[string]string "Cannot reject booking in current status"
// @Failure 500 {object} map[string]string
// @Router /bookings/{id}/reject [post]
func (h *BookingHandler) RejectBooking(c fiber.Ctx) error {
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

	var request RejectBookingRequest
	if err := c.Bind().Body(&request); err != nil {
		log.Error(ctx, common.ErrParseRequestBody, zap.Error(err))

		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": common.ErrInvalidParams,
		})
	}

	if err := h.bookingUseCase.RejectBooking(ctx, id, request.Reason); err != nil {
		log.Error(ctx, common.ErrRejectBookingByID, zap.String("id", id), zap.Error(err))

		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": common.ErrInternalServer,
		})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"status": common.MsgSuccess,
	})
}

// CancelBooking godoc
// @Summary Cancel booking
// @Description Cancel a booking by the user
// @Tags bookings
// @Accept json
// @Produce json
// @Param id path string true "Booking ID"
// @Success 200 {object} domain.Booking
// @Failure 400 {object} map[string]string
// @Failure 404 {object} map[string]string "Booking not found"
// @Failure 422 {object} map[string]string "Cannot cancel booking in current status"
// @Failure 500 {object} map[string]string
// @Router /bookings/{id}/cancel [post]
func (h *BookingHandler) CancelBooking(c fiber.Ctx) error {
	return h.handleBookingStatusChange(c, h.bookingUseCase.CancelBooking, common.ErrCancelBookingByID)
}

// CompleteBooking godoc
// @Summary Complete booking
// @Description Mark a booking as completed
// @Tags bookings
// @Accept json
// @Produce json
// @Param id path string true "Booking ID"
// @Success 200 {object} domain.Booking
// @Failure 400 {object} map[string]string
// @Failure 404 {object} map[string]string "Booking not found"
// @Failure 422 {object} map[string]string "Cannot complete booking in current status"
// @Failure 500 {object} map[string]string
// @Router /bookings/{id}/complete [post]
func (h *BookingHandler) CompleteBooking(c fiber.Ctx) error {
	return h.handleBookingStatusChange(c, h.bookingUseCase.CompleteBooking, common.ErrCompleteBookingByID)
}

func (h *BookingHandler) handleBookingStatusChange(
	c fiber.Ctx,
	action func(ctx context.Context, id string) error,
	errMsg string,
) error {
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

	if err := action(ctx, id); err != nil {
		log.Error(ctx, errMsg, zap.String("id", id), zap.Error(err))

		if err.Error() == common.ErrBookingNotFound {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"error": common.ErrBookingNotFound,
			})
		}

		if err.Error() == common.ErrInvalidBookingStatus {
			return c.Status(fiber.StatusUnprocessableEntity).JSON(fiber.Map{
				"error": common.ErrInvalidBookingStatus,
			})
		}

		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": common.ErrInternalServer,
		})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"status": common.MsgSuccess,
	})
}

type SuggestAlternativeTimeRequest struct {
	Date    time.Time `json:"date" validate:"required"`
	Time    string    `json:"time" validate:"required"`
	Message string    `json:"message"`
}

// SuggestAlternativeTime godoc
// @Summary Suggest alternative time
// @Description Restaurant suggests an alternative time for a booking
// @Tags bookings
// @Accept json
// @Produce json
// @Param id path string true "Booking ID"
// @Param alternative_time body SuggestAlternativeTimeRequest true "Alternative time data"
// @Success 201 {object} domain.Booking
// @Failure 400 {object} map[string]string
// @Failure 404 {object} map[string]string "Booking not found"
// @Failure 422 {object} map[string]string "Cannot suggest alternative time in current status"
// @Failure 500 {object} map[string]string
// @Router /bookings/{id}/alternative [post]
func (h *BookingHandler) SuggestAlternativeTime(c fiber.Ctx) error {
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

	var request SuggestAlternativeTimeRequest
	if err := c.Bind().Body(&request); err != nil {
		log.Error(ctx, common.ErrParseRequestBody, zap.Error(err))

		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": common.ErrInvalidParams,
		})
	}

	alternativeID, err := h.bookingUseCase.SuggestAlternativeTime(ctx, id, request.Date, request.Time, request.Message)
	if err != nil {
		log.Error(ctx, common.ErrSuggestAlternativeTime,
			zap.String("bookingID", id),
			zap.Time("date", request.Date),
			zap.String("time", request.Time),
			zap.Error(err))

		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": common.ErrInternalServer,
		})
	}

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"id": alternativeID,
	})
}

// AcceptAlternative godoc
// @Summary Accept alternative
// @Description User accepts the suggested alternative time
// @Tags bookings
// @Accept json
// @Produce json
// @Param id path string true "Alternative ID"
// @Success 200 {object} domain.Booking
// @Failure 400 {object} map[string]string
// @Failure 404 {object} map[string]string "Alternative not found"
// @Failure 500 {object} map[string]string
// @Router /bookings/alternatives/{id}/accept [post]
func (h *BookingHandler) AcceptAlternative(c fiber.Ctx) error {
	return h.handleAlternativeAction(c, h.bookingUseCase.AcceptAlternative, common.ErrAcceptAlternative)
}

// RejectAlternative godoc
// @Summary Reject alternative
// @Description User rejects the suggested alternative time
// @Tags bookings
// @Accept json
// @Produce json
// @Param id path string true "Alternative ID"
// @Success 200 {object} domain.Booking
// @Failure 400 {object} map[string]string
// @Failure 404 {object} map[string]string "Alternative not found"
// @Failure 500 {object} map[string]string
// @Router /bookings/alternatives/{id}/reject [post]
func (h *BookingHandler) RejectAlternative(c fiber.Ctx) error {
	return h.handleAlternativeAction(c, h.bookingUseCase.RejectAlternative, common.ErrRejectAlternative)
}

// handleAlternativeAction обрабатывает действия с альтернативными предложениями.
func (h *BookingHandler) handleAlternativeAction(
	c fiber.Ctx,
	action func(ctx context.Context, id string) error,
	errMsg string,
) error {
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

	if err := action(ctx, id); err != nil {
		log.Error(ctx, errMsg, zap.String("alternativeID", id), zap.Error(err))

		if err.Error() == common.ErrAlternativeNotFound {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"error": common.ErrAlternativeNotFound,
			})
		}

		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": common.ErrInternalServer,
		})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"status": common.MsgSuccess,
	})
}
