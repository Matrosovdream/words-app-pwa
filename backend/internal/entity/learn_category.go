package entity

import "time"

// LearnCategory is a user-defined folder under /learn.
type LearnCategory struct {
	ID        string    `gorm:"column:id;primaryKey;type:uuid"`
	Name      string    `gorm:"column:name;size:128"`
	Color     string    `gorm:"column:color;size:16"`
	SortOrder int       `gorm:"column:sort_order;default:0"`
	CreatedAt time.Time `gorm:"column:created_at"`
	UpdatedAt time.Time `gorm:"column:updated_at"`
}

func (c *LearnCategory) TableName() string { return "learn_categories" }
