package jobs

import (
	"context"
	"fmt"
	"log/slog"
	"sync"
)

type Handler func(context.Context) error

type Dispatcher struct {
	logger   *slog.Logger
	mu       sync.RWMutex
	handlers map[string]Handler
}

func NewDispatcher(logger *slog.Logger) *Dispatcher {
	return &Dispatcher{
		logger:   logger,
		handlers: make(map[string]Handler),
	}
}

func (d *Dispatcher) Register(name string, handler Handler) {
	d.mu.Lock()
	defer d.mu.Unlock()
	d.handlers[name] = handler
}

func (d *Dispatcher) Dispatch(ctx context.Context, name string) error {
	d.mu.RLock()
	handler, ok := d.handlers[name]
	d.mu.RUnlock()
	if !ok {
		return fmt.Errorf("job handler %q not registered", name)
	}

	d.logger.Info("dispatch job", "name", name)
	return handler(ctx)
}
