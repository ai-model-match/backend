package rsEngine

import (
	"github.com/ai-model-match/backend/internal/pkg/mm_log"
	"github.com/ai-model-match/backend/internal/pkg/mm_scheduler"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

type rsEngineSchedulerInterface interface {
	init()
}

type rsEngineScheduler struct {
	scheduler        *mm_scheduler.Scheduler
	storage          *gorm.DB
	singleConnection *mm_scheduler.SingleConnection
	service          rsEngineServiceInterface
}

func newRsEngineScheduler(storage *gorm.DB, scheduler *mm_scheduler.Scheduler, service rsEngineServiceInterface) rsEngineScheduler {
	singleConnection := scheduler.GetSingleConnection(storage)
	return rsEngineScheduler{
		scheduler:        scheduler,
		storage:          storage,
		singleConnection: singleConnection,
		service:          service,
	}
}

func (s rsEngineScheduler) init() {
	// Declare all jobs to be scheduled
	var jobsToSchedule []mm_scheduler.ScheduledJob = []mm_scheduler.ScheduledJob{
		{
			Schedule: "* * * * *", // Every minute
			Handler:  s.rsEngineTimeTick,
			Parameters: mm_scheduler.ScheduledJobParameter{
				JobID: 73919279,
				Title: "RsEngineTimeTick",
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
Scheduled function to run. It runs the onTimeTick of RS engine
*/
func (s rsEngineScheduler) rsEngineTimeTick(p mm_scheduler.ScheduledJobParameter) error {
	defer func() {
		if r := recover(); r != nil {
			mm_log.LogPanicError(r, "RsEngineTimeTick", "Panic occurred in cron activity")
		}
	}()
	// If this istance acquires the lock, executre the business logic
	if lockAcquired := s.scheduler.AcquireLock(s.singleConnection, p.JobID); lockAcquired {
		zap.L().Info("Starting Cron Job...", zap.String("job", p.Title))
		if err := s.service.onTimeTick(); err != nil {
			zap.L().Error("Cron Job Failed", zap.String("job", p.Title), zap.Error(err))
			return err
		}
		zap.L().Info("Cron Job executed!", zap.String("job", p.Title))
	}
	return nil
}
