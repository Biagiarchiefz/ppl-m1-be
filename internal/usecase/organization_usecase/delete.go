package organization_usecase

import (
	"context"

	"ppl-m1-be/internal/usecase/apperr"
	"ppl-m1-be/internal/usecase/audit_usecase"
	"ppl-m1-be/internal/usecase/repository/sqlc"
)

// Delete menghapus permanen organisasi yang sudah dinonaktifkan beserta
// seluruh pengguna, peran, dan divisinya, dalam satu transaksi. Audit log
// organisasi itu tidak ikut terhapus: tetap tersimpan dengan salinan nama
// organisasi dan pelakunya, dan dapat dilihat super admin.
func (u *Usecase) Delete(ctx context.Context, actor audit_usecase.Actor, id int64) error {
	return u.store.InTx(ctx, func(q *sqlc.Queries) error {
		current, err := lockOrganization(ctx, q, id)
		if err != nil {
			return err
		}
		if !current.IsActive {
			return deleteOrganization(ctx, q, actor, current)
		}
		return apperr.OrganizationStillActive
	})
}

func deleteOrganization(ctx context.Context, q *sqlc.Queries, actor audit_usecase.Actor, org sqlc.Organization) error {
	users, err := q.CountOrganizationUsers(ctx, audit_usecase.ID(org.ID))
	if err != nil {
		return err
	}
	snapshot := map[string]any{"code": org.Code, "name": org.Name, "users": users}
	if err := audit_usecase.Record(ctx, q, actor.Change(org.ID, audit_usecase.ActionOrgDeleted,
		audit_usecase.ResourceOrganizations, org.ID, snapshot, nil)); err != nil {
		return err
	}
	steps := []func(context.Context, int64) error{
		q.DeleteOrganizationBusinessUnits,
		func(ctx context.Context, id int64) error { return q.DeleteOrganizationUsers(ctx, audit_usecase.ID(id)) },
		q.DeleteOrganizationRoles, q.DeleteOrganizationDivisions, q.DeleteOrganization,
	}
	for _, step := range steps {
		if err := step(ctx, org.ID); err != nil {
			return err
		}
	}
	return nil
}
