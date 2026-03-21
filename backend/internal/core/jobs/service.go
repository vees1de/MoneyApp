package jobs

import (
	"time"

	platformjobs "moneyapp/backend/internal/platform/jobs"
)

type Service struct {
	scheduler *platformjobs.Scheduler
}

func NewService(scheduler *platformjobs.Scheduler) *Service {
	return &Service{scheduler: scheduler}
}

func (s *Service) Register(name string, interval time.Duration) {
	s.scheduler.Add(platformjobs.ScheduledJob{
		Name:     name,
		Interval: interval,
	})
}
