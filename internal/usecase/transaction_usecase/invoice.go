package transaction_usecase

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"ppl-m1-be/internal/usecase/apperr"
	"ppl-m1-be/internal/usecase/audit_usecase"
	"ppl-m1-be/internal/usecase/business_unit_usecase"
	"ppl-m1-be/internal/usecase/repository"
	"ppl-m1-be/internal/usecase/repository/sqlc"
	"ppl-m1-be/internal/usecase/utils"
)

const (
	InvoiceUnpaid = "belum_lunas"
	InvoicePaid   = "lunas"
	InvoiceVoid   = "dibatalkan"

	PaymentSucceeded = "berhasil"
	PaymentFailed    = "gagal"

	sandboxProvider = "sandbox"
)

type PaymentResult struct {
	Payment sqlc.Payment
	Invoice sqlc.InvoiceDetail
}

func (u *Usecase) ListInvoices(ctx context.Context, id Identity, status sql.NullString, limit, offset int32) (Page[sqlc.InvoiceDetail], error) {
	scope, err := u.scopeFor(ctx, id, utils.PermInvoicesManage)
	if err != nil {
		return Page[sqlc.InvoiceDetail]{}, err
	}
	items, err := u.store.ListInvoices(ctx, sqlc.ListInvoicesParams{OrganizationID: scope.OrgID(), PartnerID: scope.PartnerID, CustomerID: scope.CustomerID, Status: status, LimitRows: limit, OffsetRows: offset})
	if err != nil {
		return Page[sqlc.InvoiceDetail]{}, err
	}
	total, err := u.store.CountInvoices(ctx, sqlc.CountInvoicesParams{OrganizationID: scope.OrgID(), PartnerID: scope.PartnerID, CustomerID: scope.CustomerID, Status: status})
	return Page[sqlc.InvoiceDetail]{Items: items, Total: total}, err
}

// CreateInvoice hanya menerima quotation yang sudah disetujui.
func (u *Usecase) CreateInvoice(ctx context.Context, id Identity, quotationID int64) (sqlc.InvoiceDetail, error) {
	var invoiceID int64
	err := u.store.InTx(ctx, func(q *sqlc.Queries) error {
		scope, err := sellerScope(ctx, q, id, utils.PermInvoicesManage)
		if err != nil {
			return err
		}
		quotation, err := lockOwnQuotation(ctx, q, scope, quotationID)
		if err != nil {
			return err
		}
		if quotation.Status != QuotationApproved {
			return apperr.QuotationNotApproved
		}
		invoiceID, err = insertInvoice(ctx, q, scope, quotation)
		return err
	})
	if err != nil {
		return sqlc.InvoiceDetail{}, err
	}
	return u.invoice(ctx, id.Actor.OrganizationID, invoiceID)
}

func insertInvoice(ctx context.Context, q *sqlc.Queries, scope Scope, quotation sqlc.Quotation) (int64, error) {
	id, err := q.CreateInvoice(ctx, sqlc.CreateInvoiceParams{
		OrganizationID: scope.OrgID(), QuotationID: audit_usecase.ID(quotation.ID), PartnerID: quotation.PartnerID,
		CustomerID: quotation.CustomerID, Total: quotation.Total, CreatedBy: audit_usecase.ID(scope.Actor.UserID),
	})
	if repository.IsUniqueViolation(err) {
		return 0, apperr.InvoiceAlreadyExists
	}
	if err != nil {
		return 0, err
	}
	return id, audit_usecase.Record(ctx, q, scope.Actor.Change(scope.OrgID(), audit_usecase.ActionInvoiceCreated,
		audit_usecase.ResourceInvoices, id, nil, map[string]any{"quotation_id": quotation.ID, "total": quotation.Total, "status": InvoiceUnpaid}))
}

func (u *Usecase) invoice(ctx context.Context, orgID, id int64) (sqlc.InvoiceDetail, error) {
	invoice, err := u.store.GetInvoice(ctx, sqlc.GetInvoiceParams{ID: id, OrganizationID: orgID})
	return invoice, apperr.NotFound(err, apperr.InvoiceNotFound)
}

// gatewayResponse adalah jawaban payment gateway.
type gatewayResponse struct {
	Status    string
	Reference string
	Message   string
}

