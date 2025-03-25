// Package handlers содержит HTTP обработчики для REST API.
package handlers

import (
	"strconv"

	"github.com/flexer2006/case-back-restaurant-go/common"
	"github.com/flexer2006/case-back-restaurant-go/pkg/usecase"

	"github.com/gofiber/fiber/v3"
	"go.uber.org/zap"
)

type FactsHandler struct {
	factsUseCase usecase.FactsUseCase
}

func NewFactsHandler(factsUseCase usecase.FactsUseCase) *FactsHandler {
	return &FactsHandler{
		factsUseCase: factsUseCase,
	}
}

// GetRandomFacts godoc
// @Summary Get random facts
// @Description Get a collection of random facts about restaurants
// @Tags facts
// @Accept json
// @Produce json
// @Param count query int false "Number of facts to return" default(3)
// @Success 200 {array} domain.Fact
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /facts/random [get]
func (h *FactsHandler) GetRandomFacts(c fiber.Ctx) error {
	ctx, log, err := getContextAndLogger(c)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": common.ErrInternalServer,
		})
	}

	count, err := strconv.Atoi(c.Query("count", "3"))
	if err != nil || count < 1 || count > 10 {
		count = 3
	}

	facts, err := h.factsUseCase.GetRandomFacts(ctx, count)
	if err != nil {
		log.Error(ctx, common.ErrGetRandomFacts, zap.Error(err))

		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": common.ErrInternalError,
		})
	}

	return c.Status(fiber.StatusOK).JSON(facts)
}
