package repository

import (
	"words-app/internal/entity"

	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

type WordRelationRepository struct {
	Log *logrus.Logger
}

func NewWordRelationRepository(log *logrus.Logger) *WordRelationRepository {
	return &WordRelationRepository{Log: log}
}

func (r *WordRelationRepository) FindByWord(db *gorm.DB, out *[]entity.WordRelation, wordID string) error {
	return db.Where("word_id = ?", wordID).Order("relation_type ASC, related_text ASC").Find(out).Error
}

func (r *WordRelationRepository) Create(db *gorm.DB, rel *entity.WordRelation) error {
	return db.Create(rel).Error
}

func (r *WordRelationRepository) CountByWord(db *gorm.DB, wordID string) (int64, error) {
	var n int64
	err := db.Model(&entity.WordRelation{}).Where("word_id = ?", wordID).Count(&n).Error
	return n, err
}
