package events

import (
	"context"
	"errors"
	"os"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/segmentio/kafka-go"
)

func TestConsumerCommits(t *testing.T) {
	t.Parallel()

	brokers := os.Getenv("KAFKA_BROKERS")
	if brokers == "" {
		t.Skip("KAFKA_BROKERS is not set")
	}
	tests := []struct {
		name       string
		messages   []string
		wantEvents []string
		wantOffset int64
	}{
		{
			name: "filter and commit success, stop without committing handler failure",
			messages: []string{
				`{"event_type":"TripStarted"}`,
				`{"event_type":"TripPublished","event_id":"success"}`,
				`{"event_type":"TripPublished","event_id":"failure"}`,
			},
			wantEvents: []string{"success", "failure"}, wantOffset: 2,
		},
		{
			name:       "handler failure leaves first event uncommitted",
			messages:   []string{`{"event_type":"TripPublished","event_id":"failure"}`},
			wantEvents: []string{"failure"}, wantOffset: -1,
		},
		{
			name: "invalid JSON stops before following valid event",
			messages: []string{
				`{`, `{"event_type":"TripPublished","event_id":"failure"}`,
			},
			wantOffset: -1,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			ctx, cancel := context.WithTimeout(t.Context(), 30*time.Second)
			defer cancel()
			topic := "notification-test-" + uuid.NewString()
			client := &kafka.Client{Addr: kafka.TCP(strings.Split(brokers, ",")...)}
			t.Cleanup(func() {
				ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
				defer cancel()
				if result, err := client.DeleteGroups(ctx, &kafka.DeleteGroupsRequest{GroupIDs: []string{topic}}); err != nil {
					t.Errorf("delete test group: %v", err)
				} else if err := result.Errors[topic]; err != nil {
					t.Errorf("delete test group: %v", err)
				}
				if result, err := client.DeleteTopics(ctx, &kafka.DeleteTopicsRequest{Topics: []string{topic}}); err != nil {
					t.Errorf("delete test topic: %v", err)
				} else if err := result.Errors[topic]; err != nil {
					t.Errorf("delete test topic: %v", err)
				}
			})
			created, err := client.CreateTopics(ctx, &kafka.CreateTopicsRequest{
				Topics: []kafka.TopicConfig{{Topic: topic, NumPartitions: 1, ReplicationFactor: 1}},
			})
			if err != nil {
				t.Fatalf("create test topic: %v", err)
			}
			if err := created.Errors[topic]; err != nil {
				t.Fatalf("create test topic: %v", err)
			}
			writer := &kafka.Writer{Addr: client.Addr, Topic: topic, RequiredAcks: kafka.RequireAll}
			var messages []kafka.Message
			for _, value := range test.messages {
				messages = append(messages, kafka.Message{Value: []byte(value)})
			}
			err = writer.WriteMessages(ctx, messages...)
			closeErr := writer.Close()
			if err != nil || closeErr != nil {
				t.Fatalf("write messages: %v, close: %v", err, closeErr)
			}

			consumer := NewConsumer(strings.Split(brokers, ","), topic, topic)
			var handled []string
			handlerErr := errors.New("notification insert failed")
			err = consumer.Run(ctx, func(_ context.Context, event TripPublished) error {
				handled = append(handled, event.EventID)
				if event.EventID == "failure" {
					return handlerErr
				}
				return nil
			})
			if closeErr := consumer.Close(); closeErr != nil {
				t.Errorf("close consumer: %v", closeErr)
			}
			if err == nil || errors.Is(err, context.DeadlineExceeded) {
				t.Fatalf("expected message processing error, got %v", err)
			}
			if test.wantEvents != nil && !errors.Is(err, handlerErr) {
				t.Errorf("expected handler error, got %v", err)
			}
			if !reflect.DeepEqual(handled, test.wantEvents) {
				t.Errorf("handled events = %v, want %v", handled, test.wantEvents)
			}
			offsets, err := client.OffsetFetch(ctx, &kafka.OffsetFetchRequest{
				GroupID: topic, Topics: map[string][]int{topic: {0}},
			})
			if err != nil {
				t.Fatalf("fetch committed offsets: %v", err)
			}
			if offsets.Error != nil || len(offsets.Topics[topic]) != 1 {
				t.Fatalf("unexpected offsets response: %+v", offsets)
			}
			partition := offsets.Topics[topic][0]
			if partition.Error != nil || partition.CommittedOffset != test.wantOffset {
				t.Errorf("committed offset = %d, error = %v, want %d", partition.CommittedOffset, partition.Error, test.wantOffset)
			}
		})
	}
}
