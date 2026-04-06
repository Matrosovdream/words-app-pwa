package translation

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/sirupsen/logrus"
)

// Translator is the shared interface for any translation provider.
type Translator interface {
	Translate(ctx context.Context, text, sourceLang, targetLang string) (string, error)
	Name() string
}

// DeepLClient talks to the DeepL free/pro API.
// Free tier:  https://api-free.deepl.com/v2/translate
// Pro tier:   https://api.deepl.com/v2/translate
type DeepLClient struct {
	Log      *logrus.Logger
	APIKey   string
	Endpoint string
	HTTP     *http.Client
}

func NewDeepLClient(log *logrus.Logger, apiKey, endpoint string) *DeepLClient {
	if endpoint == "" {
		endpoint = "https://api-free.deepl.com/v2/translate"
	}
	return &DeepLClient{
		Log:      log,
		APIKey:   apiKey,
		Endpoint: endpoint,
		HTTP:     &http.Client{Timeout: 20 * time.Second},
	}
}

func (c *DeepLClient) Name() string { return "deepl" }

type deeplResponse struct {
	Translations []struct {
		DetectedSourceLanguage string `json:"detected_source_language"`
		Text                   string `json:"text"`
	} `json:"translations"`
}

func (c *DeepLClient) Translate(ctx context.Context, text, sourceLang, targetLang string) (string, error) {
	if c.APIKey == "" {
		return "", errors.New("deepl: api key not configured")
	}
	if text == "" {
		return "", nil
	}

	form := url.Values{}
	form.Set("text", text)
	form.Set("target_lang", strings.ToUpper(targetLang))
	if sourceLang != "" {
		form.Set("source_lang", strings.ToUpper(sourceLang))
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.Endpoint, strings.NewReader(form.Encode()))
	if err != nil {
		return "", err
	}
	req.Header.Set("Authorization", "DeepL-Auth-Key "+c.APIKey)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := c.HTTP.Do(req)
	if err != nil {
		return "", fmt.Errorf("deepl request failed: %w", err)
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return "", fmt.Errorf("deepl: status %d: %s", resp.StatusCode, string(body))
	}
	var out deeplResponse
	if err := json.Unmarshal(body, &out); err != nil {
		return "", fmt.Errorf("deepl: bad json: %w", err)
	}
	if len(out.Translations) == 0 {
		return "", errors.New("deepl: no translations returned")
	}
	return out.Translations[0].Text, nil
}
