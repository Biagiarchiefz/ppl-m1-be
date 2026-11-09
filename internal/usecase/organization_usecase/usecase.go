// Package organization_usecase dipakai super admin platform untuk
// mendaftarkan dan mengelola organisasi (tenant).
package organization_usecase

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

type ListFilter struct {
	Q        sql.NullString
	IsActive sql.NullBool
}

type ListResult struct {
	Items []sqlc.Organization
	Total int64
}

func (u *Usecase) List(ctx context.Context, f ListFilter, limit, offset int32) (ListResult, error) {
	items, err := u.store.ListOrganizations(ctx, sqlc.ListOrganizationsParams{Q: f.Q, IsActive: f.IsActive, LimitRows: limit, OffsetRows: offset})
	if err != nil {
		return ListResult{}, err
	}
	total, err := u.store.CountOrganizations(ctx, sqlc.CountOrganizationsParams{Q: f.Q, IsActive: f.IsActive})
	return ListResult{Items: items, Total: total}, err
}

func (u *Usecase) Get(ctx context.Context, id int64) (sqlc.Organization, error) {
	org, err := u.store.GetOrganization(ctx, id)
	return org, apperr.NotFound(err, apperr.OrganizationNotFound)
}

func (u *Usecase) Rename(ctx context.Context, actor audit_usecase.Actor, id int64, name string) (sqlc.Organization, error) {
	var org sqlc.Organization
	err := u.store.InTx(ctx, func(q *sqlc.Queries) error {
		current, err := lockOrganization(ctx, q, id)
		if err != nil {
			return err
		}
		if org, err = q.UpdateOrganization(ctx, sqlc.UpdateOrganizationParams{ID: id, Name: strings.TrimSpace(name)}); err != nil {
			return err
		}
		if org.Name == current.Name {
			return nil
		}
		return audit_usecase.Record(ctx, q, actor.Change(id, audit_usecase.ActionOrgUpdated,
			audit_usecase.ResourceOrganizations, id, map[string]string{"name": current.Name}, map[string]string{"name": org.Name}))
	})
	return org, err
}

// Deactivate menonaktifkan organisasi: semua penggunanya tidak bisa login
// lagi (ORGANIZATION_INACTIVE), tetapi datanya tetap tersimpan.
func (u *Usecase) Deactivate(ctx context.Context, actor audit_usecase.Actor, id int64) error {
	return u.store.InTx(ctx, func(q *sqlc.Queries) error {
		current, err := lockOrganization(ctx, q, id)
		if err != nil {
			return err
		}
		if !current.IsActive {
			return apperr.OrganizationAlreadyInactive
		}
		if _, err := q.DeactivateOrganization(ctx, id); err != nil {
			return err
		}
		return audit_usecase.Record(ctx, q, actor.Change(id, audit_usecase.ActionOrgDeactivated,
			audit_usecase.ResourceOrganizations, id, map[string]bool{"is_active": true}, map[string]bool{"is_active": false}))
	})
}

func lockOrganization(ctx context.Context, q *sqlc.Queries, id int64) (sqlc.Organization, error) {
	org, err := q.LockOrganization(ctx, id)
	return org, apperr.NotFound(err, apperr.OrganizationNotFound)
}
