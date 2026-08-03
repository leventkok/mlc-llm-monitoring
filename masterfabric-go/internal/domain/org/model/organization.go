package model

import "time"

const (
	MemberRoleAdmin  = "company_admin"
	MemberRoleMember = "company_member"
)

type Organization struct {
	ID        string
	Name      string
	Slug      string
	CreatedAt time.Time
}

type Member struct {
	OrgID    string
	OrgName  string
	OrgSlug  string
	UserID   string
	Email    string
	Username string
	Role     string
	JoinedAt time.Time
}

type Invite struct {
	ID        string
	OrgID     string
	OrgName   string
	Token     string
	Email     string
	Role      string
	ExpiresAt time.Time
	UsedAt    *time.Time
	CreatedAt time.Time
}

func (i Invite) IsExpired(now time.Time) bool {
	return now.After(i.ExpiresAt)
}

func (i Invite) IsUsed() bool {
	return i.UsedAt != nil
}
