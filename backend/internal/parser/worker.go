package parser

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"io"
	"net/http"
	"strings"
	"time"

	"words-app/internal/entity"
	"words-app/internal/repository"

	"github.com/PuerkitoBio/goquery"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

// Worker picks ParseJobs off the queue, fetches pages, extracts non-common words
// and creates ReviewItems.
type Worker struct {
	DB                   *gorm.DB
	Log                  *logrus.Logger
	SiteRepository       *repository.SiteRepository
	CategoryRepository   *repository.SiteCategoryRepository
	JobRepository        *repository.ParseJobRepository
	PageRepository       *repository.ParsedPageRepository
	DictRepository       *repository.DictWordRepository
	OccurrenceRepository *repository.WordOccurrenceRepository
	ReviewRepository     *repository.ReviewItemRepository
	GeoRepository        *repository.GeoRepository

	PollInterval        time.Duration
	CommonWordThreshold int // words with frequency_rank <= threshold are considered common
	GeoNames            map[string]bool // in-memory set of geographic names to exclude
	HTTP                *http.Client
}

func NewWorker(
	db *gorm.DB, log *logrus.Logger,
	siteRepo *repository.SiteRepository,
	catRepo *repository.SiteCategoryRepository,
	jobRepo *repository.ParseJobRepository,
	pageRepo *repository.ParsedPageRepository,
	dictRepo *repository.DictWordRepository,
	occRepo *repository.WordOccurrenceRepository,
	reviewRepo *repository.ReviewItemRepository,
	geoRepo *repository.GeoRepository,
	pollInterval time.Duration, commonThreshold int,
) *Worker {
	geoNames := geoRepo.LoadAllGeoNames(db)
	log.Infof("Loaded %d geographic names for word filtering", len(geoNames))

	return &Worker{
		DB: db, Log: log,
		SiteRepository: siteRepo, CategoryRepository: catRepo,
		JobRepository: jobRepo, PageRepository: pageRepo,
		DictRepository: dictRepo, OccurrenceRepository: occRepo,
		ReviewRepository:    reviewRepo,
		GeoRepository:       geoRepo,
		PollInterval:        pollInterval,
		CommonWordThreshold: commonThreshold,
		GeoNames:            geoNames,
		HTTP:                &http.Client{Timeout: 20 * time.Second},
	}
}

// Start launches the worker loop in a goroutine. It returns immediately.
func (w *Worker) Start(ctx context.Context) {
	go func() {
		w.Log.Info("Parser worker started")
		ticker := time.NewTicker(w.PollInterval)
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				w.Log.Info("Parser worker stopped")
				return
			case <-ticker.C:
				w.tick(ctx)
			}
		}
	}()
}

func (w *Worker) tick(ctx context.Context) {
	job := new(entity.ParseJob)
	if err := w.JobRepository.ClaimNextPending(w.DB.WithContext(ctx), job); err != nil {
		if !errors.Is(err, gorm.ErrRecordNotFound) {
			w.Log.Warnf("Failed to claim job : %+v", err)
		}
		return
	}
	w.processJob(ctx, job)
}

func (w *Worker) processJob(ctx context.Context, job *entity.ParseJob) {
	now := time.Now()
	defer func() {
		if job.FinishedAt == nil {
			t := time.Now()
			job.FinishedAt = &t
			_ = w.JobRepository.Update(w.DB.WithContext(ctx), job)
		}
	}()

	site := new(entity.Site)
	if err := w.SiteRepository.FindByID(w.DB.WithContext(ctx), site, job.SiteID); err != nil {
		w.failJob(ctx, job, "site not found")
		return
	}
	if !site.IsActive {
		w.skipJob(ctx, job, "site inactive")
		return
	}

	// daily hit limit
	if site.DailyHitLimit > 0 {
		n, _ := w.JobRepository.CountHitsToday(w.DB.WithContext(ctx), site.ID)
		if n >= int64(site.DailyHitLimit) {
			// reschedule in 1h
			t := time.Now().Add(1 * time.Hour)
			job.Status = entity.ParseJobStatusPending
			job.ScheduledAt = t
			job.StartedAt = nil
			job.LastError = "daily hit limit reached"
			_ = w.JobRepository.Update(w.DB.WithContext(ctx), job)
			return
		}
	}

	// check history: skip if already parsed and reparse disabled
	urlHash := hashString(job.URL)
	existing := new(entity.ParsedPage)
	pageExisted := w.PageRepository.FindByHash(w.DB.WithContext(ctx), existing, site.ID, urlHash) == nil
	if pageExisted {
		if !site.ReparseEnabled {
			w.skipJob(ctx, job, "already parsed")
			return
		}
		if site.ReparseAfterDays > 0 {
			cutoff := time.Now().Add(-time.Duration(site.ReparseAfterDays) * 24 * time.Hour)
			if existing.ParsedAt.After(cutoff) {
				w.skipJob(ctx, job, "within reparse window")
				return
			}
		}
	}

	// fetch
	html, err := w.fetchURL(ctx, job.URL, site.UserAgent)
	if err != nil {
		w.Log.Warnf("Fetch failed for %s : %+v", job.URL, err)
		w.failJob(ctx, job, err.Error())
		return
	}

	// optional category config for selectors
	var cat *entity.SiteCategory
	if job.SiteCategoryID != nil {
		c := new(entity.SiteCategory)
		if err := w.CategoryRepository.FindByID(w.DB.WithContext(ctx), c, *job.SiteCategoryID); err == nil {
			cat = c
		}
	}

	title, body := extractContent(html, cat)
	contentHash := hashString(body)

	// respect per-site delay after fetch
	defer func() {
		if site.HitDelayMs > 0 {
			time.Sleep(time.Duration(site.HitDelayMs) * time.Millisecond)
		}
	}()

	// save parsed_page (create or update)
	var page *entity.ParsedPage
	if pageExisted {
		page = existing
		page.Title = title
		page.ContentHash = contentHash
		page.ParsedAt = now
	} else {
		page = &entity.ParsedPage{
			PublicID:       uuid.NewString(),
			SiteID:         site.ID,
			SiteCategoryID: job.SiteCategoryID,
			URL:            job.URL,
			URLHash:        urlHash,
			ContentHash:    contentHash,
			Title:          title,
			ParsedAt:       now,
		}
	}

	// tokenize body
	tokens := Tokenize(body)
	sentences := SplitSentences(body)
	page.WordCount = len(tokens)

	if pageExisted {
		if err := w.PageRepository.Update(w.DB.WithContext(ctx), page); err != nil {
			w.Log.Warnf("Failed to update parsed_page : %+v", err)
		}
	} else {
		if err := w.PageRepository.Create(w.DB.WithContext(ctx), page); err != nil {
			w.Log.Warnf("Failed to create parsed_page : %+v", err)
			w.failJob(ctx, job, "persist page")
			return
		}
	}

	// extract unique rare words + create review items
	w.extractRareWords(ctx, page, tokens, sentences, cat)

	// mark done
	t := time.Now()
	job.Status = entity.ParseJobStatusDone
	job.FinishedAt = &t
	job.LastError = ""
	_ = w.JobRepository.Update(w.DB.WithContext(ctx), job)
}

