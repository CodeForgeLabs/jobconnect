package events

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/segmentio/kafka-go"
)

type Envelope struct {
	EventID        string          `json:"event_id"`
	EventType      string          `json:"event_type"`
	AggregateID    string          `json:"aggregate_id"`
	OccurredAt     time.Time       `json:"occurred_at"`
	Producer       string          `json:"producer"`
	Version        int             `json:"version"`
	Payload        json.RawMessage `json:"payload"`
	IdempotencyKey string          `json:"idempotency_key,omitempty"`
	CorrelationID  string          `json:"correlation_id,omitempty"`
}

func NewEnvelope(eventType, aggregateID, producer string, version int, payload any, idempotencyKey, correlationID string) (Envelope, error) {
	body, err := json.Marshal(payload)
	if err != nil {
		return Envelope{}, fmt.Errorf("marshal payload: %w", err)
	}
	if strings.TrimSpace(aggregateID) == "" {
		aggregateID = uuid.NewString()
	}
	if version <= 0 {
		version = 1
	}
	return Envelope{
		EventID:        uuid.NewString(),
		EventType:      eventType,
		AggregateID:    aggregateID,
		OccurredAt:     time.Now().UTC(),
		Producer:       producer,
		Version:        version,
		Payload:        body,
		IdempotencyKey: strings.TrimSpace(idempotencyKey),
		CorrelationID:  strings.TrimSpace(correlationID),
	}, nil
}

type Publisher struct {
	writer *kafka.Writer
}

func NewPublisher(brokers []string, topic string) *Publisher {
	return &Publisher{writer: &kafka.Writer{
		Addr:                   kafka.TCP(brokers...),
		Topic:                  topic,
		Balancer:               &kafka.Hash{},
		RequiredAcks:           kafka.RequireAll,
		AllowAutoTopicCreation: true,
	}}
}

func (p *Publisher) Publish(ctx context.Context, env Envelope) error {
	if p == nil || p.writer == nil {
		return fmt.Errorf("publisher is nil")
	}
	b, err := json.Marshal(env)
	if err != nil {
		return err
	}
	return p.writer.WriteMessages(ctx, kafka.Message{Key: []byte(env.AggregateID), Value: b})
}

func (p *Publisher) Close() error {
	if p == nil || p.writer == nil {
		return nil
	}
	return p.writer.Close()
}

type HandlerFunc func(ctx context.Context, env Envelope) error

type Consumer struct {
	reader   *kafka.Reader
	handlers map[string]HandlerFunc
}

func NewConsumer(brokers []string, topic, groupID string) *Consumer {
	return &Consumer{
		reader: kafka.NewReader(kafka.ReaderConfig{
			Brokers: brokers,
			Topic:   topic,
			GroupID: groupID,
		}),
		handlers: map[string]HandlerFunc{},
	}
}

func (c *Consumer) On(eventType string, handler HandlerFunc) {
	c.handlers[eventType] = handler
}

func (c *Consumer) Run(ctx context.Context) error {
	for {
		msg, err := c.reader.FetchMessage(ctx)
		if err != nil {
			return err
		}
		var env Envelope
		if err := json.Unmarshal(msg.Value, &env); err == nil {
			if h, ok := c.handlers[env.EventType]; ok {
				if _ = h(ctx, env); true {
				}
			}
		}
		if err := c.reader.CommitMessages(ctx, msg); err != nil {
			return err
		}
	}
}

func (c *Consumer) Close() error {
	if c == nil || c.reader == nil {
		return nil
	}
	return c.reader.Close()
}

func ParseBrokers(raw string) []string {
	parts := strings.Split(raw, ",")
	brokers := make([]string, 0, len(parts))
	for _, p := range parts {
		if v := strings.TrimSpace(p); v != "" {
			brokers = append(brokers, v)
		}
	}
	if len(brokers) == 0 {
		return []string{"localhost:9092"}
	}
	return brokers
}
