package division_usecase

import (
	"context"

	"ppl-m1-be/internal/usecase/apperr"
	"ppl-m1-be/internal/usecase/audit_usecase"
	"ppl-m1-be/internal/usecase/repository/sqlc"
)

// SetStatus mengaktifkan atau menonaktifkan divisi.
func (u *Usecase) SetStatus(ctx context.Context, actor audit_usecase.Actor, id int64, active bool) error {
	if !active {
		return u.Deactivate(ctx, actor, id)
	}
	return u.store.InTx(ctx, func(q *sqlc.Queries) error {
		current, err := lockDivision(ctx, q, actor.OrganizationID, id)
		if err != nil {
			return err
		}
		if current.IsActive {
			return apperr.DivisionAlreadyActive
		}
		if err := q.ActivateDivision(ctx, id); err != nil {
			return err
		}
		return audit_usecase.Record(ctx, q, actor.Change(actor.OrganizationID, audit_usecase.ActionDivisionActivated,
			audit_usecase.ResourceDivisions, id, map[string]bool{"is_active": false}, map[string]bool{"is_active": true}))
	})
}
