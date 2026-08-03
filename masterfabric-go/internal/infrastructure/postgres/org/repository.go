package org

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"regexp"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
	orgModel "github.com/leventkok/mlc-llm-monitoring/masterfabric-go/internal/domain/org/model"
)

var (
	ErrNotFound      = errors.New("not found")
	ErrAlreadyMember = errors.New("user already belongs to an organization")
	ErrInviteInvalid = errors.New("invite invalid or expired")
)

type Repository struct {
	db *pgxpool.Pool
}

func NewRepository(db *pgxpool.Pool) *Repository {
	return &Repository{db: db}
}

func Slugify(name string) string {
	s := strings.ToLower(strings.TrimSpace(name))
	re := regexp.MustCompile(`[^a-z0-9]+`)
	s = re.ReplaceAllString(s, "-")
	s = strings.Trim(s, "-")
	if s == "" {
		s = "org"
	}
	return s
}

func newToken() (string, error) {
	b := make([]byte, 24)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}

func (r *Repository) CreateOrganization(ctx context.Context, name string) (orgModel.Organization, error) {
	id := uuid.New()
	slugBase := Slugify(name)
	slug := slugBase
	for i := 0; i < 5; i++ {
		if i > 0 {
			slug = slugBase + "-" + uuid.NewString()[:8]
		}
		var created time.Time
		err := r.db.QueryRow(ctx,
			`INSERT INTO organizations (id, name, slug, created_at)
			 VALUES ($1, $2, $3, now()) RETURNING created_at`,
			id, name, slug,
		).Scan(&created)
		if err == nil {
			return orgModel.Organization{ID: id.String(), Name: name, Slug: slug, CreatedAt: created.UTC()}, nil
		}
		if !isUniqueViolation(err) {
			return orgModel.Organization{}, err
		}
	}
	return orgModel.Organization{}, errors.New("could not create organization")
}

