package database

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
)

const schema = `
CREATE TABLE IF NOT EXISTS users (
	id            UUID PRIMARY KEY,
	username      TEXT UNIQUE NOT NULL,
	password_hash TEXT NOT NULL,
	created_at    TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE IF NOT EXISTS reviews (
	id         UUID PRIMARY KEY,
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
`

func Migrate(ctx context.Context, pool *pgxpool.Pool) error {
	_, err := pool.Exec(ctx, schema)
	return err
}