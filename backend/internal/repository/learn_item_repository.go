package repository

import (
	"words-app/internal/entity"

	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

type LearnItemRepository struct {
	Log *logrus.Logger
}

func NewLearnItemRepository(log *logrus.Logger) *LearnItemRepository {
	return &LearnItemRepository{Log: log}
}

func (r *LearnItemRepository) FindByID(db *gorm.DB, out *entity.LearnItem, id string) error {
	return db.Where("id = ?", id).First(out).Error
}

func (r *LearnItemRepository) FindByWordID(db *gorm.DB, out *entity.LearnItem, wordID string) error {
	return db.Where("word_id = ?", wordID).First(out).Error
}

func (r *LearnItemRepository) FindActive(db *gorm.DB, out *[]entity.LearnItem, categoryID *string, sort string) error {
	q := db.Where("status = ?", entity.LearnStatusActive)
	if categoryID != nil {
		q = q.Where("learn_category_id = ?", *categoryID)
	}
	switch sort {
	case "oldest":
		q = q.Order("created_at ASC")
	case "mastery":
		q = q.Order("mastery_level DESC, created_at DESC")
	default:
		q = q.Order("created_at DESC")
	}
	return q.Find(out).Error
}

func (r *LearnItemRepository) FindArchived(db *gorm.DB, out *[]entity.LearnItem, limit int) error {
	return db.Where("status = ?", entity.LearnStatusArchived).
		Order("archived_at DESC").Limit(limit).Find(out).Error
}

func (r *LearnItemRepository) Create(db *gorm.DB, item *entity.LearnItem) error {
	return db.Create(item).Error
}

func (r *LearnItemRepository) Update(db *gorm.DB, item *entity.LearnItem) error {
	return db.Save(item).Error
}

func (r *LearnItemRepository) Delete(db *gorm.DB, id string) error {
	return db.Where("id = ?", id).Delete(&entity.LearnItem{}).Error
}
