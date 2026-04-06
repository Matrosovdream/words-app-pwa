package usecase

import (
	"context"
	"errors"
	"time"

	"words-app/internal/entity"
	"words-app/internal/model"
	"words-app/internal/model/converter"
	"words-app/internal/repository"

	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

type SiteUseCase struct {
	DB                     *gorm.DB
	Log                    *logrus.Logger
	Validate               *validator.Validate
	SiteRepository         *repository.SiteRepository
	SiteCategoryRepository *repository.SiteCategoryRepository
	ParseJobRepository     *repository.ParseJobRepository
}

func NewSiteUseCase(db *gorm.DB, log *logrus.Logger, validate *validator.Validate,
	siteRepo *repository.SiteRepository,
	categoryRepo *repository.SiteCategoryRepository,
	jobRepo *repository.ParseJobRepository) *SiteUseCase {
	return &SiteUseCase{
		DB:                     db,
		Log:                    log,
		Validate:               validate,
		SiteRepository:         siteRepo,
		SiteCategoryRepository: categoryRepo,
		ParseJobRepository:     jobRepo,
	}
}

func boolOrDefault(p *bool, def bool) bool {
	if p == nil {
		return def
	}
	return *p
}

func (c *SiteUseCase) List(ctx context.Context) ([]model.SiteResponse, error) {
	tx := c.DB.WithContext(ctx).Begin()
	defer tx.Rollback()

	var sites []entity.Site
	if err := c.SiteRepository.FindAll(tx, &sites); err != nil {
		c.Log.Warnf("Failed to list sites : %+v", err)
		return nil, fiber.ErrInternalServerError
	}
	if err := tx.Commit().Error; err != nil {
		c.Log.Warnf("Failed commit transaction : %+v", err)
		return nil, fiber.ErrInternalServerError
	}
	return converter.SitesToResponses(sites), nil
}

func (c *SiteUseCase) Create(ctx context.Context, req *model.CreateSiteRequest) (*model.SiteResponse, error) {
	tx := c.DB.WithContext(ctx).Begin()
	defer tx.Rollback()

	if err := c.Validate.Struct(req); err != nil {
		c.Log.Warnf("Invalid create site request : %+v", err)
		return nil, fiber.ErrBadRequest
	}

	site := &entity.Site{
		PublicID:         uuid.NewString(),
		Name:             req.Name,
		BaseURL:          req.BaseURL,
		IsActive:         boolOrDefault(req.IsActive, true),
		DailyHitLimit:    req.DailyHitLimit,
		HitDelayMs:       req.HitDelayMs,
		ReparseEnabled:   boolOrDefault(req.ReparseEnabled, false),
		ReparseAfterDays: req.ReparseAfterDays,
		UserAgent:        req.UserAgent,
		CreatedAt:        time.Now(),
		UpdatedAt:        time.Now(),
	}
	if site.DailyHitLimit == 0 {
		site.DailyHitLimit = 100
	}
	if site.HitDelayMs == 0 {
		site.HitDelayMs = 2000
	}

	if err := c.SiteRepository.Create(tx, site); err != nil {
		c.Log.Warnf("Failed to create site : %+v", err)
		return nil, fiber.ErrInternalServerError
	}
	if err := tx.Commit().Error; err != nil {
		c.Log.Warnf("Failed commit transaction : %+v", err)
		return nil, fiber.ErrInternalServerError
	}
	return converter.SiteToResponse(site), nil
}

func (c *SiteUseCase) Update(ctx context.Context, req *model.UpdateSiteRequest) (*model.SiteResponse, error) {
	tx := c.DB.WithContext(ctx).Begin()
	defer tx.Rollback()

	if err := c.Validate.Struct(req); err != nil {
		c.Log.Warnf("Invalid update site request : %+v", err)
		return nil, fiber.ErrBadRequest
	}

	site := new(entity.Site)
	if err := c.SiteRepository.FindByPublicID(tx, site, req.ID); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fiber.ErrNotFound
		}
		c.Log.Warnf("Failed to find site : %+v", err)
		return nil, fiber.ErrInternalServerError
	}
	site.Name = req.Name
	site.BaseURL = req.BaseURL
	site.IsActive = boolOrDefault(req.IsActive, site.IsActive)
	site.DailyHitLimit = req.DailyHitLimit
	site.HitDelayMs = req.HitDelayMs
	site.ReparseEnabled = boolOrDefault(req.ReparseEnabled, site.ReparseEnabled)
	site.ReparseAfterDays = req.ReparseAfterDays
	site.UserAgent = req.UserAgent
	site.UpdatedAt = time.Now()

	if err := c.SiteRepository.Update(tx, site); err != nil {
		c.Log.Warnf("Failed to update site : %+v", err)
		return nil, fiber.ErrInternalServerError
	}
	if err := tx.Commit().Error; err != nil {
		c.Log.Warnf("Failed commit transaction : %+v", err)
		return nil, fiber.ErrInternalServerError
	}
	return converter.SiteToResponse(site), nil
}

