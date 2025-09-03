package mm_pubsub

import (
	"github.com/ai-model-match/backend/internal/pkg/mm_scheduler"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

type pubsubScheduler struct {
	scheduler            *mm_scheduler.Scheduler
	storage              *gorm.DB
	persistRetentionDays int
}

func newPubsubScheduler(storage *gorm.DB, scheduler *mm_scheduler.Scheduler, persistRetentionDays int) pubsubScheduler {
	return pubsubScheduler{
		scheduler:            scheduler,
		storage:              storage,
		persistRetentionDays: persistRetentionDays,
	}
}

func (s pubsubScheduler) init() {
	// Declare all jobs to be scheduled
	var jobsToSchedule []mm_scheduler.ScheduledJob = []mm_scheduler.ScheduledJob{
		{
			Schedule: "0 * * * *", // Every hour at HH:00
			Handler:  s.cleanUpOldPubSubEvents,
			Parameters: mm_scheduler.ScheduledJobParameter{
				JobID: 83701937,
				Title: "CleanUpOldPubSubEvents",
			},
		},
	}
	// Schedule all jobs
	for _, jobToSchedule := range jobsToSchedule {
		s.scheduler.AddJob(mm_scheduler.ScheduledJob{
			Schedule:   jobToSchedule.Schedule,
			Handler:    jobToSchedule.Handler,
			Parameters: jobToSchedule.Parameters,
		})
	}

}

/*
Scheduled function to run. It cleanup expired refresh tokens
*/
func (s pubsubScheduler) cleanUpOldPubSubEvents(p mm_scheduler.ScheduledJobParameter) error {
	retentionInDays := s.persistRetentionDays
	// If this istance acquires the lock, executre the business logic
	if lockAcquired := s.scheduler.AcquireLock(s.storage, p.JobID); lockAcquired {
		zap.L().Info("Starting Cron Job...", zap.String("job", p.Title))
		// Delete all old events based on retention policy
		if err := s.storage.Where("event_date < NOW() - (? * INTERVAL '1 day')", retentionInDays).Delete(&eventModel{}).Error; err != nil {
			zap.L().Error("Cron Job Failed", zap.String("job", p.Title), zap.Error(err))
			return err
		}
		zap.L().Info("Cron Job executed!", zap.String("job", p.Title))
	}
	return nil
}
