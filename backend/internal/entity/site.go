package entity

import "time"

// Site is a source to parse words from.
type Site struct {
	ID               int64     `gorm:"column:id;primaryKey;autoIncrement"`
	GUID         string    `gorm:"column:guid;type:uuid;uniqueIndex;not null"`
	Name             string    `gorm:"column:name;size:255"`
	BaseURL          string    `gorm:"column:base_url;size:512"`
	IsActive         bool      `gorm:"column:is_active;default:true"`
	DailyHitLimit    int       `gorm:"column:daily_hit_limit;default:100"`
	HitDelayMs       int       `gorm:"column:hit_delay_ms;default:2000"`
	ReparseEnabled   bool      `gorm:"column:reparse_enabled;default:false"`
	ReparseAfterDays int       `gorm:"column:reparse_after_days;default:30"`
	UserAgent        string    `gorm:"column:user_agent;size:255"`
	CreatedAt        time.Time `gorm:"column:created_at"`
	UpdatedAt        time.Time `gorm:"column:updated_at"`
}

func (s *Site) TableName() string { return "sites" }
