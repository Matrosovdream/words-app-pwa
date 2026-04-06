package repository

import (
	"words-app/internal/entity"

	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

type SiteRepository struct {
	Log *logrus.Logger
}

func NewSiteRepository(log *logrus.Logger) *SiteRepository {
	return &SiteRepository{Log: log}
}

func (r *SiteRepository) FindAll(db *gorm.DB, sites *[]entity.Site) error {
	return db.Order("created_at DESC").Find(sites).Error
}

func (r *SiteRepository) FindByID(db *gorm.DB, site *entity.Site, id string) error {
	return db.Where("id = ?", id).First(site).Error
}

func (r *SiteRepository) Create(db *gorm.DB, site *entity.Site) error {
	return db.Create(site).Error
}

func (r *SiteRepository) Update(db *gorm.DB, site *entity.Site) error {
	return db.Save(site).Error
}

func (r *SiteRepository) Delete(db *gorm.DB, id string) error {
	return db.Where("id = ?", id).Delete(&entity.Site{}).Error
}

func (r *SiteRepository) FindActive(db *gorm.DB, sites *[]entity.Site) error {
	return db.Where("is_active = ?", true).Find(sites).Error
}
