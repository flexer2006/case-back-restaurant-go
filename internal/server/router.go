package server

import (
	"github.com/flexer2006/case-back-restaurant-go/internal/server/handlers"

	"github.com/gofiber/fiber/v3"
)

type Router struct {
	restaurantHandler *handlers.RestaurantHandler
	bookingHandler    *handlers.BookingHandler
	userHandler       *handlers.UserHandler
	factsHandler      *handlers.FactsHandler
}

func NewRouter() *Router {
	return &Router{}
}

func (r *Router) SetHandlers(
	restaurantHandler *handlers.RestaurantHandler,
	bookingHandler *handlers.BookingHandler,
	userHandler *handlers.UserHandler,
	factsHandler *handlers.FactsHandler,
) {
	r.restaurantHandler = restaurantHandler
	r.bookingHandler = bookingHandler
	r.userHandler = userHandler
	r.factsHandler = factsHandler
}

func (r *Router) RegisterRoutes(app *fiber.App) {
	api := app.Group("/api/v1")

	// Настройка Swagger
	app.Get("/swagger/*", func(c fiber.Ctx) error {
		return c.SendFile("./docs/swagger.json")
	})

	app.Get("/health", func(c fiber.Ctx) error {
		return c.Status(fiber.StatusOK).JSON(fiber.Map{
			"status": "ok",
		})
	})

	app.Get("/swagger.json", func(c fiber.Ctx) error {
		return c.SendFile("./docs/swagger.json")
	})

	app.Get("/swagger-ui", func(c fiber.Ctx) error {
		return c.SendFile("./docs/static/swagger-ui.html")
	})

	restaurants := api.Group("/restaurants")
	restaurants.Get("/", r.restaurantHandler.ListRestaurants)
	restaurants.Post("/", r.restaurantHandler.CreateRestaurant)
	restaurants.Get("/:id", r.restaurantHandler.GetRestaurant)
	restaurants.Put("/:id", r.restaurantHandler.UpdateRestaurant)
	restaurants.Delete("/:id", r.restaurantHandler.DeleteRestaurant)
	restaurants.Post("/:id/facts", r.restaurantHandler.AddFact)
	restaurants.Get("/:id/facts", r.restaurantHandler.GetFacts)
	restaurants.Post("/:id/working-hours", r.restaurantHandler.SetWorkingHours)
	restaurants.Get("/:id/working-hours", r.restaurantHandler.GetWorkingHours)
	restaurants.Post("/:id/availability", r.restaurantHandler.SetAvailability)
	restaurants.Get("/:id/availability", r.restaurantHandler.GetAvailability)
	restaurants.Get("/:id/bookings", r.restaurantHandler.GetRestaurantBookings)

	bookings := api.Group("/bookings")
	bookings.Post("/", r.bookingHandler.CreateBooking)
	bookings.Get("/:id", r.bookingHandler.GetBooking)
	bookings.Post("/:id/confirm", r.bookingHandler.ConfirmBooking)
	bookings.Post("/:id/reject", r.bookingHandler.RejectBooking)
	bookings.Post("/:id/cancel", r.bookingHandler.CancelBooking)
	bookings.Post("/:id/complete", r.bookingHandler.CompleteBooking)
	bookings.Post("/:id/alternative", r.bookingHandler.SuggestAlternativeTime)
	bookings.Post("/alternatives/:id/accept", r.bookingHandler.AcceptAlternative)
	bookings.Post("/alternatives/:id/reject", r.bookingHandler.RejectAlternative)

	users := api.Group("/users")
	users.Post("/", r.userHandler.CreateUser)
	users.Get("/:id", r.userHandler.GetUser)
	users.Put("/:id", r.userHandler.UpdateUser)
	users.Get("/:id/bookings", r.userHandler.GetUserBookings)
	users.Get("/:id/notifications", r.userHandler.GetUserNotifications)

	facts := api.Group("/facts")
	facts.Get("/random", r.factsHandler.GetRandomFacts)

}
