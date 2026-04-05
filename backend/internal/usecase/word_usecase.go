package usecase

import (
	"context"
	"errors"

	"words-app/internal/entity"
	"words-app/internal/model"
	"words-app/internal/model/converter"
	"words-app/internal/repository"

	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

type WordUseCase struct {
	DB             *gorm.DB
	Log            *logrus.Logger
	Validate       *validator.Validate
	WordRepository *repository.WordRepository
}

func NewWordUseCase(db *gorm.DB, logger *logrus.Logger, validate *validator.Validate,
	wordRepository *repository.WordRepository) *WordUseCase {
	return &WordUseCase{
		DB:             db,
		Log:            logger,
		Validate:       validate,
		WordRepository: wordRepository,
	}
}

func (c *WordUseCase) List(ctx context.Context) ([]model.WordResponse, error) {
	tx := c.DB.WithContext(ctx).Begin()
	defer tx.Rollback()

	var words []entity.Word
	if err := c.WordRepository.FindAll(tx, &words); err != nil {
		c.Log.Warnf("Failed find all words : %+v", err)
		return nil, fiber.ErrInternalServerError
	}

	if err := tx.Commit().Error; err != nil {
		c.Log.Warnf("Failed commit transaction : %+v", err)
		return nil, fiber.ErrInternalServerError
	}

	return converter.WordsToResponses(words), nil
}

func (c *WordUseCase) Get(ctx context.Context, request *model.GetWordRequest) (*model.WordResponse, error) {
	tx := c.DB.WithContext(ctx).Begin()
	defer tx.Rollback()

	if err := c.Validate.Struct(request); err != nil {
		c.Log.Warnf("Invalid request body : %+v", err)
		return nil, fiber.ErrBadRequest
	}

	word := new(entity.Word)
	if err := c.WordRepository.FindBySlug(tx, word, request.Slug); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.Log.Warnf("Word not found : slug=%s", request.Slug)
			return nil, fiber.ErrNotFound
		}
		c.Log.Warnf("Failed find word by slug : %+v", err)
		return nil, fiber.ErrInternalServerError
	}

	if err := tx.Commit().Error; err != nil {
		c.Log.Warnf("Failed commit transaction : %+v", err)
		return nil, fiber.ErrInternalServerError
	}

	return converter.WordToResponse(word), nil
}
