package handlers

import (
	"strconv"
	"time"

	"github.com/flexer2006/case-back-restaurant-go/common"
	"github.com/flexer2006/case-back-restaurant-go/internal/domain"
	"github.com/flexer2006/case-back-restaurant-go/pkg/usecase"

	"github.com/gofiber/fiber/v3"
	"go.uber.org/zap"
)

type RestaurantHandler struct {
	restaurantUseCase   usecase.RestaurantUseCase
	bookingUseCase      usecase.BookingUseCase
	availabilityUseCase usecase.AvailabilityUseCase
}

func NewRestaurantHandler(
	restaurantUseCase usecase.RestaurantUseCase,
	bookingUseCase usecase.BookingUseCase,
	availabilityUseCase usecase.AvailabilityUseCase,
) *RestaurantHandler {
	return &RestaurantHandler{
		restaurantUseCase:   restaurantUseCase,
		bookingUseCase:      bookingUseCase,
		availabilityUseCase: availabilityUseCase,
	}
}

// ListRestaurants godoc
// @Summary List restaurants
// @Description Get a list of all restaurants with optional pagination
// @Tags restaurants
// @Accept json
// @Produce json
// @Param offset query int false "Offset" default(0)
// @Param limit query int false "Limit" default(20)
// @Success 200 {array} domain.Restaurant
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /restaurants [get]
func (h *RestaurantHandler) ListRestaurants(c fiber.Ctx) error {
	ctx, log, err := getContextAndLogger(c)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": common.ErrInternalServer,
		})
	}

	offset, err := strconv.Atoi(c.Query("offset", "0"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": common.ErrInvalidParams,
		})
	}

	limit, err := strconv.Atoi(c.Query("limit", "20"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": common.ErrInvalidParams,
		})
	}

	restaurants, err := h.restaurantUseCase.ListRestaurants(ctx, offset, limit)
	if err != nil {
		log.Error(ctx, common.ErrListRestaurants, zap.Error(err))

		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": common.ErrInternalServer,
		})
	}

	return c.Status(fiber.StatusOK).JSON(restaurants)
}

// GetRestaurant godoc
// @Summary Get restaurant
// @Description Get detailed information about a restaurant by ID
// @Tags restaurants
// @Accept json
// @Produce json
// @Param id path string true "Restaurant ID"
// @Success 200 {object} domain.Restaurant
// @Failure 400 {object} map[string]string
// @Failure 404 {object} map[string]string "Restaurant not found"
// @Failure 500 {object} map[string]string
// @Router /restaurants/{id} [get]
func (h *RestaurantHandler) GetRestaurant(c fiber.Ctx) error {
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

	restaurant, err := h.restaurantUseCase.GetRestaurant(ctx, id)
	if err != nil {
		log.Error(ctx, common.ErrGetRestaurant, zap.String("id", id), zap.Error(err))

		if err.Error() == common.ErrRestaurantNotFound {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"error": common.ErrRestaurantNotFound,
			})
		}

		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": common.ErrInternalServer,
		})
	}

	if restaurant == nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": common.ErrRestaurantNotFound,
		})
	}

	return c.Status(fiber.StatusOK).JSON(restaurant)
}

type CreateRestaurantRequest struct {
	Name         string         `json:"name"          validate:"required"`
	Address      string         `json:"address"       validate:"required"`
	Cuisine      domain.Cuisine `json:"cuisine"       validate:"required"`
	Description  string         `json:"description"`
	ContactEmail string         `json:"contact_email" validate:"required,email"`
	ContactPhone string         `json:"contact_phone" validate:"required"`
	Facts        []string       `json:"facts"`
}

