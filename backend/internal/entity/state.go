package entity

// State represents a US state used for filtering geographic words.
type State struct {
	ID   string `gorm:"column:id;primaryKey;type:uuid"`
	Name string `gorm:"column:name;size:128;uniqueIndex"`
	Code string `gorm:"column:code;size:8;index"`
}

func (s *State) TableName() string { return "states" }
