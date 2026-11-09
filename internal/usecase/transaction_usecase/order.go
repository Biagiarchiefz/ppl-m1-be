package transaction_usecase

import (
	"context"
	"database/sql"

	"ppl-m1-be/internal/usecase/apperr"
	"ppl-m1-be/internal/usecase/audit_usecase"
	"ppl-m1-be/internal/usecase/business_unit_usecase"
	"ppl-m1-be/internal/usecase/repository/sqlc"
)

const (
	OrderNew        = "baru"
	OrderProcessing = "diproses"
	OrderDone       = "selesai"
	OrderCancelled  = "dibatalkan"
)

// nextOrderStatus adalah urutan status order.
var nextOrderStatus = map[string]string{OrderNew: OrderProcessing, OrderProcessing: OrderDone}

// OrderLine adalah kontrak M4 untuk membentuk Order dari modul lain.
type OrderLine struct {
	ProductID      int64
	BusinessUnitID sql.NullInt64
	PartnerID      sql.NullInt64
	Name           string
	Price          int64
	Qty            int32
}

func (l OrderLine) toLine() line {
	return line{audit_usecase.ID(l.ProductID), l.BusinessUnitID, l.Name, l.Price, l.Qty}
}

type OrderView struct {
	Order sqlc.OrderDetail
	Items []sqlc.OrderItem
}

// seller adalah pihak yang memproses Order: partner, atau unit usaha untuk
// produk milik organisasi.
type seller struct {
	Partner sql.NullInt64
	Unit    sql.NullInt64
}

func sellerOf(l OrderLine) seller {
	if l.PartnerID.Valid {
		return seller{Partner: l.PartnerID}
	}
	return seller{Unit: l.BusinessUnitID}
}

// CreateOrders memotong stok, lalu membentuk satu Order dan satu invoice per
// penjual. Panggil dalam transaksi modul pemanggil.
func CreateOrders(ctx context.Context, q *sqlc.Queries, actor audit_usecase.Actor, customerID int64, lines []OrderLine) ([]int64, error) {
	if err := reserveStock(ctx, q, lines); err != nil {
		return nil, err
	}
	groups, sellers := groupBySeller(lines)
	ids := make([]int64, 0, len(sellers))
	for _, s := range sellers {
		id, err := createOrder(ctx, q, actor, s, customerID, groups[s])
		if err != nil {
			return nil, err
		}
		ids = append(ids, id)
	}
	return ids, nil
}

// reserveStock berjalan berurutan id produk (urutan baris keranjang), supaya
// dua checkout bersamaan tidak saling mengunci.
func reserveStock(ctx context.Context, q *sqlc.Queries, lines []OrderLine) error {
	for _, l := range lines {
		rows, err := q.ReserveStock(ctx, sqlc.ReserveStockParams{Qty: l.Qty, ID: l.ProductID})
		if err != nil {
			return err
		}
		if rows == 0 {
			return apperr.OutOfStock
		}
	}
	return nil
}

func groupBySeller(lines []OrderLine) (map[seller][]line, []seller) {
	groups := map[seller][]line{}
	var sellers []seller
	for _, l := range lines {
		s := sellerOf(l)
		if _, seen := groups[s]; !seen {
			sellers = append(sellers, s)
		}
		groups[s] = append(groups[s], l.toLine())
	}
	return groups, sellers
}

func createOrder(ctx context.Context, q *sqlc.Queries, actor audit_usecase.Actor, s seller, customerID int64, lines []line) (int64, error) {
	total := totalOf(lines)
	id, err := q.CreateOrder(ctx, sqlc.CreateOrderParams{OrganizationID: actor.OrganizationID, PartnerID: s.Partner, BusinessUnitID: s.Unit, CustomerID: customerID, Total: total})
	if err != nil {
		return 0, err
	}
	for _, l := range lines {
		if err := addOrderItem(ctx, q, actor.OrganizationID, id, l); err != nil {
			return 0, err
		}
	}
	if err := billOrder(ctx, q, actor.OrganizationID, id, s, customerID, total); err != nil {
		return 0, err
	}
	return id, audit_usecase.Record(ctx, q, actor.Change(actor.OrganizationID, audit_usecase.ActionOrderCreated,
		audit_usecase.ResourceOrders, id, nil, map[string]any{"customer_id": customerID, "total": total, "status": OrderNew}))
}

