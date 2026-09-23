package api

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"

	"job4j.ru/sharetrip-notification/internal/domain"
	repo "job4j.ru/sharetrip-notification/internal/repository"
)

func TestGetNotificationByIDOK(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	pool := newIntegrationPool(t)
	storage := repo.NewPostgresNotificationRepository(pool)

	created, err := storage.Create(ctx, domain.Notification{
		ID:          uuid.New(),
		RecipientID: "client-123",
		Type:        "trip_published",
		Status:      domain.NotificationStatusCreated,
		Payload:     json.RawMessage(`{"trip_id":"trip-456"}`),
	})
	if err != nil {
		t.Fatalf("create notification: %v", err)
	}
	t.Cleanup(func() {
		if _, err := pool.Exec(context.Background(), "DELETE FROM notifications WHERE id = $1", created.ID); err != nil {
			t.Errorf("delete test notification: %v", err)
		}
	})

	app := newTestApp(pool)
	request := httptest.NewRequest(http.MethodGet, "/notifications/"+created.ID.String(), nil)

	response, err := app.Test(request)
	if err != nil {
		t.Fatalf("perform request: %v", err)
	}
	defer closeResponseBody(t, response.Body)

	if response.StatusCode != fiber.StatusOK {
		t.Fatalf("expected status %d, got %d", fiber.StatusOK, response.StatusCode)
	}

	var found GetNotificationByIDResponse
	if err := json.NewDecoder(response.Body).Decode(&found); err != nil {
		t.Fatalf("decode response: %v", err)
	}

	if found.ID != created.ID {
		t.Errorf("expected id %s, got %s", created.ID, found.ID)
	}
	if found.RecipientID != created.RecipientID {
		t.Errorf("expected recipient_id %q, got %q", created.RecipientID, found.RecipientID)
	}
	if found.Type != created.Type {
		t.Errorf("expected type %q, got %q", created.Type, found.Type)
	}
	if found.Status != created.Status {
		t.Errorf("expected status %q, got %q", created.Status, found.Status)
	}
	if !found.CreatedAt.Equal(created.CreatedAt) {
		t.Errorf("expected created_at %s, got %s", created.CreatedAt, found.CreatedAt)
	}

	var payload map[string]string
	if err := json.Unmarshal(found.Payload, &payload); err != nil {
		t.Fatalf("decode payload: %v", err)
	}
	if payload["trip_id"] != "trip-456" {
		t.Errorf("expected trip_id %q, got %q", "trip-456", payload["trip_id"])
	}
}

func TestGetNotificationByIDNotFound(t *testing.T) {
	t.Parallel()

	pool := newIntegrationPool(t)
	app := newTestApp(pool)
	missingID := uuid.New()

	request := httptest.NewRequest(http.MethodGet, "/notifications/"+missingID.String(), nil)
	response, err := app.Test(request)
	if err != nil {
		t.Fatalf("perform request: %v", err)
	}
	defer closeResponseBody(t, response.Body)

	assertErrorResponse(t, response, fiber.StatusNotFound, errorResponse{
		Code:    "NOT_FOUND",
		Message: "notification not found",
	})
}

func TestGetNotificationByIDBadRequest(t *testing.T) {
	t.Parallel()

	pool := newClosedPool(t)
	app := newTestApp(pool)

	request := httptest.NewRequest(http.MethodGet, "/notifications/not-a-uuid", nil)
	response, err := app.Test(request)
	if err != nil {
		t.Fatalf("perform request: %v", err)
	}
	defer closeResponseBody(t, response.Body)

	assertErrorResponse(t, response, fiber.StatusBadRequest, errorResponse{
		Code:    "VALIDATION_ERROR",
		Message: "notification id must be a valid UUID",
	})
}

func TestGetNotificationByIDInternalServerError(t *testing.T) {
	t.Parallel()

	pool := newClosedPool(t)
	app := newTestApp(pool)

	request := httptest.NewRequest(http.MethodGet, "/notifications/"+uuid.New().String(), nil)
	response, err := app.Test(request)
	if err != nil {
		t.Fatalf("perform request: %v", err)
	}
	defer closeResponseBody(t, response.Body)

	assertErrorResponse(t, response, fiber.StatusInternalServerError, errorResponse{
		Code:    "INTERNAL_ERROR",
		Message: "internal server error",
	})
}

func newIntegrationPool(t *testing.T) *pgxpool.Pool {
	t.Helper()

	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		t.Skip("DATABASE_URL is not set")
	}

	pool, err := pgxpool.New(context.Background(), dsn)
	if err != nil {
		t.Fatalf("create pool: %v", err)
	}
	t.Cleanup(pool.Close)

	if err := pool.Ping(context.Background()); err != nil {
		t.Fatalf("ping PostgreSQL: %v", err)
	}

	return pool
}
