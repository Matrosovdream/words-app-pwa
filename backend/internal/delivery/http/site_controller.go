package http

import (
	"words-app/internal/model"
	"words-app/internal/usecase"

	"github.com/gofiber/fiber/v2"
	"github.com/sirupsen/logrus"
)

type SiteController struct {
	Log     *logrus.Logger
	UseCase *usecase.SiteUseCase
}

func NewSiteController(useCase *usecase.SiteUseCase, log *logrus.Logger) *SiteController {
	return &SiteController{Log: log, UseCase: useCase}
}

func (c *SiteController) List(ctx *fiber.Ctx) error {
	response, err := c.UseCase.List(ctx.UserContext())
	if err != nil {
		return err
	}
	return ctx.JSON(model.WebResponse[[]model.SiteResponse]{Data: response})
}

func (c *SiteController) Create(ctx *fiber.Ctx) error {
	request := new(model.CreateSiteRequest)
	if err := ctx.BodyParser(request); err != nil {
		c.Log.Warnf("Failed to parse body : %+v", err)
		return fiber.ErrBadRequest
	}
	response, err := c.UseCase.Create(ctx.UserContext(), request)
	if err != nil {
		return err
	}
	return ctx.Status(fiber.StatusCreated).JSON(model.WebResponse[*model.SiteResponse]{Data: response})
}

func (c *SiteController) Update(ctx *fiber.Ctx) error {
	request := new(model.UpdateSiteRequest)
	if err := ctx.BodyParser(request); err != nil {
		c.Log.Warnf("Failed to parse body : %+v", err)
		return fiber.ErrBadRequest
	}
	request.ID = ctx.Params("id")
	response, err := c.UseCase.Update(ctx.UserContext(), request)
	if err != nil {
		return err
	}
	return ctx.JSON(model.WebResponse[*model.SiteResponse]{Data: response})
}

func (c *SiteController) Delete(ctx *fiber.Ctx) error {
	id := ctx.Params("id")
	if err := c.UseCase.Delete(ctx.UserContext(), id); err != nil {
		return err
	}
	return ctx.JSON(model.WebResponse[string]{Data: "deleted"})
}

func (c *SiteController) ListCategories(ctx *fiber.Ctx) error {
	siteID := ctx.Params("id")
	response, err := c.UseCase.ListCategories(ctx.UserContext(), siteID)
	if err != nil {
		return err
	}
	return ctx.JSON(model.WebResponse[[]model.SiteCategoryResponse]{Data: response})
}

func (c *SiteController) CreateCategory(ctx *fiber.Ctx) error {
	request := new(model.CreateSiteCategoryRequest)
	if err := ctx.BodyParser(request); err != nil {
		c.Log.Warnf("Failed to parse body : %+v", err)
		return fiber.ErrBadRequest
	}
	request.SiteID = ctx.Params("id")
	response, err := c.UseCase.CreateCategory(ctx.UserContext(), request)
	if err != nil {
		return err
	}
	return ctx.Status(fiber.StatusCreated).JSON(model.WebResponse[*model.SiteCategoryResponse]{Data: response})
}

func (c *SiteController) UpdateCategory(ctx *fiber.Ctx) error {
	request := new(model.UpdateSiteCategoryRequest)
	if err := ctx.BodyParser(request); err != nil {
		c.Log.Warnf("Failed to parse body : %+v", err)
		return fiber.ErrBadRequest
	}
	request.ID = ctx.Params("categoryId")
	response, err := c.UseCase.UpdateCategory(ctx.UserContext(), request)
	if err != nil {
		return err
	}
	return ctx.JSON(model.WebResponse[*model.SiteCategoryResponse]{Data: response})
}

func (c *SiteController) DeleteCategory(ctx *fiber.Ctx) error {
	id := ctx.Params("categoryId")
	if err := c.UseCase.DeleteCategory(ctx.UserContext(), id); err != nil {
		return err
	}
	return ctx.JSON(model.WebResponse[string]{Data: "deleted"})
}

func (c *SiteController) CrawlCategory(ctx *fiber.Ctx) error {
	id := ctx.Params("categoryId")
	if err := c.UseCase.CrawlCategory(ctx.UserContext(), id); err != nil {
		return err
	}
	return ctx.Status(fiber.StatusCreated).JSON(model.WebResponse[string]{Data: "queued"})
}

func (c *SiteController) RunSite(ctx *fiber.Ctx) error {
	id := ctx.Params("id")
	if err := c.UseCase.RunSite(ctx.UserContext(), id); err != nil {
		return err
	}
	return ctx.Status(fiber.StatusCreated).JSON(model.WebResponse[string]{Data: "queued"})
}

func (c *SiteController) EnqueueURL(ctx *fiber.Ctx) error {
	request := new(model.EnqueueURLRequest)
	if err := ctx.BodyParser(request); err != nil {
		c.Log.Warnf("Failed to parse body : %+v", err)
		return fiber.ErrBadRequest
	}
	request.SiteID = ctx.Params("id")
	if err := c.UseCase.EnqueueURL(ctx.UserContext(), request); err != nil {
		return err
	}
	return ctx.Status(fiber.StatusCreated).JSON(model.WebResponse[string]{Data: "queued"})
}
