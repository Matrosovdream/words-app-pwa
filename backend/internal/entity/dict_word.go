package entity

import "time"

// DictWord is the canonical dictionary entry for a word (lemma form).
// Represents the `words` table (replacing the old demo table).
type DictWord struct {
	ID            int64     `gorm:"column:id;primaryKey;autoIncrement"`
	PublicID      string    `gorm:"column:public_id;type:uuid;uniqueIndex;not null"`
	Lemma         string    `gorm:"column:lemma;size:128;index:idx_lemma_lang,priority:1,unique"`
	Language      string    `gorm:"column:language;size:8;default:en;index:idx_lemma_lang,priority:2"`
	POS           string    `gorm:"column:pos;size:32"`
	FrequencyRank *int      `gorm:"column:frequency_rank;index"`
	IPA           string    `gorm:"column:ipa;size:128"`
	Definition    string    `gorm:"column:definition;size:2048"`
	CreatedAt     time.Time `gorm:"column:created_at"`
	UpdatedAt     time.Time `gorm:"column:updated_at"`
}

func (w *DictWord) TableName() string { return "words" }
