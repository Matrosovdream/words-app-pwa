package entity

// Word is the persistence-layer representation of a word in the collection.
// It maps to the `words` table via GORM.
type Word struct {
	Slug       string `gorm:"column:slug;primaryKey"`
	Word       string `gorm:"column:word"`
	Definition string `gorm:"column:definition"`
	Example    string `gorm:"column:example"`
	Emoji      string `gorm:"column:emoji"`
}

func (w *Word) TableName() string {
	return "words"
}
