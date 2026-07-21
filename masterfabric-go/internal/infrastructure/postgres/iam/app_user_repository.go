package iam

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/leventkok/mlc-llm-monitoring/masterfabric-go/internal/domain/iam/model"
	domainErr "github.com/leventkok/mlc-llm-monitoring/masterfabric-go/internal/shared/errors"
)

var (
	ErrEmailTaken    = errors.New("email already taken")
	ErrUsernameTaken = errors.New("username already taken")
)

// AppUserRepo implements application user persistence (username-based schema).
type AppUserRepo struct {
	db *pgxpool.Pool
}

func NewAppUserRepo(db *pgxpool.Pool) *AppUserRepo {
	return &AppUserRepo{db: db}
}

func (r *AppUserRepo) Create(ctx context.Context, user *model.User) error {
	if user.ID == uuid.Nil {
		user.ID = uuid.New()
	}
	if user.CreatedAt.IsZero() {
		user.CreatedAt = time.Now().UTC()
	}

	_, err := r.db.Exec(ctx,
		`INSERT INTO users (id, email, username, password_hash, created_at)
		 VALUES ($1, $2, $3, $4, $5)`,
		user.ID, user.Email, user.Username, user.PasswordHash, user.CreatedAt,
	)
	if err != nil {
		if isUniqueViolation(err) {
			if isEmailConstraint(err) {
				return ErrEmailTaken
			}
			return ErrUsernameTaken
		}
		return domainErr.New(domainErr.ErrInternal, "failed to create user", err)
	}
	return nil
}

func (r *AppUserRepo) GetByID(ctx context.Context, id uuid.UUID) (*model.User, error) {
	return r.scanOne(ctx, `SELECT id, email, username, password_hash, created_at FROM users WHERE id = $1`, id)
}

func (r *AppUserRepo) GetByEmail(ctx context.Context, email string) (*model.User, error) {
	return r.scanOne(ctx, `SELECT id, email, username, password_hash, created_at FROM users WHERE email = $1`, email)
}

func (r *AppUserRepo) GetByUsername(ctx context.Context, username string) (*model.User, error) {
	return r.scanOne(ctx, `SELECT id, email, username, password_hash, created_at FROM users WHERE username = $1`, username)
}

func (r *AppUserRepo) scanOne(ctx context.Context, query string, arg any) (*model.User, error) {
	var u model.User
	err := r.db.QueryRow(ctx, query, arg).Scan(&u.ID, &u.Email, &u.Username, &u.PasswordHash, &u.CreatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domainErr.New(domainErr.ErrNotFound, "user not found", nil)
		}
		return nil, domainErr.New(domainErr.ErrInternal, "failed to get user", err)
	}
	return &u, nil
}

func (r *AppUserRepo) Update(ctx context.Context, user *model.User) error {
	tag, err := r.db.Exec(ctx,
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
		return domainErr.New(domainErr.ErrInternal, "failed to update user", err)
	}
	if tag.RowsAffected() == 0 {
		return domainErr.New(domainErr.ErrNotFound, "user not found", nil)
	}
	return nil
}

func (r *AppUserRepo) Delete(ctx context.Context, id uuid.UUID) error {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return domainErr.New(domainErr.ErrInternal, "failed to begin transaction", err)
	}
	defer tx.Rollback(ctx)

	if _, err := tx.Exec(ctx, `UPDATE scores SET scored_by = NULL WHERE scored_by = $1`, id); err != nil {
		return domainErr.New(domainErr.ErrInternal, "failed to clear scores", err)
	}
	tag, err := tx.Exec(ctx, `DELETE FROM users WHERE id = $1`, id)
	if err != nil {
		return domainErr.New(domainErr.ErrInternal, "failed to delete user", err)
	}
	if tag.RowsAffected() == 0 {
		return domainErr.New(domainErr.ErrNotFound, "user not found", nil)
	}
	return tx.Commit(ctx)
}

func isUniqueViolation(err error) bool {
	var pgErr *pgconn.PgError
	return errors.As(err, &pgErr) && pgErr.Code == "23505"
}

func isEmailConstraint(err error) bool {
	var pgErr *pgconn.PgError
	return errors.As(err, &pgErr) && pgErr.ConstraintName == "users_email_key"
}
