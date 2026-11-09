package role_usecase

import (
	"context"
	"database/sql"

	"ppl-m1-be/internal/usecase/apperr"
	"ppl-m1-be/internal/usecase/audit_usecase"
	"ppl-m1-be/internal/usecase/repository/sqlc"
)

func (u *Usecase) Get(ctx context.Context, actor audit_usecase.Actor, id int64) (RoleView, error) {
	role, err := u.store.GetRoleInOrg(ctx, sqlc.GetRoleInOrgParams{ID: id, OrganizationID: actor.OrganizationID})
	if err != nil {
		return RoleView{}, apperr.NotFound(err, apperr.RoleNotFound)
	}
	permissions, err := permissionsOf(ctx, u.store.Queries, role)
	return RoleView{Role: role, Permissions: permissions}, err
}

func (u *Usecase) List(ctx context.Context, actor audit_usecase.Actor, isActive sql.NullBool, limit, offset int32) (ListResult, error) {
	org := actor.OrganizationID
	roles, err := u.store.ListRoles(ctx, sqlc.ListRolesParams{OrganizationID: org, IsActive: isActive, LimitRows: limit, OffsetRows: offset})
	if err != nil {
		return ListResult{}, err
	}
	items, err := u.withPermissions(ctx, org, roles)
	if err != nil {
		return ListResult{}, err
	}
	total, err := u.store.CountRoles(ctx, sqlc.CountRolesParams{OrganizationID: org, IsActive: isActive})
	return ListResult{Items: items, Total: total}, err
}

// withPermissions memasangkan izin ke tiap peran dengan dua query saja.
func (u *Usecase) withPermissions(ctx context.Context, orgID int64, roles []sqlc.Role) ([]RoleView, error) {
	all, err := u.store.ListOrganizationPermissionCodes(ctx)
	if err != nil {
		return nil, err
	}
	pairs, err := u.store.ListOrganizationRolePermissions(ctx, orgID)
	if err != nil {
		return nil, err
	}
	byRole := groupByRole(pairs)
	items := make([]RoleView, 0, len(roles))
	for _, r := range roles {
		perms := byRole[r.ID]
		if r.IsSystem {
			perms = all
		}
		items = append(items, RoleView{Role: r, Permissions: perms})
	}
	return items, nil
}

func groupByRole(pairs []sqlc.RolePermission) map[int64][]string {
	byRole := map[int64][]string{}
	for _, p := range pairs {
		byRole[p.RoleID] = append(byRole[p.RoleID], p.PermissionCode)
	}
	return byRole
}
