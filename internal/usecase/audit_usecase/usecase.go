package audit_usecase

import (
	"context"

	"ppl-m1-be/internal/usecase/repository"
	"ppl-m1-be/internal/usecase/repository/sqlc"
)

type Usecase struct {
	store *repository.Store
}

func New(store *repository.Store) *Usecase {
	return &Usecase{store: store}
}

// List menampilkan audit log satu organisasi, atau semua organisasi bila
// f.AllOrganizations (halaman audit platform super admin).
func (u *Usecase) List(ctx context.Context, f ListFilter, limit, offset int32) (ListResult, error) {
	items, err := u.store.ListAuditLogs(ctx, sqlc.ListAuditLogsParams{
		AllOrganizations: f.AllOrganizations, OrganizationID: f.OrganizationID, FromTime: f.From, ToTime: f.To,
		UserID: f.UserID, Action: f.Action, Resource: f.Resource, Sort: f.Sort, LimitRows: limit, OffsetRows: offset,
	})
	if err != nil {
		return ListResult{}, err
	}
	total, err := u.store.CountAuditLogs(ctx, sqlc.CountAuditLogsParams{
		AllOrganizations: f.AllOrganizations, OrganizationID: f.OrganizationID, FromTime: f.From, ToTime: f.To,
		UserID: f.UserID, Action: f.Action, Resource: f.Resource,
	})
	return ListResult{Items: items, Total: total}, err
}
