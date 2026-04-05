package http

import (
	"words-app/internal/model"
	"words-app/internal/usecase"

	"github.com/gofiber/fiber/v2"
	"github.com/sirupsen/logrus"
)

type HelloController struct {
	Log     *logrus.Logger
	UseCase *usecase.HelloUseCase
}

func NewHelloController(useCase *usecase.HelloUseCase, logger *logrus.Logger) *HelloController {
	return &HelloController{Log: logger, UseCase: useCase}
}

func (c *HelloController) Get(ctx *fiber.Ctx) error {
	response, err := c.UseCase.Greet(ctx.UserContext())
	if err != nil {
		return err
	}
	return ctx.JSON(model.WebResponse[*model.HelloResponse]{Data: response})
}
