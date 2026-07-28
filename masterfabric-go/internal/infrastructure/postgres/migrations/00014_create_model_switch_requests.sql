-- +goose Up
CREATE TABLE IF NOT EXISTS model_switch_requests (
    id             UUID PRIMARY KEY,
    profile_id     TEXT NOT NULL,
    request_model  TEXT NOT NULL,
    engine_model   TEXT NOT NULL,
    local_adapter  TEXT NOT NULL DEFAULT '',
    status         TEXT NOT NULL DEFAULT 'pending',
    error_message  TEXT,
    requested_by   UUID REFERENCES users(id) ON DELETE SET NULL,
    created_at     TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at     TIMESTAMPTZ NOT NULL DEFAULT now(),
    completed_at   TIMESTAMPTZ
);

CREATE INDEX IF NOT EXISTS model_switch_requests_status_idx
    ON model_switch_requests (status, created_at DESC);

-- +goose Down
DROP TABLE IF EXISTS model_switch_requests;
