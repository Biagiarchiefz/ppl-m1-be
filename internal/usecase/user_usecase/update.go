package user_usecase

import (
	"context"
	"database/sql"
	"strings"

	"ppl-m1-be/internal/usecase/audit_usecase"
	"ppl-m1-be/internal/usecase/repository/sqlc"
)

// userState adalah nilai yang bisa diubah lewat PATCH, dalam bentuk id
// (untuk disimpan) dan kode (untuk audit log).
type userState struct {
	Name         string
	RoleID       int64
	RoleCode     string
	DivisionID   sql.NullInt64
	DivisionCode sql.NullString
	UnitUsahaID  sql.NullInt64
}

// Update mengubah name/role/division/unit_usaha_id. Hanya field yang dikirim
// yang berubah; nilai baru masuk ke token pada login berikutnya.
func (u *Usecase) Update(ctx context.Context, in UpdateInput) (sqlc.UserDetail, error) {
	var user sqlc.UserDetail
	err := u.store.InTx(ctx, func(q *sqlc.Queries) (err error) {
		user, err = applyUpdate(ctx, q, in)
		return err
	})
	return user, err
}

func applyUpdate(ctx context.Context, q *sqlc.Queries, in UpdateInput) (sqlc.UserDetail, error) {
	org := in.Actor.OrganizationID
	current, err := lockDetail(ctx, q, org, in.ID)
	if err != nil {
		return sqlc.UserDetail{}, err
	}
	next, err := nextState(ctx, q, org, stateOf(current), in)
	if err != nil {
		return sqlc.UserDetail{}, err
	}
	oldValue, newValue := diffState(stateOf(current), next)
	if len(newValue) == 0 {
		return current, nil
	}
	return saveUpdate(ctx, q, in, next, oldValue, newValue)
}

func saveUpdate(ctx context.Context, q *sqlc.Queries, in UpdateInput, next userState, oldValue, newValue map[string]any) (sqlc.UserDetail, error) {
	org := in.Actor.OrganizationID
	if err := q.UpdateUser(ctx, toUpdateParams(in.ID, next)); err != nil {
		return sqlc.UserDetail{}, err
	}
	err := audit_usecase.Record(ctx, q, in.Actor.Change(org, audit_usecase.ActionUserUpdated,
		audit_usecase.ResourceUsers, in.ID, oldValue, newValue))
	if err != nil {
		return sqlc.UserDetail{}, err
	}
	return getDetail(ctx, q, org, in.ID)
}

// lockDetail mengunci baris pengguna lalu membaca kode peran/divisinya.
func lockDetail(ctx context.Context, q *sqlc.Queries, orgID, id int64) (sqlc.UserDetail, error) {
	if _, err := lockUser(ctx, q, orgID, id); err != nil {
		return sqlc.UserDetail{}, err
	}
	return getDetail(ctx, q, orgID, id)
}

func stateOf(u sqlc.UserDetail) userState {
	return userState{
		Name: u.Name, RoleID: u.RoleID.Int64, RoleCode: u.RoleCode.String,
		DivisionID: u.DivisionID, DivisionCode: u.DivisionCode, UnitUsahaID: u.UnitUsahaID,
	}
}

func nextState(ctx context.Context, q *sqlc.Queries, orgID int64, s userState, in UpdateInput) (userState, error) {
	ids, err := resolveRefs(ctx, q, orgID, in.Role.Value, in.Division.Value)
	if err != nil {
		return userState{}, err
	}
	if in.Name.Set {
		s.Name = strings.TrimSpace(in.Name.Value)
	}
	if in.Role.Set {
		s.RoleID, s.RoleCode = ids.RoleID, in.Role.Value
	}
	if in.Division.Set {
		s.DivisionID, s.DivisionCode = ids.DivisionID, in.Division.Value
	}
	if in.UnitUsahaID.Set {
		s.UnitUsahaID = in.UnitUsahaID.Value
	}
	return s, nil
}

// diffState hanya memuat field yang benar-benar berubah.
func diffState(cur, next userState) (map[string]any, map[string]any) {
	oldValue, newValue := map[string]any{}, map[string]any{}
	add := func(field string, changed bool, before, after any) {
		if changed {
			oldValue[field], newValue[field] = before, after
		}
	}
	add("name", cur.Name != next.Name, cur.Name, next.Name)
	add("role", cur.RoleCode != next.RoleCode, cur.RoleCode, next.RoleCode)
	add("division", cur.DivisionCode != next.DivisionCode, nullableString(cur.DivisionCode), nullableString(next.DivisionCode))
	add("unit_usaha_id", cur.UnitUsahaID != next.UnitUsahaID, nullableInt(cur.UnitUsahaID), nullableInt(next.UnitUsahaID))
	return oldValue, newValue
}

func toUpdateParams(id int64, s userState) sqlc.UpdateUserParams {
	return sqlc.UpdateUserParams{
		ID: id, Name: s.Name, RoleID: audit_usecase.ID(s.RoleID),
		DivisionID: s.DivisionID, UnitUsahaID: s.UnitUsahaID,
	}
}
