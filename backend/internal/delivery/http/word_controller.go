package http

import (
	"words-app/internal/model"
	"words-app/internal/usecase"

	"github.com/gofiber/fiber/v2"
	"github.com/sirupsen/logrus"
)

type WordController struct {
	Log     *logrus.Logger
	UseCase *usecase.WordUseCase
}

func NewWordController(useCase *usecase.WordUseCase, logger *logrus.Logger) *WordController {
	return &WordController{Log: logger, UseCase: useCase}
}

func (c *WordController) List(ctx *fiber.Ctx) error {
	response, err := c.UseCase.List(ctx.UserContext())
	if err != nil {
		return err
	}
	return ctx.JSON(model.WebResponse[[]model.WordResponse]{Data: response})
}

func (c *WordController) Get(ctx *fiber.Ctx) error {
	request := &model.GetWordRequest{
		Slug: ctx.Params("slug"),
	}
	response, err := c.UseCase.Get(ctx.UserContext(), request)
	if err != nil {
		return err
	}
	return ctx.JSON(model.WebResponse[*model.WordResponse]{Data: response})
}
