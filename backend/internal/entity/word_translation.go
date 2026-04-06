package entity

import "time"

// WordTranslation stores a translation of a word in a target language.
type WordTranslation struct {
	ID             int64     `gorm:"column:id;primaryKey;autoIncrement"`
	GUID       string    `gorm:"column:guid;type:uuid;uniqueIndex;not null"`
	WordID         int64     `gorm:"column:word_id;index:idx_word_lang_trans,priority:1,unique"`
	TargetLanguage string    `gorm:"column:target_language;size:8;index:idx_word_lang_trans,priority:2"`
	Translation    string    `gorm:"column:translation;size:512;index:idx_word_lang_trans,priority:3"`
	Source         string    `gorm:"column:source;size:32;index:idx_word_lang_trans,priority:4"`
	IsPrimary      bool      `gorm:"column:is_primary;default:false"`
	FetchedAt      time.Time `gorm:"column:fetched_at"`
}

func (t *WordTranslation) TableName() string { return "word_translations" }
