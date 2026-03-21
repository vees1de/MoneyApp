package events

import "context"

type Publisher interface {
	PublishJSON(ctx context.Context, topic, key string, payload any) error
	Ping(ctx context.Context) error
	Close() error
}
