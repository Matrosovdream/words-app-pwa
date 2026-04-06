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

type LearnUseCase struct {
	DB                      *gorm.DB
	Log                     *logrus.Logger
	Validate                *validator.Validate
	LearnCategoryRepository *repository.LearnCategoryRepository
	LearnItemRepository     *repository.LearnItemRepository
	DictWordRepository      *repository.DictWordRepository
}

func NewLearnUseCase(db *gorm.DB, log *logrus.Logger, validate *validator.Validate,
	catRepo *repository.LearnCategoryRepository,
	itemRepo *repository.LearnItemRepository,
	dictRepo *repository.DictWordRepository) *LearnUseCase {
	return &LearnUseCase{
		DB:                      db,
		Log:                     log,
		Validate:                validate,
		LearnCategoryRepository: catRepo,
		LearnItemRepository:     itemRepo,
		DictWordRepository:      dictRepo,
	}
}

// Categories

func (c *LearnUseCase) ListCategories(ctx context.Context) ([]model.LearnCategoryResponse, error) {
	tx := c.DB.WithContext(ctx).Begin()
	defer tx.Rollback()

	var cats []entity.LearnCategory
	if err := c.LearnCategoryRepository.FindAll(tx, &cats); err != nil {
		c.Log.Warnf("Failed list categories : %+v", err)
		return nil, fiber.ErrInternalServerError
	}

	// aggregate counts of active items per category
	counts := make(map[string]int64, len(cats))
	type row struct {
		LearnCategoryID string
		N               int64
	}
	var rows []row
	if err := tx.Model(&entity.LearnItem{}).
		Select("learn_category_id, COUNT(*) AS n").
		Where("status = ? AND learn_category_id IS NOT NULL", entity.LearnStatusActive).
		Group("learn_category_id").
		Scan(&rows).Error; err == nil {
		for _, r := range rows {
			counts[r.LearnCategoryID] = r.N
		}
	}

	out := make([]model.LearnCategoryResponse, len(cats))
	for i := range cats {
		out[i] = model.LearnCategoryResponse{
			ID:        cats[i].ID,
			Name:      cats[i].Name,
			Color:     cats[i].Color,
			SortOrder: cats[i].SortOrder,
			ItemCount: counts[cats[i].ID],
		}
	}
	if err := tx.Commit().Error; err != nil {
		c.Log.Warnf("Failed commit transaction : %+v", err)
		return nil, fiber.ErrInternalServerError
	}
	return out, nil
}

func (c *LearnUseCase) CreateCategory(ctx context.Context, req *model.CreateLearnCategoryRequest) (*model.LearnCategoryResponse, error) {
	tx := c.DB.WithContext(ctx).Begin()
	defer tx.Rollback()

	if err := c.Validate.Struct(req); err != nil {
		c.Log.Warnf("Invalid create category request : %+v", err)
		return nil, fiber.ErrBadRequest
	}
	cat := &entity.LearnCategory{
		ID:        uuid.NewString(),
		Name:      req.Name,
		Color:     req.Color,
		SortOrder: req.SortOrder,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	if err := c.LearnCategoryRepository.Create(tx, cat); err != nil {
		c.Log.Warnf("Failed to create category : %+v", err)
		return nil, fiber.ErrInternalServerError
	}
	if err := tx.Commit().Error; err != nil {
		c.Log.Warnf("Failed commit transaction : %+v", err)
		return nil, fiber.ErrInternalServerError
	}
	return &model.LearnCategoryResponse{
		ID: cat.ID, Name: cat.Name, Color: cat.Color, SortOrder: cat.SortOrder,
	}, nil
}

func (c *LearnUseCase) UpdateCategory(ctx context.Context, req *model.UpdateLearnCategoryRequest) (*model.LearnCategoryResponse, error) {
	tx := c.DB.WithContext(ctx).Begin()
	defer tx.Rollback()

	if err := c.Validate.Struct(req); err != nil {
		c.Log.Warnf("Invalid update category request : %+v", err)
		return nil, fiber.ErrBadRequest
	}
	cat := new(entity.LearnCategory)
	if err := c.LearnCategoryRepository.FindByID(tx, cat, req.ID); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fiber.ErrNotFound
		}
		c.Log.Warnf("Failed find category : %+v", err)
		return nil, fiber.ErrInternalServerError
	}
	cat.Name = req.Name
	cat.Color = req.Color
	cat.SortOrder = req.SortOrder
	cat.UpdatedAt = time.Now()
	if err := c.LearnCategoryRepository.Update(tx, cat); err != nil {
		c.Log.Warnf("Failed to update category : %+v", err)
		return nil, fiber.ErrInternalServerError
	}
	if err := tx.Commit().Error; err != nil {
		c.Log.Warnf("Failed commit transaction : %+v", err)
		return nil, fiber.ErrInternalServerError
	}
	return &model.LearnCategoryResponse{
		ID: cat.ID, Name: cat.Name, Color: cat.Color, SortOrder: cat.SortOrder,
	}, nil
}

func (c *LearnUseCase) DeleteCategory(ctx context.Context, id string) error {
	tx := c.DB.WithContext(ctx).Begin()
	defer tx.Rollback()
	// null out items in this category
	if err := tx.Model(&entity.LearnItem{}).
		Where("learn_category_id = ?", id).
		Update("learn_category_id", nil).Error; err != nil {
		c.Log.Warnf("Failed to clear category refs : %+v", err)
		return fiber.ErrInternalServerError
	}
	if err := c.LearnCategoryRepository.Delete(tx, id); err != nil {
		c.Log.Warnf("Failed to delete category : %+v", err)
		return fiber.ErrInternalServerError
	}
	if err := tx.Commit().Error; err != nil {
		c.Log.Warnf("Failed commit transaction : %+v", err)
		return fiber.ErrInternalServerError
	}
	return nil
}

