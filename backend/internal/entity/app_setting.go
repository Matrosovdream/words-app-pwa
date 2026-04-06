package entity

import "time"

// AppSetting is a key/value row for runtime configuration (API keys, default target language, etc).
type AppSetting struct {
	Key       string    `gorm:"column:key;primaryKey;size:64"`
	Value     string    `gorm:"column:value;size:2048"`
	UpdatedAt time.Time `gorm:"column:updated_at"`
}

func (s *AppSetting) TableName() string { return "app_settings" }

const (
	SettingDeepLAPIKey        = "deepl_api_key"
	SettingDeepLEndpoint      = "deepl_endpoint"
	SettingDefaultTargetLang  = "default_target_language"
	SettingRelationsSource    = "relations_source"
)
