package transaction_usecase

import (
	"context"
	"database/sql"
	"errors"
	"strings"

	"ppl-m1-be/internal/usecase/apperr"
	"ppl-m1-be/internal/usecase/audit_usecase"
	"ppl-m1-be/internal/usecase/repository/sqlc"
	"ppl-m1-be/internal/usecase/utils"
)

// ProductInput: BusinessUnitID boleh kosong untuk produk partner. Stock wajib
// untuk barang dan selalu kosong untuk jasa.
type ProductInput struct {
	BusinessUnitID sql.NullInt64
	Type           string
	Name           string
	Category       string
	Description    sql.NullString
	Price          int64
	Stock          sql.NullInt32
}

func (u *Usecase) ListProducts(ctx context.Context, id Identity, q sql.NullString, limit, offset int32) (Page[sqlc.ProductDetail], error) {
	scope, err := u.scopeFor(ctx, id, utils.PermProductsManage)
	if err != nil {
		return Page[sqlc.ProductDetail]{}, err
	}
	items, err := u.store.ListProducts(ctx, sqlc.ListProductsParams{OrganizationID: scope.OrgID(), PartnerID: scope.PartnerID, Q: q, LimitRows: limit, OffsetRows: offset})
	if err != nil {
		return Page[sqlc.ProductDetail]{}, err
	}
	total, err := u.store.CountProducts(ctx, sqlc.CountProductsParams{OrganizationID: scope.OrgID(), PartnerID: scope.PartnerID, Q: q})
	return Page[sqlc.ProductDetail]{Items: items, Total: total}, err
}

func (u *Usecase) CreateProduct(ctx context.Context, id Identity, in ProductInput) (sqlc.ProductDetail, error) {
	var productID int64
	err := u.store.InTx(ctx, func(q *sqlc.Queries) error {
		scope, err := sellerScope(ctx, q, id, utils.PermProductsManage)
		if err != nil {
			return err
		}
		if err := checkProduct(ctx, q, scope.OrgID(), scope.PartnerID, in); err != nil {
			return err
		}
		if productID, err = insertProduct(ctx, q, scope, in); err != nil {
			return err
		}
		return audit_usecase.Record(ctx, q, scope.Actor.Change(scope.OrgID(), audit_usecase.ActionProductCreated,
			audit_usecase.ResourceProducts, productID, nil, map[string]any{"name": in.Name, "price": in.Price}))
	})
	if err != nil {
		return sqlc.ProductDetail{}, err
	}
	return u.product(ctx, id.Actor.OrganizationID, productID)
}

func insertProduct(ctx context.Context, q *sqlc.Queries, scope Scope, in ProductInput) (int64, error) {
	return q.CreateProduct(ctx, sqlc.CreateProductParams{
		OrganizationID: scope.OrgID(), BusinessUnitID: in.BusinessUnitID, PartnerID: scope.PartnerID, Type: in.Type,
		Name: strings.TrimSpace(in.Name), Category: strings.TrimSpace(in.Category), Description: in.Description, Price: in.Price, Stock: in.Stock,
	})
}

func (u *Usecase) UpdateProduct(ctx context.Context, id Identity, productID int64, in ProductInput) (sqlc.ProductDetail, error) {
	err := u.store.InTx(ctx, func(q *sqlc.Queries) error {
		scope, current, err := lockOwnProduct(ctx, q, id, productID)
		if err != nil {
			return err
		}
		if err := checkProduct(ctx, q, scope.OrgID(), current.PartnerID, in); err != nil {
			return err
		}
		if err := saveProduct(ctx, q, scope.OrgID(), productID, in); err != nil {
			return err
		}
		return audit_usecase.Record(ctx, q, scope.Actor.Change(scope.OrgID(), audit_usecase.ActionProductUpdated,
			audit_usecase.ResourceProducts, productID, map[string]any{"name": current.Name, "price": current.Price},
			map[string]any{"name": in.Name, "price": in.Price}))
	})
	if err != nil {
		return sqlc.ProductDetail{}, err
	}
	return u.product(ctx, id.Actor.OrganizationID, productID)
}

