package role_usecase

import (
	"context"
	"database/sql"
	"slices"
	"strings"

	"ppl-m1-be/internal/usecase/apperr"
	"ppl-m1-be/internal/usecase/repository"
	"ppl-m1-be/internal/usecase/repository/sqlc"
)

type Usecase struct {
	store *repository.Store
}

func New(store *repository.Store) *Usecase {
	return &Usecase{store: store}
}

// Permissions mengembalikan katalog izin yang boleh dipilih organisasi
// (tanpa izin platform).
func (u *Usecase) Permissions(ctx context.Context) ([]sqlc.Permission, error) {
	return u.store.ListPermissions(ctx)
}

// permissionsOf: peran bawaan selalu memiliki semua izin organisasi.
func permissionsOf(ctx context.Context, q *sqlc.Queries, role sqlc.Role) ([]string, error) {
	if role.IsSystem {
		return q.ListOrganizationPermissionCodes(ctx)
	}
	return q.ListRolePermissionCodes(ctx, role.ID)
}

// validatePermissions membuang duplikat dan memastikan semua kode ada di
// katalog izin organisasi.
func validatePermissions(ctx context.Context, q *sqlc.Queries, codes []string) ([]string, error) {
	catalog, err := q.ListOrganizationPermissionCodes(ctx)
	if err != nil {
		return nil, err
	}
	clean := make([]string, 0, len(codes))
	var unknown []string
	for _, code := range codes {
		switch {
		case !slices.Contains(catalog, code):
			unknown = append(unknown, code)
		case !slices.Contains(clean, code):
			clean = append(clean, code)
		}
	}
	if len(unknown) > 0 {
		return nil, apperr.FieldErrors{"permissions": "Izin tidak terdaftar: " + strings.Join(unknown, ", ")}
	}
	slices.Sort(clean)
	return clean, nil
}

func replacePermissions(ctx context.Context, q *sqlc.Queries, roleID int64, codes []string) error {
	if err := q.DeleteRolePermissions(ctx, roleID); err != nil {
		return err
	}
	for _, code := range codes {
		if err := q.AddRolePermission(ctx, sqlc.AddRolePermissionParams{RoleID: roleID, PermissionCode: code}); err != nil {
			return err
		}
	}
	return nil
}

func lockRole(ctx context.Context, q *sqlc.Queries, orgID, id int64) (sqlc.Role, error) {
	role, err := q.LockRoleInOrg(ctx, sqlc.LockRoleInOrgParams{ID: id, OrganizationID: orgID})
	return role, apperr.NotFound(err, apperr.RoleNotFound)
}

func roleSnapshot(role sqlc.Role, permissions []string) map[string]any {
	return map[string]any{
		"code": role.Code, "name": role.Name,
		"description": nullable(role.Description), "permissions": permissions,
	}
}

func nullable(s sql.NullString) any {
	if !s.Valid {
		return nil
	}
	return s.String
}
