package storage

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/leventkok/mlc-llm-monitoring/internal/models"
)

type PostgresStore struct {
	pool *pgxpool.Pool
}

func NewPostgresStore(pool *pgxpool.Pool) *PostgresStore {
	return &PostgresStore{pool: pool}
}

func (s *PostgresStore) Create(ctx context.Context, user models.User) error {
	_, err := s.pool.Exec(ctx,
		`INSERT INTO users (id, email, username, password_hash) VALUES ($1, $2, $3, $4)`,
		user.ID, user.Email, user.Username, user.PasswordHash,
	)
	if err != nil {
		if isUniqueViolation(err) {
			if isEmailConstraint(err) {
				return ErrEmailTaken
			}
			return ErrUsernameTaken
		}
		return err
	}
	return nil
}

func (s *PostgresStore) FindByEmail(ctx context.Context, email string) (models.User, error) {
	return s.findBy(ctx, `SELECT id, email, username, password_hash FROM users WHERE email = $1`, email)
}

func (s *PostgresStore) FindByUsername(ctx context.Context, username string) (models.User, error) {
	return s.findBy(ctx, `SELECT id, email, username, password_hash FROM users WHERE username = $1`, username)
}

func (s *PostgresStore) FindByID(ctx context.Context, id string) (models.User, error) {
	return s.findBy(ctx, `SELECT id, email, username, password_hash FROM users WHERE id = $1`, id)
}

func (s *PostgresStore) findBy(ctx context.Context, query string, arg string) (models.User, error) {
	var u models.User
	err := s.pool.QueryRow(ctx, query, arg).
		Scan(&u.ID, &u.Email, &u.Username, &u.PasswordHash)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return models.User{}, ErrUserNotFound
		}
		return models.User{}, err
	}
	return u, nil
}

func (s *PostgresStore) Delete(ctx context.Context, id string) error {
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	if _, err := tx.Exec(ctx, `UPDATE scores SET scored_by = NULL WHERE scored_by = $1`, id); err != nil {
		return err
	}
	tag, err := tx.Exec(ctx, `DELETE FROM users WHERE id = $1`, id)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return ErrUserNotFound
	}
	return tx.Commit(ctx)
}

func (s *PostgresStore) Update(ctx context.Context, user models.User) error {
	tag, err := s.pool.Exec(ctx,
		`UPDATE users SET email = $1, username = $2, password_hash = $3 WHERE id = $4`,
		user.Email, user.Username, user.PasswordHash, user.ID,
	)
	if err != nil {
		if isUniqueViolation(err) {
			if isEmailConstraint(err) {
				return ErrEmailTaken
			}
			return ErrUsernameTaken
		}
		return err
	}
	if tag.RowsAffected() == 0 {
		return ErrUserNotFound
	}
	return nil
}

func isUniqueViolation(err error) bool {
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		return pgErr.Code == "23505"
	}
	return false
}

func isEmailConstraint(err error) bool {
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		return pgErr.ConstraintName == "users_email_key"
	}
	return false
}
