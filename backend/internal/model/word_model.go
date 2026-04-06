package model

type WordTranslationDTO struct {
	Language    string `json:"language"`
	Translation string `json:"translation"`
	Source      string `json:"source"`
	IsPrimary   bool   `json:"is_primary"`
}

type WordRelationDTO struct {
	Type   string `json:"type"`
	Text   string `json:"text"`
	Source string `json:"source"`
}

type WordOccurrenceDTO struct {
	Sentence  string `json:"sentence"`
	SourceURL string `json:"source_url,omitempty"`
	PageTitle string `json:"page_title,omitempty"`
}

type WordDetailResponse struct {
	ID            string               `json:"id"`
	Lemma         string               `json:"lemma"`
	Language      string               `json:"language"`
	POS           string               `json:"pos,omitempty"`
	FrequencyRank *int                 `json:"frequency_rank,omitempty"`
	IPA           string               `json:"ipa,omitempty"`
	Definition    string               `json:"definition,omitempty"`
	Translations  []WordTranslationDTO `json:"translations"`
	Synonyms      []WordRelationDTO    `json:"synonyms"`
	Antonyms      []WordRelationDTO    `json:"antonyms"`
	Occurrences   []WordOccurrenceDTO  `json:"occurrences"`
}
