package partner_usecase

import (
	"context"
	"database/sql"
	"slices"
	"strings"

	"ppl-m1-be/internal/usecase/apperr"
	"ppl-m1-be/internal/usecase/audit_usecase"
	"ppl-m1-be/internal/usecase/repository/sqlc"
)

// SetStatus menjalankan verifikasi, penolakan, penangguhan, atau pengaktifan
// kembali. Penolakan dan penangguhan wajib disertai alasan.
func (u *Usecase) SetStatus(ctx context.Context, actor audit_usecase.Actor, id int64, status, reason string) (sqlc.Partner, error) {
	reason = strings.TrimSpace(reason)
	if needsReason(status) && reason == "" {
		return sqlc.Partner{}, apperr.FieldErrors{"reason": "Alasan wajib diisi"}
	}
	var partner sqlc.Partner
	err := u.store.InTx(ctx, func(q *sqlc.Queries) error {
		current, err := q.LockPartner(ctx, sqlc.LockPartnerParams{ID: id, OrganizationID: actor.OrganizationID})
		if err != nil {
			return apperr.NotFound(err, apperr.PartnerNotFound)
		}
		if !slices.Contains(transitions[current.Status], status) {
			return apperr.PartnerStatusInvalid
		}
		if partner, err = applyStatus(ctx, q, actor, id, status, reason); err != nil {
			return err
		}
		return recordStatus(ctx, q, actor, current, partner)
	})
	return partner, err
}

// ResubmitInput adalah data usaha yang diperbaiki saat mengajukan ulang.
type ResubmitInput struct {
	BusinessName string
	BusinessType string
	ContactName  string
	Phone        string
	Address      string
	Description  sql.NullString
}

// Resubmit mengembalikan pendaftaran yang ditolak ke antrean verifikasi.
func (u *Usecase) Resubmit(ctx context.Context, actor audit_usecase.Actor, in ResubmitInput) (sqlc.Partner, error) {
	var partner sqlc.Partner
	err := u.store.InTx(ctx, func(q *sqlc.Queries) error {
		current, err := q.LockPartnerByUser(ctx, audit_usecase.ID(actor.UserID))
		if err != nil {
			return apperr.NotFound(err, apperr.PartnerNotFound)
		}
		if current.Status != StatusRejected {
			return apperr.PartnerStatusInvalid
		}
		if partner, err = resubmit(ctx, q, current.ID, in); err != nil {
			return err
		}
		return audit_usecase.Record(ctx, q, actor.Change(actor.OrganizationID, audit_usecase.ActionPartnerResubmitted,
			audit_usecase.ResourcePartners, partner.ID, map[string]string{"status": current.Status}, map[string]string{"status": partner.Status}))
	})
	return partner, err
}

func resubmit(ctx context.Context, q *sqlc.Queries, id int64, in ResubmitInput) (sqlc.Partner, error) {
	return q.ResubmitPartner(ctx, sqlc.ResubmitPartnerParams{
		ID: id, BusinessName: strings.TrimSpace(in.BusinessName), BusinessType: strings.TrimSpace(in.BusinessType),
		ContactName: strings.TrimSpace(in.ContactName), Phone: strings.TrimSpace(in.Phone),
		Address: strings.TrimSpace(in.Address), Description: in.Description,
	})
}

func needsReason(status string) bool {
	return status == StatusRejected || status == StatusSuspended
}

func applyStatus(ctx context.Context, q *sqlc.Queries, actor audit_usecase.Actor, id int64, status, reason string) (sqlc.Partner, error) {
	return q.UpdatePartnerStatus(ctx, sqlc.UpdatePartnerStatusParams{
		ID: id, OrganizationID: actor.OrganizationID, Status: status,
		StatusReason: sql.NullString{String: reason, Valid: needsReason(status)},
		ActorID:      audit_usecase.ID(actor.UserID),
	})
}

func recordStatus(ctx context.Context, q *sqlc.Queries, actor audit_usecase.Actor, before, after sqlc.Partner) error {
	return audit_usecase.Record(ctx, q, actor.Change(actor.OrganizationID, audit_usecase.ActionPartnerStatus,
		audit_usecase.ResourcePartners, after.ID, map[string]string{"status": before.Status},
		map[string]string{"status": after.Status, "reason": after.StatusReason.String}))
}
