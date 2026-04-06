package usecase

import (
	"context"
	"time"

	"words-app/internal/entity"
	"words-app/internal/model"
	"words-app/internal/repository"

	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

type SettingUseCase struct {
	DB             *gorm.DB
	Log            *logrus.Logger
	Validate       *validator.Validate
	SettingRepository *repository.AppSettingRepository
}

func NewSettingUseCase(db *gorm.DB, log *logrus.Logger, validate *validator.Validate,
	settingRepo *repository.AppSettingRepository) *SettingUseCase {
	return &SettingUseCase{
		DB: db, Log: log, Validate: validate, SettingRepository: settingRepo,
	}
}

func (c *SettingUseCase) GetTranslation(ctx context.Context) (*model.TranslationSettingsResponse, error) {
	tx := c.DB.WithContext(ctx).Begin()
	defer tx.Rollback()

	apiKey, _ := c.SettingRepository.Get(tx, entity.SettingDeepLAPIKey)
	endpoint, _ := c.SettingRepository.Get(tx, entity.SettingDeepLEndpoint)
	target, _ := c.SettingRepository.Get(tx, entity.SettingDefaultTargetLang)
	relSource, _ := c.SettingRepository.Get(tx, entity.SettingRelationsSource)

	if err := tx.Commit().Error; err != nil {
		c.Log.Warnf("Failed commit transaction : %+v", err)
		return nil, fiber.ErrInternalServerError
	}
	return &model.TranslationSettingsResponse{
		DeepLAPIKeySet:        apiKey != "",
		DeepLEndpoint:         endpoint,
		DefaultTargetLanguage: target,
		RelationsSource:       relSource,
	}, nil
}

func (c *SettingUseCase) UpdateTranslation(ctx context.Context, req *model.UpdateTranslationSettingsRequest) (*model.TranslationSettingsResponse, error) {
	tx := c.DB.WithContext(ctx).Begin()
	defer tx.Rollback()

	if err := c.Validate.Struct(req); err != nil {
		c.Log.Warnf("Invalid update settings : %+v", err)
		return nil, fiber.ErrBadRequest
	}

	if req.DeepLAPIKey != nil {
		if err := c.SettingRepository.Set(tx, &entity.AppSetting{
			Key: entity.SettingDeepLAPIKey, Value: *req.DeepLAPIKey, UpdatedAt: time.Now(),
		}); err != nil {
			c.Log.Warnf("Failed to save deepl key : %+v", err)
			return nil, fiber.ErrInternalServerError
		}
	}
	if req.DeepLEndpoint != nil {
		if err := c.SettingRepository.Set(tx, &entity.AppSetting{
			Key: entity.SettingDeepLEndpoint, Value: *req.DeepLEndpoint, UpdatedAt: time.Now(),
		}); err != nil {
			c.Log.Warnf("Failed to save deepl endpoint : %+v", err)
			return nil, fiber.ErrInternalServerError
		}
	}
	if req.DefaultTargetLanguage != nil {
		if err := c.SettingRepository.Set(tx, &entity.AppSetting{
			Key: entity.SettingDefaultTargetLang, Value: *req.DefaultTargetLanguage, UpdatedAt: time.Now(),
		}); err != nil {
			c.Log.Warnf("Failed to save default target lang : %+v", err)
			return nil, fiber.ErrInternalServerError
		}
	}
	if req.RelationsSource != nil {
		if err := c.SettingRepository.Set(tx, &entity.AppSetting{
			Key: entity.SettingRelationsSource, Value: *req.RelationsSource, UpdatedAt: time.Now(),
		}); err != nil {
			c.Log.Warnf("Failed to save relations source : %+v", err)
			return nil, fiber.ErrInternalServerError
		}
	}

	if err := tx.Commit().Error; err != nil {
		c.Log.Warnf("Failed commit transaction : %+v", err)
		return nil, fiber.ErrInternalServerError
	}
	return c.GetTranslation(ctx)
}
