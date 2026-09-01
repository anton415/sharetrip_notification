package api

import (
	"encoding/json"
	"errors"
	"log"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"

	"job4j.ru/sharetrip-notification/internal/domain"
)

type GetNotificationByIDResponse struct {
	ID          uuid.UUID                 `json:"id"`
	RecipientID string                    `json:"recipient_id"`
	Type        string                    `json:"type"`
	Status      domain.NotificationStatus `json:"status"`
	Payload     json.RawMessage           `json:"payload"`
	CreatedAt   time.Time                 `json:"created_at"`
}

func (s *Server) getNotificationByID(c *fiber.Ctx) error {
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.SendStatus(fiber.StatusBadRequest)
	}

	notification, err := s.notifications.GetNotificationByID(c.UserContext(), id)
	if errors.Is(err, domain.ErrNotificationNotFound) {
		return c.SendStatus(fiber.StatusNotFound)
	}
	if err != nil {
		log.Printf("get notification by id: %v", err)
		return c.SendStatus(fiber.StatusInternalServerError)
	}

	return c.Status(fiber.StatusOK).JSON(newGetNotificationByIDResponse(notification))
}

func newGetNotificationByIDResponse(notification domain.Notification) GetNotificationByIDResponse {
	return GetNotificationByIDResponse{
		ID:          notification.ID,
		RecipientID: notification.RecipientID,
		Type:        notification.Type,
		Status:      notification.Status,
		Payload:     notification.Payload,
		CreatedAt:   notification.CreatedAt,
	}
}
