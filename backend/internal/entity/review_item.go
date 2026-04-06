package entity

import "time"

const (
	ReviewStatusPending = "pending"
	ReviewStatusAdded   = "added"
	ReviewStatusDenied  = "denied"
)

// ReviewItem is a word awaiting user decision (add to /learn or deny forever).
type ReviewItem struct {
	ID              int64      `gorm:"column:id;primaryKey;autoIncrement"`
	GUID        string     `gorm:"column:guid;type:uuid;uniqueIndex;not null"`
	WordID          int64      `gorm:"column:word_id;uniqueIndex"`
	FirstSeenPageID *int64     `gorm:"column:first_seen_page_id"`
	SiteCategoryID  *int64     `gorm:"column:site_category_id"`
	Status          string     `gorm:"column:status;size:16;index;default:pending"`
	ReviewedAt      *time.Time `gorm:"column:reviewed_at"`
	CreatedAt       time.Time  `gorm:"column:created_at"`
	UpdatedAt       time.Time  `gorm:"column:updated_at"`
}

func (r *ReviewItem) TableName() string { return "review_items" }
