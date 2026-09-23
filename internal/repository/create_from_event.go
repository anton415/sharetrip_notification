package repo

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"job4j.ru/sharetrip-notification/internal/domain"
)

func (r *PostgresNotificationRepository) CreateFromEvent(
	ctx context.Context,
	eventID uuid.UUID,
	notification domain.Notification,
) error {
	err := pgx.BeginFunc(ctx, r.pool, func(tx pgx.Tx) error {
		result, err := tx.Exec(ctx, `
			INSERT INTO processed_events (event_id) VALUES ($1)
			ON CONFLICT (event_id) DO NOTHING
		`, eventID)
		if err != nil {
			return fmt.Errorf("insert processed event: %w", err)
		}
		if result.RowsAffected() == 0 {
			return nil
		}

		_, err = tx.Exec(ctx, `
			INSERT INTO notifications (id, recipient_id, type, status, payload)
			VALUES ($1, $2, $3, $4, $5)
		`, notification.ID, notification.RecipientID, notification.Type, notification.Status, notification.Payload)
		if err != nil {
			return fmt.Errorf("insert notification: %w", err)
		}

		return nil
	})
	if err != nil {
		return fmt.Errorf("create notification from event: %w", err)
	}

	return nil
}
