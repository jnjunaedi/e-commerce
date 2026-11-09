package kafka

import (
	"context"
	"e-commerce/notification_service/internal/domain"
	"encoding/json"
	"time"

	"github.com/twmb/franz-go/pkg/kgo"
)

type Consumer struct {
	client *kgo.Client
	uc     domain.Usecase
}

func NewConsumer(brokers []string, groupID string, topics []string, uc domain.Usecase) (*Consumer, error) {
	client, err := kgo.NewClient(kgo.SeedBrokers(brokers...), kgo.ConsumerGroup(groupID), kgo.ConsumeTopics(topics...), kgo.ConsumeResetOffset(kgo.NewOffset().AtStart()), kgo.AllowAutoTopicCreation(), kgo.RecordDeliveryTimeout(5*time.Second))
	if err != nil {
		return nil, err
	}

	return &Consumer{client: client, uc: uc}, nil
}

func (c *Consumer) Start(ctx context.Context, errChan chan<- error) {
	defer close(errChan)

	for {
		fetches := c.client.PollFetches(ctx)
		if fetches.IsClientClosed() || ctx.Err() != nil {
			break
		}

		if errs := fetches.Errors(); len(errs) > 0 {
			for _, e := range errs {
				select {
				case errChan <- e.Err:
				case <-ctx.Done():
					return
				}
			}
			continue
		}

		var processErr error
		fetches.EachRecord(func(r *kgo.Record) {
			if processErr != nil {
				return
			}
			var notification domain.Notification

			if err := json.Unmarshal(r.Value, &notification); err != nil {
				processErr = err
				return
			}

			err := c.uc.SendAndSaveNotification(ctx, notification)
			if err != nil {
				processErr = err
				return
			}
		})

		if processErr != nil {
			select {
			case errChan <- processErr:
			case <-ctx.Done():
				return
			}
		}
	}
}

func (c *Consumer) Close() {
	if c.client != nil {
		c.client.Close()
	}
}
