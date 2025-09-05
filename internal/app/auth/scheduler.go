package auth

import (
	"github.com/ai-model-match/backend/internal/pkg/mm_log"
	"github.com/ai-model-match/backend/internal/pkg/mm_scheduler"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

type authSchedulerInterface interface {
	init()
}

type authScheduler struct {
	scheduler        *mm_scheduler.Scheduler
	storage          *gorm.DB
	singleConnection *mm_scheduler.SingleConnection
	repository       authRepositoryInterface
}

func newAuthScheduler(storage *gorm.DB, scheduler *mm_scheduler.Scheduler, repository authRepositoryInterface) authScheduler {
	singleConnection := scheduler.GetSingleConnection(storage)
	return authScheduler{
		scheduler:        scheduler,
		storage:          storage,
		singleConnection: singleConnection,
		repository:       repository,
	}
}

func (s authScheduler) init() {
	// Declare all jobs to be scheduled
	var jobsToSchedule []mm_scheduler.ScheduledJob = []mm_scheduler.ScheduledJob{
		{
			Schedule: "10 * * * *", // Every hour at HH:10
			Handler:  s.cleanUpExpiredRefreshToken,
			Parameters: mm_scheduler.ScheduledJobParameter{
				JobID: 29347129,
				Title: "CleanUpExpiredRefreshToken",
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
func (s authScheduler) cleanUpExpiredRefreshToken(p mm_scheduler.ScheduledJobParameter) error {
	defer func() {
		if r := recover(); r != nil {
			mm_log.LogPanicError(r, "CleanUpExpiredRefreshToken", "Panic occurred in cron activity")
		}
	}()
	// If this istance acquires the lock, executre the business logic
	if lockAcquired := s.scheduler.AcquireLock(s.singleConnection, p.JobID); lockAcquired {
		zap.L().Info("Starting Cron Job...", zap.String("job", p.Title))
		if err := s.repository.cleanUpExpiredRefreshToken(s.storage); err != nil {
			zap.L().Error("Cron Job Failed", zap.String("job", p.Title), zap.Error(err))
			return err
		}
		zap.L().Info("Cron Job executed!", zap.String("job", p.Title))
	}
	return nil
}
