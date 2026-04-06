package entity

import "time"

const (
	RelationTypeSynonym = "synonym"
	RelationTypeAntonym = "antonym"
)

// WordRelation is a synonym / antonym between two dictionary words.
// For simplicity we store the related word as free text (it may not yet be in the dict).
type WordRelation struct {
	ID           string    `gorm:"column:id;primaryKey;type:uuid"`
	WordID       string    `gorm:"column:word_id;type:uuid;index:idx_word_rel,priority:1,unique"`
	RelatedText  string    `gorm:"column:related_text;size:128;index:idx_word_rel,priority:2"`
	RelationType string    `gorm:"column:relation_type;size:16;index:idx_word_rel,priority:3"`
	Source       string    `gorm:"column:source;size:32;index:idx_word_rel,priority:4"`
	FetchedAt    time.Time `gorm:"column:fetched_at"`
}

func (r *WordRelation) TableName() string { return "word_relations" }
