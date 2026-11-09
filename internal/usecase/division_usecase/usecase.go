package division_usecase

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

type ListResult struct {
	Items []sqlc.Division
	Total int64
}

func (u *Usecase) List(ctx context.Context, actor audit_usecase.Actor, isActive sql.NullBool, limit, offset int32) (ListResult, error) {
	org := actor.OrganizationID
	items, err := u.store.ListDivisions(ctx, sqlc.ListDivisionsParams{OrganizationID: org, IsActive: isActive, LimitRows: limit, OffsetRows: offset})
	if err != nil {
		return ListResult{}, err
	}
	total, err := u.store.CountDivisions(ctx, sqlc.CountDivisionsParams{OrganizationID: org, IsActive: isActive})
	return ListResult{Items: items, Total: total}, err
}

func (u *Usecase) Get(ctx context.Context, actor audit_usecase.Actor, id int64) (sqlc.Division, error) {
	d, err := u.store.GetDivisionInOrg(ctx, sqlc.GetDivisionInOrgParams{ID: id, OrganizationID: actor.OrganizationID})
	return d, apperr.NotFound(err, apperr.DivisionNotFound)
}

func (u *Usecase) Create(ctx context.Context, actor audit_usecase.Actor, code, name string) (sqlc.Division, error) {
	var d sqlc.Division
	err := u.store.InTx(ctx, func(q *sqlc.Queries) (err error) {
		d, err = q.CreateDivision(ctx, sqlc.CreateDivisionParams{
			OrganizationID: actor.OrganizationID, Code: code, Name: strings.TrimSpace(name),
		})
		if repository.IsUniqueViolation(err) {
			return apperr.DivisionCodeTaken
		}
		if err != nil {
			return err
		}
		return audit_usecase.Record(ctx, q, actor.Change(actor.OrganizationID, audit_usecase.ActionDivisionCreated,
			audit_usecase.ResourceDivisions, d.ID, nil, map[string]string{"code": d.Code, "name": d.Name}))
	})
	return d, err
}

func (u *Usecase) Rename(ctx context.Context, actor audit_usecase.Actor, id int64, name string) (sqlc.Division, error) {
	var d sqlc.Division
	err := u.store.InTx(ctx, func(q *sqlc.Queries) error {
		current, err := lockDivision(ctx, q, actor.OrganizationID, id)
		if err != nil {
			return err
		}
		if d, err = q.UpdateDivision(ctx, sqlc.UpdateDivisionParams{ID: id, Name: strings.TrimSpace(name)}); err != nil {
			return err
		}
		if d.Name == current.Name {
			return nil
		}
		return audit_usecase.Record(ctx, q, actor.Change(actor.OrganizationID, audit_usecase.ActionDivisionUpdated,
			audit_usecase.ResourceDivisions, id, map[string]string{"name": current.Name}, map[string]string{"name": d.Name}))
	})
	return d, err
}

// Deactivate menonaktifkan divisi yang tidak dipakai pengguna aktif.
func (u *Usecase) Deactivate(ctx context.Context, actor audit_usecase.Actor, id int64) error {
	return u.store.InTx(ctx, func(q *sqlc.Queries) error {
		current, err := lockDivision(ctx, q, actor.OrganizationID, id)
		if err != nil {
			return err
		}
		if err := checkDeactivate(ctx, q, current); err != nil {
			return err
		}
		if _, err := q.DeactivateDivision(ctx, id); err != nil {
			return err
		}
		return audit_usecase.Record(ctx, q, actor.Change(actor.OrganizationID, audit_usecase.ActionDivisionDeactivated,
			audit_usecase.ResourceDivisions, id, map[string]bool{"is_active": true}, map[string]bool{"is_active": false}))
	})
}

func checkDeactivate(ctx context.Context, q *sqlc.Queries, d sqlc.Division) error {
	if !d.IsActive {
		return apperr.DivisionAlreadyInactive
	}
	inUse, err := q.CountActiveUsersWithDivision(ctx, audit_usecase.ID(d.ID))
	if err == nil && inUse > 0 {
		return apperr.DivisionInUse
	}
	return err
}

func lockDivision(ctx context.Context, q *sqlc.Queries, orgID, id int64) (sqlc.Division, error) {
	d, err := q.LockDivisionInOrg(ctx, sqlc.LockDivisionInOrgParams{ID: id, OrganizationID: orgID})
	return d, apperr.NotFound(err, apperr.DivisionNotFound)
}
