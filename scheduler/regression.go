package scheduler

var regressionSchedulerSingleton IScheduler

func Regression() IScheduler {
	if regressionSchedulerSingleton == nil {
		regressionSchedulerSingleton = &SimpleScheduler{
			Scheduler: Scheduler{
				SchedType:      SchedSimple,
				InstanceNumber: 0,
			},
			Settings: SimpleSchedSettings{
				ExecutionTime:   -1,
				Iterations:      1,
				ConcurrentUsers: 1,
				RampupDelay:     0.5,
			},
		}
	}
	return regressionSchedulerSingleton
}
