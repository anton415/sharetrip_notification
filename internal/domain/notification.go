package domain

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

type Notification struct {
	ID          uuid.UUID
	RecipientID string
	Type        string
	Status      NotificationStatus
	Payload     json.RawMessage
	CreatedAt   time.Time
}

type NotificationStatus string

const (
	NotificationStatusCreated NotificationStatus = "created"
)
