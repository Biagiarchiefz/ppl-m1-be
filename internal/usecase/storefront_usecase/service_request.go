package storefront_usecase

import (
	"context"
	"database/sql"

	"ppl-m1-be/internal/usecase/apperr"
	"ppl-m1-be/internal/usecase/audit_usecase"
	"ppl-m1-be/internal/usecase/repository/sqlc"
)

// nextServiceStatus adalah urutan tindak lanjut permintaan layanan.
var nextServiceStatus = map[string]string{"baru": "diproses", "diproses": "selesai"}

// SetServiceRequestStatus memajukan permintaan satu langkah. `unit` membatasi
// staf ke permintaan unit usahanya.
func (u *Usecase) SetServiceRequestStatus(ctx context.Context, actor audit_usecase.Actor, unit sql.NullInt64, id int64, status string) error {
	return u.store.InTx(ctx, func(q *sqlc.Queries) error {
		current, err := q.LockServiceRequest(ctx, sqlc.LockServiceRequestParams{ID: id, OrganizationID: actor.OrganizationID})
		if err != nil {
			return apperr.NotFound(err, apperr.ServiceRequestNotFound)
		}
		if unit.Valid && unit.Int64 != current.BusinessUnitID {
			return apperr.ServiceRequestNotFound
		}
		if nextServiceStatus[current.Status] != status {
			return apperr.ServiceRequestStatusInvalid
		}
		err = q.UpdateServiceRequestStatus(ctx, sqlc.UpdateServiceRequestStatusParams{ID: id, OrganizationID: actor.OrganizationID, Status: status})
		if err != nil {
			return err
		}
		return audit_usecase.Record(ctx, q, actor.Change(actor.OrganizationID, audit_usecase.ActionServiceRequestStatus,
			audit_usecase.ResourceServiceRequests, id, map[string]string{"status": current.Status}, map[string]string{"status": status}))
	})
}

// MyServiceRequests adalah riwayat permintaan milik customer yang login.
func (u *Usecase) MyServiceRequests(ctx context.Context, userID int64, limit, offset int32) (ServiceRequestPage, error) {
	id := audit_usecase.ID(userID)
	items, err := u.store.ListMyServiceRequests(ctx, sqlc.ListMyServiceRequestsParams{UserID: id, LimitRows: limit, OffsetRows: offset})
	if err != nil {
		return ServiceRequestPage{}, err
	}
	total, err := u.store.CountMyServiceRequests(ctx, id)
	return ServiceRequestPage{Items: items, Total: total}, err
}
