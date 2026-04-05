package repository

import (
	"words-app/internal/entity"

	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

type WordRepository struct {
	Log *logrus.Logger
}

func NewWordRepository(log *logrus.Logger) *WordRepository {
	return &WordRepository{Log: log}
}

func (r *WordRepository) FindAll(db *gorm.DB, words *[]entity.Word) error {
	return db.Order("word ASC").Find(words).Error
}

func (r *WordRepository) FindBySlug(db *gorm.DB, word *entity.Word, slug string) error {
	return db.Where("slug = ?", slug).First(word).Error
}