// sandboxCharge mensimulasikan payment gateway. `simulate` menentukan hasil
// sehingga skenario berhasil dan gagal sama-sama bisa diuji.
func sandboxCharge(invoice sqlc.Invoice, simulate string) gatewayResponse {
	reference := fmt.Sprintf("SBX-%d-%d", invoice.ID, time.Now().UnixNano())
	if simulate == PaymentFailed {
		return gatewayResponse{PaymentFailed, reference, "Pembayaran ditolak oleh gateway (simulasi)"}
	}
	return gatewayResponse{PaymentSucceeded, reference, "Pembayaran berhasil (simulasi)"}
}

// PayInvoice mencatat setiap percobaan. Invoice hanya berubah menjadi lunas
// bila gateway menjawab berhasil.
func (u *Usecase) PayInvoice(ctx context.Context, id Identity, invoiceID int64, simulate string) (PaymentResult, error) {
	var res PaymentResult
	err := u.store.InTx(ctx, func(q *sqlc.Queries) error {
		scope, invoice, err := lockPayableInvoice(ctx, q, id, invoiceID)
		if err != nil {
			return err
		}
		res.Payment, err = chargeInvoice(ctx, q, scope, invoice, simulate)
		return err
	})
	if err != nil {
		return PaymentResult{}, err
	}
	res.Invoice, err = u.invoice(ctx, id.Actor.OrganizationID, invoiceID)
	return res, err
}

func chargeInvoice(ctx context.Context, q *sqlc.Queries, scope Scope, invoice sqlc.Invoice, simulate string) (sqlc.Payment, error) {
	charge := sandboxCharge(invoice, simulate)
	payment, err := q.CreatePayment(ctx, sqlc.CreatePaymentParams{
		InvoiceID: invoice.ID, Provider: sandboxProvider, Reference: charge.Reference,
		Amount: invoice.Total, Status: charge.Status, Message: charge.Message,
	})
	if err != nil {
		return payment, err
	}
	return payment, applyCharge(ctx, q, scope, invoice, charge)
}

// lockPayableInvoice hanya meloloskan customer pemilik invoice.
func lockPayableInvoice(ctx context.Context, q *sqlc.Queries, id Identity, invoiceID int64) (Scope, sqlc.Invoice, error) {
	scope, err := customerScope(ctx, q, Scope{Actor: id.Actor})
	if err != nil {
		return scope, sqlc.Invoice{}, err
	}
	invoice, err := q.LockInvoice(ctx, sqlc.LockInvoiceParams{ID: invoiceID, OrganizationID: scope.OrgID()})
	if err != nil {
		return scope, invoice, apperr.NotFound(err, apperr.InvoiceNotFound)
	}
	if !scope.Owns(invoice.PartnerID, invoice.CustomerID) {
		return scope, invoice, apperr.InvoiceNotFound
	}
	switch invoice.Status {
	case InvoicePaid:
		return scope, invoice, apperr.InvoiceAlreadyPaid
	case InvoiceVoid:
		return scope, invoice, apperr.InvoiceCancelled
	}
	return scope, invoice, nil
}

func applyCharge(ctx context.Context, q *sqlc.Queries, scope Scope, invoice sqlc.Invoice, charge gatewayResponse) error {
	if charge.Status != PaymentSucceeded {
		return audit_usecase.Record(ctx, q, scope.Actor.Change(scope.OrgID(), audit_usecase.ActionPaymentFailed,
			audit_usecase.ResourceInvoices, invoice.ID, nil, map[string]string{"reference": charge.Reference, "status": charge.Status}))
	}
	if err := q.MarkInvoicePaid(ctx, invoice.ID); err != nil {
		return err
	}
	if err := pushRevenue(ctx, q, invoice); err != nil {
		return err
	}
	return audit_usecase.Record(ctx, q, scope.Actor.Change(scope.OrgID(), audit_usecase.ActionInvoicePaid,
		audit_usecase.ResourceInvoices, invoice.ID, map[string]string{"status": invoice.Status},
		map[string]string{"status": InvoicePaid, "reference": charge.Reference}))
}

func pushRevenue(ctx context.Context, q *sqlc.Queries, invoice sqlc.Invoice) error {
	rows, err := q.ListInvoiceUnitRevenue(ctx, sqlc.ListInvoiceUnitRevenueParams{QuotationID: invoice.QuotationID, OrderID: invoice.OrderID})
	if err != nil {
		return err
	}
	for _, row := range rows {
		err := business_unit_usecase.Push(ctx, q, business_unit_usecase.Metric{
			OrganizationID: invoice.OrganizationID, BusinessUnitID: row.BusinessUnitID, SourceModule: sourceModule,
			Key: "revenue", Label: "Pendapatan lunas", Unit: "rupiah", Value: row.Revenue,
		})
		if err != nil {
			return err
		}
	}
	return nil
}
