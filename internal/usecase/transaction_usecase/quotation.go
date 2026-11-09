package transaction_usecase

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"

	"ppl-m1-be/internal/usecase/apperr"
	"ppl-m1-be/internal/usecase/audit_usecase"
	"ppl-m1-be/internal/usecase/repository/sqlc"
	"ppl-m1-be/internal/usecase/utils"
)

const (
	QuotationWaiting  = "menunggu"
	QuotationApproved = "disetujui"
	QuotationRejected = "ditolak"

	ResponseByCustomer = "customer"
	ResponseOffline    = "offline"
)

type ItemInput struct {
	ProductID int64
	Qty       int32
}

type QuotationInput struct {
	CustomerID int64
	Notes      sql.NullString
	Items      []ItemInput
}

type QuotationView struct {
	Quotation sqlc.QuotationDetail
	Items     []sqlc.QuotationItem
}

// line adalah salinan produk saat transaksi.
type line struct {
	ProductID      sql.NullInt64
	BusinessUnitID sql.NullInt64
	Name           string
	Price          int64
	Qty            int32
}

func (l line) subtotal() int64 {
	return l.Price * int64(l.Qty)
}

func (u *Usecase) ListQuotations(ctx context.Context, id Identity, status sql.NullString, limit, offset int32) (Page[sqlc.QuotationDetail], error) {
	scope, err := u.scopeFor(ctx, id, utils.PermQuotationsManage)
	if err != nil {
		return Page[sqlc.QuotationDetail]{}, err
	}
	items, err := u.store.ListQuotations(ctx, sqlc.ListQuotationsParams{OrganizationID: scope.OrgID(), PartnerID: scope.PartnerID, CustomerID: scope.CustomerID, Status: status, LimitRows: limit, OffsetRows: offset})
	if err != nil {
		return Page[sqlc.QuotationDetail]{}, err
	}
	total, err := u.store.CountQuotations(ctx, sqlc.CountQuotationsParams{OrganizationID: scope.OrgID(), PartnerID: scope.PartnerID, CustomerID: scope.CustomerID, Status: status})
	return Page[sqlc.QuotationDetail]{Items: items, Total: total}, err
}

func (u *Usecase) GetQuotation(ctx context.Context, id Identity, quotationID int64) (QuotationView, error) {
	scope, err := u.scopeFor(ctx, id, utils.PermQuotationsManage)
	if err != nil {
		return QuotationView{}, err
	}
	quotation, err := u.store.GetQuotation(ctx, sqlc.GetQuotationParams{ID: quotationID, OrganizationID: scope.OrgID()})
	if err != nil {
		return QuotationView{}, apperr.NotFound(err, apperr.QuotationNotFound)
	}
	if !scope.Owns(quotation.PartnerID, quotation.CustomerID) {
		return QuotationView{}, apperr.QuotationNotFound
	}
	items, err := u.store.ListQuotationItems(ctx, quotationID)
	return QuotationView{Quotation: quotation, Items: items}, err
}

func (u *Usecase) CreateQuotation(ctx context.Context, id Identity, in QuotationInput) (QuotationView, error) {
	var quotationID int64
	err := u.store.InTx(ctx, func(q *sqlc.Queries) error {
		scope, err := sellerScope(ctx, q, id, utils.PermQuotationsManage)
		if err != nil {
			return err
		}
		if err := checkCustomer(ctx, q, scope, in.CustomerID); err != nil {
			return err
		}
		lines, err := quotationLines(ctx, q, scope, in.Items)
		if err != nil {
			return err
		}
		quotationID, err = insertQuotation(ctx, q, scope, in, lines)
		return err
	})
	if err != nil {
		return QuotationView{}, err
	}
	return u.GetQuotation(ctx, id, quotationID)
}

func checkCustomer(ctx context.Context, q *sqlc.Queries, scope Scope, customerID int64) error {
	unknown := apperr.FieldErrors{"customer_id": "Pelanggan tidak terdaftar"}
	_, err := q.GetCustomer(ctx, sqlc.GetCustomerParams{ID: customerID, OrganizationID: scope.OrgID()})
	if errors.Is(err, sql.ErrNoRows) {
		return unknown
	}
	if err != nil {
		return err
	}
	return requireLinked(ctx, q, scope, customerID, unknown)
}

// quotationLines mengambil harga dari database, bukan dari klien. Partner
// hanya boleh menawarkan produknya sendiri.
func quotationLines(ctx context.Context, q *sqlc.Queries, scope Scope, items []ItemInput) ([]line, error) {
	lines := make([]line, 0, len(items))
	for i, item := range items {
		p, err := q.GetProduct(ctx, sqlc.GetProductParams{ID: item.ProductID, OrganizationID: scope.OrgID()})
		if errors.Is(err, sql.ErrNoRows) || (err == nil && (!p.IsActive || !scope.Owns(p.PartnerID, 0))) {
			return nil, apperr.FieldErrors{fmt.Sprintf("items.%d.product_id", i): "Produk tidak tersedia"}
		}
		if err != nil {
			return nil, err
		}
		lines = append(lines, line{audit_usecase.ID(p.ID), p.BusinessUnitID, p.Name, p.Price, item.Qty})
	}
	return lines, nil
}