// CreateRestaurant godoc
// @Summary Create restaurant
// @Description Create a new restaurant
// @Tags restaurants
// @Accept json
// @Produce json
// @Param restaurant body CreateRestaurantRequest true "Restaurant data"
// @Success 201 {object} domain.Restaurant
// @Failure 400 {object} map[string]string "Invalid data"
// @Failure 500 {object} map[string]string
// @Router /restaurants [post]
func (h *RestaurantHandler) CreateRestaurant(c fiber.Ctx) error {
	ctx, log, err := getContextAndLogger(c)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": common.ErrInternalServer,
		})
	}

	var request CreateRestaurantRequest
	if err := c.Bind().Body(&request); err != nil {
		log.Error(ctx, common.ErrParseRequestBody, zap.Error(err))

		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": common.ErrInvalidParams,
		})
	}

	restaurant := &domain.Restaurant{
		Name:         request.Name,
		Address:      request.Address,
		Cuisine:      request.Cuisine,
		Description:  request.Description,
		ContactEmail: request.ContactEmail,
		ContactPhone: request.ContactPhone,
	}

	restaurantID, err := h.restaurantUseCase.CreateRestaurant(ctx, restaurant)
	if err != nil {
		log.Error(ctx, common.ErrCreateRestaurant, zap.Error(err))

		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": common.ErrInternalServer,
		})
	}

	for _, factContent := range request.Facts {
		if _, err := h.restaurantUseCase.AddFact(ctx, restaurantID, factContent); err != nil {
			log.Warn(ctx, common.ErrAddFact,
				zap.String("restaurantID", restaurantID),
				zap.Error(err))
		}
	}

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"id": restaurantID,
	})
}

type UpdateRestaurantRequest struct {
	Name         string         `json:"name"          validate:"required"`
	Address      string         `json:"address"       validate:"required"`
	Cuisine      domain.Cuisine `json:"cuisine"       validate:"required"`
	Description  string         `json:"description"`
	ContactEmail string         `json:"contact_email" validate:"required,email"`
	ContactPhone string         `json:"contact_phone" validate:"required"`
}