// Items

func (c *LearnUseCase) ListItems(ctx context.Context, categoryID *string, sort string) ([]model.LearnItemResponse, error) {
	tx := c.DB.WithContext(ctx).Begin()
	defer tx.Rollback()

	var items []entity.LearnItem
	if err := c.LearnItemRepository.FindActive(tx, &items, categoryID, sort); err != nil {
		c.Log.Warnf("Failed list learn items : %+v", err)
		return nil, fiber.ErrInternalServerError
	}
	out, err := c.hydrateItems(tx, items)
	if err != nil {
		return nil, err
	}
	if err := tx.Commit().Error; err != nil {
		c.Log.Warnf("Failed commit transaction : %+v", err)
		return nil, fiber.ErrInternalServerError
	}
	return out, nil
}

func (c *LearnUseCase) ListArchived(ctx context.Context, limit int) ([]model.LearnItemResponse, error) {
	tx := c.DB.WithContext(ctx).Begin()
	defer tx.Rollback()

	if limit <= 0 || limit > 500 {
		limit = 100
	}
	var items []entity.LearnItem
	if err := c.LearnItemRepository.FindArchived(tx, &items, limit); err != nil {
		c.Log.Warnf("Failed list archived : %+v", err)
		return nil, fiber.ErrInternalServerError
	}
	out, err := c.hydrateItems(tx, items)
	if err != nil {
		return nil, err
	}
	if err := tx.Commit().Error; err != nil {
		c.Log.Warnf("Failed commit transaction : %+v", err)
		return nil, fiber.ErrInternalServerError
	}
	return out, nil
}

func (c *LearnUseCase) hydrateItems(tx *gorm.DB, items []entity.LearnItem) ([]model.LearnItemResponse, error) {
	// fetch all needed categories + words in one go
	catMap := map[string]string{}
	var cats []entity.LearnCategory
	if err := tx.Find(&cats).Error; err == nil {
		for _, c := range cats {
			catMap[c.ID] = c.Name
		}
	}

	out := make([]model.LearnItemResponse, 0, len(items))
	for i := range items {
		it := &items[i]
		w := new(entity.DictWord)
		if err := c.DictWordRepository.FindByID(tx, w, it.WordID); err != nil {
			continue
		}
		resp := model.LearnItemResponse{
			ID:              it.ID,
			WordID:          it.WordID,
			Lemma:           w.Lemma,
			Language:        w.Language,
			LearnCategoryID: it.LearnCategoryID,
			Status:          it.Status,
			MasteryLevel:    it.MasteryLevel,
			CreatedAt:       it.CreatedAt.Unix(),
		}
		if it.LearnCategoryID != nil {
			resp.CategoryName = catMap[*it.LearnCategoryID]
		}
		if it.ArchivedAt != nil {
			t := it.ArchivedAt.Unix()
			resp.ArchivedAt = &t
		}
		out = append(out, resp)
	}
	return out, nil
}

func (c *LearnUseCase) UpdateItem(ctx context.Context, req *model.UpdateLearnItemRequest) error {
	tx := c.DB.WithContext(ctx).Begin()
	defer tx.Rollback()
	if err := c.Validate.Struct(req); err != nil {
		c.Log.Warnf("Invalid update learn item : %+v", err)
		return fiber.ErrBadRequest
	}
	item := new(entity.LearnItem)
	if err := c.LearnItemRepository.FindByID(tx, item, req.ID); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return fiber.ErrNotFound
		}
		c.Log.Warnf("Failed find learn item : %+v", err)
		return fiber.ErrInternalServerError
	}
	if req.LearnCategoryID != nil {
		if *req.LearnCategoryID == "" {
			item.LearnCategoryID = nil
		} else {
			cid := *req.LearnCategoryID
			item.LearnCategoryID = &cid
		}
	}
	if req.MasteryLevel != nil {
		item.MasteryLevel = *req.MasteryLevel
	}
	item.UpdatedAt = time.Now()
	if err := c.LearnItemRepository.Update(tx, item); err != nil {
		c.Log.Warnf("Failed to update learn item : %+v", err)
		return fiber.ErrInternalServerError
	}
	if err := tx.Commit().Error; err != nil {
		c.Log.Warnf("Failed commit transaction : %+v", err)
		return fiber.ErrInternalServerError
	}
	return nil
}

// Done archives the item (point #6: Done -> archive, never shown again).
func (c *LearnUseCase) Done(ctx context.Context, id string) error {
	tx := c.DB.WithContext(ctx).Begin()
	defer tx.Rollback()
	item := new(entity.LearnItem)
	if err := c.LearnItemRepository.FindByID(tx, item, id); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return fiber.ErrNotFound
		}
		c.Log.Warnf("Failed find learn item : %+v", err)
		return fiber.ErrInternalServerError
	}
	now := time.Now()
	item.Status = entity.LearnStatusArchived
	item.ArchivedAt = &now
	item.UpdatedAt = now
	if err := c.LearnItemRepository.Update(tx, item); err != nil {
		c.Log.Warnf("Failed to archive learn item : %+v", err)
		return fiber.ErrInternalServerError
	}
	if err := tx.Commit().Error; err != nil {
		c.Log.Warnf("Failed commit transaction : %+v", err)
		return fiber.ErrInternalServerError
	}
	return nil
}
