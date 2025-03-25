// Package server contains the HTTP server implementation of the application.
package server

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/flexer2006/case-back-restaurant-go/common"
	"github.com/flexer2006/case-back-restaurant-go/configs"
	"github.com/flexer2006/case-back-restaurant-go/internal/logger"
	"github.com/flexer2006/case-back-restaurant-go/internal/server/handlers"
	"github.com/flexer2006/case-back-restaurant-go/internal/server/middleware"
	"github.com/flexer2006/case-back-restaurant-go/pkg/usecase"

	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/middleware/cors"
	"github.com/gofiber/fiber/v3/middleware/recover"
	"go.uber.org/zap"
)

type ErrCustomServerShutdown struct {
	Err error
}

func (e *ErrCustomServerShutdown) Error() string {
	return fmt.Sprintf("%s: %v", common.ErrServerShutdownRu, e.Err)
}

func (e *ErrCustomServerShutdown) Unwrap() error {
	return e.Err
}

type ErrCustomLoggerFromContext struct {
	Err error
}

func (e *ErrCustomLoggerFromContext) Error() string {
	return fmt.Sprintf("%s: %v", common.ErrGetLoggerFromContextRu, e.Err)
}

func (e *ErrCustomLoggerFromContext) Unwrap() error {
	return e.Err
}

type Server struct {
	config *configs.Config
	app    *fiber.App
	router *Router
}

func NewServer(
	ctx context.Context,
	config *configs.Config,
	restaurantUseCase usecase.RestaurantUseCase,
	bookingUseCase usecase.BookingUseCase,
	userUseCase usecase.UserUseCase,
	factsUseCase usecase.FactsUseCase,
	availabilityUseCase usecase.AvailabilityUseCase,
	notificationUseCase usecase.NotificationUseCase,
) (*Server, error) {
	app := fiber.New(fiber.Config{
		AppName: "Restaurant Booking API",
		ErrorHandler: func(c fiber.Ctx, err error) error {
			code := fiber.StatusInternalServerError

			ctxValue, ok := c.Locals("ctx").(context.Context)
			if !ok {
				return wrapFiberError(c.Status(code).JSON(fiber.Map{
					"error": common.ErrContextNotFoundRu,
				}))
			}

			log, logErr := logger.FromContext(ctxValue)
			if logErr != nil {
				defaultLog, defLogErr := logger.NewLogger()
				if defLogErr != nil {
					return wrapFiberError(c.Status(code).JSON(fiber.Map{
						"error": fmt.Sprintf("%s: %v", common.ErrLoggerCreationRu, defLogErr),
					}))
				}
				log = defaultLog
			}

			var fiberErr *fiber.Error
			if errors.As(err, &fiberErr) {
				code = fiberErr.Code
			}

			log.Error(ctx, common.MsgHTTPError,
				zap.Error(err),
				zap.Int("status", code),
				zap.String("path", c.Path()),
				zap.String("method", c.Method()))

			return wrapFiberError(c.Status(code).JSON(fiber.Map{
				"error": err.Error(),
			}))
		},
	})

	app.Use(recover.New())
	app.Use(cors.New())
	app.Use(middleware.LoggingMiddleware())

	restaurantHandler := handlers.NewRestaurantHandler(restaurantUseCase, bookingUseCase, availabilityUseCase)
	bookingHandler := handlers.NewBookingHandler(bookingUseCase)
	userHandler := handlers.NewUserHandler(userUseCase, bookingUseCase, notificationUseCase)
	factsHandler := handlers.NewFactsHandler(factsUseCase)

	router := NewRouter()
	router.SetHandlers(restaurantHandler, bookingHandler, userHandler, factsHandler)

	s := &Server{
		config: config,
		app:    app,
		router: router,
	}

	return s, nil
}

func wrapFiberError(err error) error {
	if err != nil {
		return fmt.Errorf("ошибка HTTP: %w", err)
	}
	return nil
}

func (s *Server) RegisterRoutes() {
	s.router.RegisterRoutes(s.app)
}

func (s *Server) Start(ctx context.Context) error {
	log, err := logger.FromContext(ctx)
	if err != nil {
		return &ErrCustomLoggerFromContext{Err: err}
	}

	s.RegisterRoutes()

	serverAddr := fmt.Sprintf("%s:%d", s.config.Server.Host, s.config.Server.Port)
	go func() {
		log.Info(ctx, common.MsgServerStarting,
			zap.String("address", serverAddr))

		if err := s.app.Listen(serverAddr); err != nil {
			log.Error(ctx, common.MsgServerStartError, zap.Error(err))
			os.Exit(1)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)
	<-quit

	log.Info(ctx, common.MsgShutdownSignal)

	shutdownTimeout := s.config.Shutdown.Timeout
	ctx, cancel := context.WithTimeout(ctx, shutdownTimeout)
	defer cancel()

	log.Info(ctx, common.MsgServerShuttingDown, zap.Duration("timeout", shutdownTimeout))

	if err := s.app.ShutdownWithContext(ctx); err != nil {
		log.Error(ctx, common.MsgServerForcedShutdown, zap.Error(err))

		return &ErrCustomServerShutdown{Err: err}
	}

	log.Info(ctx, common.MsgServerGracefulStop)

	return nil
}

func (s *Server) Stop(ctx context.Context) error {
	log, err := logger.FromContext(ctx)
	if err != nil {
		return &ErrCustomLoggerFromContext{Err: err}
	}

	log.Info(ctx, common.MsgServerStopping)

	if err := s.app.ShutdownWithContext(ctx); err != nil {
		return &ErrCustomServerShutdown{Err: err}
	}

	return nil
}
