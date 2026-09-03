package api

import (
	"encoding/json"
	"log/slog"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"

	"job4j.ru/sharetrip-notification/internal/domain"
	"job4j.ru/sharetrip-notification/internal/service"
)

type CreateNotificationRequest struct {
	RecipientID string          `json:"recipient_id"`
	Type        string          `json:"type"`
	Payload     json.RawMessage `json:"payload"`
}

type CreateNotificationResponse struct {
	ID          uuid.UUID                 `json:"id"`
	RecipientID string                    `json:"recipient_id"`
	Type        string                    `json:"type"`
	Status      domain.NotificationStatus `json:"status"`
	Payload     json.RawMessage           `json:"payload"`
	CreatedAt   time.Time                 `json:"created_at"`
}

func (s *Server) createNotification(c *fiber.Ctx) error {
	var request CreateNotificationRequest
	if err := c.BodyParser(&request); err != nil {
		return writeErrorResponse(
			c,
			fiber.StatusBadRequest,
			errorCodeValidation,
			"invalid request body",
		)
	}

	switch {
	case strings.TrimSpace(request.RecipientID) == "":
		return writeErrorResponse(
			c,
			fiber.StatusBadRequest,
			errorCodeValidation,
			"recipient_id is required",
		)
	case strings.TrimSpace(request.Type) == "":
		return writeErrorResponse(
			c,
			fiber.StatusBadRequest,
			errorCodeValidation,
			"type is required",
		)
	case len(request.Payload) == 0:
		return writeErrorResponse(
			c,
			fiber.StatusBadRequest,
			errorCodeValidation,
			"payload is required",
		)
	}

	notification, err := s.notifications.CreateNotification(
		c.UserContext(),
		service.CreateNotificationCommand{
			RecipientID: request.RecipientID,
			Type:        request.Type,
			Payload:     request.Payload,
		},
	)
	if err != nil {
		s.logger.ErrorContext(
			c.UserContext(),
			"create notification failed",
			slog.Any("error", err),
			slog.String("recipient_id", request.RecipientID),
			slog.String("notification_type", request.Type),
		)
		return writeErrorResponse(
			c,
			fiber.StatusInternalServerError,
			errorCodeInternal,
			"internal server error",
		)
	}

	return c.Status(fiber.StatusCreated).JSON(newCreateNotificationResponse(notification))
}

func newCreateNotificationResponse(notification domain.Notification) CreateNotificationResponse {
	return CreateNotificationResponse{
		ID:          notification.ID,
		RecipientID: notification.RecipientID,
		Type:        notification.Type,
		Status:      notification.Status,
		Payload:     notification.Payload,
		CreatedAt:   notification.CreatedAt,
	}
}
