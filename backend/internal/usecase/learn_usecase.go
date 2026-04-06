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
	DB                          *gorm.DB
	Log                         *logrus.Logger
	Validate                    *validator.Validate
	LearnCategoryRepository     *repository.LearnCategoryRepository
	LearnItemRepository         *repository.LearnItemRepository
	LearnItemCategoryRepository *repository.LearnItemCategoryRepository
	DictWordRepository          *repository.DictWordRepository
}

func NewLearnUseCase(db *gorm.DB, log *logrus.Logger, validate *validator.Validate,
	catRepo *repository.LearnCategoryRepository,
	itemRepo *repository.LearnItemRepository,
	itemCatRepo *repository.LearnItemCategoryRepository,
	dictRepo *repository.DictWordRepository) *LearnUseCase {
	return &LearnUseCase{
		DB:                          db,
		Log:                         log,
		Validate:                    validate,
		LearnCategoryRepository:     catRepo,
		LearnItemRepository:         itemRepo,
		LearnItemCategoryRepository: itemCatRepo,
		DictWordRepository:          dictRepo,
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

	// aggregate counts via join table
	counts := make(map[int64]int64, len(cats))
	type row struct {
		LearnCategoryID int64
		N               int64
	}
	var rows []row
	if err := tx.Table("learn_item_categories lic").
		Select("lic.learn_category_id, COUNT(*) AS n").
		Joins("JOIN learn_items li ON li.id = lic.learn_item_id").
		Where("li.status = ?", entity.LearnStatusActive).
		Group("lic.learn_category_id").
		Scan(&rows).Error; err == nil {
		for _, r := range rows {
			counts[r.LearnCategoryID] = r.N
		}
	}

	out := make([]model.LearnCategoryResponse, len(cats))
	for i := range cats {
		out[i] = model.LearnCategoryResponse{
			ID:        cats[i].GUID,
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
		GUID:  uuid.NewString(),
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
		ID: cat.GUID, Name: cat.Name, Color: cat.Color, SortOrder: cat.SortOrder,
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
	if err := c.LearnCategoryRepository.FindByGUID(tx, cat, req.ID); err != nil {
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
		ID: cat.GUID, Name: cat.Name, Color: cat.Color, SortOrder: cat.SortOrder,
	}, nil
}

func (c *LearnUseCase) DeleteCategory(ctx context.Context, publicID string) error {
	tx := c.DB.WithContext(ctx).Begin()
	defer tx.Rollback()

	cat := new(entity.LearnCategory)
	if err := c.LearnCategoryRepository.FindByGUID(tx, cat, publicID); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return fiber.ErrNotFound
		}
		c.Log.Warnf("Failed find category : %+v", err)
		return fiber.ErrInternalServerError
	}

	// null out items in this category
	if err := tx.Model(&entity.LearnItem{}).
		Where("learn_category_id = ?", cat.ID).
		Update("learn_category_id", nil).Error; err != nil {
		c.Log.Warnf("Failed to clear category refs : %+v", err)
		return fiber.ErrInternalServerError
	}
	if err := c.LearnCategoryRepository.Delete(tx, cat.ID); err != nil {
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

func (c *LearnUseCase) ListItems(ctx context.Context, categoryGUID *string, sort string) ([]model.LearnItemResponse, error) {
	tx := c.DB.WithContext(ctx).Begin()
	defer tx.Rollback()

	var categoryID *int64
	if categoryGUID != nil {
		cat := new(entity.LearnCategory)
		if err := c.LearnCategoryRepository.FindByGUID(tx, cat, *categoryGUID); err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return nil, fiber.ErrNotFound
			}
			c.Log.Warnf("Failed find category : %+v", err)
			return nil, fiber.ErrInternalServerError
		}
		categoryID = &cat.ID
	}

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
	// build lookup maps for all categories
	type catInfo struct {
		GUID  string
		Name  string
		Color string
	}
	catByID := map[int64]catInfo{}
	var cats []entity.LearnCategory
	if err := tx.Find(&cats).Error; err == nil {
		for _, cat := range cats {
			catByID[cat.ID] = catInfo{GUID: cat.GUID, Name: cat.Name, Color: cat.Color}
		}
	}

	// batch-load join table rows for all items
	itemIDs := make([]int64, 0, len(items))
	for i := range items {
		itemIDs = append(itemIDs, items[i].ID)
	}
	var links []entity.LearnItemCategory
	if len(itemIDs) > 0 {
		_ = c.LearnItemCategoryRepository.FindByItemIDs(tx, &links, itemIDs)
	}
	// group by item ID
	itemCats := make(map[int64][]model.CategoryRef, len(items))
	for _, lnk := range links {
		if info, ok := catByID[lnk.LearnCategoryID]; ok {
			itemCats[lnk.LearnItemID] = append(itemCats[lnk.LearnItemID], model.CategoryRef{
				ID: info.GUID, Name: info.Name, Color: info.Color,
			})
		}
	}

	out := make([]model.LearnItemResponse, 0, len(items))
	for i := range items {
		it := &items[i]
		w := new(entity.DictWord)
		if err := c.DictWordRepository.FindByID(tx, w, it.WordID); err != nil {
			continue
		}
		cats := itemCats[it.ID]
		if cats == nil {
			cats = []model.CategoryRef{}
		}
		resp := model.LearnItemResponse{
			ID:           it.GUID,
			WordID:       w.GUID,
			Lemma:        w.Lemma,
			Language:     w.Language,
			Categories:   cats,
			Status:       it.Status,
			MasteryLevel: it.MasteryLevel,
			CreatedAt:    it.CreatedAt.Unix(),
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
	if err := c.LearnItemRepository.FindByGUID(tx, item, req.ID); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return fiber.ErrNotFound
		}
		c.Log.Warnf("Failed find learn item : %+v", err)
		return fiber.ErrInternalServerError
	}
	if req.CategoryIDs != nil {
		// replace all category associations
		if err := c.LearnItemCategoryRepository.DeleteByItem(tx, item.ID); err != nil {
			c.Log.Warnf("Failed to clear item categories : %+v", err)
			return fiber.ErrInternalServerError
		}
		for _, catGUID := range req.CategoryIDs {
			cat := new(entity.LearnCategory)
			if err := c.LearnCategoryRepository.FindByGUID(tx, cat, catGUID); err != nil {
				if errors.Is(err, gorm.ErrRecordNotFound) {
					return fiber.ErrNotFound
				}
				c.Log.Warnf("Failed find category : %+v", err)
				return fiber.ErrInternalServerError
			}
			if err := c.LearnItemCategoryRepository.Create(tx, &entity.LearnItemCategory{
				LearnItemID: item.ID, LearnCategoryID: cat.ID,
			}); err != nil {
				c.Log.Warnf("Failed to add item category : %+v", err)
				return fiber.ErrInternalServerError
			}
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
func (c *LearnUseCase) Done(ctx context.Context, publicID string) error {
	tx := c.DB.WithContext(ctx).Begin()
	defer tx.Rollback()
	item := new(entity.LearnItem)
	if err := c.LearnItemRepository.FindByGUID(tx, item, publicID); err != nil {
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
