package user_usecase

import (
	"context"
	"database/sql"

	"ppl-m1-be/internal/usecase/apperr"
	"ppl-m1-be/internal/usecase/audit_usecase"
	"ppl-m1-be/internal/usecase/repository"
	"ppl-m1-be/internal/usecase/repository/sqlc"
)

type Usecase struct {
	store *repository.Store
}

func New(store *repository.Store) *Usecase {
	return &Usecase{store: store}
}

func (u *Usecase) Get(ctx context.Context, actor audit_usecase.Actor, id int64) (sqlc.UserDetail, error) {
	return getDetail(ctx, u.store.Queries, actor.OrganizationID, id)
}

func getDetail(ctx context.Context, q *sqlc.Queries, orgID, id int64) (sqlc.UserDetail, error) {
	user, err := q.GetUserDetailInOrg(ctx, sqlc.GetUserDetailInOrgParams{ID: id, OrganizationID: audit_usecase.ID(orgID)})
	return user, apperr.NotFound(err, apperr.UserNotFound)
}

func (u *Usecase) List(ctx context.Context, actor audit_usecase.Actor, f ListFilter, limit, offset int32) (ListResult, error) {
	org := audit_usecase.ID(actor.OrganizationID)
	items, err := u.store.ListUsers(ctx, sqlc.ListUsersParams{
		OrganizationID: org, Q: f.Q, Role: f.Role, Division: f.Division, UnitUsahaID: f.UnitUsahaID,
		IsActive: f.IsActive, Sort: f.Sort, LimitRows: limit, OffsetRows: offset,
	})
	if err != nil {
		return ListResult{}, err
	}
	total, err := u.store.CountUsers(ctx, sqlc.CountUsersParams{
		OrganizationID: org, Q: f.Q, Role: f.Role, Division: f.Division,
		UnitUsahaID: f.UnitUsahaID, IsActive: f.IsActive,
	})
	return ListResult{Items: items, Total: total}, err
}

// userSnapshot adalah isi akun yang aman dicatat di audit log
// (tanpa password_hash).
func userSnapshot(user sqlc.UserDetail) map[string]any {
	return map[string]any{
		"name":          user.Name,
		"email":         user.Email,
		"role":          user.RoleCode.String,
		"division":      nullableString(user.DivisionCode),
		"unit_usaha_id": nullableInt(user.UnitUsahaID),
		"is_active":     user.IsActive,
	}
}

func nullableString(s sql.NullString) any {
	if !s.Valid {
		return nil
	}
	return s.String
}

func nullableInt(n sql.NullInt64) any {
	if !n.Valid {
		return nil
	}
	return n.Int64
}
