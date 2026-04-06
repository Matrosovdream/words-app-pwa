package route

import (
	"words-app/internal/delivery/http"
	"words-app/internal/delivery/http/middleware"

	"github.com/gofiber/fiber/v2"
)

type RouteConfig struct {
	App               *fiber.App
	HelloController   *http.HelloController
	AuthController    *http.AuthController
	SiteController    *http.SiteController
	ReviewController  *http.ReviewController
	LearnController   *http.LearnController
	WordController    *http.WordController
	SettingController *http.SettingController
	StatsController   *http.StatsController
	JWTSecret         string
	StaticDir         string
}

func (c *RouteConfig) Setup() {
	api := c.App.Group("/api")

	// --- Public ---
	api.Get("/hello", c.HelloController.Get)
	api.Post("/auth/login", c.AuthController.Login)

	// --- Protected ---
	auth := api.Use(middleware.NewJWTMiddleware(c.JWTSecret))
	auth.Get("/auth/me", c.AuthController.Me)

	// Admin: sites
	auth.Get("/admin/sites", c.SiteController.List)
	auth.Post("/admin/sites", c.SiteController.Create)
	auth.Put("/admin/sites/:id", c.SiteController.Update)
	auth.Delete("/admin/sites/:id", c.SiteController.Delete)
	auth.Get("/admin/sites/:id/categories", c.SiteController.ListCategories)
	auth.Post("/admin/sites/:id/categories", c.SiteController.CreateCategory)
	auth.Put("/admin/sites/:id/categories/:categoryId", c.SiteController.UpdateCategory)
	auth.Delete("/admin/sites/:id/categories/:categoryId", c.SiteController.DeleteCategory)
	auth.Post("/admin/sites/:id/categories/:categoryId/crawl", c.SiteController.CrawlCategory)
	auth.Post("/admin/sites/:id/enqueue", c.SiteController.EnqueueURL)

	// Admin: settings
	auth.Get("/admin/settings/translation", c.SettingController.GetTranslation)
	auth.Put("/admin/settings/translation", c.SettingController.UpdateTranslation)

	// Admin: parser stats
	auth.Get("/admin/parser/stats", c.StatsController.ParserStats)

	// Review
	auth.Get("/review", c.ReviewController.List)
	auth.Get("/review/count", c.ReviewController.Count)
	auth.Post("/review/:id/add", c.ReviewController.Add)
	auth.Post("/review/:id/deny", c.ReviewController.Deny)

	// Learn
	auth.Get("/learn/categories", c.LearnController.ListCategories)
	auth.Post("/learn/categories", c.LearnController.CreateCategory)
	auth.Put("/learn/categories/:id", c.LearnController.UpdateCategory)
	auth.Delete("/learn/categories/:id", c.LearnController.DeleteCategory)
	auth.Get("/learn/items", c.LearnController.ListItems)
	auth.Get("/learn/archived", c.LearnController.ListArchived)
	auth.Put("/learn/items/:id", c.LearnController.UpdateItem)
	auth.Post("/learn/items/:id/done", c.LearnController.Done)

	// Words
	auth.Get("/words/:id", c.WordController.Detail)

	// --- Static SPA ---
	if c.StaticDir != "" {
		c.App.Static("/", c.StaticDir)
		c.App.Use(func(ctx *fiber.Ctx) error {
			return ctx.SendFile(c.StaticDir + "/index.html")
		})
	}
}
