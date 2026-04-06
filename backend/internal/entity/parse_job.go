package entity

import "time"

const (
	ParseJobStatusPending = "pending"
	ParseJobStatusRunning = "running"
	ParseJobStatusDone    = "done"
	ParseJobStatusFailed  = "failed"
	ParseJobStatusSkipped = "skipped"
)

// ParseJob is a queued URL to fetch and parse.
type ParseJob struct {
	ID             int64      `gorm:"column:id;primaryKey;autoIncrement"`
	GUID       string     `gorm:"column:guid;type:uuid;uniqueIndex;not null"`
	SiteID         int64      `gorm:"column:site_id;index"`
	SiteCategoryID *int64     `gorm:"column:site_category_id"`
	URL            string     `gorm:"column:url;size:1024"`
	Status         string     `gorm:"column:status;size:32;index;default:pending"`
	ScheduledAt    time.Time  `gorm:"column:scheduled_at;index"`
	StartedAt      *time.Time `gorm:"column:started_at"`
	FinishedAt     *time.Time `gorm:"column:finished_at"`
	Depth          int        `gorm:"column:depth;default:0"`
	AttemptCount   int        `gorm:"column:attempt_count;default:0"`
	LastError      string     `gorm:"column:last_error;size:1024"`
	CreatedAt      time.Time  `gorm:"column:created_at"`
	UpdatedAt      time.Time  `gorm:"column:updated_at"`
}

func (p *ParseJob) TableName() string { return "parse_jobs" }
