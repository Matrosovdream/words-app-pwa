package http

import (
	"words-app/internal/model"
	"words-app/internal/usecase"

	"github.com/gofiber/fiber/v2"
	"github.com/sirupsen/logrus"
)

type SettingController struct {
	Log     *logrus.Logger
	UseCase *usecase.SettingUseCase
}

func NewSettingController(useCase *usecase.SettingUseCase, log *logrus.Logger) *SettingController {
	return &SettingController{Log: log, UseCase: useCase}
}

func (c *SettingController) GetTranslation(ctx *fiber.Ctx) error {
	response, err := c.UseCase.GetTranslation(ctx.UserContext())
	if err != nil {
		return err
	}
	return ctx.JSON(model.WebResponse[*model.TranslationSettingsResponse]{Data: response})
}

func (c *SettingController) UpdateTranslation(ctx *fiber.Ctx) error {
	request := new(model.UpdateTranslationSettingsRequest)
	if err := ctx.BodyParser(request); err != nil {
		c.Log.Warnf("Failed to parse body : %+v", err)
		return fiber.ErrBadRequest
	}
	response, err := c.UseCase.UpdateTranslation(ctx.UserContext(), request)
	if err != nil {
		return err
	}
	return ctx.JSON(model.WebResponse[*model.TranslationSettingsResponse]{Data: response})
}