func totalOf(lines []line) int64 {
	var total int64
	for _, l := range lines {
		total += l.subtotal()
	}
	return total
}

func insertQuotation(ctx context.Context, q *sqlc.Queries, scope Scope, in QuotationInput, lines []line) (int64, error) {
	total := totalOf(lines)
	id, err := q.CreateQuotation(ctx, sqlc.CreateQuotationParams{
		OrganizationID: scope.OrgID(), PartnerID: scope.PartnerID, CustomerID: in.CustomerID,
		Notes: in.Notes, Total: total, CreatedBy: audit_usecase.ID(scope.Actor.UserID),
	})
	if err != nil {
		return 0, err
	}
	for _, l := range lines {
		err := q.CreateQuotationItem(ctx, sqlc.CreateQuotationItemParams{
			QuotationID: id, ProductID: l.ProductID, BusinessUnitID: l.BusinessUnitID,
			ProductName: l.Name, Price: l.Price, Qty: l.Qty, Subtotal: l.subtotal(),
		})
		if err != nil {
			return 0, err
		}
	}
	return id, audit_usecase.Record(ctx, q, scope.Actor.Change(scope.OrgID(), audit_usecase.ActionQuotationCreated,
		audit_usecase.ResourceQuotations, id, nil, map[string]int64{"customer_id": in.CustomerID, "total": total}))
}

// RespondQuotation menyetujui atau menolak quotation yang masih menunggu.
// Customer menanggapi langsung; penjual hanya boleh mencatat tanggapan yang
// diterima di luar sistem dan wajib menyertakan catatan buktinya.
func (u *Usecase) RespondQuotation(ctx context.Context, id Identity, quotationID int64, status, note string) (QuotationView, error) {
	err := u.store.InTx(ctx, func(q *sqlc.Queries) error {
		return respondQuotation(ctx, q, id, quotationID, status, strings.TrimSpace(note))
	})
	if err != nil {
		return QuotationView{}, err
	}
	return u.GetQuotation(ctx, id, quotationID)
}

func respondQuotation(ctx context.Context, q *sqlc.Queries, id Identity, quotationID int64, status, note string) error {
	scope, err := resolveScope(ctx, q, id, utils.PermQuotationsManage)
	if err != nil {
		return err
	}
	source, err := responseSource(scope, note)
	if err != nil {
		return err
	}
	current, err := lockOwnQuotation(ctx, q, scope, quotationID)
	if err != nil {
		return err
	}
	if current.Status != QuotationWaiting {
		return apperr.QuotationStatusInvalid
	}
	return saveResponse(ctx, q, scope, current, status, source, note)
}

func responseSource(scope Scope, note string) (string, error) {
	if scope.CustomerID.Valid {
		return ResponseByCustomer, nil
	}
	if err := scope.requireRespondable(); err != nil {
		return "", err
	}
	if note == "" {
		return "", apperr.FieldErrors{"note": "Catatan bukti persetujuan wajib diisi"}
	}
	return ResponseOffline, nil
}

func saveResponse(ctx context.Context, q *sqlc.Queries, scope Scope, current sqlc.Quotation, status, source, note string) error {
	err := q.UpdateQuotationStatus(ctx, sqlc.UpdateQuotationStatusParams{
		ID: current.ID, OrganizationID: scope.OrgID(), Status: status,
		ResponseSource: sql.NullString{String: source, Valid: true}, ResponseNote: sql.NullString{String: note, Valid: note != ""},
	})
	if err != nil {
		return err
	}
	return audit_usecase.Record(ctx, q, scope.Actor.Change(scope.OrgID(), audit_usecase.ActionQuotationStatus,
		audit_usecase.ResourceQuotations, current.ID, map[string]string{"status": current.Status},
		map[string]string{"status": status, "source": source, "note": note}))
}

func lockOwnQuotation(ctx context.Context, q *sqlc.Queries, scope Scope, quotationID int64) (sqlc.Quotation, error) {
	quotation, err := q.LockQuotation(ctx, sqlc.LockQuotationParams{ID: quotationID, OrganizationID: scope.OrgID()})
	if err != nil {
		return quotation, apperr.NotFound(err, apperr.QuotationNotFound)
	}
	if !scope.Owns(quotation.PartnerID, quotation.CustomerID) {
		return quotation, apperr.QuotationNotFound
	}
	return quotation, nil
}
