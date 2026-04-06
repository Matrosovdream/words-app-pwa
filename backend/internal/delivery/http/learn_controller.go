package http

import (
	"strconv"

	"words-app/internal/model"
	"words-app/internal/usecase"

	"github.com/gofiber/fiber/v2"
	"github.com/sirupsen/logrus"
)

type LearnController struct {
	Log     *logrus.Logger
	UseCase *usecase.LearnUseCase
}

func NewLearnController(useCase *usecase.LearnUseCase, log *logrus.Logger) *LearnController {
	return &LearnController{Log: log, UseCase: useCase}
}

// Categories

func (c *LearnController) ListCategories(ctx *fiber.Ctx) error {
	response, err := c.UseCase.ListCategories(ctx.UserContext())
	if err != nil {
		return err
	}
	return ctx.JSON(model.WebResponse[[]model.LearnCategoryResponse]{Data: response})
}

func (c *LearnController) CreateCategory(ctx *fiber.Ctx) error {
	request := new(model.CreateLearnCategoryRequest)
	if err := ctx.BodyParser(request); err != nil {
		c.Log.Warnf("Failed to parse body : %+v", err)
		return fiber.ErrBadRequest
	}
	response, err := c.UseCase.CreateCategory(ctx.UserContext(), request)
	if err != nil {
		return err
	}
	return ctx.Status(fiber.StatusCreated).JSON(model.WebResponse[*model.LearnCategoryResponse]{Data: response})
}

func (c *LearnController) UpdateCategory(ctx *fiber.Ctx) error {
	request := new(model.UpdateLearnCategoryRequest)
	if err := ctx.BodyParser(request); err != nil {
		c.Log.Warnf("Failed to parse body : %+v", err)
		return fiber.ErrBadRequest
	}
	request.ID = ctx.Params("id")
	response, err := c.UseCase.UpdateCategory(ctx.UserContext(), request)
	if err != nil {
		return err
	}
	return ctx.JSON(model.WebResponse[*model.LearnCategoryResponse]{Data: response})
}

func (c *LearnController) DeleteCategory(ctx *fiber.Ctx) error {
	id := ctx.Params("id")
	if err := c.UseCase.DeleteCategory(ctx.UserContext(), id); err != nil {
		return err
	}
	return ctx.JSON(model.WebResponse[string]{Data: "deleted"})
}

// Items

func (c *LearnController) ListItems(ctx *fiber.Ctx) error {
	var categoryID *string
	if q := ctx.Query("category_id"); q != "" {
		categoryID = &q
	}
	sort := ctx.Query("sort", "newest")
	response, err := c.UseCase.ListItems(ctx.UserContext(), categoryID, sort)
	if err != nil {
		return err
	}
	return ctx.JSON(model.WebResponse[[]model.LearnItemResponse]{Data: response})
}

func (c *LearnController) ListArchived(ctx *fiber.Ctx) error {
	limit, _ := strconv.Atoi(ctx.Query("limit", "100"))
	response, err := c.UseCase.ListArchived(ctx.UserContext(), limit)
	if err != nil {
		return err
	}
	return ctx.JSON(model.WebResponse[[]model.LearnItemResponse]{Data: response})
}

func (c *LearnController) UpdateItem(ctx *fiber.Ctx) error {
	request := new(model.UpdateLearnItemRequest)
	if err := ctx.BodyParser(request); err != nil {
		c.Log.Warnf("Failed to parse body : %+v", err)
		return fiber.ErrBadRequest
	}
	request.ID = ctx.Params("id")
	if err := c.UseCase.UpdateItem(ctx.UserContext(), request); err != nil {
		return err
	}
	return ctx.JSON(model.WebResponse[string]{Data: "updated"})
}

func (c *LearnController) Done(ctx *fiber.Ctx) error {
	id := ctx.Params("id")
	if err := c.UseCase.Done(ctx.UserContext(), id); err != nil {
		return err
	}
	return ctx.JSON(model.WebResponse[string]{Data: "archived"})
}
