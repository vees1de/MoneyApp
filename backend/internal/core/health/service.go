package health

import "context"

type CheckFunc func(context.Context) error

type Service struct {
	checks map[string]CheckFunc
}

func NewService(checks map[string]CheckFunc) *Service {
	return &Service{checks: checks}
}

func (s *Service) Checks() map[string]CheckFunc {
	return s.checks
}
