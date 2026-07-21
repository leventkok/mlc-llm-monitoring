-- +goose Up
CREATE TABLE IF NOT EXISTS users (
    id            UUID PRIMARY KEY,
    email         TEXT UNIQUE NOT NULL,
    username      TEXT UNIQUE NOT NULL,
    password_hash TEXT NOT NULL,
    created_at    TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE IF NOT EXISTS reviews (
    id         UUID PRIMARY KEY,
    user_id    UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    app_name   TEXT NOT NULL,
    store      TEXT NOT NULL,
    rating     INT,
    text       TEXT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE IF NOT EXISTS decisions (
    id         UUID PRIMARY KEY,
    review_id  UUID NOT NULL REFERENCES reviews(id) ON DELETE CASCADE,
    category   TEXT NOT NULL,
    sentiment  TEXT NOT NULL,
    raw_output TEXT,
    latency_ms INT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE IF NOT EXISTS scores (
    id               UUID PRIMARY KEY,
    decision_id      UUID NOT NULL REFERENCES decisions(id) ON DELETE CASCADE,
    quality          INT NOT NULL,
    correct_category TEXT,
    scored_by        UUID REFERENCES users(id),
    created_at       TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS reviews_user_id_idx ON reviews (user_id);
CREATE INDEX IF NOT EXISTS decisions_review_id_idx ON decisions (review_id);
CREATE INDEX IF NOT EXISTS scores_decision_id_idx ON scores (decision_id);
CREATE INDEX IF NOT EXISTS scores_scored_by_idx ON scores (scored_by);
CREATE UNIQUE INDEX IF NOT EXISTS scores_one_per_decision_idx ON scores (decision_id);
CREATE UNIQUE INDEX IF NOT EXISTS decisions_one_per_review_idx ON decisions (review_id);

-- +goose Down
DROP TABLE IF EXISTS scores;
DROP TABLE IF EXISTS decisions;
DROP TABLE IF EXISTS reviews;
DROP TABLE IF EXISTS users;
