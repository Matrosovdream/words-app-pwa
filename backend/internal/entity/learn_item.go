package entity

import "time"

const (
	LearnStatusActive   = "active"
	LearnStatusArchived = "archived"
)

// LearnItem is a word the user accepted into their learning list.
type LearnItem struct {
	ID              string     `gorm:"column:id;primaryKey;type:uuid"`
	WordID          string     `gorm:"column:word_id;type:uuid;uniqueIndex"`
	LearnCategoryID *string    `gorm:"column:learn_category_id;type:uuid;index"`
	Status          string     `gorm:"column:status;size:16;index;default:active"`
	MasteryLevel    int        `gorm:"column:mastery_level;default:0"`
	LastReviewedAt  *time.Time `gorm:"column:last_reviewed_at"`
	ArchivedAt      *time.Time `gorm:"column:archived_at"`
	CreatedAt       time.Time  `gorm:"column:created_at"`
	UpdatedAt       time.Time  `gorm:"column:updated_at"`
}

func (l *LearnItem) TableName() string { return "learn_items" }
