package model

type SiteResponse struct {
	ID               string `json:"id"`
	Name             string `json:"name"`
	BaseURL          string `json:"base_url"`
	IsActive         bool   `json:"is_active"`
	DailyHitLimit    int    `json:"daily_hit_limit"`
	HitDelayMs       int    `json:"hit_delay_ms"`
	ReparseEnabled   bool   `json:"reparse_enabled"`
	ReparseAfterDays int    `json:"reparse_after_days"`
	UserAgent        string `json:"user_agent"`
}

type CreateSiteRequest struct {
	Name             string `json:"name" validate:"required,max=255"`
	BaseURL          string `json:"base_url" validate:"required,url,max=512"`
	IsActive         *bool  `json:"is_active"`
	DailyHitLimit    int    `json:"daily_hit_limit" validate:"min=0,max=100000"`
	HitDelayMs       int    `json:"hit_delay_ms" validate:"min=0,max=600000"`
	ReparseEnabled   *bool  `json:"reparse_enabled"`
	ReparseAfterDays int    `json:"reparse_after_days" validate:"min=0,max=3650"`
	UserAgent        string `json:"user_agent" validate:"max=255"`
}

type UpdateSiteRequest struct {
	ID               string `json:"-" validate:"required,uuid"`
	Name             string `json:"name" validate:"required,max=255"`
	BaseURL          string `json:"base_url" validate:"required,url,max=512"`
	IsActive         *bool  `json:"is_active"`
	DailyHitLimit    int    `json:"daily_hit_limit" validate:"min=0,max=100000"`
	HitDelayMs       int    `json:"hit_delay_ms" validate:"min=0,max=600000"`
	ReparseEnabled   *bool  `json:"reparse_enabled"`
	ReparseAfterDays int    `json:"reparse_after_days" validate:"min=0,max=3650"`
	UserAgent        string `json:"user_agent" validate:"max=255"`
}

type SiteCategoryResponse struct {
	ID             string `json:"id"`
	SiteID         string `json:"site_id"`
	Name           string `json:"name"`
	StartURL       string `json:"start_url"`
	URLPattern     string `json:"url_pattern"`
	SelectorTitle  string `json:"selector_title"`
	SelectorBody   string `json:"selector_body"`
	SourceLanguage string `json:"source_language"`
	IsActive       bool   `json:"is_active"`
}

type CreateSiteCategoryRequest struct {
	SiteID         string `json:"-" validate:"required,uuid"`
	Name           string `json:"name" validate:"required,max=255"`
	StartURL       string `json:"start_url" validate:"required,url,max=512"`
	URLPattern     string `json:"url_pattern" validate:"max=512"`
	SelectorTitle  string `json:"selector_title" validate:"max=255"`
	SelectorBody   string `json:"selector_body" validate:"required,max=255"`
	SourceLanguage string `json:"source_language" validate:"max=8"`
	IsActive       *bool  `json:"is_active"`
}

type UpdateSiteCategoryRequest struct {
	ID             string `json:"-" validate:"required,uuid"`
	Name           string `json:"name" validate:"required,max=255"`
	StartURL       string `json:"start_url" validate:"required,url,max=512"`
	URLPattern     string `json:"url_pattern" validate:"max=512"`
	SelectorTitle  string `json:"selector_title" validate:"max=255"`
	SelectorBody   string `json:"selector_body" validate:"required,max=255"`
	SourceLanguage string `json:"source_language" validate:"max=8"`
	IsActive       *bool  `json:"is_active"`
}

type EnqueueURLRequest struct {
	SiteID         string `json:"-" validate:"required,uuid"`
	URL            string `json:"url" validate:"required,url,max=1024"`
	SiteCategoryID string `json:"site_category_id" validate:"omitempty,uuid"`
}
