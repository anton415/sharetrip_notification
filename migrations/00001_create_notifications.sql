-- +goose Up
-- +goose StatementBegin

CREATE TABLE notifications (
    id UUID PRIMARY KEY,
    recipient_id TEXT NOT NULL,
    type TEXT NOT NULL,
    status TEXT NOT NULL,
    payload JSONB NOT NULL DEFAULT '{}'::jsonb,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

-- +goose StatementEnd


-- +goose Down
-- +goose StatementBegin

DROP TABLE IF EXISTS notifications;

-- +goose StatementEnd
