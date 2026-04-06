package model

type JobStatusCounts struct {
	Pending int64 `json:"pending"`
	Running int64 `json:"running"`
	Done    int64 `json:"done"`
	Failed  int64 `json:"failed"`
	Skipped int64 `json:"skipped"`
}

type RecentJobDTO struct {
	ID          string  `json:"id"`
	SiteName    string  `json:"site_name"`
	URL         string  `json:"url"`
	Status      string  `json:"status"`
	ScheduledAt int64   `json:"scheduled_at"`
	StartedAt   *int64  `json:"started_at,omitempty"`
	FinishedAt  *int64  `json:"finished_at,omitempty"`
	LastError   string  `json:"last_error,omitempty"`
}

type SiteHitDTO struct {
	ID            string `json:"id"`
	Name          string `json:"name"`
	IsActive      bool   `json:"is_active"`
	HitsToday     int64  `json:"hits_today"`
	DailyHitLimit int    `json:"daily_hit_limit"`
}

type WorkerConfigDTO struct {
	Enabled             bool `json:"enabled"`
	PollIntervalSeconds int  `json:"poll_interval_seconds"`
	CommonWordThreshold int  `json:"common_word_threshold"`
}

type CountsDTO struct {
	DictWords      int64 `json:"dict_words"`
	PendingReview  int64 `json:"pending_review"`
	DeniedReview   int64 `json:"denied_review"`
	ActiveLearn    int64 `json:"active_learn"`
	ArchivedLearn  int64 `json:"archived_learn"`
	ParsedPages    int64 `json:"parsed_pages"`
}

type ParserStatsResponse struct {
	Worker     WorkerConfigDTO `json:"worker"`
	Jobs       JobStatusCounts `json:"jobs"`
	RecentJobs []RecentJobDTO  `json:"recent_jobs"`
	Sites      []SiteHitDTO    `json:"sites"`
	Counts     CountsDTO       `json:"counts"`
}
