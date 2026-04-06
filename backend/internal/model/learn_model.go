package model

type LearnCategoryResponse struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	Color     string `json:"color"`
	SortOrder int    `json:"sort_order"`
	ItemCount int64  `json:"item_count"`
}

type CreateLearnCategoryRequest struct {
	Name      string `json:"name" validate:"required,max=128"`
	Color     string `json:"color" validate:"max=16"`
	SortOrder int    `json:"sort_order" validate:"min=0,max=10000"`
}

type UpdateLearnCategoryRequest struct {
	ID        string `json:"-" validate:"required,uuid"`
	Name      string `json:"name" validate:"required,max=128"`
	Color     string `json:"color" validate:"max=16"`
	SortOrder int    `json:"sort_order" validate:"min=0,max=10000"`
}

type CategoryRef struct {
	ID    string `json:"id"`
	Name  string `json:"name"`
	Color string `json:"color"`
}

type LearnItemResponse struct {
	ID           string        `json:"id"`
	WordID       string        `json:"word_id"`
	Lemma        string        `json:"lemma"`
	Language     string        `json:"language"`
	Categories   []CategoryRef `json:"categories"`
	Status       string        `json:"status"`
	MasteryLevel int           `json:"mastery_level"`
	CreatedAt    int64         `json:"created_at"`
	ArchivedAt   *int64        `json:"archived_at,omitempty"`
}

type UpdateLearnItemRequest struct {
	ID           string   `json:"-" validate:"required,uuid"`
	CategoryIDs  []string `json:"category_ids" validate:"omitempty,dive,uuid"`
	MasteryLevel *int     `json:"mastery_level" validate:"omitempty,min=0,max=5"`
}
