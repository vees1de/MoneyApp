package events

import (
	"context"
	"encoding/json"
	"time"

	"moneyapp/backend/internal/config"

	"github.com/segmentio/kafka-go"
)

type KafkaPublisher struct {
	brokers []string
	writer  *kafka.Writer
}

func NewKafkaPublisher(ctx context.Context, cfg config.KafkaConfig) (*KafkaPublisher, error) {
	writer := &kafka.Writer{
		Addr:                   kafka.TCP(cfg.Brokers...),
		AllowAutoTopicCreation: true,
		Balancer:               &kafka.LeastBytes{},
		RequiredAcks:           kafka.RequireOne,
		BatchTimeout:           50 * time.Millisecond,
		WriteTimeout:           cfg.WriteTimeout,
		ReadTimeout:            cfg.WriteTimeout,
	}

	publisher := &KafkaPublisher{
		brokers: cfg.Brokers,
		writer:  writer,
	}

	pingCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	if err := publisher.Ping(pingCtx); err != nil {
		_ = publisher.Close()
		return nil, err
	}

	return publisher, nil
}

func (p *KafkaPublisher) PublishJSON(ctx context.Context, topic, key string, payload any) error {
	value, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	return p.writer.WriteMessages(ctx, kafka.Message{
		Topic: topic,
		Key:   []byte(key),
		Value: value,
		Time:  time.Now().UTC(),
	})
}

func (p *KafkaPublisher) Ping(ctx context.Context) error {
	conn, err := kafka.DialContext(ctx, "tcp", p.brokers[0])
	if err != nil {
		return err
	}
	defer conn.Close()

	_, err = conn.ReadPartitions()
	return err
}

func (p *KafkaPublisher) Close() error {
	return p.writer.Close()
}
