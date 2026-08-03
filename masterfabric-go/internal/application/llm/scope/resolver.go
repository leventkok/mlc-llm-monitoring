package scope

import (
	"context"

	orgUC "github.com/leventkok/mlc-llm-monitoring/masterfabric-go/internal/application/org/usecase"
	"github.com/leventkok/mlc-llm-monitoring/masterfabric-go/internal/domain/llm/model"
)

// Resolver builds review access scope from the authenticated user.
type Resolver struct {
	org *orgUC.Service
}

func NewResolver(org *orgUC.Service) *Resolver {
	return &Resolver{org: org}
}

func (r *Resolver) ForUser(ctx context.Context, userID string) model.ReviewScope {
	scope := model.ReviewScope{UserID: userID}
	if r.org == nil {
		return scope
	}
	if m := r.org.MembershipForUser(ctx, userID); m != nil {
		scope.OrgID = m.ID
	}
	return scope
}
