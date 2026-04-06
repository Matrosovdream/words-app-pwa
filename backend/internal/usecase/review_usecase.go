package usecase

import (
	"context"
	"errors"
	"time"

	"words-app/internal/entity"
	"words-app/internal/model"
	"words-app/internal/repository"

	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

type ReviewUseCase struct {
	DB                       *gorm.DB
	Log                      *logrus.Logger
	Validate                 *validator.Validate
	ReviewRepository         *repository.ReviewItemRepository
	DictWordRepository       *repository.DictWordRepository
	OccurrenceRepository     *repository.WordOccurrenceRepository
	ParsedPageRepository     *repository.ParsedPageRepository
	LearnItemRepository      *repository.LearnItemRepository
	LearnCategoryRepository  *repository.LearnCategoryRepository
}

func NewReviewUseCase(db *gorm.DB, log *logrus.Logger, validate *validator.Validate,
	reviewRepo *repository.ReviewItemRepository,
	dictRepo *repository.DictWordRepository,
	occRepo *repository.WordOccurrenceRepository,
	pageRepo *repository.ParsedPageRepository,
	learnRepo *repository.LearnItemRepository,
	catRepo *repository.LearnCategoryRepository) *ReviewUseCase {
	return &ReviewUseCase{
		DB:                      db,
		Log:                     log,
		Validate:                validate,
		ReviewRepository:        reviewRepo,
		DictWordRepository:      dictRepo,
		OccurrenceRepository:    occRepo,
		ParsedPageRepository:    pageRepo,
		LearnItemRepository:     learnRepo,
		LearnCategoryRepository: catRepo,
	}
}

func (c *ReviewUseCase) ListPending(ctx context.Context, limit int) ([]model.ReviewItemResponse, error) {
	tx := c.DB.WithContext(ctx).Begin()
	defer tx.Rollback()

	if limit <= 0 || limit > 500 {
		limit = 50
	}

	var items []entity.ReviewItem
	if err := c.ReviewRepository.FindPending(tx, &items, limit); err != nil {
		c.Log.Warnf("Failed list pending reviews : %+v", err)
		return nil, fiber.ErrInternalServerError
	}

	responses := make([]model.ReviewItemResponse, 0, len(items))
	for i := range items {
		r := &items[i]
		word := new(entity.DictWord)
		if err := c.DictWordRepository.FindByID(tx, word, r.WordID); err != nil {
			continue
		}
		sample := ""
		sourceURL := ""
		var occs []entity.WordOccurrence
		if err := c.OccurrenceRepository.FindByWord(tx, &occs, r.WordID, 1); err == nil && len(occs) > 0 {
			sample = occs[0].SampleSentence
			page := new(entity.ParsedPage)
			if err := tx.Where("id = ?", occs[0].ParsedPageID).First(page).Error; err == nil {
				sourceURL = page.URL
			}
		}
		responses = append(responses, model.ReviewItemResponse{
			ID:              r.ID,
			WordID:          r.WordID,
			Lemma:           word.Lemma,
			Language:        word.Language,
			SampleSentence:  sample,
			FirstSeenPageID: r.FirstSeenPageID,
			SourceURL:       sourceURL,
			Status:          r.Status,
			CreatedAt:       r.CreatedAt.Unix(),
		})
	}

	if err := tx.Commit().Error; err != nil {
		c.Log.Warnf("Failed commit transaction : %+v", err)
		return nil, fiber.ErrInternalServerError
	}
	return responses, nil
}

func (c *ReviewUseCase) Add(ctx context.Context, req *model.ReviewDecisionRequest) error {
	tx := c.DB.WithContext(ctx).Begin()
	defer tx.Rollback()

	if err := c.Validate.Struct(req); err != nil {
		c.Log.Warnf("Invalid review decision : %+v", err)
		return fiber.ErrBadRequest
	}

	item := new(entity.ReviewItem)
	if err := c.ReviewRepository.FindByID(tx, item, req.ID); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return fiber.ErrNotFound
		}
		c.Log.Warnf("Failed find review item : %+v", err)
		return fiber.ErrInternalServerError
	}

	// validate optional category
	var categoryID *string
	if req.LearnCategoryID != "" {
		cat := new(entity.LearnCategory)
		if err := c.LearnCategoryRepository.FindByID(tx, cat, req.LearnCategoryID); err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return fiber.ErrNotFound
			}
			c.Log.Warnf("Failed find learn category : %+v", err)
			return fiber.ErrInternalServerError
		}
		cid := cat.ID
		categoryID = &cid
	}

	// upsert learn item
	existing := new(entity.LearnItem)
	if err := c.LearnItemRepository.FindByWordID(tx, existing, item.WordID); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			li := &entity.LearnItem{
				ID:              uuid.NewString(),
				WordID:          item.WordID,
				LearnCategoryID: categoryID,
				Status:          entity.LearnStatusActive,
				MasteryLevel:    0,
				CreatedAt:       time.Now(),
				UpdatedAt:       time.Now(),
			}
			if err := c.LearnItemRepository.Create(tx, li); err != nil {
				c.Log.Warnf("Failed to create learn item : %+v", err)
				return fiber.ErrInternalServerError
			}
		} else {
			c.Log.Warnf("Failed find learn item : %+v", err)
			return fiber.ErrInternalServerError
		}
	} else {
		// reactivate / reassign
		existing.Status = entity.LearnStatusActive
		existing.LearnCategoryID = categoryID
		existing.ArchivedAt = nil
		existing.UpdatedAt = time.Now()
		if err := c.LearnItemRepository.Update(tx, existing); err != nil {
			c.Log.Warnf("Failed to update learn item : %+v", err)
			return fiber.ErrInternalServerError
		}
	}

	now := time.Now()
	item.Status = entity.ReviewStatusAdded
	item.ReviewedAt = &now
	item.UpdatedAt = now
	if err := c.ReviewRepository.Update(tx, item); err != nil {
		c.Log.Warnf("Failed to update review item : %+v", err)
		return fiber.ErrInternalServerError
	}

	if err := tx.Commit().Error; err != nil {
		c.Log.Warnf("Failed commit transaction : %+v", err)
		return fiber.ErrInternalServerError
	}
	return nil
}

func (c *ReviewUseCase) Deny(ctx context.Context, id string) error {
	tx := c.DB.WithContext(ctx).Begin()
	defer tx.Rollback()

	item := new(entity.ReviewItem)
	if err := c.ReviewRepository.FindByID(tx, item, id); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return fiber.ErrNotFound
		}
		c.Log.Warnf("Failed find review item : %+v", err)
		return fiber.ErrInternalServerError
	}
	now := time.Now()
	item.Status = entity.ReviewStatusDenied
	item.ReviewedAt = &now
	item.UpdatedAt = now
	if err := c.ReviewRepository.Update(tx, item); err != nil {
		c.Log.Warnf("Failed to update review item : %+v", err)
		return fiber.ErrInternalServerError
	}
	if err := tx.Commit().Error; err != nil {
		c.Log.Warnf("Failed commit transaction : %+v", err)
		return fiber.ErrInternalServerError
	}
	return nil
}

func (c *ReviewUseCase) CountPending(ctx context.Context) (int64, error) {
	n, err := c.ReviewRepository.CountPending(c.DB.WithContext(ctx))
	if err != nil {
		c.Log.Warnf("Failed count pending : %+v", err)
		return 0, fiber.ErrInternalServerError
	}
	return n, nil
}