func (r *Repository) ListOrganizations(ctx context.Context) ([]orgModel.Organization, error) {
	rows, err := r.db.Query(ctx,
		`SELECT id, name, slug, created_at FROM organizations ORDER BY created_at DESC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []orgModel.Organization
	for rows.Next() {
		var o orgModel.Organization
		var id uuid.UUID
		if err := rows.Scan(&id, &o.Name, &o.Slug, &o.CreatedAt); err != nil {
			return nil, err
		}
		o.ID = id.String()
		out = append(out, o)
	}
	return out, rows.Err()
}

func (r *Repository) CreateInvite(ctx context.Context, orgID, role, email, createdBy string, days int) (orgModel.Invite, error) {
	if days <= 0 {
		days = 14
	}
	if role == "" {
		role = orgModel.MemberRoleMember
	}
	token, err := newToken()
	if err != nil {
		return orgModel.Invite{}, err
	}
	id := uuid.New()
	orgUUID, err := uuid.Parse(orgID)
	if err != nil {
		return orgModel.Invite{}, ErrNotFound
	}
	var orgName string
	if err := r.db.QueryRow(ctx, `SELECT name FROM organizations WHERE id = $1`, orgUUID).Scan(&orgName); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return orgModel.Invite{}, ErrNotFound
		}
		return orgModel.Invite{}, err
	}

	var createdByUUID *uuid.UUID
	if createdBy != "" {
		parsed, err := uuid.Parse(createdBy)
		if err == nil {
			createdByUUID = &parsed
		}
	}

	var inv orgModel.Invite
	var inviteID uuid.UUID
	err = r.db.QueryRow(ctx,
		`INSERT INTO organization_invites (id, org_id, token, email, role, created_by, expires_at, created_at)
		 VALUES ($1, $2, $3, NULLIF($4, ''), $5, $6, now() + make_interval(days => $7), now())
		 RETURNING id, expires_at, created_at`,
		id, orgUUID, token, normalizeInviteEmail(email), role, createdByUUID, days,
	).Scan(&inviteID, &inv.ExpiresAt, &inv.CreatedAt)
	if err != nil {
		return orgModel.Invite{}, err
	}
	inv.ID = inviteID.String()
	inv.OrgID = orgID
	inv.OrgName = orgName
	inv.Token = token
	inv.Email = normalizeInviteEmail(email)
	inv.Role = role
	return inv, nil
}

func (r *Repository) ListInvites(ctx context.Context, orgID string) ([]orgModel.Invite, error) {
	orgUUID, err := uuid.Parse(orgID)
	if err != nil {
		return nil, ErrNotFound
	}
	rows, err := r.db.Query(ctx,
		`SELECT i.id, i.org_id, o.name, i.token, COALESCE(i.email, ''), i.role, i.expires_at, i.used_at, i.created_at
		 FROM organization_invites i
		 JOIN organizations o ON o.id = i.org_id
		 WHERE i.org_id = $1
		 ORDER BY i.created_at DESC`,
		orgUUID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []orgModel.Invite
	for rows.Next() {
		var inv orgModel.Invite
		var id, oid uuid.UUID
		var usedAt *time.Time
		if err := rows.Scan(&id, &oid, &inv.OrgName, &inv.Token, &inv.Email, &inv.Role, &inv.ExpiresAt, &usedAt, &inv.CreatedAt); err != nil {
			return nil, err
		}
		inv.ID = id.String()
		inv.OrgID = oid.String()
		inv.UsedAt = usedAt
		out = append(out, inv)
	}
	return out, rows.Err()
}

func (r *Repository) GetInviteByToken(ctx context.Context, token string) (orgModel.Invite, error) {
	var inv orgModel.Invite
	var id, orgID uuid.UUID
	var usedAt *time.Time
	err := r.db.QueryRow(ctx,
		`SELECT i.id, i.org_id, o.name, i.token, COALESCE(i.email, ''), i.role, i.expires_at, i.used_at, i.created_at
		 FROM organization_invites i
		 JOIN organizations o ON o.id = i.org_id
		 WHERE i.token = $1`,
		token,
	).Scan(&id, &orgID, &inv.OrgName, &inv.Token, &inv.Email, &inv.Role, &inv.ExpiresAt, &usedAt, &inv.CreatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return orgModel.Invite{}, ErrNotFound
		}
		return orgModel.Invite{}, err
	}
	inv.ID = id.String()
	inv.OrgID = orgID.String()
	inv.UsedAt = usedAt
	return inv, nil
}

func (r *Repository) AcceptInvite(ctx context.Context, token, userID string) (orgModel.Member, error) {
	userUUID, err := uuid.Parse(userID)
	if err != nil {
		return orgModel.Member{}, errors.New("invalid user")
	}

	tx, err := r.db.Begin(ctx)
	if err != nil {
		return orgModel.Member{}, err
	}
	defer tx.Rollback(ctx)

	var existing int
	if err := tx.QueryRow(ctx,
		`SELECT COUNT(*) FROM organization_members WHERE user_id = $1`, userUUID,
	).Scan(&existing); err != nil {
		return orgModel.Member{}, err
	}
	if existing > 0 {
		return orgModel.Member{}, ErrAlreadyMember
	}

	var inv orgModel.Invite
	var inviteID, orgID uuid.UUID
	var slug string
	var usedAt *time.Time
	err = tx.QueryRow(ctx,
		`SELECT i.id, i.org_id, o.name, o.slug, i.token, COALESCE(i.email, ''), i.role, i.expires_at, i.used_at
		 FROM organization_invites i
		 JOIN organizations o ON o.id = i.org_id
		 WHERE i.token = $1 FOR UPDATE`,
		token,
	).Scan(&inviteID, &orgID, &inv.OrgName, &slug, &inv.Token, &inv.Email, &inv.Role, &inv.ExpiresAt, &usedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return orgModel.Member{}, ErrInviteInvalid
		}
		return orgModel.Member{}, err
	}
	now := time.Now().UTC()
	if usedAt != nil || now.After(inv.ExpiresAt) {
		return orgModel.Member{}, ErrInviteInvalid
	}

	if inv.Email != "" {
		var userEmail string
		if err := tx.QueryRow(ctx, `SELECT email FROM users WHERE id = $1`, userUUID).Scan(&userEmail); err != nil {
			return orgModel.Member{}, err
		}
		locked := strings.ToLower(strings.TrimSpace(inv.Email))
		account := strings.ToLower(strings.TrimSpace(userEmail))
		if account != locked {
			return orgModel.Member{}, errors.New("invite email does not match your account — sign in with " + locked + " or ask your admin for a new invite")
		}
	}

	var joined time.Time
	if err := tx.QueryRow(ctx,
		`INSERT INTO organization_members (org_id, user_id, role, joined_at)
		 VALUES ($1, $2, $3, now()) RETURNING joined_at`,
		orgID, userUUID, inv.Role,
	).Scan(&joined); err != nil {
		return orgModel.Member{}, err
	}

	if _, err := tx.Exec(ctx,
		`UPDATE organization_invites SET used_at = now(), used_by = $1 WHERE id = $2`,
		userUUID, inviteID,
	); err != nil {
		return orgModel.Member{}, err
	}

	if err := tx.Commit(ctx); err != nil {
		return orgModel.Member{}, err
	}

	return orgModel.Member{
		OrgID:    orgID.String(),
		OrgName:  inv.OrgName,
		OrgSlug:  slug,
		UserID:   userID,
		Role:     inv.Role,
		JoinedAt: joined.UTC(),
	}, nil
}

func (r *Repository) GetMemberByUserID(ctx context.Context, userID string) (*orgModel.Member, error) {
	userUUID, err := uuid.Parse(userID)
	if err != nil {
		return nil, nil
	}
	var m orgModel.Member
	var orgID uuid.UUID
	err = r.db.QueryRow(ctx,
		`SELECT m.org_id, o.name, o.slug, m.user_id, m.role, m.joined_at
		 FROM organization_members m
		 JOIN organizations o ON o.id = m.org_id
		 WHERE m.user_id = $1`,
		userUUID,
	).Scan(&orgID, &m.OrgName, &m.OrgSlug, &m.UserID, &m.Role, &m.JoinedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	m.OrgID = orgID.String()
	return &m, nil
}

func (r *Repository) ListMembers(ctx context.Context, orgID string) ([]orgModel.Member, error) {
	orgUUID, err := uuid.Parse(orgID)
	if err != nil {
		return nil, ErrNotFound
	}
	rows, err := r.db.Query(ctx,
		`SELECT m.user_id, u.email, u.username, m.role, m.joined_at
		 FROM organization_members m
		 JOIN users u ON u.id = m.user_id
		 WHERE m.org_id = $1
		 ORDER BY m.joined_at ASC`,
		orgUUID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []orgModel.Member
	for rows.Next() {
		var m orgModel.Member
		var uid uuid.UUID
		if err := rows.Scan(&uid, &m.Email, &m.Username, &m.Role, &m.JoinedAt); err != nil {
			return nil, err
		}
		m.UserID = uid.String()
		out = append(out, m)
	}
	return out, rows.Err()
}

func normalizeInviteEmail(email string) string {
	return strings.ToLower(strings.TrimSpace(email))
}

func isUniqueViolation(err error) bool {
	var pgErr *pgconn.PgError
	return errors.As(err, &pgErr) && pgErr.Code == "23505"
}
