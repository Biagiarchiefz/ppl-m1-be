// Package transaction_usecase adalah modul M4: produk/jasa, pelanggan,
// quotation, invoice, pembayaran, dan order.
package transaction_usecase

import (
	"context"
	"database/sql"
	"errors"
	"slices"

	"ppl-m1-be/internal/usecase/apperr"
	"ppl-m1-be/internal/usecase/audit_usecase"
	"ppl-m1-be/internal/usecase/partner_usecase"
	"ppl-m1-be/internal/usecase/repository"
	"ppl-m1-be/internal/usecase/repository/sqlc"
	"ppl-m1-be/internal/usecase/utils"
)

const sourceModule = "m4"

type Usecase struct {
	store *repository.Store
}

func New(store *repository.Store) *Usecase {
	return &Usecase{store: store}
}

// Identity adalah pemanggil menurut token.
type Identity struct {
	Actor       audit_usecase.Actor
	Role        string
	Permissions []string
	UnitUsahaID *int64
}

// orderUnit mengembalikan unit usaha yang membatasi Order pemanggil. Staf unit
// tanpa unit mendapat id 0, yang tidak cocok dengan Order mana pun.
func (id Identity) orderUnit() sql.NullInt64 {
	switch {
	case id.Role == utils.SystemRoleCode:
		return sql.NullInt64{}
	case id.UnitUsahaID != nil:
		return sql.NullInt64{Int64: *id.UnitUsahaID, Valid: true}
	case id.Role == utils.RoleUnitStaff:
		return sql.NullInt64{Valid: true}
	}
	return sql.NullInt64{}
}

// Page adalah satu halaman data beserta total keseluruhannya.
type Page[T any] struct {
	Items []T
	Total int64
}

// Scope membatasi data yang boleh dilihat pemanggil: partner hanya miliknya,
// customer hanya transaksinya, pengelola organisasi seluruh data organisasi.
type Scope struct {
	Actor      audit_usecase.Actor
	PartnerID  sql.NullInt64
	CustomerID sql.NullInt64
	// BusinessUnitID membatasi staf unit usaha ke Order unitnya.
	BusinessUnitID sql.NullInt64
	verified       bool
}

func (s Scope) ownsOrder(partnerID, unitID sql.NullInt64, customerID int64) bool {
	if s.BusinessUnitID.Valid {
		return unitID == s.BusinessUnitID
	}
	return s.Owns(partnerID, customerID)
}

func (s Scope) OrgID() int64 {
	return s.Actor.OrganizationID
}

// Owns memeriksa apakah data dengan pemilik tersebut masuk cakupan pemanggil.
func (s Scope) Owns(partnerID sql.NullInt64, customerID int64) bool {
	switch {
	case s.PartnerID.Valid:
		return partnerID == s.PartnerID
	case s.CustomerID.Valid:
		return customerID == s.CustomerID.Int64
	}
	return true
}

// RequireSeller memastikan pemanggil boleh mengelola transaksi: pengelola
// organisasi atau partner yang berstatus terverifikasi.
func (s Scope) RequireSeller() error {
	if s.CustomerID.Valid {
		return apperr.Forbidden
	}
	if s.PartnerID.Valid && !s.verified {
		return apperr.PartnerNotVerified
	}
	return nil
}

// requireRespondable menolak partner yang ditangguhkan atau belum
// terverifikasi mencatat tanggapan quotation.
func (s Scope) requireRespondable() error {
	if s.PartnerID.Valid && !s.verified {
		return apperr.PartnerNotVerified
	}
	return nil
}

func resolveScope(ctx context.Context, q *sqlc.Queries, id Identity, managePerm string) (Scope, error) {
	scope := Scope{Actor: id.Actor}
	switch {
	case id.Role == utils.RolePartner:
		return partnerScope(ctx, q, scope)
	case slices.Contains(id.Permissions, managePerm):
		return scope, nil
	}
	return customerScope(ctx, q, scope)
}

func resolveOrderScope(ctx context.Context, q *sqlc.Queries, id Identity) (Scope, error) {
	scope, err := resolveScope(ctx, q, id, utils.PermOrdersUpdate)
	if err != nil {
		return scope, err
	}
	if !scope.PartnerID.Valid && !scope.CustomerID.Valid {
		scope.BusinessUnitID = id.orderUnit()
	}
	return scope, nil
}

func partnerScope(ctx context.Context, q *sqlc.Queries, scope Scope) (Scope, error) {
	partner, err := q.GetPartnerByUser(ctx, audit_usecase.ID(scope.Actor.UserID))
	if errors.Is(err, sql.ErrNoRows) {
		return scope, apperr.PartnerNotVerified
	}
	if err != nil {
		return scope, err
	}
	scope.PartnerID = audit_usecase.ID(partner.ID)
	scope.verified = partner.Status == partner_usecase.StatusVerified
	return scope, nil
}

// customerScope tidak menulis apa pun. Akun tanpa data pelanggan mendapat
// CustomerID 0, yang tidak cocok dengan transaksi mana pun.
func customerScope(ctx context.Context, q *sqlc.Queries, scope Scope) (Scope, error) {
	customer, err := q.GetCustomerForUser(ctx, scope.Actor.UserID)
	if errors.Is(err, sql.ErrNoRows) {
		scope.CustomerID = sql.NullInt64{Valid: true}
		return scope, nil
	}
	if err != nil {
		return scope, err
	}
	scope.CustomerID = sql.NullInt64{Int64: customer.ID, Valid: true}
	return scope, nil
}

// scopeFor menghitung cakupan di luar transaksi, untuk endpoint baca.
func (u *Usecase) scopeFor(ctx context.Context, id Identity, managePerm string) (Scope, error) {
	return resolveScope(ctx, u.store.Queries, id, managePerm)
}
