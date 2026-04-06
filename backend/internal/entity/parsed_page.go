package entity

import "time"

// ParsedPage is the history record for a URL that was already processed.
type ParsedPage struct {
	ID             int64   `gorm:"column:id;primaryKey;autoIncrement"`
	PublicID       string  `gorm:"column:public_id;type:uuid;uniqueIndex;not null"`
	SiteID         int64   `gorm:"column:site_id;index:idx_site_url,priority:1"`
	SiteCategoryID *int64  `gorm:"column:site_category_id"`
	URL            string  `gorm:"column:url;size:1024"`
	URLHash        string  `gorm:"column:url_hash;size:64;index:idx_site_url,priority:2,unique"`
	ContentHash    string  `gorm:"column:content_hash;size:64"`
	Title          string  `gorm:"column:title;size:512"`
	WordCount      int     `gorm:"column:word_count"`
	ParsedAt       time.Time `gorm:"column:parsed_at;index"`
}

func (p *ParsedPage) TableName() string { return "parsed_pages" }
