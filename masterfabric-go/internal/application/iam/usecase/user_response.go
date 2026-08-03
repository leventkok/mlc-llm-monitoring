package usecase

import (
	"context"

	"github.com/leventkok/mlc-llm-monitoring/masterfabric-go/internal/application/iam/dto"
	orgUC "github.com/leventkok/mlc-llm-monitoring/masterfabric-go/internal/application/org/usecase"
	"github.com/leventkok/mlc-llm-monitoring/masterfabric-go/internal/domain/iam/model"
	"github.com/leventkok/mlc-llm-monitoring/masterfabric-go/internal/shared/admin"
)

func toUserResponse(user *model.User, org *orgUC.Service, ctx context.Context) dto.UserResponse {
	resp := dto.UserResponse{
		ID:           user.ID.String(),
		Email:        user.Email,
		Username:     user.Username,
		IsAdmin:      admin.IsPlatformAdmin(user),
		PlatformRole: user.PlatformRole,
		AccountKind:  user.AccountKind,
	}
	if resp.PlatformRole == "" {
		resp.PlatformRole = model.PlatformRoleUser
	}
	if resp.AccountKind == "" {
		resp.AccountKind = model.AccountKindIndividual
	}
	if org != nil {
		if m := org.MembershipForUser(ctx, user.ID.String()); m != nil {
			resp.AccountKind = model.AccountKindCompany
			resp.Organization = &dto.OrganizationSummary{
				ID: m.ID, Name: m.Name, Slug: m.Slug, Role: m.Role,
			}
		}
	}
	return resp
}
