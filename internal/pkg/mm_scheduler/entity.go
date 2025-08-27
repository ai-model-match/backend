package mm_scheduler

/*
Represents the parameters to pass to the scheduler handler function
*/
type ScheduledJobParameter struct {
	Title string
	JobID int64 // Unique numeric identifier to acquire lock
}

/*
Represents a scheduled function based on the Cron configuration
*/
type ScheduledJob struct {
	Schedule   string // crontab format (* * * * *)
	Handler    any
	Parameters ScheduledJobParameter
}
