package service

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/google/uuid"

	"job4j.ru/sharetrip-notification/internal/domain"
)

type CreateNotificationCommand struct {
	RecipientID string
	Type        string
	Payload     json.RawMessage
}

func (s *NotificationService) CreateNotification(
	ctx context.Context,
	command CreateNotificationCommand,
) (domain.Notification, error) {
	notification := domain.Notification{
		ID:          uuid.New(),
		RecipientID: command.RecipientID,
		Type:        command.Type,
		Status:      domain.NotificationStatusCreated,
		Payload:     command.Payload,
	}

	created, err := s.repo.Create(ctx, notification)
	if err != nil {
		return domain.Notification{}, fmt.Errorf("create notification: %w", err)
	}

	return created, nil
}
