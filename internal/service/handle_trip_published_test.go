package service_test

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"reflect"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"

	"job4j.ru/sharetrip-notification/internal/domain"
	"job4j.ru/sharetrip-notification/internal/events"
	repo "job4j.ru/sharetrip-notification/internal/repository"
	"job4j.ru/sharetrip-notification/internal/service"
)

func TestTripPublishedProcessing(t *testing.T) {
	t.Parallel()

	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		t.Skip("DATABASE_URL is not set")
	}
	ctx := t.Context()
	pool, err := pgxpool.New(ctx, dsn)
	if err != nil {
		t.Fatalf("create pool: %v", err)
	}
	t.Cleanup(pool.Close)
	notificationRepo := repo.NewPostgresNotificationRepository(pool)
	notifications := service.NewNotificationService(notificationRepo)

	for _, workers := range []int{1, 8} {
		t.Run(fmt.Sprintf("%d concurrent deliveries and a retry", workers), func(t *testing.T) {
			t.Parallel()

			event := events.TripPublished{
				EventID: uuid.NewString(), EventType: "TripPublished", TripID: uuid.NewString(),
				DriverID: uuid.NewString(), CompanyID: uuid.NewString(), OccurredAt: time.Now().UTC(),
			}
			cleanupEvent(t, pool, event.EventID, event.DriverID)
			results := make(chan error, workers)
			for range workers {
				go func() { results <- notifications.HandleTripPublished(ctx, event) }()
			}
			for range workers {
				if err := <-results; err != nil {
					t.Errorf("handle event: %v", err)
				}
			}
			if err := notifications.HandleTripPublished(ctx, event); err != nil {
				t.Fatalf("handle duplicate: %v", err)
			}
			assertEventCounts(t, pool, event.EventID, event.DriverID, 1)

			var notification domain.Notification
			err := pool.QueryRow(ctx, `
				SELECT id, type, status, payload, created_at FROM notifications WHERE recipient_id = $1
			`, event.DriverID).Scan(&notification.ID, &notification.Type, &notification.Status,
				&notification.Payload, &notification.CreatedAt)
			if err != nil {
				t.Fatalf("read notification: %v", err)
			}
			if notification.ID == uuid.Nil || notification.CreatedAt.IsZero() ||
				notification.Type != "trip_published" || notification.Status != domain.NotificationStatusCreated {
				t.Errorf("unexpected notification: %+v", notification)
			}
			var payload events.TripPublished
			if err := json.Unmarshal(notification.Payload, &payload); err != nil {
				t.Fatalf("decode payload: %v", err)
			}
			if !reflect.DeepEqual(event, payload) {
				t.Errorf("payload = %+v, want %+v", payload, event)
			}
		})
	}

	t.Run("notification failure rolls back event and allows retry", func(t *testing.T) {
		t.Parallel()

		eventID := uuid.New()
		notification := domain.Notification{
			ID: uuid.New(), RecipientID: uuid.NewString(), Type: "trip_published",
			Status: domain.NotificationStatusCreated, Payload: json.RawMessage(`{`),
		}
		cleanupEvent(t, pool, eventID.String(), notification.RecipientID)
		if err := notificationRepo.CreateFromEvent(ctx, eventID, notification); err == nil {
			t.Fatal("expected invalid JSON to fail notification insertion")
		}
		assertEventCounts(t, pool, eventID.String(), notification.RecipientID, 0)

		notification.Payload = json.RawMessage(`{"trip_id":"synthetic-trip"}`)
		if err := notificationRepo.CreateFromEvent(ctx, eventID, notification); err != nil {
			t.Fatalf("retry after rollback: %v", err)
		}
		assertEventCounts(t, pool, eventID.String(), notification.RecipientID, 1)
	})
}

func assertEventCounts(t *testing.T, pool *pgxpool.Pool, eventID, recipientID string, want int) {
	t.Helper()
	var processed, notifications int
	err := pool.QueryRow(t.Context(), `
		SELECT (SELECT COUNT(*) FROM processed_events WHERE event_id = $1),
		       (SELECT COUNT(*) FROM notifications WHERE recipient_id = $2)
	`, eventID, recipientID).Scan(&processed, &notifications)
	if err != nil {
		t.Fatalf("count records: %v", err)
	}
	if processed != want || notifications != want {
		t.Errorf("processed_events = %d, notifications = %d, want %d each", processed, notifications, want)
	}
}

func cleanupEvent(t *testing.T, pool *pgxpool.Pool, eventID, recipientID string) {
	t.Helper()
	t.Cleanup(func() {
		ctx := context.Background()
		if _, err := pool.Exec(ctx, `DELETE FROM notifications WHERE recipient_id = $1`, recipientID); err != nil {
			t.Errorf("delete test notifications: %v", err)
		}
		if _, err := pool.Exec(ctx, `DELETE FROM processed_events WHERE event_id = $1`, eventID); err != nil {
			t.Errorf("delete test event: %v", err)
		}
	})
}
