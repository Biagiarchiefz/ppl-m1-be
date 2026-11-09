// Package business_unit_usecase adalah modul M2: data unit usaha dan
// Dashboard Operasional Dasar.
package business_unit_usecase

import (
	"context"
	"database/sql"
	"strings"

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

// Input adalah isian profil unit usaha. Code hanya dipakai saat membuat.
type Input struct {
	DivisionID  sql.NullInt64
	Code        string
	Name        string
	Category    string
	Description sql.NullString
	Address     sql.NullString
	Phone       sql.NullString
	Email       sql.NullString
}

type ListFilter struct {
	IsActive sql.NullBool
	Q        sql.NullString
}

type ListResult struct {
	Items []sqlc.BusinessUnitDetail
	Total int64
}

func (u *Usecase) List(ctx context.Context, orgID int64, f ListFilter, limit, offset int32) (ListResult, error) {
	items, err := u.store.ListBusinessUnits(ctx, sqlc.ListBusinessUnitsParams{
		OrganizationID: orgID, IsActive: f.IsActive, Q: f.Q, LimitRows: limit, OffsetRows: offset,
	})
	if err != nil {
		return ListResult{}, err
	}
	total, err := u.store.CountBusinessUnits(ctx, sqlc.CountBusinessUnitsParams{OrganizationID: orgID, IsActive: f.IsActive, Q: f.Q})
	return ListResult{Items: items, Total: total}, err
}

func (u *Usecase) Get(ctx context.Context, orgID, id int64) (sqlc.BusinessUnitDetail, error) {
	unit, err := u.store.GetBusinessUnit(ctx, sqlc.GetBusinessUnitParams{ID: id, OrganizationID: orgID})
	return unit, apperr.NotFound(err, apperr.BusinessUnitNotFound)
}

func (u *Usecase) Create(ctx context.Context, actor audit_usecase.Actor, in Input) (sqlc.BusinessUnitDetail, error) {
	var id int64
	err := u.store.InTx(ctx, func(q *sqlc.Queries) error {
		if err := checkDivision(ctx, q, actor.OrganizationID, in.DivisionID); err != nil {
			return err
		}
		var err error
		if id, err = insertUnit(ctx, q, actor.OrganizationID, in); err != nil {
			return err
		}
		return audit_usecase.Record(ctx, q, actor.Change(actor.OrganizationID, audit_usecase.ActionBusinessUnitCreated,
			audit_usecase.ResourceBusinessUnits, id, nil, map[string]string{"code": in.Code, "name": in.Name}))
	})
	if err != nil {
		return sqlc.BusinessUnitDetail{}, err
	}
	return u.Get(ctx, actor.OrganizationID, id)
}

func insertUnit(ctx context.Context, q *sqlc.Queries, orgID int64, in Input) (int64, error) {
	id, err := q.CreateBusinessUnit(ctx, sqlc.CreateBusinessUnitParams{
		OrganizationID: orgID, DivisionID: in.DivisionID, Code: in.Code, Name: strings.TrimSpace(in.Name),
		Category: strings.TrimSpace(in.Category), Description: in.Description, Address: in.Address, Phone: in.Phone, Email: in.Email,
	})
	if repository.IsUniqueViolation(err) {
		return 0, apperr.BusinessUnitCodeTaken
	}
	return id, err
}

func (u *Usecase) Update(ctx context.Context, actor audit_usecase.Actor, id int64, in Input) (sqlc.BusinessUnitDetail, error) {
	err := u.store.InTx(ctx, func(q *sqlc.Queries) error {
		current, err := lock(ctx, q, actor.OrganizationID, id)
		if err != nil {
			return err
		}
		if err := checkDivision(ctx, q, actor.OrganizationID, in.DivisionID); err != nil {
			return err
		}
		if err := update(ctx, q, actor.OrganizationID, id, in); err != nil {
			return err
		}
		return audit_usecase.Record(ctx, q, actor.Change(actor.OrganizationID, audit_usecase.ActionBusinessUnitUpdated,
			audit_usecase.ResourceBusinessUnits, id, map[string]string{"name": current.Name, "category": current.Category},
			map[string]string{"name": in.Name, "category": in.Category}))
	})
	if err != nil {
		return sqlc.BusinessUnitDetail{}, err
	}
	return u.Get(ctx, actor.OrganizationID, id)
}

func update(ctx context.Context, q *sqlc.Queries, orgID, id int64, in Input) error {
	return q.UpdateBusinessUnit(ctx, sqlc.UpdateBusinessUnitParams{
		ID: id, OrganizationID: orgID, DivisionID: in.DivisionID, Name: strings.TrimSpace(in.Name),
		Category: strings.TrimSpace(in.Category), Description: in.Description, Address: in.Address, Phone: in.Phone, Email: in.Email,
	})
}

// SetStatus mengubah status aktif tanpa menyentuh data lain, sehingga unit
// yang diaktifkan kembali tampil dengan data semula.
func (u *Usecase) SetStatus(ctx context.Context, actor audit_usecase.Actor, id int64, active bool) error {
	return u.store.InTx(ctx, func(q *sqlc.Queries) error {
		current, err := lock(ctx, q, actor.OrganizationID, id)
		if err != nil {
			return err
		}
		if current.IsActive == active {
			return apperr.BusinessUnitStatusUnchanged
		}
		err = q.SetBusinessUnitActive(ctx, sqlc.SetBusinessUnitActiveParams{ID: id, OrganizationID: actor.OrganizationID, IsActive: active})
		if err != nil {
			return err
		}
		return audit_usecase.Record(ctx, q, actor.Change(actor.OrganizationID, audit_usecase.ActionBusinessUnitStatus,
			audit_usecase.ResourceBusinessUnits, id, map[string]bool{"is_active": current.IsActive}, map[string]bool{"is_active": active}))
	})
}

func lock(ctx context.Context, q *sqlc.Queries, orgID, id int64) (sqlc.BusinessUnit, error) {
	unit, err := q.LockBusinessUnit(ctx, sqlc.LockBusinessUnitParams{ID: id, OrganizationID: orgID})
	return unit, apperr.NotFound(err, apperr.BusinessUnitNotFound)
}

func checkDivision(ctx context.Context, q *sqlc.Queries, orgID int64, division sql.NullInt64) error {
	if !division.Valid {
		return nil
	}
	exists, err := q.ActiveDivisionExists(ctx, sqlc.ActiveDivisionExistsParams{ID: division.Int64, OrganizationID: orgID})
	if err != nil {
		return err
	}
	if !exists {
		return apperr.FieldErrors{"division_id": "Divisi tidak terdaftar"}
	}
	return nil
}
