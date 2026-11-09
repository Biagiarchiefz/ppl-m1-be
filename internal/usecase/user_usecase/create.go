package user_usecase

import (
	"context"
	"strings"

	"ppl-m1-be/internal/usecase/apperr"
	"ppl-m1-be/internal/usecase/audit_usecase"
	"ppl-m1-be/internal/usecase/repository"
	"ppl-m1-be/internal/usecase/repository/sqlc"
	"ppl-m1-be/internal/usecase/utils"
)

// Create menambah akun aktif di organisasi administrator (M1-US3).
func (u *Usecase) Create(ctx context.Context, in CreateInput) (sqlc.UserDetail, error) {
	hash, err := utils.HashPassword(in.Password)
	if err != nil {
		return sqlc.UserDetail{}, err
	}
	var user sqlc.UserDetail
	err = u.store.InTx(ctx, func(q *sqlc.Queries) error {
		ids, err := resolveRefs(ctx, q, in.Actor.OrganizationID, in.Role, in.Division)
		if err != nil {
			return err
		}
		if user, err = insertUser(ctx, q, in, ids, hash); err != nil {
			return err
		}
		return audit_usecase.Record(ctx, q, in.Actor.Change(in.Actor.OrganizationID,
			audit_usecase.ActionUserCreated, audit_usecase.ResourceUsers, user.ID, nil, userSnapshot(user)))
	})
	return user, err
}

func insertUser(ctx context.Context, q *sqlc.Queries, in CreateInput, ids refs, hash string) (sqlc.UserDetail, error) {
	org := in.Actor.OrganizationID
	id, err := q.CreateUser(ctx, sqlc.CreateUserParams{
		OrganizationID: audit_usecase.ID(org),
		Name:           strings.TrimSpace(in.Name),
		Email:          strings.ToLower(strings.TrimSpace(in.Email)),
		PasswordHash:   hash,
		RoleID:         audit_usecase.ID(ids.RoleID),
		DivisionID:     ids.DivisionID,
		UnitUsahaID:    in.UnitUsahaID,
		CreatedBy:      audit_usecase.ID(in.Actor.UserID),
	})
	if repository.IsUniqueViolation(err) {
		return sqlc.UserDetail{}, apperr.EmailTaken
	}
	if err != nil {
		return sqlc.UserDetail{}, err
	}
	return getDetail(ctx, q, org, id)
}
