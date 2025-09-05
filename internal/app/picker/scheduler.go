package picker

import (
	"github.com/ai-model-match/backend/internal/pkg/mm_log"
	"github.com/ai-model-match/backend/internal/pkg/mm_scheduler"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

type pickerSchedulerInterface interface {
	init()
}

type pickerScheduler struct {
	scheduler        *mm_scheduler.Scheduler
	storage          *gorm.DB
	singleConnection *mm_scheduler.SingleConnection
	repository       pickerRepositoryInterface
}

func newPickerScheduler(storage *gorm.DB, scheduler *mm_scheduler.Scheduler, repository pickerRepositoryInterface) pickerScheduler {
	singleConnection := scheduler.GetSingleConnection(storage)
	return pickerScheduler{
		scheduler:        scheduler,
		storage:          storage,
		singleConnection: singleConnection,
		repository:       repository,
	}
}

func (s pickerScheduler) init() {
	// Declare all jobs to be scheduled
	var jobsToSchedule []mm_scheduler.ScheduledJob = []mm_scheduler.ScheduledJob{
		{
			Schedule: "0,10,20,30,40,50 * * * *", // Every 10 minutes of the hour
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
	defer func() {
		if r := recover(); r != nil {
			mm_log.LogPanicError(r, "CleanUpExpiredPickerCorrelations", "Panic occurred in cron activity")
		}
	}()
	// If this istance acquires the lock, executre the business logic
	if lockAcquired := s.scheduler.AcquireLock(s.singleConnection, p.JobID); lockAcquired {
		zap.L().Info("Starting Cron Job...", zap.String("job", p.Title))
		if err := s.repository.cleanUpExpiredPickerCorrelations(s.storage); err != nil {
			zap.L().Error("Cron Job Failed", zap.String("job", p.Title), zap.Error(err))
			return err
		}
		zap.L().Info("Cron Job executed!", zap.String("job", p.Title))
	}
	return nil
}
