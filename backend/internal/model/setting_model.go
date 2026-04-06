package model

// TranslationSettingsResponse represents the translation config (API key is
// redacted — the UI gets a presence flag and masked value).
type TranslationSettingsResponse struct {
	DeepLAPIKeySet         bool   `json:"deepl_api_key_set"`
	DeepLEndpoint          string `json:"deepl_endpoint"`
	DefaultTargetLanguage  string `json:"default_target_language"`
	RelationsSource        string `json:"relations_source"`
}

type UpdateTranslationSettingsRequest struct {
	DeepLAPIKey            *string `json:"deepl_api_key" validate:"omitempty,max=256"`
	DeepLEndpoint          *string `json:"deepl_endpoint" validate:"omitempty,url,max=512"`
	DefaultTargetLanguage  *string `json:"default_target_language" validate:"omitempty,min=2,max=8"`
	RelationsSource        *string `json:"relations_source" validate:"omitempty,max=32"`
}
