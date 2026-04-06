package entity

// Country represents a country used for filtering geographic words.
type Country struct {
	ID   string `gorm:"column:id;primaryKey;type:uuid"`
	Name string `gorm:"column:name;size:128;uniqueIndex"`
	Code string `gorm:"column:code;size:8;index"`
}

func (c *Country) TableName() string { return "countries" }
