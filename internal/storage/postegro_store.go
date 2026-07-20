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

func (s *PostgresStore) Create(user models.User) error {
	_, err := s.pool.Exec(
		context.Background(),
		`INSERT INTO users (id, username, password_hash) VALUES ($1, $2, $3)`,
		user.ID, user.Username, user.PasswordHash,
	)
	if err != nil {
		if isUniqueViolation(err) {
			return ErrUsernameTaken
		}
		return err
	}
	return nil
}

func (s *PostgresStore) FindByUsername(username string) (models.User, error) {
	return s.findBy(`SELECT id, username, password_hash FROM users WHERE username = $1`, username)
}

func (s *PostgresStore) FindByID(id string) (models.User, error) {
	return s.findBy(`SELECT id, username, password_hash FROM users WHERE id = $1`, id)
}

func (s *PostgresStore) findBy(query string, arg string) (models.User, error) {
	var u models.User
	err := s.pool.QueryRow(context.Background(), query, arg).
		Scan(&u.ID, &u.Username, &u.PasswordHash)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return models.User{}, ErrUserNotFound
		}
		return models.User{}, err
	}
	return u, nil
}

func (s *PostgresStore) Update(user models.User) error {
	tag, err := s.pool.Exec(
		context.Background(),
		`UPDATE users SET username = $1, password_hash = $2 WHERE id = $3`,
		user.Username, user.PasswordHash, user.ID,
	)
	if err != nil {
		if isUniqueViolation(err) {
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