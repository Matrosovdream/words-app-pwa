package repository

import (
	"words-app/internal/entity"

	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

type LearnCategoryRepository struct {
	Log *logrus.Logger
}

func NewLearnCategoryRepository(log *logrus.Logger) *LearnCategoryRepository {
	return &LearnCategoryRepository{Log: log}
}

func (r *LearnCategoryRepository) FindAll(db *gorm.DB, out *[]entity.LearnCategory) error {
	return db.Order("sort_order ASC, name ASC").Find(out).Error
}

func (r *LearnCategoryRepository) FindByID(db *gorm.DB, out *entity.LearnCategory, id string) error {
	return db.Where("id = ?", id).First(out).Error
}

func (r *LearnCategoryRepository) Create(db *gorm.DB, c *entity.LearnCategory) error {
	return db.Create(c).Error
}

func (r *LearnCategoryRepository) Update(db *gorm.DB, c *entity.LearnCategory) error {
	return db.Save(c).Error
}

func (r *LearnCategoryRepository) Delete(db *gorm.DB, id string) error {
	return db.Where("id = ?", id).Delete(&entity.LearnCategory{}).Error
}
