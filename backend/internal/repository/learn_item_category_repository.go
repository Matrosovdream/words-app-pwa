package repository

import (
	"words-app/internal/entity"

	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type LearnItemCategoryRepository struct {
	Log *logrus.Logger
}

func NewLearnItemCategoryRepository(log *logrus.Logger) *LearnItemCategoryRepository {
	return &LearnItemCategoryRepository{Log: log}
}

func (r *LearnItemCategoryRepository) Create(db *gorm.DB, lc *entity.LearnItemCategory) error {
	return db.Clauses(clause.OnConflict{DoNothing: true}).Create(lc).Error
}

func (r *LearnItemCategoryRepository) FindByItemIDs(db *gorm.DB, out *[]entity.LearnItemCategory, itemIDs []int64) error {
	return db.Where("learn_item_id IN ?", itemIDs).Find(out).Error
}

func (r *LearnItemCategoryRepository) DeleteByItem(db *gorm.DB, itemID int64) error {
	return db.Where("learn_item_id = ?", itemID).Delete(&entity.LearnItemCategory{}).Error
}