// billOrder menerbitkan invoice agar Order dari checkout bisa langsung dibayar.
func billOrder(ctx context.Context, q *sqlc.Queries, orgID, orderID int64, s seller, customerID, total int64) error {
	err := q.CreateOrderInvoice(ctx, sqlc.CreateOrderInvoiceParams{
		OrganizationID: orgID, OrderID: audit_usecase.ID(orderID), PartnerID: s.Partner, CustomerID: customerID, Total: total,
	})
	if err != nil {
		return err
	}
	return linkPartner(ctx, q, orgID, s.Partner, customerID)
}

func addOrderItem(ctx context.Context, q *sqlc.Queries, orgID, orderID int64, l line) error {
	err := q.CreateOrderItem(ctx, sqlc.CreateOrderItemParams{
		OrderID: orderID, ProductID: l.ProductID, BusinessUnitID: l.BusinessUnitID,
		ProductName: l.Name, Price: l.Price, Qty: l.Qty, Subtotal: l.subtotal(),
	})
	if err != nil {
		return err
	}
	return business_unit_usecase.Push(ctx, q, business_unit_usecase.Metric{
		OrganizationID: orgID, BusinessUnitID: l.BusinessUnitID, SourceModule: sourceModule,
		Key: "items_ordered", Label: "Item dipesan", Unit: "item", Value: int64(l.Qty),
	})
}

func (u *Usecase) ListOrders(ctx context.Context, id Identity, status sql.NullString, limit, offset int32) (Page[sqlc.OrderDetail], error) {
	scope, err := resolveOrderScope(ctx, u.store.Queries, id)
	if err != nil {
		return Page[sqlc.OrderDetail]{}, err
	}
	items, err := u.store.ListOrders(ctx, sqlc.ListOrdersParams{OrganizationID: scope.OrgID(), BusinessUnitID: scope.BusinessUnitID, PartnerID: scope.PartnerID, CustomerID: scope.CustomerID, Status: status, LimitRows: limit, OffsetRows: offset})
	if err != nil {
		return Page[sqlc.OrderDetail]{}, err
	}
	total, err := u.store.CountOrders(ctx, sqlc.CountOrdersParams{OrganizationID: scope.OrgID(), BusinessUnitID: scope.BusinessUnitID, PartnerID: scope.PartnerID, CustomerID: scope.CustomerID, Status: status})
	return Page[sqlc.OrderDetail]{Items: items, Total: total}, err
}

func (u *Usecase) GetOrder(ctx context.Context, id Identity, orderID int64) (OrderView, error) {
	scope, err := resolveOrderScope(ctx, u.store.Queries, id)
	if err != nil {
		return OrderView{}, err
	}
	order, err := u.store.GetOrder(ctx, sqlc.GetOrderParams{ID: orderID, OrganizationID: scope.OrgID()})
	if err != nil {
		return OrderView{}, apperr.NotFound(err, apperr.OrderNotFound)
	}
	if !scope.ownsOrder(order.PartnerID, order.BusinessUnitID, order.CustomerID) {
		return OrderView{}, apperr.OrderNotFound
	}
	items, err := u.store.ListOrderItems(ctx, orderID)
	return OrderView{Order: order, Items: items}, err
}

// UpdateOrderStatus memajukan order satu langkah atau membatalkannya. Partner
// yang ditangguhkan tetap boleh menyelesaikan Order yang sudah berjalan.
func (u *Usecase) UpdateOrderStatus(ctx context.Context, id Identity, orderID int64, status string) (OrderView, error) {
	err := u.store.InTx(ctx, func(q *sqlc.Queries) error {
		return changeOrderStatus(ctx, q, id, orderID, status)
	})
	if err != nil {
		return OrderView{}, err
	}
	return u.GetOrder(ctx, id, orderID)
}

