package role_usecase

import (
	"context"

	"ppl-m1-be/internal/usecase/apperr"
	"ppl-m1-be/internal/usecase/audit_usecase"
	"ppl-m1-be/internal/usecase/repository/sqlc"
)

// SetStatus mengaktifkan atau menonaktifkan peran. Izin peran tidak berubah,
// jadi peran aktif kembali dengan izin terakhirnya.
func (u *Usecase) SetStatus(ctx context.Context, actor audit_usecase.Actor, id int64, active bool) error {
	if !active {
		return u.Deactivate(ctx, actor, id)
	}
	return u.store.InTx(ctx, func(q *sqlc.Queries) error {
		role, err := lockRole(ctx, q, actor.OrganizationID, id)
		if err != nil {
			return err
		}
		if role.IsActive {
			return apperr.RoleAlreadyActive
		}
		if err := q.ActivateRole(ctx, id); err != nil {
			return err
		}
		return audit_usecase.Record(ctx, q, actor.Change(actor.OrganizationID, audit_usecase.ActionRoleActivated,
			audit_usecase.ResourceRoles, id, map[string]bool{"is_active": false}, map[string]bool{"is_active": true}))
	})
}
