package main

// @title Restaurant Booking API
// @version 1.0
// @description API для системы бронирования ресторанов
// @termsOfService http://swagger.io/terms/

// @contact.name API Support
// @contact.email support@restaurant-booking.com

// @license.name Apache 2.0
// @license.url http://www.apache.org/licenses/LICENSE-2.0.html

// @host localhost:8080
// @BasePath /api/v1
// @schemes http https

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"go.uber.org/zap"

	"github.com/flexer2006/case-back-restaurant-go/common"
	"github.com/flexer2006/case-back-restaurant-go/configs"
	pgdb "github.com/flexer2006/case-back-restaurant-go/db/postgres"
	"github.com/flexer2006/case-back-restaurant-go/internal/logger"
	"github.com/flexer2006/case-back-restaurant-go/internal/logger/ports"
	"github.com/flexer2006/case-back-restaurant-go/internal/repository/postgres"
	"github.com/flexer2006/case-back-restaurant-go/internal/server"
	"github.com/flexer2006/case-back-restaurant-go/pkg/usecase"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go catchSignals(cancel)

	if err := run(ctx); err != nil {
		if _, printErr := fmt.Fprintf(os.Stderr, common.ErrAppStartup+": %v\n", err); printErr != nil {
			log.Printf(common.ErrAppStartup+": %v ("+common.ErrWriteToStderr+": %v)", err, printErr)
		}

		os.Exit(1)
	}
}

func catchSignals(cancel context.CancelFunc) {
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
	<-sigCh
	cancel()
}

func run(ctx context.Context) error {
	zapLogger, err := logger.NewLogger()
	if err != nil {
		return fmt.Errorf(common.ErrInitLogger+": %w", err)
	}

	ctx = logger.NewContext(ctx, zapLogger)

	cfg, err := configs.Load(ctx)
	if err != nil {
		zapLogger.Fatal(ctx, common.ErrConfigLoad, zap.Error(err))

		return err
	}

	zapLogger.Info(ctx, common.MsgConnectingToPostgres)

	db, err := pgdb.New(ctx, &cfg.Database)
	if err != nil {
		zapLogger.Fatal(ctx, common.ErrPostgresConnect, zap.Error(err))

		return err
	}

	defer closeDB(ctx, zapLogger, db)

	if err := db.Ping(ctx); err != nil {
		zapLogger.Fatal(ctx, common.ErrPingPostgresPool, zap.Error(err))

		return err
	}

	zapLogger.Info(ctx, common.MsgPostgresConnected)

	useCases, err := setupUseCases(db)
	if err != nil {
		return err
	}

	srv, err := server.NewServer(
		ctx,
		cfg,
		useCases.restaurant,
		useCases.booking,
		useCases.user,
		useCases.facts,
		useCases.availability,
		useCases.notification,
	)
	if err != nil {
		zapLogger.Fatal(ctx, common.ErrCreateServer, zap.Error(err))

		return err
	}

	err = srv.Start(ctx)
	if err != nil {
		return fmt.Errorf("%w", err)
	}
	return nil
}

type useCases struct {
	restaurant   usecase.RestaurantUseCase
	booking      usecase.BookingUseCase
	user         usecase.UserUseCase
	facts        usecase.FactsUseCase
	availability usecase.AvailabilityUseCase
	notification usecase.NotificationUseCase
}

func setupUseCases(db pgdb.Database) (*useCases, error) {
	repoFactory := postgres.NewRepositoryFactory(db)

	restaurantRepo := repoFactory.Restaurant()
	workingHoursRepo := repoFactory.WorkingHours()
	availabilityRepo := repoFactory.Availability()
	bookingRepo := repoFactory.Booking()
	userRepo := repoFactory.User()
	notificationRepo := repoFactory.Notification()

	notificationService := postgres.NewNotificationService(notificationRepo)

	// Using mock email service
	// smtpConfig, err := configs.NewSMTPConfig()
	// if err != nil {
	// 	return nil, fmt.Errorf("%s: %w", common.ErrSMTPInvalidConfig, err)
	// }

	// emailService := notification.NewSMTPMailer(smtpConfig)
	emailService := postgres.NewMockEmailService()

	return &useCases{
		restaurant:   usecase.NewRestaurantUseCase(restaurantRepo, workingHoursRepo),
		facts:        usecase.NewFactsUseCase(restaurantRepo),
		availability: usecase.NewAvailabilityUseCase(availabilityRepo, restaurantRepo, workingHoursRepo),
		notification: usecase.NewNotificationUseCase(emailService, notificationService),
		booking:      usecase.NewBookingUseCase(bookingRepo, availabilityRepo, notificationService),
		user:         usecase.NewUserUseCase(userRepo),
	}, nil
}

func closeDB(ctx context.Context, log ports.LoggerPort, db pgdb.Database) {
	log.Info(ctx, common.MsgClosingPostgresPool)

	if err := db.Close(ctx); err != nil {
		log.Error(ctx, common.ErrDBClose, zap.Error(err))
	}
}
