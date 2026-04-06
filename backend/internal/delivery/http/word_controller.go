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

func NewWordController(useCase *usecase.WordUseCase, log *logrus.Logger) *WordController {
	return &WordController{Log: log, UseCase: useCase}
}

func (c *WordController) Detail(ctx *fiber.Ctx) error {
	id := ctx.Params("id")
	lang := ctx.Query("lang")
	response, err := c.UseCase.Detail(ctx.UserContext(), id, lang)
	if err != nil {
		return err
	}
	return ctx.JSON(model.WebResponse[*model.WordDetailResponse]{Data: response})
}
