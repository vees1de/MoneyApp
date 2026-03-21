package jobs

import "time"

type JobType string

const (
	JobTypeWeeklyReview   JobType = "weekly_review"
	JobTypeRebuildSummary JobType = "rebuild_summary"
	JobTypeSavingsAlert   JobType = "savings_alert"
)

type Job struct {
	Name      string
	Type      JobType
	RunAt     time.Time
	Payload   map[string]any
	CreatedAt time.Time
}
