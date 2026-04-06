package repository

import (
	"words-app/internal/entity"

	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

type AppSettingRepository struct {
	Log *logrus.Logger
}

func NewAppSettingRepository(log *logrus.Logger) *AppSettingRepository {
	return &AppSettingRepository{Log: log}
}

func (r *AppSettingRepository) FindAll(db *gorm.DB, out *[]entity.AppSetting) error {
	return db.Order("key ASC").Find(out).Error
}

func (r *AppSettingRepository) Get(db *gorm.DB, key string) (string, error) {
	var s entity.AppSetting
	if err := db.Where("key = ?", key).First(&s).Error; err != nil {
		return "", err
	}
	return s.Value, nil
}

func (r *AppSettingRepository) Set(db *gorm.DB, s *entity.AppSetting) error {
	return db.Save(s).Error
}
