package middleware_test

import (
	"context"
	"io"
	"net/http"
	"testing"

	"github.com/flexer2006/case-back-restaurant-go/internal/logger"
	"github.com/flexer2006/case-back-restaurant-go/internal/logger/utils"
	"github.com/flexer2006/case-back-restaurant-go/internal/server/middleware"

	"github.com/gofiber/fiber/v3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLoggingMiddleware(t *testing.T) {
	app := fiber.New()

	app.Use(middleware.LoggingMiddleware())

	app.Get("/test", func(c fiber.Ctx) error {
		ctx := c.Locals("ctx")
		if ctx == nil {
			return c.Status(500).SendString("context not found")
		}

		ctxTyped, ok := ctx.(context.Context)
		if !ok {
			return c.Status(500).SendString("context type assertion failed")
		}

		requestID, ok := utils.GetRequestID(ctxTyped)
		if !ok || requestID == "" {
			return c.Status(500).SendString("request id not found")
		}

		return c.SendString("ok")
	})

	tests := []struct {
		name           string
		route          string
		expectedStatus int
		expectedBody   string
	}{
		{
			name:           "should add context to locals",
			route:          "/test",
			expectedStatus: http.StatusOK,
			expectedBody:   "ok",
		},
		{
			name:           "should handle non-existent route",
			route:          "/not-found",
			expectedStatus: http.StatusNotFound,
			expectedBody:   "Cannot GET /not-found",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req, err := http.NewRequest(http.MethodGet, tt.route, nil)
			require.NoError(t, err)

			resp, err := app.Test(req)
			require.NoError(t, err)

			assert.Equal(t, tt.expectedStatus, resp.StatusCode)

			body, err := io.ReadAll(resp.Body)
			require.NoError(t, err)

			assert.Contains(t, string(body), tt.expectedBody)
		})
	}
}

func TestLoggingMiddlewareContextPropagation(t *testing.T) {
	app := fiber.New()

	app.Use(middleware.LoggingMiddleware())

	app.Get("/check-logger", func(c fiber.Ctx) error {
		ctx := c.Locals("ctx")
		require.NotNil(t, ctx, "context should not be nil")

		ctxTyped, ok := ctx.(context.Context)
		if !ok {
			return c.Status(500).SendString("context type assertion failed")
		}

		log, err := logger.FromContext(ctxTyped)
		if err != nil {
			return c.Status(500).SendString("logger not found in context")
		}

		log.Info(ctxTyped, "logger successfully retrieved from context")

		return c.SendString("logger found")
	})

	req, err := http.NewRequest(http.MethodGet, "/check-logger", nil)
	require.NoError(t, err)

	resp, err := app.Test(req)
	require.NoError(t, err)

	assert.Equal(t, http.StatusOK, resp.StatusCode)

	body, err := io.ReadAll(resp.Body)
	require.NoError(t, err)

	assert.Equal(t, "logger found", string(body))
}

func TestLoggingMiddlewareWithDifferentMethods(t *testing.T) {
	app := fiber.New()

	app.Use(middleware.LoggingMiddleware())

	methods := []string{http.MethodGet, http.MethodPost, http.MethodPut, http.MethodDelete}

	for _, method := range methods {
		switch method {
		case http.MethodGet:
			app.Get("/method-test", func(c fiber.Ctx) error {
				ctx := c.Locals("ctx")
				if ctx == nil {
					return c.Status(500).SendString("context not found")
				}
				return c.SendString("method: " + method)
			})
		case http.MethodPost:
			app.Post("/method-test", func(c fiber.Ctx) error {
				ctx := c.Locals("ctx")
				if ctx == nil {
					return c.Status(500).SendString("context not found")
				}
				return c.SendString("method: " + method)
			})
		case http.MethodPut:
			app.Put("/method-test", func(c fiber.Ctx) error {
				ctx := c.Locals("ctx")
				if ctx == nil {
					return c.Status(500).SendString("context not found")
				}
				return c.SendString("method: " + method)
			})
		case http.MethodDelete:
			app.Delete("/method-test", func(c fiber.Ctx) error {
				ctx := c.Locals("ctx")
				if ctx == nil {
					return c.Status(500).SendString("context not found")
				}
				return c.SendString("method: " + method)
			})
		}
	}

	for _, method := range methods {
		t.Run("http method "+method, func(t *testing.T) {
			req, err := http.NewRequest(method, "/method-test", nil)
			require.NoError(t, err)

			resp, err := app.Test(req)
			require.NoError(t, err)

			assert.Equal(t, http.StatusOK, resp.StatusCode)

			body, err := io.ReadAll(resp.Body)
			require.NoError(t, err)

			assert.Equal(t, "method: "+method, string(body))
		})
	}
}
