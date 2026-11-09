package user_usecase

import (
	"context"

	"ppl-m1-be/internal/usecase/apperr"
	"ppl-m1-be/internal/usecase/audit_usecase"
	"ppl-m1-be/internal/usecase/repository/sqlc"
)

// Delete menghapus permanen akun yang sudah dinonaktifkan. Audit log tetap
// tersimpan: data terakhir akun dicatat di old_value, dan log yang dibuat
// akun ini tetap menampilkan salinan nama/email pelakunya.
func (u *Usecase) Delete(ctx context.Context, actor audit_usecase.Actor, id int64) error {
	return u.store.InTx(ctx, func(q *sqlc.Queries) error {
		current, err := lockUser(ctx, q, actor.OrganizationID, id)
		if err != nil {
			return err
		}
		if current.IsActive {
			return apperr.UserStillActive
		}
		detail, err := q.GetUserDetailInOrg(ctx, sqlc.GetUserDetailInOrgParams{ID: id, OrganizationID: current.OrganizationID})
		if err != nil {
			return err
		}
		if err := audit_usecase.Record(ctx, q, actor.Change(actor.OrganizationID, audit_usecase.ActionUserDeleted,
			audit_usecase.ResourceUsers, id, userSnapshot(detail), nil)); err != nil {
			return err
		}
		return q.DeleteUser(ctx, id)
	})
}
