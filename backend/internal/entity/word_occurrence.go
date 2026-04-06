package entity

import "time"

// WordOccurrence links a word to the page where it was found.
type WordOccurrence struct {
	ID             int64     `gorm:"column:id;primaryKey;autoIncrement"`
	GUID       string    `gorm:"column:guid;type:uuid;uniqueIndex;not null"`
	WordID         int64     `gorm:"column:word_id;index:idx_word_page,priority:1,unique"`
	ParsedPageID   int64     `gorm:"column:parsed_page_id;index:idx_word_page,priority:2"`
	Count          int       `gorm:"column:count;default:1"`
	SampleSentence string    `gorm:"column:sample_sentence;size:1024"`
	CreatedAt      time.Time `gorm:"column:created_at"`
}

func (o *WordOccurrence) TableName() string { return "word_occurrences" }
