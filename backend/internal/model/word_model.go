package model

// GetWordRequest is the request DTO for fetching a single word by slug.
type GetWordRequest struct {
	Slug string `json:"slug" validate:"required,max=100"`
}

// WordResponse is the outward shape of a word — never expose entity.Word directly.
type WordResponse struct {
	Slug       string `json:"slug"`
	Word       string `json:"word"`
	Definition string `json:"definition"`
	Example    string `json:"example"`
	Emoji      string `json:"emoji"`
}
