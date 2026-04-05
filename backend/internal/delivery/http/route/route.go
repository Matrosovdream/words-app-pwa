package route

import (
	"words-app/internal/delivery/http"

	"github.com/gofiber/fiber/v2"
)

type RouteConfig struct {
	App             *fiber.App
	HelloController *http.HelloController
	WordController  *http.WordController
	StaticDir       string
}

func (c *RouteConfig) Setup() {
	// --- JSON API ---
	api := c.App.Group("/api")
	api.Get("/hello", c.HelloController.Get)
	api.Get("/words", c.WordController.List)
	api.Get("/words/:slug", c.WordController.Get)

	// --- Static SPA assets + fallback ---
	// In dev the frontend is served by Vite, so StaticDir may be empty.
	if c.StaticDir != "" {
		c.App.Static("/", c.StaticDir)
		c.App.Use(func(ctx *fiber.Ctx) error {
			return ctx.SendFile(c.StaticDir + "/index.html")
		})
	}
}
