package usecase

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/google/uuid"
	orgDTO "github.com/leventkok/mlc-llm-monitoring/masterfabric-go/internal/application/org/dto"
	orgModel "github.com/leventkok/mlc-llm-monitoring/masterfabric-go/internal/domain/org/model"
	iamModel "github.com/leventkok/mlc-llm-monitoring/masterfabric-go/internal/domain/iam/model"
	"github.com/leventkok/mlc-llm-monitoring/masterfabric-go/internal/domain/iam/repository"
	pgOrg "github.com/leventkok/mlc-llm-monitoring/masterfabric-go/internal/infrastructure/postgres/org"
)

type Service struct {
	repo  *pgOrg.Repository
	users repository.UserRepository
}

func NewService(repo *pgOrg.Repository) *Service {
	return &Service{repo: repo}
}

func (s *Service) WithUsers(users repository.UserRepository) *Service {
	s.users = users
	return s
}

func (s *Service) CreateOrganization(ctx context.Context, name string) (orgDTO.OrganizationResponse, error) {
	name = strings.TrimSpace(name)
	if name == "" {
		return orgDTO.OrganizationResponse{}, errors.New("organization name is required")
	}
	o, err := s.repo.CreateOrganization(ctx, name)
	if err != nil {
		return orgDTO.OrganizationResponse{}, errors.New("could not create organization")
	}
	return toOrgResponse(o), nil
}

func (s *Service) ListOrganizations(ctx context.Context) ([]orgDTO.OrganizationResponse, error) {
	list, err := s.repo.ListOrganizations(ctx)
	if err != nil {
		return nil, errors.New("could not list organizations")
	}
	out := make([]orgDTO.OrganizationResponse, 0, len(list))
	for _, o := range list {
		out = append(out, toOrgResponse(o))
	}
	return out, nil
}

func (s *Service) CreateInvite(ctx context.Context, orgID, createdBy string, req orgDTO.CreateInviteRequest) (orgDTO.InviteResponse, error) {
	role := strings.TrimSpace(req.Role)
	if role != orgModel.MemberRoleAdmin && role != orgModel.MemberRoleMember {
		return orgDTO.InviteResponse{}, errors.New("role must be company_admin or company_member")
	}
	days := req.Days
	if days <= 0 {
		days = 14
	}
	inv, err := s.repo.CreateInvite(ctx, orgID, role, req.Email, createdBy, days)
	if err != nil {
		if errors.Is(err, pgOrg.ErrNotFound) {
			return orgDTO.InviteResponse{}, errors.New("organization not found")
		}
		return orgDTO.InviteResponse{}, errors.New("could not create invite")
	}
	return toInviteResponse(inv), nil
}

func (s *Service) ListInvites(ctx context.Context, orgID string) ([]orgDTO.InviteResponse, error) {
	list, err := s.repo.ListInvites(ctx, orgID)
	if err != nil {
		return nil, errors.New("could not list invites")
	}
	out := make([]orgDTO.InviteResponse, 0, len(list))
	for _, inv := range list {
		out = append(out, toInviteResponse(inv))
	}
	return out, nil
}

func (s *Service) PreviewInvite(ctx context.Context, token string) (orgDTO.InvitePreviewResponse, error) {
	inv, err := s.repo.GetInviteByToken(ctx, token)
	if err != nil {
		return orgDTO.InvitePreviewResponse{Valid: false, Reason: "invite not found"}, nil
	}
	now := time.Now().UTC()
	if inv.IsUsed() {
		return orgDTO.InvitePreviewResponse{
			OrgName: inv.OrgName, Role: inv.Role, Email: inv.Email,
			ExpiresAt: inv.ExpiresAt.Format(time.RFC3339), Valid: false, Reason: "already used",
		}, nil
	}
	if inv.IsExpired(now) {
		return orgDTO.InvitePreviewResponse{
			OrgName: inv.OrgName, Role: inv.Role, Email: inv.Email,
			ExpiresAt: inv.ExpiresAt.Format(time.RFC3339), Valid: false, Reason: "expired",
		}, nil
	}
	return orgDTO.InvitePreviewResponse{
		OrgName:   inv.OrgName,
		Role:      inv.Role,
		Email:     inv.Email,
		ExpiresAt: inv.ExpiresAt.Format(time.RFC3339),
		Valid:     true,
	}, nil
}

