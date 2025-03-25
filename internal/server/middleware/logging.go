package middleware

import (
	"context"

	"github.com/flexer2006/case-back-restaurant-go/common"
	"github.com/flexer2006/case-back-restaurant-go/internal/logger"
	"github.com/flexer2006/case-back-restaurant-go/internal/logger/utils"

	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

func LoggingMiddleware() fiber.Handler {
	return func(c fiber.Ctx) error {
		ctx := context.Background()

		requestID := uuid.New().String()
		ctx = context.WithValue(ctx, utils.RequestID, requestID)

		log, err := logger.FromContext(ctx)
		if err != nil {
			log, err = logger.NewLogger()
			if err != nil {
				return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
					"error": common.ErrInternalServer,
				})
			}
			ctx = logger.NewContext(ctx, log)
		}

		log.Info(ctx, common.MsgIncomingRequest,
			zap.String("method", c.Method()),
			zap.String("url", c.OriginalURL()))

		c.Locals("ctx", ctx)

		return c.Next()
	}
}
