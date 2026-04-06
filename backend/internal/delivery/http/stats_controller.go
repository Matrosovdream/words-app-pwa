package http

import (
	"words-app/internal/model"
	"words-app/internal/usecase"

	"github.com/gofiber/fiber/v2"
	"github.com/sirupsen/logrus"
)

type StatsController struct {
	Log     *logrus.Logger
	UseCase *usecase.StatsUseCase
}

func NewStatsController(useCase *usecase.StatsUseCase, log *logrus.Logger) *StatsController {
	return &StatsController{Log: log, UseCase: useCase}
}

func (c *StatsController) ParserStats(ctx *fiber.Ctx) error {
	response, err := c.UseCase.ParserStats(ctx.UserContext())
	if err != nil {
		return err
	}
	return ctx.JSON(model.WebResponse[*model.ParserStatsResponse]{Data: response})
}