func (c *SiteUseCase) Delete(ctx context.Context, publicID string) error {
	tx := c.DB.WithContext(ctx).Begin()
	defer tx.Rollback()

	site := new(entity.Site)
	if err := c.SiteRepository.FindByPublicID(tx, site, publicID); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return fiber.ErrNotFound
		}
		c.Log.Warnf("Failed to find site : %+v", err)
		return fiber.ErrInternalServerError
	}

	if err := c.SiteRepository.Delete(tx, site.ID); err != nil {
		c.Log.Warnf("Failed to delete site : %+v", err)
		return fiber.ErrInternalServerError
	}
	if err := tx.Commit().Error; err != nil {
		c.Log.Warnf("Failed commit transaction : %+v", err)
		return fiber.ErrInternalServerError
	}
	return nil
}

// Categories

func (c *SiteUseCase) ListCategories(ctx context.Context, sitePublicID string) ([]model.SiteCategoryResponse, error) {
	tx := c.DB.WithContext(ctx).Begin()
	defer tx.Rollback()

	site := new(entity.Site)
	if err := c.SiteRepository.FindByPublicID(tx, site, sitePublicID); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fiber.ErrNotFound
		}
		c.Log.Warnf("Failed to find site : %+v", err)
		return nil, fiber.ErrInternalServerError
	}

	var cats []entity.SiteCategory
	if err := c.SiteCategoryRepository.FindBySite(tx, &cats, site.ID); err != nil {
		c.Log.Warnf("Failed list categories : %+v", err)
		return nil, fiber.ErrInternalServerError
	}
	if err := tx.Commit().Error; err != nil {
		c.Log.Warnf("Failed commit transaction : %+v", err)
		return nil, fiber.ErrInternalServerError
	}
	return converter.SiteCategoriesToResponses(cats, site.PublicID), nil
}

func (c *SiteUseCase) CreateCategory(ctx context.Context, req *model.CreateSiteCategoryRequest) (*model.SiteCategoryResponse, error) {
	tx := c.DB.WithContext(ctx).Begin()
	defer tx.Rollback()

	if err := c.Validate.Struct(req); err != nil {
		c.Log.Warnf("Invalid create category request : %+v", err)
		return nil, fiber.ErrBadRequest
	}

	// verify site exists
	site := new(entity.Site)
	if err := c.SiteRepository.FindByPublicID(tx, site, req.SiteID); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fiber.ErrNotFound
		}
		c.Log.Warnf("Failed to find site : %+v", err)
		return nil, fiber.ErrInternalServerError
	}

	lang := req.SourceLanguage
	if lang == "" {
		lang = "en"
	}

	cat := &entity.SiteCategory{
		PublicID:       uuid.NewString(),
		SiteID:         site.ID,
		Name:           req.Name,
		StartURL:       req.StartURL,
		URLPattern:     req.URLPattern,
		SelectorTitle:  req.SelectorTitle,
		SelectorBody:   req.SelectorBody,
		SourceLanguage: lang,
		IsActive:       boolOrDefault(req.IsActive, true),
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
	}
	if err := c.SiteCategoryRepository.Create(tx, cat); err != nil {
		c.Log.Warnf("Failed to create category : %+v", err)
		return nil, fiber.ErrInternalServerError
	}

	// auto-enqueue the start URL so parsing begins immediately
	catID := cat.ID
	job := &entity.ParseJob{
		PublicID:       uuid.NewString(),
		SiteID:         site.ID,
		SiteCategoryID: &catID,
		URL:            cat.StartURL,
		Status:         entity.ParseJobStatusPending,
		ScheduledAt:    time.Now(),
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
	}
	if err := c.ParseJobRepository.Create(tx, job); err != nil {
		c.Log.Warnf("Failed to auto-enqueue category start URL : %+v", err)
		return nil, fiber.ErrInternalServerError
	}

	if err := tx.Commit().Error; err != nil {
		c.Log.Warnf("Failed commit transaction : %+v", err)
		return nil, fiber.ErrInternalServerError
	}
	return converter.SiteCategoryToResponse(cat, site.PublicID), nil
}

// CrawlCategory re-enqueues the category's start URL for parsing.
func (c *SiteUseCase) CrawlCategory(ctx context.Context, categoryPublicID string) error {
	tx := c.DB.WithContext(ctx).Begin()
	defer tx.Rollback()

	cat := new(entity.SiteCategory)
	if err := c.SiteCategoryRepository.FindByPublicID(tx, cat, categoryPublicID); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return fiber.ErrNotFound
		}
		c.Log.Warnf("Failed find category : %+v", err)
		return fiber.ErrInternalServerError
	}

	catID := cat.ID
	job := &entity.ParseJob{
		PublicID:       uuid.NewString(),
		SiteID:         cat.SiteID,
		SiteCategoryID: &catID,
		URL:            cat.StartURL,
		Status:         entity.ParseJobStatusPending,
		ScheduledAt:    time.Now(),
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
	}
	if err := c.ParseJobRepository.Create(tx, job); err != nil {
		c.Log.Warnf("Failed to enqueue crawl job : %+v", err)
		return fiber.ErrInternalServerError
	}
	if err := tx.Commit().Error; err != nil {
		c.Log.Warnf("Failed commit transaction : %+v", err)
		return fiber.ErrInternalServerError
	}
	return nil
}