// UpdateRestaurant godoc
// @Summary Update restaurant
// @Description Update an existing restaurant
// @Tags restaurants
// @Accept json
// @Produce json
// @Param id path string true "Restaurant ID"
// @Param restaurant body UpdateRestaurantRequest true "Restaurant data"
// @Success 200 {object} domain.Restaurant
// @Failure 400 {object} map[string]string "Invalid data"
// @Failure 404 {object} map[string]string "Restaurant not found"
// @Failure 500 {object} map[string]string
// @Router /restaurants/{id} [put]
func (h *RestaurantHandler) UpdateRestaurant(c fiber.Ctx) error {
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

	var request UpdateRestaurantRequest
	if err := c.Bind().Body(&request); err != nil {
		log.Error(ctx, common.ErrParseRequestBody, zap.Error(err))

		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": common.ErrInvalidParams,
		})
	}

	restaurant, err := h.restaurantUseCase.GetRestaurant(ctx, id)
	if err != nil {
		log.Error(ctx, common.ErrGetRestaurant, zap.String("id", id), zap.Error(err))

		if err.Error() == common.ErrRestaurantNotFound {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"error": common.ErrRestaurantNotFound,
			})
		}

		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": common.ErrInternalServer,
		})
	}

	if restaurant == nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": common.ErrRestaurantNotFound,
		})
	}

	restaurant.Name = request.Name
	restaurant.Address = request.Address
	restaurant.Cuisine = request.Cuisine
	restaurant.Description = request.Description
	restaurant.ContactEmail = request.ContactEmail
	restaurant.ContactPhone = request.ContactPhone

	if err := h.restaurantUseCase.UpdateRestaurant(ctx, restaurant); err != nil {
		log.Error(ctx, common.ErrUpdateRestaurant, zap.String("id", id), zap.Error(err))

		if err.Error() == common.ErrRestaurantNotFound {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"error": common.ErrRestaurantNotFound,
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

// DeleteRestaurant godoc
// @Summary Delete restaurant
// @Description Delete restaurant by ID
// @Tags restaurants
// @Accept json
// @Produce json
// @Param id path string true "Restaurant ID"
// @Success 200 {object} map[string]string
// @Failure 400 {object} map[string]string
// @Failure 404 {object} map[string]string "Restaurant not found"
// @Failure 500 {object} map[string]string
// @Router /restaurants/{id} [delete]
func (h *RestaurantHandler) DeleteRestaurant(c fiber.Ctx) error {
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

	if err := h.restaurantUseCase.DeleteRestaurant(ctx, id); err != nil {
		log.Error(ctx, common.ErrDeleteRestaurant, zap.String("id", id), zap.Error(err))

		if err.Error() == common.ErrRestaurantNotFound {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"error": common.ErrRestaurantNotFound,
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

type AddFactRequest struct {
	Content string `json:"content" validate:"required"`
}

// AddFact godoc
// @Summary Add fact
// @Description Add an interesting fact about a restaurant
// @Tags restaurants,facts
// @Accept json
// @Produce json
// @Param id path string true "Restaurant ID"
// @Param fact body AddFactRequest true "Fact content"
// @Success 201 {object} domain.Fact
// @Failure 400 {object} map[string]string "Invalid data"
// @Failure 404 {object} map[string]string "Restaurant not found"
// @Failure 500 {object} map[string]string
// @Router /restaurants/{id}/facts [post]
func (h *RestaurantHandler) AddFact(c fiber.Ctx) error {
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

	var request AddFactRequest
	if err := c.Bind().Body(&request); err != nil {
		log.Error(ctx, common.ErrParseRequestBody, zap.Error(err))

		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": common.ErrInvalidParams,
		})
	}

	restaurant, err := h.restaurantUseCase.GetRestaurant(ctx, id)
	if err != nil {
		log.Error(ctx, common.ErrGetRestaurant, zap.String("id", id), zap.Error(err))

		if err.Error() == common.ErrRestaurantNotFound {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"error": common.ErrRestaurantNotFound,
			})
		}

		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": common.ErrInternalServer,
		})
	}

	if restaurant == nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": common.ErrRestaurantNotFound,
		})
	}

	fact, err := h.restaurantUseCase.AddFact(ctx, id, request.Content)
	if err != nil {
		log.Error(ctx, common.ErrAddFact,
			zap.String("restaurantID", id),
			zap.Error(err))

		if err.Error() == common.ErrRestaurantNotFound {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"error": common.ErrRestaurantNotFound,
			})
		}

		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": common.ErrInternalServer,
		})
	}

	return c.Status(fiber.StatusCreated).JSON(fact)
}

// GetFacts godoc
// @Summary Get facts
// @Description Get all interesting facts about a restaurant
// @Tags restaurants,facts
// @Accept json
// @Produce json
// @Param id path string true "Restaurant ID"
// @Success 200 {array} domain.Fact
// @Failure 400 {object} map[string]string
// @Failure 404 {object} map[string]string "Restaurant not found"
// @Failure 500 {object} map[string]string
// @Router /restaurants/{id}/facts [get]
func (h *RestaurantHandler) GetFacts(c fiber.Ctx) error {
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

	facts, err := h.restaurantUseCase.GetFacts(ctx, id)
	if err != nil {
		log.Error(ctx, common.ErrGetFacts, zap.String("restaurantID", id), zap.Error(err))

		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": common.ErrInternalServer,
		})
	}

	return c.Status(fiber.StatusOK).JSON(facts)
}

type SetWorkingHoursRequest struct {
	WeekDay   domain.WeekDay `json:"week_day"   validate:"required"`
	OpenTime  string         `json:"open_time"  validate:"required"`
	CloseTime string         `json:"close_time" validate:"required"`
	ValidFrom time.Time      `json:"valid_from"`
	ValidTo   time.Time      `json:"valid_to"`
}

