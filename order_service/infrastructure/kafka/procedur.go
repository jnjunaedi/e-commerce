package kafka

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/twmb/franz-go/pkg/kgo"
)

type OrderCreatedEvent struct {
	ID               string `json:"id"`
	UserID           string `json:"user_id"`
	OrderID          string `json:"order_id"`
	NotificationType string `json:"notification_type"`
	Title            string `json:"title"`
	Email            string `json:"email"`
	Message          string `json:"message"`
	Status           string `json:"status"`
}

type OrderProducer interface {
	PublishOrderCreated(ctx context.Context, event OrderCreatedEvent) error
	Close()
}

type OrderProducerImpl struct {
	client *kgo.Client
}

func NewOrderProducer(brokers []string) (OrderProducer, error) {
	client, err := kgo.NewClient(
		kgo.SeedBrokers(brokers...),
	)
	if err != nil {
		return nil, err
	}

	return &OrderProducerImpl{client: client}, nil
}

func (p *OrderProducerImpl) PublishOrderCreated(ctx context.Context, event OrderCreatedEvent) error {
	payload, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("gagal melakukan marshalling pada peristiwa pesanan: %w", err)
	}
	record := &kgo.Record{
		Topic: "notification-events",
		Value: payload,
	}

	results := p.client.ProduceSync(ctx, record)
	if err := results.FirstErr(); err != nil {
		return fmt.Errorf("gagal mengirim pesan ke Kafka: %w", err)
	}

	return nil
}

func (p *OrderProducerImpl) Close() {
	if p.client != nil {
		p.client.Close()
	}
}
