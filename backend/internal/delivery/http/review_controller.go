package http

import (
	"strconv"

	"words-app/internal/model"
	"words-app/internal/usecase"

	"github.com/gofiber/fiber/v2"
	"github.com/sirupsen/logrus"
)

type ReviewController struct {
	Log     *logrus.Logger
	UseCase *usecase.ReviewUseCase
}

func NewReviewController(useCase *usecase.ReviewUseCase, log *logrus.Logger) *ReviewController {
	return &ReviewController{Log: log, UseCase: useCase}
}

func (c *ReviewController) List(ctx *fiber.Ctx) error {
	limit, _ := strconv.Atoi(ctx.Query("limit", "50"))
	response, err := c.UseCase.ListPending(ctx.UserContext(), limit)
	if err != nil {
		return err
	}
	return ctx.JSON(model.WebResponse[[]model.ReviewItemResponse]{Data: response})
}

func (c *ReviewController) Add(ctx *fiber.Ctx) error {
	request := new(model.ReviewDecisionRequest)
	_ = ctx.BodyParser(request)
	request.ID = ctx.Params("id")
	if err := c.UseCase.Add(ctx.UserContext(), request); err != nil {
		return err
	}
	return ctx.JSON(model.WebResponse[string]{Data: "added"})
}

func (c *ReviewController) Deny(ctx *fiber.Ctx) error {
	id := ctx.Params("id")
	if err := c.UseCase.Deny(ctx.UserContext(), id); err != nil {
		return err
	}
	return ctx.JSON(model.WebResponse[string]{Data: "denied"})
}

func (c *ReviewController) Count(ctx *fiber.Ctx) error {
	n, err := c.UseCase.CountPending(ctx.UserContext())
	if err != nil {
		return err
	}
	return ctx.JSON(model.WebResponse[int64]{Data: n})
}
