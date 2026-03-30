package app

import (
	"context"
	"errors"
	"time"

	"moneyapp/backend/internal/config"
)

type Worker struct {
	container *Container
}

func NewWorker(cfg *config.Config) (*Worker, error) {
	container, err := NewContainer(cfg)
	if err != nil {
		return nil, err
	}
	return &Worker{container: container}, nil
}

func (w *Worker) Run(ctx context.Context) error {
	defer func() {
		_ = w.container.DB.Close()
	}()

	workerID := "worker-" + time.Now().UTC().Format("20060102150405")
	err := w.container.Queue.Run(ctx, workerID, 2*time.Second)
	if errors.Is(err, context.Canceled) {
		return nil
	}
	return err
}
