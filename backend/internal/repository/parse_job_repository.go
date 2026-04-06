package repository

import (
	"time"

	"words-app/internal/entity"

	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

type ParseJobRepository struct {
	Log *logrus.Logger
}

func NewParseJobRepository(log *logrus.Logger) *ParseJobRepository {
	return &ParseJobRepository{Log: log}
}

func (r *ParseJobRepository) Create(db *gorm.DB, j *entity.ParseJob) error {
	return db.Create(j).Error
}

func (r *ParseJobRepository) Update(db *gorm.DB, j *entity.ParseJob) error {
	return db.Save(j).Error
}

// ClaimNextPending picks the oldest pending job whose scheduled_at <= now and marks it running.
// Returns gorm.ErrRecordNotFound if nothing to do.
func (r *ParseJobRepository) ClaimNextPending(db *gorm.DB, j *entity.ParseJob) error {
	err := db.Where("status = ? AND scheduled_at <= ?", entity.ParseJobStatusPending, time.Now()).
		Order("scheduled_at ASC").First(j).Error
	if err != nil {
		return err
	}
	now := time.Now()
	j.Status = entity.ParseJobStatusRunning
	j.StartedAt = &now
	j.AttemptCount++
	return db.Save(j).Error
}

func (r *ParseJobRepository) FindRecent(db *gorm.DB, out *[]entity.ParseJob, limit int) error {
	return db.Order("created_at DESC").Limit(limit).Find(out).Error
}

func (r *ParseJobRepository) CountHitsToday(db *gorm.DB, siteID int64) (int64, error) {
	since := time.Now().Add(-24 * time.Hour)
	var n int64
	err := db.Model(&entity.ParseJob{}).
		Where("site_id = ? AND finished_at IS NOT NULL AND finished_at >= ?", siteID, since).
		Count(&n).Error
	return n, err
}
