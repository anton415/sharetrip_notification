package service

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/google/uuid"

	"job4j.ru/sharetrip-notification/internal/domain"
	"job4j.ru/sharetrip-notification/internal/events"
)

func (s *NotificationService) HandleTripPublished(ctx context.Context, event events.TripPublished) error {
	eventID, err := uuid.Parse(event.EventID)
	if err != nil {
		return fmt.Errorf("parse event_id: %w", err)
	}

	payload, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("marshal TripPublished: %w", err)
	}

	err = s.repo.CreateFromEvent(ctx, eventID, domain.Notification{
		ID:          uuid.New(),
		RecipientID: event.DriverID,
		Type:        "trip_published",
		Status:      domain.NotificationStatusCreated,
		Payload:     payload,
	})
	if err != nil {
		return fmt.Errorf("handle TripPublished: %w", err)
	}

	return nil
}
