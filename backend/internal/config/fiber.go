package config

import (
	"github.com/gofiber/fiber/v2"
	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

func NewFiber(v *viper.Viper, log *logrus.Logger) *fiber.App {
	app := fiber.New(fiber.Config{
		AppName:      v.GetString("app.name"),
		Prefork:      v.GetBool("web.prefork"),
		ErrorHandler: newErrorHandler(log),
	})
	return app
}

func newErrorHandler(log *logrus.Logger) fiber.ErrorHandler {
	return func(ctx *fiber.Ctx, err error) error {
		code := fiber.StatusInternalServerError
		if e, ok := err.(*fiber.Error); ok {
			code = e.Code
		} else {
			log.Warnf("Unhandled error : %+v", err)
		}
		return ctx.Status(code).JSON(fiber.Map{"errors": err.Error()})
	}
}
