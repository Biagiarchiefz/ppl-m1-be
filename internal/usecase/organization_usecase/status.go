package organization_usecase

import (
	"context"

	"ppl-m1-be/internal/usecase/apperr"
	"ppl-m1-be/internal/usecase/audit_usecase"
	"ppl-m1-be/internal/usecase/repository/sqlc"
)

// SetStatus mengaktifkan atau menonaktifkan organisasi. Penonaktifan tidak
// mengubah data di dalamnya, jadi setelah aktif kembali penggunanya bisa login
// seperti semula.
func (u *Usecase) SetStatus(ctx context.Context, actor audit_usecase.Actor, id int64, active bool) error {
	if !active {
		return u.Deactivate(ctx, actor, id)
	}
	return u.store.InTx(ctx, func(q *sqlc.Queries) error {
		current, err := lockOrganization(ctx, q, id)
		if err != nil {
			return err
		}
		if current.IsActive {
			return apperr.OrganizationAlreadyActive
		}
		if err := q.ActivateOrganization(ctx, id); err != nil {
			return err
		}
		return audit_usecase.Record(ctx, q, actor.Change(id, audit_usecase.ActionOrgActivated,
			audit_usecase.ResourceOrganizations, id, map[string]bool{"is_active": false}, map[string]bool{"is_active": true}))
	})
}
