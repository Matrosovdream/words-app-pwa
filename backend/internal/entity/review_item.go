package entity

import "time"

const (
	ReviewStatusPending = "pending"
	ReviewStatusAdded   = "added"
	ReviewStatusDenied  = "denied"
)

// ReviewItem is a word awaiting user decision (add to /learn or deny forever).
type ReviewItem struct {
	ID              string     `gorm:"column:id;primaryKey;type:uuid"`
	WordID          string     `gorm:"column:word_id;type:uuid;uniqueIndex"`
	FirstSeenPageID *string    `gorm:"column:first_seen_page_id;type:uuid"`
	Status          string     `gorm:"column:status;size:16;index;default:pending"`
	ReviewedAt      *time.Time `gorm:"column:reviewed_at"`
	CreatedAt       time.Time  `gorm:"column:created_at"`
	UpdatedAt       time.Time  `gorm:"column:updated_at"`
}

func (r *ReviewItem) TableName() string { return "review_items" }
