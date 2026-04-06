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
	DB                          *gorm.DB
	Log                         *logrus.Logger
	Validate                    *validator.Validate
	ReviewRepository            *repository.ReviewItemRepository
	DictWordRepository          *repository.DictWordRepository
	OccurrenceRepository        *repository.WordOccurrenceRepository
	ParsedPageRepository        *repository.ParsedPageRepository
	LearnItemRepository         *repository.LearnItemRepository
	LearnCategoryRepository     *repository.LearnCategoryRepository
	LearnItemCategoryRepository *repository.LearnItemCategoryRepository
	SiteCategoryRepository      *repository.SiteCategoryRepository
}

func NewReviewUseCase(db *gorm.DB, log *logrus.Logger, validate *validator.Validate,
	reviewRepo *repository.ReviewItemRepository,
	dictRepo *repository.DictWordRepository,
	occRepo *repository.WordOccurrenceRepository,
	pageRepo *repository.ParsedPageRepository,
	learnRepo *repository.LearnItemRepository,
	catRepo *repository.LearnCategoryRepository,
	itemCatRepo *repository.LearnItemCategoryRepository,
	siteCatRepo *repository.SiteCategoryRepository) *ReviewUseCase {
	return &ReviewUseCase{
		DB:                          db,
		Log:                         log,
		Validate:                    validate,
		ReviewRepository:            reviewRepo,
		DictWordRepository:          dictRepo,
		OccurrenceRepository:        occRepo,
		ParsedPageRepository:        pageRepo,
		LearnItemRepository:         learnRepo,
		LearnCategoryRepository:     catRepo,
		LearnItemCategoryRepository: itemCatRepo,
		SiteCategoryRepository:      siteCatRepo,
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
			if err := c.ParsedPageRepository.FindByID(tx, page, occs[0].ParsedPageID); err == nil {
				sourceURL = page.URL
			}
		}
		var firstSeenPageGUID *string
		if r.FirstSeenPageID != nil {
			page := new(entity.ParsedPage)
			if err := c.ParsedPageRepository.FindByID(tx, page, *r.FirstSeenPageID); err == nil {
				firstSeenPageGUID = &page.GUID
			}
		}
		var siteCategoryName *string
		if r.SiteCategoryID != nil {
			sc := new(entity.SiteCategory)
			if err := c.SiteCategoryRepository.FindByID(tx, sc, *r.SiteCategoryID); err == nil {
				siteCategoryName = &sc.Name
			}
		}
		responses = append(responses, model.ReviewItemResponse{
			ID:               r.GUID,
			WordID:           word.GUID,
			Lemma:            word.Lemma,
			Language:         word.Language,
			SampleSentence:   sample,
			FirstSeenPageID:  firstSeenPageGUID,
			SourceURL:        sourceURL,
			SiteCategoryName: siteCategoryName,
			Status:           r.Status,
			CreatedAt:        r.CreatedAt.Unix(),
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
	if err := c.ReviewRepository.FindByGUID(tx, item, req.ID); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return fiber.ErrNotFound
		}
		c.Log.Warnf("Failed find review item : %+v", err)
		return fiber.ErrInternalServerError
	}

	// resolve explicit category from request (may be empty)
	var explicitCatID *int64
	if req.LearnCategoryID != "" {
		cat := new(entity.LearnCategory)
		if err := c.LearnCategoryRepository.FindByGUID(tx, cat, req.LearnCategoryID); err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return fiber.ErrNotFound
			}
			c.Log.Warnf("Failed find learn category : %+v", err)
			return fiber.ErrInternalServerError
		}
		explicitCatID = &cat.ID
	}

	// fall back to the site category's learn_category_id when no explicit pick
	if explicitCatID == nil && item.SiteCategoryID != nil {
		sc := new(entity.SiteCategory)
		if err := c.SiteCategoryRepository.FindByID(tx, sc, *item.SiteCategoryID); err == nil && sc.LearnCategoryID != nil {
			explicitCatID = sc.LearnCategoryID
		}
	}

	// upsert learn item
	var learnItemID int64
	existing := new(entity.LearnItem)
	if err := c.LearnItemRepository.FindByWordID(tx, existing, item.WordID); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			li := &entity.LearnItem{
				GUID:         uuid.NewString(),
				WordID:       item.WordID,
				Status:       entity.LearnStatusActive,
				MasteryLevel: 0,
				CreatedAt:    time.Now(),
				UpdatedAt:    time.Now(),
			}
			if err := c.LearnItemRepository.Create(tx, li); err != nil {
				c.Log.Warnf("Failed to create learn item : %+v", err)
				return fiber.ErrInternalServerError
			}
			learnItemID = li.ID
		} else {
			c.Log.Warnf("Failed find learn item : %+v", err)
			return fiber.ErrInternalServerError
		}
	} else {
		// reactivate
		existing.Status = entity.LearnStatusActive
		existing.ArchivedAt = nil
		existing.UpdatedAt = time.Now()
		if err := c.LearnItemRepository.Update(tx, existing); err != nil {
			c.Log.Warnf("Failed to update learn item : %+v", err)
			return fiber.ErrInternalServerError
		}
		learnItemID = existing.ID
	}

	// assign category via join table
	if explicitCatID != nil {
		_ = c.LearnItemCategoryRepository.Create(tx, &entity.LearnItemCategory{
			LearnItemID: learnItemID, LearnCategoryID: *explicitCatID,
		})
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

func (c *ReviewUseCase) Deny(ctx context.Context, publicID string) error {
	tx := c.DB.WithContext(ctx).Begin()
	defer tx.Rollback()

	item := new(entity.ReviewItem)
	if err := c.ReviewRepository.FindByGUID(tx, item, publicID); err != nil {
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
