package config

import (
	"words-app/internal/delivery/http"
	"words-app/internal/delivery/http/route"
	"words-app/internal/repository"
	"words-app/internal/usecase"

	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

type BootstrapConfig struct {
	DB        *gorm.DB
	App       *fiber.App
	Log       *logrus.Logger
	Validate  *validator.Validate
	StaticDir string
}

func Bootstrap(c *BootstrapConfig) {
	// Repositories
	wordRepository := repository.NewWordRepository(c.Log)

	// UseCases
	helloUseCase := usecase.NewHelloUseCase(c.Log)
	wordUseCase := usecase.NewWordUseCase(c.DB, c.Log, c.Validate, wordRepository)

	// Controllers
	helloController := http.NewHelloController(helloUseCase, c.Log)
	wordController := http.NewWordController(wordUseCase, c.Log)

	// Routes
	routeConfig := route.RouteConfig{
		App:             c.App,
		HelloController: helloController,
		WordController:  wordController,
		StaticDir:       c.StaticDir,
	}
	routeConfig.Setup()
}