func (c *SiteUseCase) UpdateCategory(ctx context.Context, req *model.UpdateSiteCategoryRequest) (*model.SiteCategoryResponse, error) {
	tx := c.DB.WithContext(ctx).Begin()
	defer tx.Rollback()

	if err := c.Validate.Struct(req); err != nil {
		c.Log.Warnf("Invalid update category request : %+v", err)
		return nil, fiber.ErrBadRequest
	}

	cat := new(entity.SiteCategory)
	if err := c.SiteCategoryRepository.FindByPublicID(tx, cat, req.ID); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fiber.ErrNotFound
		}
		c.Log.Warnf("Failed to find category : %+v", err)
		return nil, fiber.ErrInternalServerError
	}

	// resolve site PublicID for the response
	site := new(entity.Site)
	if err := c.SiteRepository.FindByID(tx, site, cat.SiteID); err != nil {
		c.Log.Warnf("Failed to find site for category : %+v", err)
		return nil, fiber.ErrInternalServerError
	}

	cat.Name = req.Name
	cat.StartURL = req.StartURL
	cat.URLPattern = req.URLPattern
	cat.SelectorTitle = req.SelectorTitle
	cat.SelectorBody = req.SelectorBody
	if req.SourceLanguage != "" {
		cat.SourceLanguage = req.SourceLanguage
	}
	cat.IsActive = boolOrDefault(req.IsActive, cat.IsActive)
	cat.UpdatedAt = time.Now()

	if err := c.SiteCategoryRepository.Update(tx, cat); err != nil {
		c.Log.Warnf("Failed to update category : %+v", err)
		return nil, fiber.ErrInternalServerError
	}
	if err := tx.Commit().Error; err != nil {
		c.Log.Warnf("Failed commit transaction : %+v", err)
		return nil, fiber.ErrInternalServerError
	}
	return converter.SiteCategoryToResponse(cat, site.PublicID), nil
}

func (c *SiteUseCase) DeleteCategory(ctx context.Context, publicID string) error {
	tx := c.DB.WithContext(ctx).Begin()
	defer tx.Rollback()

	cat := new(entity.SiteCategory)
	if err := c.SiteCategoryRepository.FindByPublicID(tx, cat, publicID); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return fiber.ErrNotFound
		}
		c.Log.Warnf("Failed to find category : %+v", err)
		return fiber.ErrInternalServerError
	}

	if err := c.SiteCategoryRepository.Delete(tx, cat.ID); err != nil {
		c.Log.Warnf("Failed to delete category : %+v", err)
		return fiber.ErrInternalServerError
	}
	if err := tx.Commit().Error; err != nil {
		c.Log.Warnf("Failed commit transaction : %+v", err)
		return fiber.ErrInternalServerError
	}
	return nil
}

// EnqueueURL manually adds a URL to the parse queue for a site.
func (c *SiteUseCase) EnqueueURL(ctx context.Context, req *model.EnqueueURLRequest) error {
	tx := c.DB.WithContext(ctx).Begin()
	defer tx.Rollback()

	if err := c.Validate.Struct(req); err != nil {
		c.Log.Warnf("Invalid enqueue request : %+v", err)
		return fiber.ErrBadRequest
	}
	site := new(entity.Site)
	if err := c.SiteRepository.FindByPublicID(tx, site, req.SiteID); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return fiber.ErrNotFound
		}
		c.Log.Warnf("Failed to find site : %+v", err)
		return fiber.ErrInternalServerError
	}

	job := &entity.ParseJob{
		PublicID:    uuid.NewString(),
		SiteID:      site.ID,
		URL:         req.URL,
		Status:      entity.ParseJobStatusPending,
		ScheduledAt: time.Now(),
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}
	if req.SiteCategoryID != "" {
		cat := new(entity.SiteCategory)
		if err := c.SiteCategoryRepository.FindByPublicID(tx, cat, req.SiteCategoryID); err == nil {
			catID := cat.ID
			job.SiteCategoryID = &catID
		}
	}
	if err := c.ParseJobRepository.Create(tx, job); err != nil {
		c.Log.Warnf("Failed to enqueue job : %+v", err)
		return fiber.ErrInternalServerError
	}
	if err := tx.Commit().Error; err != nil {
		c.Log.Warnf("Failed commit transaction : %+v", err)
		return fiber.ErrInternalServerError
	}
	return nil
}
