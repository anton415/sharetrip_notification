package api

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"

	"job4j.ru/sharetrip-notification/internal/domain"
	repo "job4j.ru/sharetrip-notification/internal/repository"
	"job4j.ru/sharetrip-notification/internal/service"
)

func TestCreateNotificationBadRequest(t *testing.T) {
	pool := newClosedPool(t)
	app := newTestApp(pool)

	tests := []struct {
		name string
		body string
	}{
		{name: "malformed JSON", body: `{"recipient_id":`},
		{name: "missing payload", body: `{"recipient_id":"client-123","type":"trip_published"}`},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			request := httptest.NewRequest(http.MethodPost, "/notifications", strings.NewReader(test.body))
			request.Header.Set(fiber.HeaderContentType, fiber.MIMEApplicationJSON)

			response, err := app.Test(request)
			if err != nil {
				t.Fatalf("perform request: %v", err)
			}
			defer response.Body.Close()

			if response.StatusCode != fiber.StatusBadRequest {
				t.Fatalf("expected status %d, got %d", fiber.StatusBadRequest, response.StatusCode)
			}
		})
	}
}

func TestCreateNotificationInternalServerError(t *testing.T) {
	pool := newClosedPool(t)
	app := newTestApp(pool)

	request := httptest.NewRequest(
		http.MethodPost,
		"/notifications",
		strings.NewReader(`{"recipient_id":"client-123","type":"trip_published","payload":{"trip_id":"trip-456"}}`),
	)
	request.Header.Set(fiber.HeaderContentType, fiber.MIMEApplicationJSON)

	response, err := app.Test(request)
	if err != nil {
		t.Fatalf("perform request: %v", err)
	}
	defer response.Body.Close()

	if response.StatusCode != fiber.StatusInternalServerError {
		t.Fatalf("expected status %d, got %d", fiber.StatusInternalServerError, response.StatusCode)
	}
}

func TestCreateNotificationCreated(t *testing.T) {
	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		t.Skip("DATABASE_URL is not set")
	}

	ctx := context.Background()
	pool, err := pgxpool.New(ctx, dsn)
	if err != nil {
		t.Fatalf("create pool: %v", err)
	}
	t.Cleanup(pool.Close)

	if err := pool.Ping(ctx); err != nil {
		t.Fatalf("ping PostgreSQL: %v", err)
	}

	app := newTestApp(pool)
	request := httptest.NewRequest(
		http.MethodPost,
		"/notifications",
		strings.NewReader(`{"recipient_id":"client-123","type":"trip_published","payload":{"trip_id":"trip-456"}}`),
	)
	request.Header.Set(fiber.HeaderContentType, fiber.MIMEApplicationJSON)

	response, err := app.Test(request)
	if err != nil {
		t.Fatalf("perform request: %v", err)
	}
	defer response.Body.Close()

	if response.StatusCode != fiber.StatusCreated {
		t.Fatalf("expected status %d, got %d", fiber.StatusCreated, response.StatusCode)
	}

	var created CreateNotificationResponse
	if err := json.NewDecoder(response.Body).Decode(&created); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	t.Cleanup(func() {
		if _, err := pool.Exec(context.Background(), "DELETE FROM notifications WHERE id = $1", created.ID); err != nil {
			t.Errorf("delete test notification: %v", err)
		}
	})

	if created.ID == uuid.Nil {
		t.Error("expected generated UUID")
	}
	if created.RecipientID != "client-123" {
		t.Errorf("expected recipient_id %q, got %q", "client-123", created.RecipientID)
	}
	if created.Type != "trip_published" {
		t.Errorf("expected type %q, got %q", "trip_published", created.Type)
	}
	if created.Status != domain.NotificationStatusCreated {
		t.Errorf("expected status %q, got %q", domain.NotificationStatusCreated, created.Status)
	}
	if created.CreatedAt.IsZero() {
		t.Error("expected created_at from PostgreSQL")
	}

	var payload map[string]string
	if err := json.Unmarshal(created.Payload, &payload); err != nil {
		t.Fatalf("decode payload: %v", err)
	}
	if payload["trip_id"] != "trip-456" {
		t.Errorf("expected trip_id %q, got %q", "trip-456", payload["trip_id"])
	}

	var exists bool
	if err := pool.QueryRow(ctx, "SELECT EXISTS (SELECT 1 FROM notifications WHERE id = $1)", created.ID).Scan(&exists); err != nil {
		t.Fatalf("check stored notification: %v", err)
	}
	if !exists {
		t.Error("created notification was not stored")
	}
}

func newClosedPool(t *testing.T) *pgxpool.Pool {
	t.Helper()

	pool, err := pgxpool.New(context.Background(), "postgres://localhost/postgres?sslmode=disable")
	if err != nil {
		t.Fatalf("create pool: %v", err)
	}
	pool.Close()
	return pool
}

func newTestApp(pool *pgxpool.Pool) *fiber.App {
	notificationRepo := repo.NewPostgresNotificationRepository(pool)
	notificationService := service.NewNotificationService(notificationRepo)
	server := NewServer(notificationService)

	app := fiber.New()
	server.RegisterRoutes(app)
	return app
}
