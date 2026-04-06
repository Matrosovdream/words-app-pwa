package repository

import (
	"words-app/internal/entity"

	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

type ParsedPageRepository struct {
	Log *logrus.Logger
}

func NewParsedPageRepository(log *logrus.Logger) *ParsedPageRepository {
	return &ParsedPageRepository{Log: log}
}

func (r *ParsedPageRepository) FindByHash(db *gorm.DB, out *entity.ParsedPage, siteID, urlHash string) error {
	return db.Where("site_id = ? AND url_hash = ?", siteID, urlHash).First(out).Error
}

func (r *ParsedPageRepository) Create(db *gorm.DB, p *entity.ParsedPage) error {
	return db.Create(p).Error
}

func (r *ParsedPageRepository) Update(db *gorm.DB, p *entity.ParsedPage) error {
	return db.Save(p).Error
}

func (r *ParsedPageRepository) FindRecent(db *gorm.DB, out *[]entity.ParsedPage, limit int) error {
	return db.Order("parsed_at DESC").Limit(limit).Find(out).Error
}
