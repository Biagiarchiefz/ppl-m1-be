// Package storefront_usecase adalah modul M5: katalog, form layanan,
// keranjang, checkout, dan pencarian.
package storefront_usecase

import (
	"context"
	"database/sql"
	"errors"
	"strings"

	"ppl-m1-be/internal/usecase/apperr"
	"ppl-m1-be/internal/usecase/audit_usecase"
	"ppl-m1-be/internal/usecase/business_unit_usecase"
	"ppl-m1-be/internal/usecase/repository"
	"ppl-m1-be/internal/usecase/repository/sqlc"
	"ppl-m1-be/internal/usecase/transaction_usecase"
)

type Usecase struct {
	store *repository.Store
}

func New(store *repository.Store) *Usecase {
	return &Usecase{store: store}
}

type CatalogFilter struct {
	Q              sql.NullString
	Category       sql.NullString
	BusinessUnitID sql.NullInt64
	PartnerID      sql.NullInt64
}

type ProductPage struct {
	Items []sqlc.ProductDetail
	Total int64
}

type Cart struct {
	Items []sqlc.ListCartItemsRow
	Count int64
	Total int64
}

type ServiceRequestInput struct {
	BusinessUnitID int64
	ContactName    string
	ContactEmail   string
	ContactPhone   string
	Subject        string
	Message        string
}

type ServiceRequestPage struct {
	Items []sqlc.ServiceRequestDetail
	Total int64
}

func (u *Usecase) BusinessUnits(ctx context.Context, orgID int64) ([]sqlc.BusinessUnitDetail, error) {
	return u.store.ListCatalogBusinessUnits(ctx, orgID)
}

func (u *Usecase) Categories(ctx context.Context, orgID int64) ([]string, error) {
	return u.store.ListCatalogCategories(ctx, orgID)
}

func (u *Usecase) Products(ctx context.Context, orgID int64, f CatalogFilter, limit, offset int32) (ProductPage, error) {
	items, err := u.store.ListCatalogProducts(ctx, sqlc.ListCatalogProductsParams{
		OrganizationID: orgID, Q: f.Q, Category: f.Category, BusinessUnitID: f.BusinessUnitID, PartnerID: f.PartnerID, LimitRows: limit, OffsetRows: offset,
	})
	if err != nil {
		return ProductPage{}, err
	}
	total, err := u.store.CountCatalogProducts(ctx, sqlc.CountCatalogProductsParams{OrganizationID: orgID, Q: f.Q, Category: f.Category, BusinessUnitID: f.BusinessUnitID, PartnerID: f.PartnerID})
	return ProductPage{Items: items, Total: total}, err
}

func (u *Usecase) GetCart(ctx context.Context, userID int64) (Cart, error) {
	items, err := u.store.ListCartItems(ctx, userID)
	if err != nil {
		return Cart{}, err
	}
	cart := Cart{Items: items}
	for _, item := range items {
		cart.Count += int64(item.Qty)
		cart.Total += item.Price * int64(item.Qty)
	}
	return cart, nil
}

// AddToCart menambah qty bila produk sudah ada di keranjang.
func (u *Usecase) AddToCart(ctx context.Context, actor audit_usecase.Actor, productID int64, qty int32) (Cart, error) {
	product, err := u.store.GetCatalogProduct(ctx, sqlc.GetCatalogProductParams{ID: productID, OrganizationID: actor.OrganizationID})
	if err != nil {
		return Cart{}, apperr.NotFound(err, apperr.ProductNotFound)
	}
	if product.Stock.Valid && product.Stock.Int32 < qty {
		return Cart{}, apperr.OutOfStock
	}
	if err := u.store.AddCartItem(ctx, sqlc.AddCartItemParams{UserID: actor.UserID, ProductID: productID, Qty: qty}); err != nil {
		return Cart{}, err
	}
	return u.GetCart(ctx, actor.UserID)
}

func (u *Usecase) SetCartQty(ctx context.Context, userID, productID int64, qty int32) (Cart, error) {
	rows, err := u.store.SetCartItemQty(ctx, sqlc.SetCartItemQtyParams{UserID: userID, ProductID: productID, Qty: qty})
	return u.afterCartChange(ctx, userID, rows, err)
}

func (u *Usecase) RemoveFromCart(ctx context.Context, userID, productID int64) (Cart, error) {
	rows, err := u.store.DeleteCartItem(ctx, sqlc.DeleteCartItemParams{UserID: userID, ProductID: productID})
	return u.afterCartChange(ctx, userID, rows, err)
}

func (u *Usecase) afterCartChange(ctx context.Context, userID, rows int64, err error) (Cart, error) {
	if err != nil {
		return Cart{}, err
	}
	if rows == 0 {
		return Cart{}, apperr.CartItemNotFound
	}
	return u.GetCart(ctx, userID)
}

