package dto

type OrganizationResponse struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	Slug      string `json:"slug"`
	CreatedAt string `json:"created_at"`
}

type OrganizationSummary struct {
	ID   string `json:"id"`
	Name string `json:"name"`
	Slug string `json:"slug"`
	Role string `json:"role,omitempty"`
}

type CreateOrganizationRequest struct {
	Name string `json:"name"`
}

type CreateInviteRequest struct {
	Role  string `json:"role"`
	Email string `json:"email,omitempty"`
	Days  int    `json:"days,omitempty"`
}

type InviteResponse struct {
	ID         string `json:"id"`
	OrgID      string `json:"org_id"`
	OrgName    string `json:"org_name"`
	Token      string `json:"token"`
	Role       string `json:"role"`
	Email      string `json:"email,omitempty"`
	ExpiresAt  string `json:"expires_at"`
	UsedAt     string `json:"used_at,omitempty"`
	InvitePath string `json:"invite_path"`
}

type InvitePreviewResponse struct {
	OrgName   string `json:"org_name"`
	Role      string `json:"role"`
	Email     string `json:"email,omitempty"`
	ExpiresAt string `json:"expires_at"`
	Valid     bool   `json:"valid"`
	Reason    string `json:"reason,omitempty"`
}

type AcceptInviteResponse struct {
	Message      string              `json:"message"`
	Organization OrganizationSummary `json:"organization"`
}

type MemberResponse struct {
	UserID   string `json:"user_id"`
	Email    string `json:"email"`
	Username string `json:"username"`
	Role     string `json:"role"`
	JoinedAt string `json:"joined_at"`
}
