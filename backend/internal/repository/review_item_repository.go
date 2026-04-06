package repository

import (
	"words-app/internal/entity"

	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

type ReviewItemRepository struct {
	Log *logrus.Logger
}

func NewReviewItemRepository(log *logrus.Logger) *ReviewItemRepository {
	return &ReviewItemRepository{Log: log}
}

func (r *ReviewItemRepository) FindPending(db *gorm.DB, out *[]entity.ReviewItem, limit int) error {
	return db.Where("status = ?", entity.ReviewStatusPending).
		Order("created_at DESC").Limit(limit).Find(out).Error
}

func (r *ReviewItemRepository) FindByGUID(db *gorm.DB, out *entity.ReviewItem, publicID string) error {
	return db.Where("guid = ?", publicID).First(out).Error
}

func (r *ReviewItemRepository) FindByWordID(db *gorm.DB, out *entity.ReviewItem, wordID int64) error {
	return db.Where("word_id = ?", wordID).First(out).Error
}

func (r *ReviewItemRepository) Create(db *gorm.DB, item *entity.ReviewItem) error {
	return db.Create(item).Error
}

func (r *ReviewItemRepository) Update(db *gorm.DB, item *entity.ReviewItem) error {
	return db.Save(item).Error
}

func (r *ReviewItemRepository) CountPending(db *gorm.DB) (int64, error) {
	var n int64
	err := db.Model(&entity.ReviewItem{}).Where("status = ?", entity.ReviewStatusPending).Count(&n).Error
	return n, err
}
