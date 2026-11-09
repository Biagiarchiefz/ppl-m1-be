package role_usecase

import (
	"context"
	"fmt"
	"strings"

	"ppl-m1-be/internal/usecase/repository/sqlc"
	"ppl-m1-be/internal/usecase/utils"
)

// ProvisionDefaults menyiapkan peran partner, customer, dan manajer untuk
// satu organisasi. Aman dijalankan berulang.
func ProvisionDefaults(ctx context.Context, q *sqlc.Queries, orgID int64) error {
	for _, role := range utils.DefaultRoles {
		id, err := q.EnsureRole(ctx, sqlc.EnsureRoleParams{OrganizationID: orgID, Code: role.Code, Name: role.Name})
		if err != nil {
			return fmt.Errorf("siapkan peran %s: %w", role.Code, err)
		}
		err = q.GrantDefaultPermissions(ctx, sqlc.GrantDefaultPermissionsParams{RoleID: id, Codes: strings.Join(role.Permissions, ",")})
		if err != nil {
			return fmt.Errorf("beri izin peran %s: %w", role.Code, err)
		}
	}
	return nil
}
