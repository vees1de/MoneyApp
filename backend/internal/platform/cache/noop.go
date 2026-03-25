package cache

import (
	"context"
	"time"
)

type NoopStore struct{}

func NewNoopStore() *NoopStore {
	return &NoopStore{}
}

func (s *NoopStore) GetJSON(context.Context, string, any) (bool, error) {
	return false, nil
}

func (s *NoopStore) SetJSON(context.Context, string, any, time.Duration) error {
	return nil
}

func (s *NoopStore) Delete(context.Context, ...string) error {
	return nil
}

func (s *NoopStore) Ping(context.Context) error {
	return nil
}

func (s *NoopStore) Close() error {
	return nil
}
