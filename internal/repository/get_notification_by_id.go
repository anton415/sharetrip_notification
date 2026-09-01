package repo

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"job4j.ru/sharetrip-notification/internal/domain"
)

func (r *PostgresNotificationRepository) GetByID(
	ctx context.Context,
	id uuid.UUID,
) (domain.Notification, error) {
	sql := `
		SELECT id, recipient_id, type, status, payload, created_at
		FROM notifications
		WHERE id = $1
	`
	var result domain.Notification
	err := r.pool.QueryRow(ctx, sql, id).Scan(
		&result.ID,
		&result.RecipientID,
		&result.Type,
		&result.Status,
		&result.Payload,
		&result.CreatedAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return domain.Notification{}, domain.ErrNotificationNotFound
	}
	if err != nil {
		return domain.Notification{}, fmt.Errorf("get notification by id: %w", err)
	}

	return result, nil
}
