package dictionary

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"

	"github.com/sirupsen/logrus"
)

// RelationsClient looks up synonyms/antonyms for English words.
type RelationsClient interface {
	Synonyms(ctx context.Context, word string) ([]string, error)
	Antonyms(ctx context.Context, word string) ([]string, error)
	Name() string
}

// DatamuseClient uses https://www.datamuse.com/api/ (free, no key, English only).
//   rel_syn=foo  -> synonyms of foo
//   rel_ant=foo  -> antonyms of foo
type DatamuseClient struct {
	Log  *logrus.Logger
	HTTP *http.Client
	Base string
}

func NewDatamuseClient(log *logrus.Logger) *DatamuseClient {
	return &DatamuseClient{
		Log:  log,
		HTTP: &http.Client{Timeout: 10 * time.Second},
		Base: "https://api.datamuse.com/words",
	}
}

func (c *DatamuseClient) Name() string { return "datamuse" }

type datamuseItem struct {
	Word string `json:"word"`
}

func (c *DatamuseClient) fetch(ctx context.Context, param, word string) ([]string, error) {
	q := url.Values{}
	q.Set(param, word)
	q.Set("max", "15")

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.Base+"?"+q.Encode(), nil)
	if err != nil {
		return nil, err
	}
	resp, err := c.HTTP.Do(req)
	if err != nil {
		return nil, fmt.Errorf("datamuse request failed: %w", err)
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("datamuse: status %d", resp.StatusCode)
	}
	var items []datamuseItem
	if err := json.Unmarshal(body, &items); err != nil {
		return nil, fmt.Errorf("datamuse: bad json: %w", err)
	}
	out := make([]string, 0, len(items))
	for _, it := range items {
		if it.Word != "" {
			out = append(out, it.Word)
		}
	}
	return out, nil
}

func (c *DatamuseClient) Synonyms(ctx context.Context, word string) ([]string, error) {
	return c.fetch(ctx, "rel_syn", word)
}

func (c *DatamuseClient) Antonyms(ctx context.Context, word string) ([]string, error) {
	return c.fetch(ctx, "rel_ant", word)
}
