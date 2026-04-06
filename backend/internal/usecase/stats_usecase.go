package usecase

import (
	"context"

	"words-app/internal/entity"
	"words-app/internal/model"
	"words-app/internal/repository"

	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

type StatsUseCase struct {
	DB                  *gorm.DB
	Log                 *logrus.Logger
	SiteRepository      *repository.SiteRepository
	JobRepository       *repository.ParseJobRepository
	WorkerEnabled       bool
	PollIntervalSeconds int
	CommonWordThreshold int
}

func NewStatsUseCase(db *gorm.DB, log *logrus.Logger,
	siteRepo *repository.SiteRepository, jobRepo *repository.ParseJobRepository,
	workerEnabled bool, pollSeconds, threshold int) *StatsUseCase {
	return &StatsUseCase{
		DB: db, Log: log,
		SiteRepository: siteRepo, JobRepository: jobRepo,
		WorkerEnabled: workerEnabled, PollIntervalSeconds: pollSeconds,
		CommonWordThreshold: threshold,
	}
}

func (c *StatsUseCase) ParserStats(ctx context.Context) (*model.ParserStatsResponse, error) {
	db := c.DB.WithContext(ctx)

	// job status counts
	type statusRow struct {
		Status string
		N      int64
	}
	var rows []statusRow
	_ = db.Model(&entity.ParseJob{}).
		Select("status, COUNT(*) AS n").
		Group("status").Scan(&rows).Error
	jobs := model.JobStatusCounts{}
	for _, r := range rows {
		switch r.Status {
		case entity.ParseJobStatusPending:
			jobs.Pending = r.N
		case entity.ParseJobStatusRunning:
			jobs.Running = r.N
		case entity.ParseJobStatusDone:
			jobs.Done = r.N
		case entity.ParseJobStatusFailed:
			jobs.Failed = r.N
		case entity.ParseJobStatusSkipped:
			jobs.Skipped = r.N
		}
	}

	// recent jobs
	var recent []entity.ParseJob
	_ = c.JobRepository.FindRecent(db, &recent, 10)

	// index site names
	var sites []entity.Site
	_ = c.SiteRepository.FindAll(db, &sites)
	siteNames := make(map[string]string, len(sites))
	for _, s := range sites {
		siteNames[s.ID] = s.Name
	}

	recentDTOs := make([]model.RecentJobDTO, 0, len(recent))
	for _, j := range recent {
		dto := model.RecentJobDTO{
			ID:          j.ID,
			SiteName:    siteNames[j.SiteID],
			URL:         j.URL,
			Status:      j.Status,
			ScheduledAt: j.ScheduledAt.Unix(),
			LastError:   j.LastError,
		}
		if j.StartedAt != nil {
			t := j.StartedAt.Unix()
			dto.StartedAt = &t
		}
		if j.FinishedAt != nil {
			t := j.FinishedAt.Unix()
			dto.FinishedAt = &t
		}
		recentDTOs = append(recentDTOs, dto)
	}

	// per-site hits today
	siteHits := make([]model.SiteHitDTO, 0, len(sites))
	for _, s := range sites {
		n, _ := c.JobRepository.CountHitsToday(db, s.ID)
		siteHits = append(siteHits, model.SiteHitDTO{
			ID:            s.ID,
			Name:          s.Name,
			IsActive:      s.IsActive,
			HitsToday:     n,
			DailyHitLimit: s.DailyHitLimit,
		})
	}

	// counts
	counts := model.CountsDTO{}
	db.Model(&entity.DictWord{}).Count(&counts.DictWords)
	db.Model(&entity.ReviewItem{}).Where("status = ?", entity.ReviewStatusPending).Count(&counts.PendingReview)
	db.Model(&entity.ReviewItem{}).Where("status = ?", entity.ReviewStatusDenied).Count(&counts.DeniedReview)
	db.Model(&entity.LearnItem{}).Where("status = ?", entity.LearnStatusActive).Count(&counts.ActiveLearn)
	db.Model(&entity.LearnItem{}).Where("status = ?", entity.LearnStatusArchived).Count(&counts.ArchivedLearn)
	db.Model(&entity.ParsedPage{}).Count(&counts.ParsedPages)

	return &model.ParserStatsResponse{
		Worker: model.WorkerConfigDTO{
			Enabled:             c.WorkerEnabled,
			PollIntervalSeconds: c.PollIntervalSeconds,
			CommonWordThreshold: c.CommonWordThreshold,
		},
		Jobs:       jobs,
		RecentJobs: recentDTOs,
		Sites:      siteHits,
		Counts:     counts,
	}, nil
}