// Checkout mengubah keranjang menjadi Order lewat kontrak M4, lalu
// mengosongkannya, semua dalam satu transaksi.
func (u *Usecase) Checkout(ctx context.Context, actor audit_usecase.Actor) ([]int64, error) {
	var orderIDs []int64
	err := u.store.InTx(ctx, func(q *sqlc.Queries) error {
		lines, err := checkoutLines(ctx, q, actor.UserID)
		if err != nil {
			return err
		}
		customerID, err := transaction_usecase.EnsureCustomerForUser(ctx, q, actor.UserID)
		if err != nil {
			return err
		}
		if orderIDs, err = transaction_usecase.CreateOrders(ctx, q, actor, customerID, lines); err != nil {
			return err
		}
		return q.ClearCart(ctx, actor.UserID)
	})
	return orderIDs, err
}

func checkoutLines(ctx context.Context, q *sqlc.Queries, userID int64) ([]transaction_usecase.OrderLine, error) {
	items, err := q.LockCartItems(ctx, userID)
	if err != nil {
		return nil, err
	}
	if len(items) == 0 {
		return nil, apperr.CartEmpty
	}
	lines := make([]transaction_usecase.OrderLine, 0, len(items))
	for _, item := range items {
		if !item.Available {
			return nil, apperr.CartItemUnavailable
		}
		lines = append(lines, transaction_usecase.OrderLine{
			ProductID: item.ProductID, BusinessUnitID: item.BusinessUnitID, PartnerID: item.PartnerID,
			Name: item.Name, Price: item.Price, Qty: item.Qty,
		})
	}
	return lines, nil
}

// SendServiceRequest meneruskan permintaan ke unit usaha yang dipilih.
func (u *Usecase) SendServiceRequest(ctx context.Context, actor audit_usecase.Actor, in ServiceRequestInput) (sqlc.ServiceRequest, error) {
	var created sqlc.ServiceRequest
	err := u.store.InTx(ctx, func(q *sqlc.Queries) error {
		if err := checkCatalogUnit(ctx, q, actor.OrganizationID, in.BusinessUnitID); err != nil {
			return err
		}
		var err error
		if created, err = insertServiceRequest(ctx, q, actor, in); err != nil {
			return err
		}
		if err := pushServiceRequest(ctx, q, actor.OrganizationID, in.BusinessUnitID); err != nil {
			return err
		}
		return audit_usecase.Record(ctx, q, actor.Change(actor.OrganizationID, audit_usecase.ActionServiceRequestSent,
			audit_usecase.ResourceServiceRequests, created.ID, nil, map[string]any{"business_unit_id": in.BusinessUnitID, "subject": created.Subject}))
	})
	return created, err
}

func checkCatalogUnit(ctx context.Context, q *sqlc.Queries, orgID, unitID int64) error {
	unit, err := q.GetBusinessUnit(ctx, sqlc.GetBusinessUnitParams{ID: unitID, OrganizationID: orgID})
	if errors.Is(err, sql.ErrNoRows) || (err == nil && !unit.IsActive) {
		return apperr.FieldErrors{"business_unit_id": "Unit usaha tidak terdaftar atau nonaktif"}
	}
	return err
}

func insertServiceRequest(ctx context.Context, q *sqlc.Queries, actor audit_usecase.Actor, in ServiceRequestInput) (sqlc.ServiceRequest, error) {
	return q.CreateServiceRequest(ctx, sqlc.CreateServiceRequestParams{
		OrganizationID: actor.OrganizationID, BusinessUnitID: in.BusinessUnitID, UserID: audit_usecase.ID(actor.UserID),
		ContactName: strings.TrimSpace(in.ContactName), ContactEmail: strings.ToLower(strings.TrimSpace(in.ContactEmail)),
		ContactPhone: strings.TrimSpace(in.ContactPhone), Subject: strings.TrimSpace(in.Subject), Message: strings.TrimSpace(in.Message),
	})
}

func pushServiceRequest(ctx context.Context, q *sqlc.Queries, orgID, unitID int64) error {
	return business_unit_usecase.Push(ctx, q, business_unit_usecase.Metric{
		OrganizationID: orgID, BusinessUnitID: audit_usecase.ID(unitID), SourceModule: "m5",
		Key: "service_requests", Label: "Permintaan layanan", Unit: "permintaan", Value: 1,
	})
}

func (u *Usecase) ServiceRequests(ctx context.Context, orgID int64, unitID sql.NullInt64, limit, offset int32) (ServiceRequestPage, error) {
	items, err := u.store.ListServiceRequests(ctx, sqlc.ListServiceRequestsParams{OrganizationID: orgID, BusinessUnitID: unitID, LimitRows: limit, OffsetRows: offset})
	if err != nil {
		return ServiceRequestPage{}, err
	}
	total, err := u.store.CountServiceRequests(ctx, sqlc.CountServiceRequestsParams{OrganizationID: orgID, BusinessUnitID: unitID})
	return ServiceRequestPage{Items: items, Total: total}, err
}
