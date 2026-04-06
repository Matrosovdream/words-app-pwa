package entity

// State represents a US state used for filtering geographic words.
type State struct {
	ID   int64  `gorm:"column:id;primaryKey;autoIncrement"`
	Name string `gorm:"column:name;size:128;uniqueIndex"`
	Code string `gorm:"column:code;size:8;index"`
}

func (s *State) TableName() string { return "states" }
