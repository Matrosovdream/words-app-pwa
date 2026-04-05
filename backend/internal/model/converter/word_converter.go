package converter

import (
	"words-app/internal/entity"
	"words-app/internal/model"
)

func WordToResponse(word *entity.Word) *model.WordResponse {
	return &model.WordResponse{
		Slug:       word.Slug,
		Word:       word.Word,
		Definition: word.Definition,
		Example:    word.Example,
		Emoji:      word.Emoji,
	}
}

func WordsToResponses(words []entity.Word) []model.WordResponse {
	out := make([]model.WordResponse, len(words))
	for i := range words {
		out[i] = *WordToResponse(&words[i])
	}
	return out
}
