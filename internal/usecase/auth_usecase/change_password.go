package auth_usecase

import (
	"context"
	"time"

	"ppl-m1-be/internal/usecase/apperr"
	"ppl-m1-be/internal/usecase/audit_usecase"
	"ppl-m1-be/internal/usecase/repository/sqlc"
	"ppl-m1-be/internal/usecase/utils"
)

// ChangePassword mengganti kata sandi pengguna yang sedang login. Aturan
// kata sandi baru diperiksa lebih dulu oleh handler.
func (u *Usecase) ChangePassword(ctx context.Context, in ChangePasswordInput) (time.Time, error) {
	newHash, err := utils.HashPassword(in.NewPassword)
	if err != nil {
		return time.Time{}, err
	}
	var changedAt time.Time
	err = u.store.InTx(ctx, func(q *sqlc.Queries) error {
		if err := verifyOldPassword(ctx, q, in); err != nil {
			return err
		}
		changedAt, err = q.UpdatePassword(ctx, sqlc.UpdatePasswordParams{ID: in.Actor.UserID, PasswordHash: newHash})
		if err != nil {
			return err
		}
		return audit_usecase.Record(ctx, q, in.Actor.Change(in.Actor.OrganizationID,
			audit_usecase.ActionUserPasswordChanged, audit_usecase.ResourceUsers, in.Actor.UserID, nil, nil))
	})
	return changedAt, err
}

func verifyOldPassword(ctx context.Context, q *sqlc.Queries, in ChangePasswordInput) error {
	user, err := q.LockUserByID(ctx, in.Actor.UserID)
	switch {
	case err != nil:
		return apperr.NotFound(err, apperr.Unauthenticated)
	case !user.IsActive:
		return apperr.AccountInactive
	case utils.CheckPassword(user.PasswordHash, in.OldPassword) != nil:
		return apperr.OldPasswordInvalid
	}
	return nil
}