// SetWorkingHours godoc
// @Summary Set working hours
// @Description Set working hours for a restaurant
// @Tags restaurants,working-hours
// @Accept json
// @Produce json
// @Param id path string true "Restaurant ID"
// @Param working_hours body SetWorkingHoursRequest true "Working hours data"
// @Success 201 {object} domain.WorkingHours
// @Failure 400 {object} map[string]string "Invalid data"
// @Failure 404 {object} map[string]string "Restaurant not found"
// @Failure 500 {object} map[string]string
// @Router /restaurants/{id}/working-hours [post]
func (h *RestaurantHandler) SetWorkingHours(c fiber.Ctx) error {
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

	var request SetWorkingHoursRequest
	if err := c.Bind().Body(&request); err != nil {
		log.Error(ctx, common.ErrParseRequestBody, zap.Error(err))

		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": common.ErrInvalidParams,
		})
	}

	if _, err := time.Parse("15:04", request.OpenTime); err != nil {
		log.Error(ctx, common.ErrParseRequestBody, zap.String("openTime", request.OpenTime), zap.Error(err))

		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": common.ErrInvalidParams,
		})
	}

	if _, err := time.Parse("15:04", request.CloseTime); err != nil {
		log.Error(ctx, common.ErrParseRequestBody, zap.String("closeTime", request.CloseTime), zap.Error(err))

		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": common.ErrInvalidParams,
		})
	}

	if request.ValidFrom.IsZero() {
		request.ValidFrom = time.Now()
	}

	if request.ValidTo.IsZero() {
		request.ValidTo = time.Date(9999, 12, 31, 23, 59, 59, 0, time.UTC)
	}

	workingHours := &domain.WorkingHours{
		RestaurantID: id,
		WeekDay:      request.WeekDay,
		OpenTime:     request.OpenTime,
		CloseTime:    request.CloseTime,
		ValidFrom:    request.ValidFrom,
		ValidTo:      request.ValidTo,
	}

	if err := h.restaurantUseCase.SetWorkingHours(ctx, id, workingHours); err != nil {
		log.Error(ctx, common.ErrSetWorkingHours,
			zap.String("restaurantID", id),
			zap.Int("weekDay", int(request.WeekDay)),
			zap.Error(err))

		if err.Error() == common.ErrRestaurantNotFound {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"error": common.ErrRestaurantNotFound,
			})
		}

		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": common.ErrInternalServer,
		})
	}

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"status": common.MsgSuccess,
	})
}

// GetWorkingHours godoc
// @Summary Get working hours
// @Description Get working hours of a restaurant
// @Tags restaurants,working-hours
// @Accept json
// @Produce json
// @Param id path string true "Restaurant ID"
// @Success 200 {array} domain.WorkingHours
// @Failure 400 {object} map[string]string
// @Failure 404 {object} map[string]string "Restaurant not found"
// @Failure 500 {object} map[string]string
// @Router /restaurants/{id}/working-hours [get]
func (h *RestaurantHandler) GetWorkingHours(c fiber.Ctx) error {
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

	workingHours, err := h.restaurantUseCase.GetWorkingHours(ctx, id)
	if err != nil {
		log.Error(ctx, common.ErrGetWorkingHours, zap.String("restaurantID", id), zap.Error(err))

		if err.Error() == common.ErrRestaurantNotFound {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"error": common.ErrRestaurantNotFound,
			})
		}

		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": common.ErrInternalServer,
		})
	}

	return c.Status(fiber.StatusOK).JSON(workingHours)
}

type SetAvailabilityRequest struct {
	Date     time.Time `json:"date"     validate:"required"`
	TimeSlot string    `json:"time_slot" validate:"required"`
	Capacity int       `json:"capacity"  validate:"required,min=1"`
}

