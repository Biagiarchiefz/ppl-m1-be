package user_usecase

import (
	"context"

	"ppl-m1-be/internal/usecase/apperr"
	"ppl-m1-be/internal/usecase/audit_usecase"
	"ppl-m1-be/internal/usecase/repository/sqlc"
)

// Deactivate menonaktifkan akun tanpa menghapus fisik (soft delete), sehingga
// riwayat dan audit log tetap utuh (M1-US3).
func (u *Usecase) Deactivate(ctx context.Context, actor audit_usecase.Actor, id int64) error {
	if id == actor.UserID {
		return apperr.CannotDeactivateSelf
	}
	return u.store.InTx(ctx, func(q *sqlc.Queries) error {
		current, err := lockUser(ctx, q, actor.OrganizationID, id)
		if err != nil {
			return err
		}
		if !current.IsActive {
			return apperr.UserAlreadyInactive
		}
		if err := q.DeactivateUser(ctx, id); err != nil {
			return err
		}
		return audit_usecase.Record(ctx, q, actor.Change(actor.OrganizationID, audit_usecase.ActionUserDeactivated,
			audit_usecase.ResourceUsers, id, map[string]bool{"is_active": true}, map[string]bool{"is_active": false}))
	})
}

func lockUser(ctx context.Context, q *sqlc.Queries, orgID, id int64) (sqlc.User, error) {
	user, err := q.LockUserInOrg(ctx, sqlc.LockUserInOrgParams{ID: id, OrganizationID: audit_usecase.ID(orgID)})
	return user, apperr.NotFound(err, apperr.UserNotFound)
}
