package entity

import "time"

// SiteCategory is a named section inside a site with its own crawl config.
type SiteCategory struct {
	ID             int64     `gorm:"column:id;primaryKey;autoIncrement"`
	GUID       string    `gorm:"column:guid;type:uuid;uniqueIndex;not null"`
	SiteID         int64     `gorm:"column:site_id;index"`
	Name           string    `gorm:"column:name;size:255"`
	StartURL       string    `gorm:"column:start_url;size:512"`
	URLPattern     string    `gorm:"column:url_pattern;size:512"`
	SelectorTitle  string    `gorm:"column:selector_title;size:255"`
	SelectorBody   string    `gorm:"column:selector_body;size:255"`
	SourceLanguage  string  `gorm:"column:source_language;size:8;default:en"`
	IsActive        bool    `gorm:"column:is_active;default:true"`
	LearnCategoryID *int64  `gorm:"column:learn_category_id"`
	CreatedAt      time.Time `gorm:"column:created_at"`
	UpdatedAt      time.Time `gorm:"column:updated_at"`
}

func (c *SiteCategory) TableName() string { return "site_categories" }
