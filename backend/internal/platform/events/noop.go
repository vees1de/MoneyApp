package events

import "context"

type NoopPublisher struct{}

func NewNoopPublisher() *NoopPublisher {
	return &NoopPublisher{}
}

func (p *NoopPublisher) PublishJSON(context.Context, string, string, any) error {
	return nil
}

func (p *NoopPublisher) Ping(context.Context) error {
	return nil
}

func (p *NoopPublisher) Close() error {
	return nil
}
