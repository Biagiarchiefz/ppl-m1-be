package auth_usecase

import (
	"context"

	"ppl-m1-be/internal/usecase/repository/sqlc"
	"ppl-m1-be/internal/usecase/utils"
)

// permissionsOf menentukan izin yang masuk ke token:
//   - super admin: hanya izin platform;
//   - peran bawaan administrator: semua izin non-platform (termasuk izin
//     modul lain yang ditambahkan kemudian);
//   - peran lain: izin yang dipilih administrator organisasi.
func permissionsOf(ctx context.Context, q *sqlc.Queries, user sqlc.UserDetail) ([]string, error) {
	switch {
	case user.IsSuperAdmin:
		return []string{utils.PermOrganizationsManage}, nil
	case user.RoleIsSystem.Bool:
		return q.ListOrganizationPermissionCodes(ctx)
	}
	return q.ListRolePermissionCodes(ctx, user.RoleID.Int64)
}

func roleOf(user sqlc.UserDetail) string {
	if user.IsSuperAdmin {
		return utils.RoleSuperAdmin
	}
	return user.RoleCode.String
}

// accessError memeriksa status akun dan organisasinya.
func accessError(user sqlc.UserDetail) string {
	switch {
	case !user.IsActive:
		return reasonInactive
	case !user.IsSuperAdmin && !user.OrganizationIsActive.Bool:
		return reasonOrgInactive
	}
	return ""
}

func toClaims(user sqlc.UserDetail, permissions []string) utils.Claims {
	claims := utils.Claims{UserID: user.ID, Name: user.Name, Role: roleOf(user), Permissions: permissions}
	if user.OrganizationID.Valid {
		claims.OrganizationID = &user.OrganizationID.Int64
	}
	if user.DivisionCode.Valid {
		claims.Division = &user.DivisionCode.String
	}
	if user.UnitUsahaID.Valid {
		claims.UnitUsahaID = &user.UnitUsahaID.Int64
	}
	if claims.Permissions == nil {
		claims.Permissions = []string{}
	}
	return claims
}
