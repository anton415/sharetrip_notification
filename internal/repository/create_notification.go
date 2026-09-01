package repo

import (
	"context"
	"fmt"

	"job4j.ru/sharetrip-notification/internal/domain"
)

func (r *PostgresNotificationRepository) Create(
	ctx context.Context,
	notification domain.Notification,
) (domain.Notification, error) {
	sql := `
		INSERT INTO notifications (id, recipient_id, type, status, payload)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id, recipient_id, type, status, payload, created_at
	`
	var result domain.Notification
	err := r.pool.QueryRow(ctx, sql,
		notification.ID,
		notification.RecipientID,
		notification.Type,
		notification.Status,
		notification.Payload,
	).Scan(
		&result.ID,
		&result.RecipientID,
		&result.Type,
		&result.Status,
		&result.Payload,
		&result.CreatedAt,
	)
	if err != nil {
		return domain.Notification{}, fmt.Errorf("insert notification: %w", err)
	}

	return result, nil
}
