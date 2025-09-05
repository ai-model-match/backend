package mm_scheduler

import (
	"context"

	"github.com/go-co-op/gocron/v2"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

/*
Scheduler represents a service that handle cron activities. It wraps the go-cron library
*/
type Scheduler struct {
	scheduler *gocron.Scheduler
}

/*
NewScheduler initialies a new scheduler service.
*/
func NewScheduler() *Scheduler {
	zap.L().Info("Start creating new Scheduler service...", zap.String("service", "scheduler"))
	scheduler, _ := gocron.NewScheduler()
	zap.L().Info("Scheduler service created!", zap.String("service", "scheduler"))
	return &Scheduler{
		scheduler: &scheduler,
	}

}

/*
Start the Scheduler service. Similar to a CRON system in Linux
*/
func (s Scheduler) Init() error {
	instance := *s.scheduler
	instance.Start()
	zap.L().Info("Scheduler service started!", zap.String("service", "scheduler"))
	return nil
}

/*
Close and stop the Scheduler service. Executes during the application shutdown
*/
func (s Scheduler) Close() error {
	zap.L().Info("Closing Scheduler service...", zap.String("service", "scheduler"))
	instance := *s.scheduler
	if err := instance.Shutdown(); err != nil {
		return err
	}
	zap.L().Info("Scheduler service closed!", zap.String("service", "scheduler"))
	return nil
}

/*
Add in the queue a new job that contains the schedule configuration (CRON format like * * * * *),
the function to call and parameters to pass during execution
*/
func (s Scheduler) AddJob(job ScheduledJob) error {
	instance := *s.scheduler
	if j, err := instance.NewJob(
		gocron.CronJob(job.Schedule, false),
		gocron.NewTask(job.Handler, job.Parameters),
	); err != nil {
		zap.L().Error("Failed to schedule a new Job", zap.Error(err), zap.String("service", "scheduler"))
		return err
	} else {
		zap.L().Info("New Job scheduled!", zap.String("Job ID", j.ID().String()), zap.String("service", "scheduler"))
	}
	return nil
}

/*
Given a generic DB Connection, get a specific low-level connection to DB.
It is used for low-level locks for scheduled activities
*/
func (s Scheduler) GetSingleConnection(tx *gorm.DB) *SingleConnection {
	sqlDB, _ := tx.DB()
	ctx := context.Background()
	conn, _ := sqlDB.Conn(ctx)
	return &SingleConnection{
		ctx:  ctx,
		conn: conn,
	}
}

/*
For distributed systems, it ensure scheduled tasks are executed once.
One instance of the application will aquire a exclusive DB Lock preventing other instances to run the same job.
Scheduled tasks need to be idenpotent, so multiple executions, even in parallel, should not break.
This approach is useful to avoid DB overload.
*/
func (s Scheduler) AcquireLock(sc *SingleConnection, jobID int64) bool {
	// Try to release the lock based on the unique JobID. Unless this app instance acquired the lock before, it will not succeed
	// So, we can ignore it because the lock could not exist (first execution) or it is acquired by another DB session (generally another app instance)
	var released bool
	var connectionID int
	if err := sc.conn.QueryRowContext(sc.ctx, "SELECT pg_backend_pid()").Scan(&connectionID); err != nil {
		zap.L().Error("Error in getting PID. Connection already in use. Check your configuration", zap.Int64("jobID", jobID), zap.Error(err), zap.String("service", "scheduler"))
		return false
	}
	if err := sc.conn.QueryRowContext(sc.ctx, "SELECT pg_advisory_unlock($1);", jobID).Scan(&released); err != nil {
		zap.L().Error("Lock release failed!", zap.Int("Connection ID", connectionID), zap.Int64("jobID", jobID), zap.Error(err), zap.String("service", "scheduler"))
		return false
	} else if !released {
		zap.L().Debug("Lock not released... ignore it!", zap.Int("Connection ID", connectionID), zap.Int64("jobID", jobID), zap.String("service", "scheduler"))
	} else {
		zap.L().Debug("Lock released!", zap.Int("Connection ID", connectionID), zap.Int64("jobID", jobID), zap.String("service", "scheduler"))
	}
	// Now, try to acquire the lock with the same JobID. If fails, it means that someone else already taken the lock, so there is
	// another app instance that will execute this specific job
	var acquired bool
	if err := sc.conn.QueryRowContext(sc.ctx, "SELECT pg_try_advisory_lock($1);", jobID).Scan(&acquired); err != nil {
		zap.L().Error("Lock acquisition failed!", zap.Int("Connection ID", connectionID), zap.Int64("jobID", jobID), zap.Error(err), zap.String("service", "scheduler"))
		return false
	} else if !acquired {
		zap.L().Debug("Lock not acquired... another service is managing it!", zap.Int("Connection ID", connectionID), zap.Int64("jobID", jobID), zap.String("service", "scheduler"))
		return false
	} else {
		zap.L().Debug("Lock acquired!", zap.Int("Connection ID", connectionID), zap.Int64("jobID", jobID), zap.String("service", "scheduler"))
		return true
	}
}