func changeOrderStatus(ctx context.Context, q *sqlc.Queries, id Identity, orderID int64, status string) error {
	scope, current, err := lockOwnOrder(ctx, q, id, orderID)
	if err != nil {
		return err
	}
	if status == OrderCancelled {
		return cancelOrder(ctx, q, scope, current)
	}
	if scope.CustomerID.Valid {
		return apperr.Forbidden
	}
	if err := checkOrderTransition(current.Status, status); err != nil {
		return err
	}
	return saveOrderStatus(ctx, q, scope, current, status, audit_usecase.ActionOrderStatus)
}

func saveOrderStatus(ctx context.Context, q *sqlc.Queries, scope Scope, current sqlc.Order, status, action string) error {
	err := q.UpdateOrderStatus(ctx, sqlc.UpdateOrderStatusParams{ID: current.ID, OrganizationID: scope.OrgID(), Status: status})
	if err != nil {
		return err
	}
	return audit_usecase.Record(ctx, q, scope.Actor.Change(scope.OrgID(), action,
		audit_usecase.ResourceOrders, current.ID, map[string]string{"status": current.Status}, map[string]string{"status": status}))
}

// cancelOrder hanya berlaku saat Order masih baru dan belum dibayar. Invoice
// ikut dibatalkan dan stok dikembalikan.
func cancelOrder(ctx context.Context, q *sqlc.Queries, scope Scope, current sqlc.Order) error {
	if current.Status == OrderDone {
		return apperr.OrderCompleted
	}
	if current.Status != OrderNew {
		return apperr.OrderCancelNotAllowed
	}
	if err := voidOrderInvoice(ctx, q, current.ID); err != nil {
		return err
	}
	if err := restoreOrderStock(ctx, q, current.ID); err != nil {
		return err
	}
	return saveOrderStatus(ctx, q, scope, current, OrderCancelled, audit_usecase.ActionOrderCancelled)
}

func voidOrderInvoice(ctx context.Context, q *sqlc.Queries, orderID int64) error {
	invoice, err := q.LockInvoiceByOrder(ctx, audit_usecase.ID(orderID))
	if err != nil {
		return err
	}
	if invoice.Status == InvoicePaid {
		return apperr.OrderAlreadyPaid
	}
	return q.CancelInvoice(ctx, invoice.ID)
}

func restoreOrderStock(ctx context.Context, q *sqlc.Queries, orderID int64) error {
	items, err := q.ListOrderItems(ctx, orderID)
	if err != nil {
		return err
	}
	for _, item := range items {
		if !item.ProductID.Valid {
			continue
		}
		if err := q.RestoreStock(ctx, sqlc.RestoreStockParams{Qty: item.Qty, ID: item.ProductID.Int64}); err != nil {
			return err
		}
	}
	return nil
}

func checkOrderTransition(current, target string) error {
	if current == OrderDone {
		return apperr.OrderCompleted
	}
	if nextOrderStatus[current] != target {
		return apperr.OrderStatusInvalid
	}
	return nil
}

func lockOwnOrder(ctx context.Context, q *sqlc.Queries, id Identity, orderID int64) (Scope, sqlc.Order, error) {
	scope, err := resolveOrderScope(ctx, q, id)
	if err != nil {
		return scope, sqlc.Order{}, err
	}
	order, err := q.LockOrder(ctx, sqlc.LockOrderParams{ID: orderID, OrganizationID: scope.OrgID()})
	if err != nil {
		return scope, order, apperr.NotFound(err, apperr.OrderNotFound)
	}
	if !scope.ownsOrder(order.PartnerID, order.BusinessUnitID, order.CustomerID) {
		return scope, order, apperr.OrderNotFound
	}
	return scope, order, nil
}
