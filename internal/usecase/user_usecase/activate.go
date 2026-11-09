package user_usecase

import (
	"context"
	"errors"

	"ppl-m1-be/internal/usecase/apperr"
	"ppl-m1-be/internal/usecase/audit_usecase"
	"ppl-m1-be/internal/usecase/repository/sqlc"
)

// SetStatus mengaktifkan atau menonaktifkan akun.
func (u *Usecase) SetStatus(ctx context.Context, actor audit_usecase.Actor, id int64, active bool) error {
	if active {
		return u.activate(ctx, actor, id)
	}
	return u.Deactivate(ctx, actor, id)
}

func (u *Usecase) activate(ctx context.Context, actor audit_usecase.Actor, id int64) error {
	return u.store.InTx(ctx, func(q *sqlc.Queries) error {
		current, err := lockDetail(ctx, q, actor.OrganizationID, id)
		if err != nil {
			return err
		}
		if current.IsActive {
			return apperr.UserAlreadyActive
		}
		if err := requireActiveRefs(ctx, q, actor.OrganizationID, current); err != nil {
			return err
		}
		if err := q.ActivateUser(ctx, id); err != nil {
			return err
		}
		return audit_usecase.Record(ctx, q, actor.Change(actor.OrganizationID, audit_usecase.ActionUserActivated,
			audit_usecase.ResourceUsers, id, map[string]bool{"is_active": false}, map[string]bool{"is_active": true}))
	})
}

// requireActiveRefs menjaga aturan bahwa pengguna aktif hanya memakai peran
// dan divisi yang aktif. Peran/divisi bisa saja dinonaktifkan selama akun ini
// nonaktif, dan harus diganti dulu lewat PATCH /users/{id}.
func requireActiveRefs(ctx context.Context, q *sqlc.Queries, orgID int64, user sqlc.UserDetail) error {
	_, err := resolveRefs(ctx, q, orgID, user.RoleCode.String, user.DivisionCode)
	var fields apperr.FieldErrors
	if errors.As(err, &fields) {
		return apperr.UserReferenceInactive
	}
	return err
}