func (s *Service) AcceptInvite(ctx context.Context, token, userID string) (orgDTO.AcceptInviteResponse, error) {
	m, err := s.repo.AcceptInvite(ctx, token, userID)
	if err != nil {
		switch {
		case errors.Is(err, pgOrg.ErrInviteInvalid):
			return orgDTO.AcceptInviteResponse{}, errors.New("invite invalid or expired")
		case errors.Is(err, pgOrg.ErrAlreadyMember):
			return orgDTO.AcceptInviteResponse{}, errors.New("you already belong to an organization")
		default:
			return orgDTO.AcceptInviteResponse{}, err
		}
	}
	if s.users != nil {
		if uid, parseErr := uuid.Parse(userID); parseErr == nil {
			_ = s.users.SetAccountKind(ctx, uid, iamModel.AccountKindCompany)
		}
	}
	return orgDTO.AcceptInviteResponse{
		Message: "joined organization",
		Organization: orgDTO.OrganizationSummary{
			ID: m.OrgID, Name: m.OrgName, Slug: m.OrgSlug, Role: m.Role,
		},
	}, nil
}

func (s *Service) ListMembersForUser(ctx context.Context, userID string) ([]orgDTO.MemberResponse, error) {
	m, err := s.repo.GetMemberByUserID(ctx, userID)
	if err != nil || m == nil {
		return nil, errors.New("organization membership required")
	}
	return s.ListMembers(ctx, m.OrgID)
}

func (s *Service) ListMembers(ctx context.Context, orgID string) ([]orgDTO.MemberResponse, error) {
	list, err := s.repo.ListMembers(ctx, orgID)
	if err != nil {
		return nil, errors.New("could not list members")
	}
	out := make([]orgDTO.MemberResponse, 0, len(list))
	for _, m := range list {
		out = append(out, orgDTO.MemberResponse{
			UserID:   m.UserID,
			Email:    m.Email,
			Username: m.Username,
			Role:     m.Role,
			JoinedAt: m.JoinedAt.Format(time.RFC3339),
		})
	}
	return out, nil
}

func (s *Service) CreateInviteForMember(ctx context.Context, userID string, req orgDTO.CreateInviteRequest) (orgDTO.InviteResponse, error) {
	m, err := s.repo.GetMemberByUserID(ctx, userID)
	if err != nil || m == nil {
		return orgDTO.InviteResponse{}, errors.New("organization membership required")
	}
	if m.Role != orgModel.MemberRoleAdmin {
		return orgDTO.InviteResponse{}, errors.New("company admin role required")
	}
	return s.CreateInvite(ctx, m.OrgID, userID, req)
}

func (s *Service) ListInvitesForMember(ctx context.Context, userID string) ([]orgDTO.InviteResponse, error) {
	m, err := s.repo.GetMemberByUserID(ctx, userID)
	if err != nil || m == nil {
		return nil, errors.New("organization membership required")
	}
	if m.Role != orgModel.MemberRoleAdmin {
		return nil, errors.New("company admin role required")
	}
	return s.ListInvites(ctx, m.OrgID)
}

func (s *Service) MembershipForUser(ctx context.Context, userID string) *orgDTO.OrganizationSummary {
	m, err := s.repo.GetMemberByUserID(ctx, userID)
	if err != nil || m == nil {
		return nil
	}
	return &orgDTO.OrganizationSummary{
		ID: m.OrgID, Name: m.OrgName, Slug: m.OrgSlug, Role: m.Role,
	}
}

func toOrgResponse(o orgModel.Organization) orgDTO.OrganizationResponse {
	return orgDTO.OrganizationResponse{
		ID: o.ID, Name: o.Name, Slug: o.Slug, CreatedAt: o.CreatedAt.Format(time.RFC3339),
	}
}

func toInviteResponse(inv orgModel.Invite) orgDTO.InviteResponse {
	resp := orgDTO.InviteResponse{
		ID:         inv.ID,
		OrgID:      inv.OrgID,
		OrgName:    inv.OrgName,
		Token:      inv.Token,
		Role:       inv.Role,
		Email:      inv.Email,
		ExpiresAt:  inv.ExpiresAt.Format(time.RFC3339),
		InvitePath: "/invite/" + inv.Token,
	}
	if inv.UsedAt != nil {
		resp.UsedAt = inv.UsedAt.Format(time.RFC3339)
	}
	return resp
}
