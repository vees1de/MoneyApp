package jobs

import (
	"context"
	"log/slog"
	"sync"
	"time"
)

type ScheduledJob struct {
	Name     string
	Interval time.Duration
}

type Scheduler struct {
	logger     *slog.Logger
	dispatcher *Dispatcher
	jobs       []ScheduledJob
	cancel     context.CancelFunc
	wg         sync.WaitGroup
}

func NewScheduler(logger *slog.Logger, dispatcher *Dispatcher) *Scheduler {
	return &Scheduler{
		logger:     logger,
		dispatcher: dispatcher,
	}
}

func (s *Scheduler) Add(job ScheduledJob) {
	s.jobs = append(s.jobs, job)
}

func (s *Scheduler) Start(ctx context.Context) {
	runCtx, cancel := context.WithCancel(ctx)
	s.cancel = cancel

	for _, job := range s.jobs {
		s.wg.Add(1)
		go func(job ScheduledJob) {
			defer s.wg.Done()
			ticker := time.NewTicker(job.Interval)
			defer ticker.Stop()

			for {
				select {
				case <-runCtx.Done():
					return
				case <-ticker.C:
					if err := s.dispatcher.Dispatch(runCtx, job.Name); err != nil {
						s.logger.Error("run scheduled job", "job", job.Name, "error", err)
					}
				}
			}
		}(job)
	}
}

func (s *Scheduler) Stop() {
	if s.cancel != nil {
		s.cancel()
	}
	s.wg.Wait()
}
