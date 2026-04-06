package converter

import (
	"words-app/internal/entity"
	"words-app/internal/model"
)

func SiteToResponse(s *entity.Site) *model.SiteResponse {
	return &model.SiteResponse{
		ID:               s.ID,
		Name:             s.Name,
		BaseURL:          s.BaseURL,
		IsActive:         s.IsActive,
		DailyHitLimit:    s.DailyHitLimit,
		HitDelayMs:       s.HitDelayMs,
		ReparseEnabled:   s.ReparseEnabled,
		ReparseAfterDays: s.ReparseAfterDays,
		UserAgent:        s.UserAgent,
	}
}

func SitesToResponses(items []entity.Site) []model.SiteResponse {
	out := make([]model.SiteResponse, len(items))
	for i := range items {
		out[i] = *SiteToResponse(&items[i])
	}
	return out
}

func SiteCategoryToResponse(c *entity.SiteCategory) *model.SiteCategoryResponse {
	return &model.SiteCategoryResponse{
		ID:             c.ID,
		SiteID:         c.SiteID,
		Name:           c.Name,
		StartURL:       c.StartURL,
		URLPattern:     c.URLPattern,
		SelectorTitle:  c.SelectorTitle,
		SelectorBody:   c.SelectorBody,
		SourceLanguage: c.SourceLanguage,
		IsActive:       c.IsActive,
	}
}

func SiteCategoriesToResponses(items []entity.SiteCategory) []model.SiteCategoryResponse {
	out := make([]model.SiteCategoryResponse, len(items))
	for i := range items {
		out[i] = *SiteCategoryToResponse(&items[i])
	}
	return out
}