func (w *Worker) extractRareWords(ctx context.Context, page *entity.ParsedPage, tokens, sentences []string, cat *entity.SiteCategory) {
	counts := make(map[string]int)
	for _, t := range tokens {
		counts[t]++
	}

	language := "en"
	if cat != nil && cat.SourceLanguage != "" {
		language = cat.SourceLanguage
	}

	for lemma, count := range counts {
		if len(lemma) < 4 {
			continue
		}
		if strings.Contains(lemma, "'") {
			continue
		}
		if w.GeoNames[lemma] {
			continue
		}
		// check if word already in dict and whether it's common
		existing := new(entity.DictWord)
		err := w.DictRepository.FindByLemma(w.DB.WithContext(ctx), existing, lemma, language)
		if err == nil {
			if existing.FrequencyRank != nil && *existing.FrequencyRank <= w.CommonWordThreshold {
				continue // common word, skip
			}
		} else {
			// create new dict word (rare, not in frequency seed)
			existing = &entity.DictWord{
				PublicID:  uuid.NewString(),
				Lemma:    lemma,
				Language: language,
			}
			if err := w.DictRepository.Create(w.DB.WithContext(ctx), existing); err != nil {
				continue
			}
		}

		// create occurrence (one per page per word)
		sample := FindSampleSentence(sentences, lemma)
		_ = w.OccurrenceRepository.Create(w.DB.WithContext(ctx), &entity.WordOccurrence{
			PublicID:       uuid.NewString(),
			WordID:         existing.ID,
			ParsedPageID:   page.ID,
			Count:          count,
			SampleSentence: sample,
			CreatedAt:      time.Now(),
		})

		// create review item if not already present
		ri := new(entity.ReviewItem)
		if err := w.ReviewRepository.FindByWordID(w.DB.WithContext(ctx), ri, existing.ID); err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				pageID := page.ID
				_ = w.ReviewRepository.Create(w.DB.WithContext(ctx), &entity.ReviewItem{
					PublicID:        uuid.NewString(),
					WordID:          existing.ID,
					FirstSeenPageID: &pageID,
					Status:          entity.ReviewStatusPending,
					CreatedAt:       time.Now(),
					UpdatedAt:       time.Now(),
				})
			}
		}
	}
}

func (w *Worker) failJob(ctx context.Context, job *entity.ParseJob, msg string) {
	t := time.Now()
	job.Status = entity.ParseJobStatusFailed
	job.FinishedAt = &t
	job.LastError = msg
	_ = w.JobRepository.Update(w.DB.WithContext(ctx), job)
}

func (w *Worker) skipJob(ctx context.Context, job *entity.ParseJob, msg string) {
	t := time.Now()
	job.Status = entity.ParseJobStatusSkipped
	job.FinishedAt = &t
	job.LastError = msg
	_ = w.JobRepository.Update(w.DB.WithContext(ctx), job)
}

func (w *Worker) fetchURL(ctx context.Context, url, userAgent string) (string, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return "", err
	}
	if userAgent == "" {
		userAgent = "Mozilla/5.0 (compatible; WordsBot/0.1)"
	}
	req.Header.Set("User-Agent", userAgent)
	resp, err := w.HTTP.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return "", errors.New("http status " + resp.Status)
	}
	body, err := io.ReadAll(io.LimitReader(resp.Body, 5*1024*1024))
	if err != nil {
		return "", err
	}
	return string(body), nil
}

func extractContent(html string, cat *entity.SiteCategory) (title, body string) {
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(html))
	if err != nil {
		return "", ""
	}
	// title
	titleSel := "title"
	if cat != nil && cat.SelectorTitle != "" {
		titleSel = cat.SelectorTitle
	}
	title = strings.TrimSpace(doc.Find(titleSel).First().Text())

	// body
	bodySel := "article, main, body"
	if cat != nil && cat.SelectorBody != "" {
		bodySel = cat.SelectorBody
	}
	// strip scripts/styles
	doc.Find("script, style, nav, footer, header").Remove()
	body = strings.TrimSpace(doc.Find(bodySel).First().Text())
	if body == "" {
		body = strings.TrimSpace(doc.Find("body").First().Text())
	}
	return title, body
}

func hashString(s string) string {
	h := sha256.Sum256([]byte(s))
	return hex.EncodeToString(h[:])
}