// SetAvailability godoc
// @Summary Set availability
// @Description Set availability for a specific date and time
// @Tags restaurants,availability
// @Accept json
// @Produce json
// @Param id path string true "Restaurant ID"
// @Param availability body SetAvailabilityRequest true "Availability data"
// @Success 201 {object} domain.Availability
// @Failure 400 {object} map[string]string "Invalid data"
// @Failure 404 {object} map[string]string "Restaurant not found"
// @Failure 500 {object} map[string]string
// @Router /restaurants/{id}/availability [post]
func (h *RestaurantHandler) SetAvailability(c fiber.Ctx) error {
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

	var request SetAvailabilityRequest
	if err := c.Bind().Body(&request); err != nil {
		log.Error(ctx, common.ErrParseRequestBody, zap.Error(err))

		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": common.ErrInvalidParams,
		})
	}

	availability := &domain.Availability{
		RestaurantID: id,
		Date:         request.Date,
		TimeSlot:     request.TimeSlot,
		Capacity:     request.Capacity,
		Reserved:     0,
	}

	log.Info(ctx, common.MsgUpdateAvailability,
		zap.String("restaurantID", id),
		zap.Time("date", availability.Date),
		zap.String("timeSlot", availability.TimeSlot),
		zap.Int("capacity", availability.Capacity))

	if err := h.availabilityUseCase.SetAvailability(ctx, availability); err != nil {
		log.Error(ctx, common.ErrUpdateAvailability,
			zap.String("restaurantID", id),
			zap.Error(err))

		if err.Error() == common.ErrRestaurantNotFound {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"error": common.ErrRestaurantNotFound,
			})
		}

		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": common.ErrInternalServer,
		})
	}

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"status": common.MsgSuccess,
	})
}

// GetAvailability godoc
// @Summary Get availability
// @Description Get availability for a restaurant on a specific date
// @Tags restaurants,availability
// @Accept json
// @Produce json
// @Param id path string true "Restaurant ID"
// @Param date query string false "Date (YYYY-MM-DD)"
// @Success 200 {array} domain.Availability
// @Failure 400 {object} map[string]string
// @Failure 404 {object} map[string]string "Restaurant not found"
// @Failure 500 {object} map[string]string
// @Router /restaurants/{id}/availability [get]
func (h *RestaurantHandler) GetAvailability(c fiber.Ctx) error {
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

	dateStr := c.Query("date")
	if dateStr == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": common.ErrInvalidParams,
		})
	}

	date, err := time.Parse("2006-01-02", dateStr)
	if err != nil {
		log.Error(ctx, common.ErrParseRequestBody, zap.String("date", dateStr), zap.Error(err))

		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": common.ErrInvalidParams,
		})
	}

	availability, err := h.availabilityUseCase.GetAvailability(ctx, id, date)
	if err != nil {
		log.Error(ctx, common.ErrGetCurrentAvailability, zap.String("restaurantID", id), zap.Error(err))

		if err.Error() == common.ErrRestaurantNotFound {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"error": common.ErrRestaurantNotFound,
			})
		}

		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": common.ErrInternalServer,
		})
	}

	return c.Status(fiber.StatusOK).JSON(availability)
}

// GetRestaurantBookings godoc
// @Summary Get restaurant bookings
// @Description Get all bookings for a specific restaurant
// @Tags restaurants,bookings
// @Accept json
// @Produce json
// @Param id path string true "Restaurant ID"
// @Param status query string false "Booking status (pending,confirmed,rejected,canceled,completed)"
// @Param date query string false "Date (YYYY-MM-DD)"
// @Success 200 {array} domain.Booking
// @Failure 400 {object} map[string]string
// @Failure 404 {object} map[string]string "Restaurant not found"
// @Failure 500 {object} map[string]string
// @Router /restaurants/{id}/bookings [get]
func (h *RestaurantHandler) GetRestaurantBookings(c fiber.Ctx) error {
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

	bookings, err := h.bookingUseCase.GetRestaurantBookings(ctx, id)
	if err != nil {
		log.Error(ctx, common.ErrGetRestaurantBookings, zap.String("restaurantID", id), zap.Error(err))

		if err.Error() == common.ErrRestaurantNotFound {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"error": common.ErrRestaurantNotFound,
			})
		}

		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": common.ErrInternalServer,
		})
	}

	return c.Status(fiber.StatusOK).JSON(bookings)
}
