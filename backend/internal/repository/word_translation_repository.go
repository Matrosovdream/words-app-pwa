package repository

import (
	"words-app/internal/entity"

	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

type WordTranslationRepository struct {
	Log *logrus.Logger
}

func NewWordTranslationRepository(log *logrus.Logger) *WordTranslationRepository {
	return &WordTranslationRepository{Log: log}
}

func (r *WordTranslationRepository) FindByWordAndLang(db *gorm.DB, out *[]entity.WordTranslation, wordID, lang string) error {
	return db.Where("word_id = ? AND target_language = ?", wordID, lang).
		Order("is_primary DESC, fetched_at DESC").Find(out).Error
}

func (r *WordTranslationRepository) Create(db *gorm.DB, t *entity.WordTranslation) error {
	return db.Create(t).Error
}

func (r *WordTranslationRepository) DeletePrimaryFlag(db *gorm.DB, wordID, lang string) error {
	return db.Model(&entity.WordTranslation{}).
		Where("word_id = ? AND target_language = ?", wordID, lang).
		Update("is_primary", false).Error
}
