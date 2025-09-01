package picker

import (
	"github.com/ai-model-match/backend/internal/pkg/mm_scheduler"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

type pickerSchedulerInterface interface {
	init()
}

type pickerScheduler struct {
	scheduler  *mm_scheduler.Scheduler
	storage    *gorm.DB
	repository pickerRepositoryInterface
}

func newPickerScheduler(storage *gorm.DB, scheduler *mm_scheduler.Scheduler, repository pickerRepositoryInterface) pickerScheduler {
	return pickerScheduler{
		scheduler:  scheduler,
		storage:    storage,
		repository: repository,
	}
}

func (s pickerScheduler) init() {
	// Declare all jobs to be scheduled
	var jobsToSchedule []mm_scheduler.ScheduledJob = []mm_scheduler.ScheduledJob{
		{
			Schedule: "0 * * * *", // Every hour at HH:00
			Handler:  s.cleanUpExpiredPickerCorrelations,
			Parameters: mm_scheduler.ScheduledJobParameter{
				JobID: 14387371,
				Title: "CleanUpExpiredPickerCorrelations",
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
func (s pickerScheduler) cleanUpExpiredPickerCorrelations(p mm_scheduler.ScheduledJobParameter) error {
	// If this istance acquires the lock, executre the business logic
	if lockAcquired := s.scheduler.AcquireLock(s.storage, p.JobID); lockAcquired {
		zap.L().Info("Starting Cron Job...", zap.String("job", p.Title))
		if err := s.repository.cleanUpExpiredPickerCorrelations(s.storage); err != nil {
			zap.L().Error("Cron Job Failed", zap.String("job", p.Title), zap.Error(err))
			return err
		}
		zap.L().Info("Cron Job executed!", zap.String("job", p.Title))
	}
	return nil
}
