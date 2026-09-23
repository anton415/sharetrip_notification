package events

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/segmentio/kafka-go"
)

type Consumer struct {
	reader *kafka.Reader
}

func NewConsumer(brokers []string, topic string, groupID string) *Consumer {
	return &Consumer{
		reader: kafka.NewReader(kafka.ReaderConfig{
			Brokers: brokers,
			Topic:   topic,
			GroupID: groupID,
		}),
	}
}

func (c *Consumer) Run(ctx context.Context, handler func(context.Context, TripPublished) error) error {
	for {
		msg, err := c.reader.FetchMessage(ctx)
		if err != nil {
			return err
		}

		var event TripPublished
		if err := json.Unmarshal(msg.Value, &event); err != nil {
			return fmt.Errorf("decode Kafka message: %w", err)
		}

		if event.EventType != "TripPublished" {
			if err := c.reader.CommitMessages(ctx, msg); err != nil {
				return err
			}
			continue
		}

		if err := handler(ctx, event); err != nil {
			return err
		}

		if err := c.reader.CommitMessages(ctx, msg); err != nil {
			return err
		}
	}
}

func (c *Consumer) Close() error {
	return c.reader.Close()
}
