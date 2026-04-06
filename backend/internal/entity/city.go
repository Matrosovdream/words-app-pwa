package entity

// City represents a US city used for filtering geographic words.
type City struct {
	ID   int64  `gorm:"column:id;primaryKey;autoIncrement"`
	Name string `gorm:"column:name;size:128;index"`
}

func (c *City) TableName() string { return "cities" }
