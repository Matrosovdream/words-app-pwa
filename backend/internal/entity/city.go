package entity

// City represents a US city used for filtering geographic words.
type City struct {
	ID   string `gorm:"column:id;primaryKey;type:uuid"`
	Name string `gorm:"column:name;size:128;index"`
}

func (c *City) TableName() string { return "cities" }
