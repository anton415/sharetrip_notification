package service

import (
	"context"
	"fmt"

	"github.com/google/uuid"

	"job4j.ru/sharetrip-notification/internal/domain"
)

func (s *NotificationService) GetNotificationByID(
	ctx context.Context,
	id uuid.UUID,
) (domain.Notification, error) {
	notification, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return domain.Notification{}, fmt.Errorf("get notification by id: %w", err)
	}

	return notification, nil
}
