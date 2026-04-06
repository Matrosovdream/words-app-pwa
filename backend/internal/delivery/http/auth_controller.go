package http

import (
	"words-app/internal/delivery/http/middleware"
	"words-app/internal/model"
	"words-app/internal/usecase"

	"github.com/gofiber/fiber/v2"
	"github.com/sirupsen/logrus"
)

type AuthController struct {
	Log     *logrus.Logger
	UseCase *usecase.AuthUseCase
}

func NewAuthController(useCase *usecase.AuthUseCase, log *logrus.Logger) *AuthController {
	return &AuthController{Log: log, UseCase: useCase}
}

func (c *AuthController) Login(ctx *fiber.Ctx) error {
	request := new(model.LoginRequest)
	if err := ctx.BodyParser(request); err != nil {
		c.Log.Warnf("Failed to parse login body : %+v", err)
		return fiber.ErrBadRequest
	}
	response, err := c.UseCase.Login(ctx.UserContext(), request)
	if err != nil {
		return err
	}
	return ctx.JSON(model.WebResponse[*model.LoginResponse]{Data: response})
}

func (c *AuthController) Me(ctx *fiber.Ctx) error {
	userID := middleware.GetUserID(ctx)
	if userID == "" {
		return fiber.ErrUnauthorized
	}
	response, err := c.UseCase.Me(ctx.UserContext(), userID)
	if err != nil {
		return err
	}
	return ctx.JSON(model.WebResponse[*model.UserResponse]{Data: response})
}
