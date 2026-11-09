package role_usecase

import (
	"context"
	"strings"

	"ppl-m1-be/internal/usecase/apperr"
	"ppl-m1-be/internal/usecase/audit_usecase"
	"ppl-m1-be/internal/usecase/repository"
	"ppl-m1-be/internal/usecase/repository/sqlc"
)

func (u *Usecase) Create(ctx context.Context, in CreateInput) (RoleView, error) {
	var view RoleView
	err := u.store.InTx(ctx, func(q *sqlc.Queries) error {
		perms, err := validatePermissions(ctx, q, in.Permissions)
		if err != nil {
			return err
		}
		role, err := q.CreateRole(ctx, sqlc.CreateRoleParams{
			OrganizationID: in.Actor.OrganizationID, Code: in.Code,
			Name: strings.TrimSpace(in.Name), Description: in.Description,
		})
		if repository.IsUniqueViolation(err) {
			return apperr.RoleCodeTaken
		}
		if err != nil {
			return err
		}
		view = RoleView{Role: role, Permissions: perms}
		return saveCreated(ctx, q, in.Actor, view)
	})
	return view, err
}

func saveCreated(ctx context.Context, q *sqlc.Queries, actor audit_usecase.Actor, view RoleView) error {
	if err := replacePermissions(ctx, q, view.Role.ID, view.Permissions); err != nil {
		return err
	}
	return audit_usecase.Record(ctx, q, actor.Change(actor.OrganizationID, audit_usecase.ActionRoleCreated,
		audit_usecase.ResourceRoles, view.Role.ID, nil, roleSnapshot(view.Role, view.Permissions)))
}

// Update mengubah nama, deskripsi, dan/atau izin. Peran bawaan dikunci.
func (u *Usecase) Update(ctx context.Context, in UpdateInput) (RoleView, error) {
	var view RoleView
	err := u.store.InTx(ctx, func(q *sqlc.Queries) error {
		current, err := lockEditableRole(ctx, q, in.Actor.OrganizationID, in.ID)
		if err != nil {
			return err
		}
		before, err := q.ListRolePermissionCodes(ctx, current.ID)
		if err != nil {
			return err
		}
		view, err = applyUpdate(ctx, q, current, in)
		if err != nil {
			return err
		}
		return audit_usecase.Record(ctx, q, in.Actor.Change(in.Actor.OrganizationID, audit_usecase.ActionRoleUpdated,
			audit_usecase.ResourceRoles, in.ID, roleSnapshot(current, before), roleSnapshot(view.Role, view.Permissions)))
	})
	return view, err
}

func lockEditableRole(ctx context.Context, q *sqlc.Queries, orgID, id int64) (sqlc.Role, error) {
	role, err := lockRole(ctx, q, orgID, id)
	if err == nil && role.IsSystem {
		return sqlc.Role{}, apperr.RoleSystemLocked
	}
	return role, err
}

func applyUpdate(ctx context.Context, q *sqlc.Queries, current sqlc.Role, in UpdateInput) (RoleView, error) {
	params := sqlc.UpdateRoleParams{ID: current.ID, Name: current.Name, Description: current.Description}
	if in.Name != nil {
		params.Name = strings.TrimSpace(*in.Name)
	}
	if in.Description != nil {
		params.Description = *in.Description
	}
	role, err := q.UpdateRole(ctx, params)
	if err != nil {
		return RoleView{}, err
	}
	if in.Permissions == nil {
		perms, err := q.ListRolePermissionCodes(ctx, role.ID)
		return RoleView{Role: role, Permissions: perms}, err
	}
	perms, err := validatePermissions(ctx, q, *in.Permissions)
	if err != nil {
		return RoleView{}, err
	}
	return RoleView{Role: role, Permissions: perms}, replacePermissions(ctx, q, role.ID, perms)
}

// Deactivate menonaktifkan peran yang tidak dipakai pengguna aktif.
func (u *Usecase) Deactivate(ctx context.Context, actor audit_usecase.Actor, id int64) error {
	return u.store.InTx(ctx, func(q *sqlc.Queries) error {
		role, err := lockRole(ctx, q, actor.OrganizationID, id)
		if err != nil {
			return err
		}
		if err := checkDeactivate(ctx, q, role); err != nil {
			return err
		}
		if _, err := q.DeactivateRole(ctx, id); err != nil {
			return err
		}
		return audit_usecase.Record(ctx, q, actor.Change(actor.OrganizationID, audit_usecase.ActionRoleDeactivated,
			audit_usecase.ResourceRoles, id, map[string]bool{"is_active": true}, map[string]bool{"is_active": false}))
	})
}

func checkDeactivate(ctx context.Context, q *sqlc.Queries, role sqlc.Role) error {
	switch {
	case role.IsSystem:
		return apperr.RoleSystemLocked
	case !role.IsActive:
		return apperr.RoleAlreadyInactive
	}
	inUse, err := q.CountActiveUsersWithRole(ctx, audit_usecase.ID(role.ID))
	if err == nil && inUse > 0 {
		return apperr.RoleInUse
	}
	return err
}
