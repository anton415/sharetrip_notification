package api

import (
	"encoding/json"
	"log"
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
		return c.SendStatus(fiber.StatusBadRequest)
	}

	if request.RecipientID == "" ||
		request.Type == "" ||
		len(request.Payload) == 0 ||
		!json.Valid(request.Payload) {
		return c.SendStatus(fiber.StatusBadRequest)
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
		log.Printf("create notification: %v", err)
		return c.SendStatus(fiber.StatusInternalServerError)
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
