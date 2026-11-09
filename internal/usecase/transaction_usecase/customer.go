package transaction_usecase

import (
	"context"
	"database/sql"
	"errors"
	"strings"

	"ppl-m1-be/internal/usecase/apperr"
	"ppl-m1-be/internal/usecase/audit_usecase"
	"ppl-m1-be/internal/usecase/repository"
	"ppl-m1-be/internal/usecase/repository/sqlc"
	"ppl-m1-be/internal/usecase/utils"
)

type CustomerInput struct {
	Name    string
	Email   string
	Phone   sql.NullString
	Address sql.NullString
}

func (i CustomerInput) email() string {
	return strings.ToLower(strings.TrimSpace(i.Email))
}

// ListCustomers: partner hanya melihat pelanggan yang terhubung dengannya.
func (u *Usecase) ListCustomers(ctx context.Context, id Identity, q sql.NullString, limit, offset int32) (Page[sqlc.Customer], error) {
	scope, err := sellerScope(ctx, u.store.Queries, id, utils.PermCustomersManage)
	if err != nil {
		return Page[sqlc.Customer]{}, err
	}
	items, err := u.store.ListCustomers(ctx, sqlc.ListCustomersParams{OrganizationID: scope.OrgID(), Q: q, PartnerID: scope.PartnerID, LimitRows: limit, OffsetRows: offset})
	if err != nil {
		return Page[sqlc.Customer]{}, err
	}
	total, err := u.store.CountCustomers(ctx, sqlc.CountCustomersParams{OrganizationID: scope.OrgID(), Q: q, PartnerID: scope.PartnerID})
	return Page[sqlc.Customer]{Items: items, Total: total}, err
}

func (u *Usecase) CreateCustomer(ctx context.Context, id Identity, in CustomerInput) (sqlc.Customer, error) {
	var customer sqlc.Customer
	err := u.store.InTx(ctx, func(q *sqlc.Queries) error {
		scope, err := sellerScope(ctx, q, id, utils.PermCustomersManage)
		if err != nil {
			return err
		}
		if customer, err = saveNewCustomer(ctx, q, scope, in); err != nil {
			return err
		}
		return audit_usecase.Record(ctx, q, scope.Actor.Change(scope.OrgID(), audit_usecase.ActionCustomerCreated,
			audit_usecase.ResourceCustomers, customer.ID, nil, map[string]string{"name": customer.Name, "email": customer.Email}))
	})
	return customer, err
}

// saveNewCustomer: bila email sudah terdaftar di organisasi dan pemanggilnya
// partner, pelanggan itu dihubungkan tanpa mengubah datanya.
func saveNewCustomer(ctx context.Context, q *sqlc.Queries, scope Scope, in CustomerInput) (sqlc.Customer, error) {
	if scope.PartnerID.Valid {
		existing, err := q.GetCustomerByEmail(ctx, sqlc.GetCustomerByEmailParams{OrganizationID: scope.OrgID(), Lower: in.email()})
		if err == nil {
			return existing, linkPartner(ctx, q, scope.OrgID(), scope.PartnerID, existing.ID)
		}
		if !errors.Is(err, sql.ErrNoRows) {
			return existing, err
		}
	}
	customer, err := q.CreateCustomer(ctx, sqlc.CreateCustomerParams{
		OrganizationID: scope.OrgID(), Name: strings.TrimSpace(in.Name), Email: in.email(), Phone: in.Phone, Address: in.Address,
	})
	if err != nil {
		return customer, customerWriteError(err)
	}
	return customer, linkPartner(ctx, q, scope.OrgID(), scope.PartnerID, customer.ID)
}

func linkPartner(ctx context.Context, q *sqlc.Queries, orgID int64, partner sql.NullInt64, customerID int64) error {
	if !partner.Valid {
		return nil
	}
	return q.LinkPartnerCustomer(ctx, sqlc.LinkPartnerCustomerParams{OrganizationID: orgID, PartnerID: partner.Int64, CustomerID: customerID})
}

func (u *Usecase) UpdateCustomer(ctx context.Context, id Identity, customerID int64, in CustomerInput) (sqlc.Customer, error) {
	var customer sqlc.Customer
	err := u.store.InTx(ctx, func(q *sqlc.Queries) error {
		scope, err := sellerScope(ctx, q, id, utils.PermCustomersManage)
		if err != nil {
			return err
		}
		if err := requireLinked(ctx, q, scope, customerID, apperr.CustomerNotFound); err != nil {
			return err
		}
		customer, err = q.UpdateCustomer(ctx, sqlc.UpdateCustomerParams{
			ID: customerID, OrganizationID: scope.OrgID(), Name: strings.TrimSpace(in.Name), Email: in.email(), Phone: in.Phone, Address: in.Address,
		})
		if err != nil {
			return customerWriteError(err)
		}
		return audit_usecase.Record(ctx, q, scope.Actor.Change(scope.OrgID(), audit_usecase.ActionCustomerUpdated,
			audit_usecase.ResourceCustomers, customer.ID, nil, map[string]string{"name": customer.Name, "email": customer.Email}))
	})
	return customer, err
}

// requireLinked menolak partner yang tidak terhubung dengan pelanggan itu.
func requireLinked(ctx context.Context, q *sqlc.Queries, scope Scope, customerID int64, denied error) error {
	if !scope.PartnerID.Valid {
		return nil
	}
	linked, err := q.PartnerHasCustomer(ctx, sqlc.PartnerHasCustomerParams{PartnerID: scope.PartnerID.Int64, CustomerID: customerID})
	if err != nil {
		return err
	}
	if !linked {
		return denied
	}
	return nil
}

func customerWriteError(err error) error {
	if repository.IsUniqueViolation(err) {
		return apperr.CustomerEmailTaken
	}
	return apperr.NotFound(err, apperr.CustomerNotFound)
}

// EnsureCustomerForUser mengembalikan data pelanggan milik akun, membuat atau
// menautkannya bila belum ada. Dipakai checkout dan registrasi M5.
func EnsureCustomerForUser(ctx context.Context, q *sqlc.Queries, userID int64) (int64, error) {
	customer, err := q.GetCustomerForUser(ctx, userID)
	if errors.Is(err, sql.ErrNoRows) {
		created, err := q.CreateCustomerFromUser(ctx, userID)
		return created.ID, err
	}
	if err != nil {
		return 0, err
	}
	if !customer.UserID.Valid {
		err = q.LinkCustomerUser(ctx, sqlc.LinkCustomerUserParams{ID: customer.ID, UserID: audit_usecase.ID(userID)})
	}
	return customer.ID, err
}