func saveProduct(ctx context.Context, q *sqlc.Queries, orgID, productID int64, in ProductInput) error {
	return q.UpdateProduct(ctx, sqlc.UpdateProductParams{
		ID: productID, OrganizationID: orgID, BusinessUnitID: in.BusinessUnitID, Type: in.Type,
		Name: strings.TrimSpace(in.Name), Category: strings.TrimSpace(in.Category), Description: in.Description, Price: in.Price, Stock: in.Stock,
	})
}

func (u *Usecase) SetProductStatus(ctx context.Context, id Identity, productID int64, active bool) error {
	return u.store.InTx(ctx, func(q *sqlc.Queries) error {
		scope, current, err := lockOwnProduct(ctx, q, id, productID)
		if err != nil {
			return err
		}
		if current.IsActive == active {
			return apperr.ProductStatusUnchanged
		}
		err = q.SetProductActive(ctx, sqlc.SetProductActiveParams{ID: productID, OrganizationID: scope.OrgID(), IsActive: active})
		if err != nil {
			return err
		}
		return audit_usecase.Record(ctx, q, scope.Actor.Change(scope.OrgID(), audit_usecase.ActionProductStatus,
			audit_usecase.ResourceProducts, productID, map[string]bool{"is_active": current.IsActive}, map[string]bool{"is_active": active}))
	})
}

func lockOwnProduct(ctx context.Context, q *sqlc.Queries, id Identity, productID int64) (Scope, sqlc.Product, error) {
	scope, err := sellerScope(ctx, q, id, utils.PermProductsManage)
	if err != nil {
		return scope, sqlc.Product{}, err
	}
	product, err := q.LockProduct(ctx, sqlc.LockProductParams{ID: productID, OrganizationID: scope.OrgID()})
	if err != nil {
		return scope, product, apperr.NotFound(err, apperr.ProductNotFound)
	}
	if !scope.Owns(product.PartnerID, 0) {
		return scope, product, apperr.ProductNotFound
	}
	return scope, product, nil
}

func sellerScope(ctx context.Context, q *sqlc.Queries, id Identity, managePerm string) (Scope, error) {
	scope, err := resolveScope(ctx, q, id, managePerm)
	if err != nil {
		return scope, err
	}
	return scope, scope.RequireSeller()
}

// checkProduct: barang wajib punya stok; produk milik organisasi wajib punya
// unit usaha, sedangkan produk partner boleh tanpa unit usaha.
func checkProduct(ctx context.Context, q *sqlc.Queries, orgID int64, owner sql.NullInt64, in ProductInput) error {
	if in.Type == "produk" && !in.Stock.Valid {
		return apperr.FieldErrors{"stock": "Stok wajib diisi untuk produk"}
	}
	if in.BusinessUnitID.Valid {
		return checkUnit(ctx, q, orgID, in.BusinessUnitID.Int64)
	}
	if !owner.Valid {
		return apperr.FieldErrors{"business_unit_id": "Unit usaha wajib dipilih untuk produk organisasi"}
	}
	return nil
}

func checkUnit(ctx context.Context, q *sqlc.Queries, orgID, unitID int64) error {
	unit, err := q.GetBusinessUnit(ctx, sqlc.GetBusinessUnitParams{ID: unitID, OrganizationID: orgID})
	if errors.Is(err, sql.ErrNoRows) || (err == nil && !unit.IsActive) {
		return apperr.FieldErrors{"business_unit_id": "Unit usaha tidak terdaftar atau nonaktif"}
	}
	return err
}

func (u *Usecase) product(ctx context.Context, orgID, id int64) (sqlc.ProductDetail, error) {
	product, err := u.store.GetProduct(ctx, sqlc.GetProductParams{ID: id, OrganizationID: orgID})
	return product, apperr.NotFound(err, apperr.ProductNotFound)
}
