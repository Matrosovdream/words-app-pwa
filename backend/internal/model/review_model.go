package model

type ReviewItemResponse struct {
	ID               string  `json:"id"`
	WordID           string  `json:"word_id"`
	Lemma            string  `json:"lemma"`
	Language         string  `json:"language"`
	SampleSentence   string  `json:"sample_sentence"`
	FirstSeenPageID  *string `json:"first_seen_page_id,omitempty"`
	SourceURL        string  `json:"source_url,omitempty"`
	SiteCategoryName *string `json:"site_category_name,omitempty"`
	Status           string  `json:"status"`
	CreatedAt        int64   `json:"created_at"`
}

type ReviewDecisionRequest struct {
	ID              string `json:"-" validate:"required,uuid"`
	LearnCategoryID string `json:"learn_category_id" validate:"omitempty,uuid"`
}
