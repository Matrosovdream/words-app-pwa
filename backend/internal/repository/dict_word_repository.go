package repository

import (
	"words-app/internal/entity"

	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

type DictWordRepository struct {
	Log *logrus.Logger
}

func NewDictWordRepository(log *logrus.Logger) *DictWordRepository {
	return &DictWordRepository{Log: log}
}

func (r *DictWordRepository) FindByGUID(db *gorm.DB, out *entity.DictWord, publicID string) error {
	return db.Where("guid = ?", publicID).First(out).Error
}

func (r *DictWordRepository) FindByID(db *gorm.DB, out *entity.DictWord, id int64) error {
	return db.Where("id = ?", id).First(out).Error
}

func (r *DictWordRepository) FindByLemma(db *gorm.DB, out *entity.DictWord, lemma, language string) error {
	return db.Where("lemma = ? AND language = ?", lemma, language).First(out).Error
}

func (r *DictWordRepository) Create(db *gorm.DB, w *entity.DictWord) error {
	return db.Create(w).Error
}

func (r *DictWordRepository) Update(db *gorm.DB, w *entity.DictWord) error {
	return db.Save(w).Error
}

// FindOrCreate looks up by (lemma, language); creates a rank-less entry if absent.
func (r *DictWordRepository) FindOrCreate(db *gorm.DB, w *entity.DictWord) error {
	return db.Where(entity.DictWord{Lemma: w.Lemma, Language: w.Language}).
		FirstOrCreate(w).Error
}
