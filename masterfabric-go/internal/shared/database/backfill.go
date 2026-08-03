package database

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/leventkok/mlc-llm-monitoring/masterfabric-go/internal/shared/admin"
)

// BackfillRoles syncs platform admins from ADMIN_USERNAMES and attaches org_id to existing reviews.
func BackfillRoles(ctx context.Context, pool *pgxpool.Pool) error {
	for _, username := range admin.AdminUsernames() {
		if _, err := pool.Exec(ctx,
			`UPDATE users SET platform_role = 'platform_admin'
			 WHERE lower(username) = lower($1) AND platform_role <> 'platform_admin'`,
			username,
		); err != nil {
			return fmt.Errorf("backfill platform admin %q: %w", username, err)
		}
	}

	if _, err := pool.Exec(ctx,
		`UPDATE reviews r
		 SET org_id = m.org_id
		 FROM organization_members m
		 WHERE r.user_id = m.user_id AND r.org_id IS NULL`,
	); err != nil {
		return fmt.Errorf("backfill review org_id: %w", err)
	}

	if _, err := pool.Exec(ctx,
		`UPDATE users u
		 SET account_kind = 'company'
		 FROM organization_members m
		 WHERE u.id = m.user_id AND u.account_kind <> 'company'`,
	); err != nil {
		return fmt.Errorf("backfill account_kind: %w", err)
	}

	return nil
}
