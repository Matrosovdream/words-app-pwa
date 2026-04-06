package entity

// LearnItemCategory is the join table linking a learn item to one or more categories.
type LearnItemCategory struct {
	LearnItemID     int64 `gorm:"column:learn_item_id;primaryKey"`
	LearnCategoryID int64 `gorm:"column:learn_category_id;primaryKey"`
}

func (l *LearnItemCategory) TableName() string { return "learn_item_categories" }
