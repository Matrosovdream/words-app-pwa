package config

import (
	"context"
	"time"

	"words-app/internal/delivery/http"
	"words-app/internal/delivery/http/route"
	"words-app/internal/gateway/dictionary"
	"words-app/internal/parser"
	"words-app/internal/repository"
	"words-app/internal/usecase"

	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"
	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"gorm.io/gorm"
)

type BootstrapConfig struct {
	DB        *gorm.DB
	App       *fiber.App
	Log       *logrus.Logger
	Validate  *validator.Validate
	Viper     *viper.Viper
	StaticDir string
}

func Bootstrap(c *BootstrapConfig) {
	// Repositories
	userRepo := repository.NewUserRepository(c.Log)
	siteRepo := repository.NewSiteRepository(c.Log)
	categoryRepo := repository.NewSiteCategoryRepository(c.Log)
	jobRepo := repository.NewParseJobRepository(c.Log)
	pageRepo := repository.NewParsedPageRepository(c.Log)
	dictRepo := repository.NewDictWordRepository(c.Log)
	occRepo := repository.NewWordOccurrenceRepository(c.Log)
	transRepo := repository.NewWordTranslationRepository(c.Log)
	relRepo := repository.NewWordRelationRepository(c.Log)
	reviewRepo := repository.NewReviewItemRepository(c.Log)
	learnCatRepo := repository.NewLearnCategoryRepository(c.Log)
	learnItemRepo := repository.NewLearnItemRepository(c.Log)
	learnItemCatRepo := repository.NewLearnItemCategoryRepository(c.Log)
	settingRepo := repository.NewAppSettingRepository(c.Log)
	geoRepo := repository.NewGeoRepository(c.Log)

	// Gateways
	datamuse := dictionary.NewDatamuseClient(c.Log)

	// UseCases
	helloUC := usecase.NewHelloUseCase(c.Log)
	jwtSecret := c.Viper.GetString("auth.jwt_secret")
	jwtTTL := time.Duration(c.Viper.GetInt("auth.jwt_ttl_hours")) * time.Hour
	if jwtTTL == 0 {
		jwtTTL = 168 * time.Hour
	}
	authUC := usecase.NewAuthUseCase(c.DB, c.Log, c.Validate, userRepo, jwtSecret, jwtTTL)
	siteUC := usecase.NewSiteUseCase(c.DB, c.Log, c.Validate, siteRepo, categoryRepo, jobRepo, learnCatRepo)
	reviewUC := usecase.NewReviewUseCase(c.DB, c.Log, c.Validate,
		reviewRepo, dictRepo, occRepo, pageRepo, learnItemRepo, learnCatRepo, learnItemCatRepo, categoryRepo)
	learnUC := usecase.NewLearnUseCase(c.DB, c.Log, c.Validate, learnCatRepo, learnItemRepo, learnItemCatRepo, dictRepo)
	wordUC := usecase.NewWordUseCase(c.DB, c.Log, dictRepo, transRepo, relRepo, occRepo, pageRepo, settingRepo, datamuse)
	settingUC := usecase.NewSettingUseCase(c.DB, c.Log, c.Validate, settingRepo)

	workerEnabled := c.Viper.GetBool("parser.enabled")
	pollSeconds := c.Viper.GetInt("parser.poll_interval_seconds")
	if pollSeconds <= 0 {
		pollSeconds = 5
	}
	threshold := c.Viper.GetInt("parser.common_word_threshold")
	if threshold <= 0 {
		threshold = 5000
	}
	statsUC := usecase.NewStatsUseCase(c.DB, c.Log, siteRepo, jobRepo, workerEnabled, pollSeconds, threshold)

	// Controllers
	helloC := http.NewHelloController(helloUC, c.Log)
	authC := http.NewAuthController(authUC, c.Log)
	siteC := http.NewSiteController(siteUC, c.Log)
	reviewC := http.NewReviewController(reviewUC, c.Log)
	learnC := http.NewLearnController(learnUC, c.Log)
	wordC := http.NewWordController(wordUC, c.Log)
	settingC := http.NewSettingController(settingUC, c.Log)
	statsC := http.NewStatsController(statsUC, c.Log)

	// Routes
	routes := route.RouteConfig{
		App:               c.App,
		HelloController:   helloC,
		AuthController:    authC,
		SiteController:    siteC,
		ReviewController:  reviewC,
		LearnController:   learnC,
		WordController:    wordC,
		SettingController: settingC,
		StatsController:   statsC,
		JWTSecret:         jwtSecret,
		StaticDir:         c.StaticDir,
	}
	routes.Setup()

	// Background worker
	if workerEnabled {
		pollInterval := time.Duration(pollSeconds) * time.Second
		worker := parser.NewWorker(c.DB, c.Log,
			siteRepo, categoryRepo, jobRepo, pageRepo,
			dictRepo, occRepo, reviewRepo, geoRepo,
			pollInterval, threshold)
		worker.Start(context.Background())
	} else {
		c.Log.Info("Parser worker disabled")
	}
}
