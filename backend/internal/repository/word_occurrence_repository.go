package repository

import (
	"words-app/internal/entity"

	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

type WordOccurrenceRepository struct {
	Log *logrus.Logger
}

func NewWordOccurrenceRepository(log *logrus.Logger) *WordOccurrenceRepository {
	return &WordOccurrenceRepository{Log: log}
}

func (r *WordOccurrenceRepository) Create(db *gorm.DB, o *entity.WordOccurrence) error {
	return db.Create(o).Error
}

func (r *WordOccurrenceRepository) FindByWord(db *gorm.DB, out *[]entity.WordOccurrence, wordID string, limit int) error {
	return db.Where("word_id = ?", wordID).Order("created_at DESC").Limit(limit).Find(out).Error
}
