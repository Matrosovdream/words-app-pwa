package repository

import (
	"words-app/internal/entity"

	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

type SiteCategoryRepository struct {
	Log *logrus.Logger
}

func NewSiteCategoryRepository(log *logrus.Logger) *SiteCategoryRepository {
	return &SiteCategoryRepository{Log: log}
}

func (r *SiteCategoryRepository) FindBySite(db *gorm.DB, out *[]entity.SiteCategory, siteID int64) error {
	return db.Where("site_id = ?", siteID).Order("name ASC").Find(out).Error
}

func (r *SiteCategoryRepository) FindByGUID(db *gorm.DB, out *entity.SiteCategory, publicID string) error {
	return db.Where("guid = ?", publicID).First(out).Error
}

func (r *SiteCategoryRepository) FindByID(db *gorm.DB, out *entity.SiteCategory, id int64) error {
	return db.Where("id = ?", id).First(out).Error
}

func (r *SiteCategoryRepository) Create(db *gorm.DB, c *entity.SiteCategory) error {
	return db.Create(c).Error
}

func (r *SiteCategoryRepository) Update(db *gorm.DB, c *entity.SiteCategory) error {
	return db.Save(c).Error
}

func (r *SiteCategoryRepository) Delete(db *gorm.DB, id int64) error {
	return db.Where("id = ?", id).Delete(&entity.SiteCategory{}).Error
}
