package transaction_route

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"ppl-m1-be/internal/controller/http/v1/respond"
	"ppl-m1-be/internal/usecase/repository/sqlc"
	"ppl-m1-be/internal/usecase/transaction_usecase"
	"ppl-m1-be/pkg/httpserver"
)

type orderStatusRequest struct {
	Status string `json:"status" binding:"required,oneof=diproses selesai dibatalkan"`
}

type orderResponse struct {
	ID              int64   `json:"id"`
	CustomerID      int64   `json:"customer_id"`
	CustomerName    string  `json:"customer_name"`
	PartnerID       *int64  `json:"partner_id"`
	PartnerName     *string `json:"partner_name"`
	Status          string  `json:"status"`
	Total           int64   `json:"total"`
	StatusChangedAt string  `json:"status_changed_at"`
	CreatedAt       string  `json:"created_at"`

	BusinessUnitName *string `json:"business_unit_name"`
	InvoiceID        *int64  `json:"invoice_id"`
	InvoiceStatus    *string `json:"invoice_status"`
}

type orderDetailResponse struct {
	orderResponse
	Items []lineResponse `json:"items"`
}

func toOrderResponse(o sqlc.OrderDetail) orderResponse {
	return orderResponse{
		ID: o.ID, CustomerID: o.CustomerID, CustomerName: o.CustomerName, PartnerID: httpserver.NullInt64Ptr(o.PartnerID),
		PartnerName: httpserver.NullStringPtr(o.PartnerName), Status: o.Status, Total: o.Total,
		StatusChangedAt: httpserver.FormatTime(o.StatusChangedAt), CreatedAt: httpserver.FormatTime(o.CreatedAt),
		BusinessUnitName: httpserver.NullStringPtr(o.BusinessUnitName), InvoiceID: httpserver.NullInt64Ptr(o.InvoiceID),
		InvoiceStatus: httpserver.NullStringPtr(o.InvoiceStatus),
	}
}

func toOrderItem(i sqlc.OrderItem) lineResponse {
	return lineResponse{httpserver.NullInt64Ptr(i.ProductID), i.ProductName, i.Price, i.Qty, i.Subtotal}
}

func toOrderDetail(v transaction_usecase.OrderView) orderDetailResponse {
	return orderDetailResponse{toOrderResponse(v.Order), httpserver.Map(v.Items, toOrderItem)}
}

func (h *Handler) ListOrders(c *gin.Context) {
	page, ok := httpserver.RequirePage(c)
	status, okStatus := httpserver.QueryEnum(c, "status", transaction_usecase.OrderNew, transaction_usecase.OrderProcessing, transaction_usecase.OrderDone, transaction_usecase.OrderCancelled)
	if !ok {
		return
	}
	if !okStatus {
		httpserver.Fail(c, httpserver.ErrInvalidParameter)
		return
	}
	result, err := h.uc.ListOrders(c.Request.Context(), identity(c), status, page.Limit(), page.Offset())
	if err != nil {
		respond.Error(c, "daftar order gagal", err)
		return
	}
	httpserver.SuccessList(c, httpserver.Map(result.Items, toOrderResponse), page.Meta(result.Total))
}

func (h *Handler) GetOrder(c *gin.Context) {
	id, ok := httpserver.RequireID(c)
	if !ok {
		return
	}
	view, err := h.uc.GetOrder(c.Request.Context(), identity(c), id)
	if err != nil {
		respond.Error(c, "ambil order gagal", err)
		return
	}
	httpserver.Success(c, http.StatusOK, toOrderDetail(view))
}

func (h *Handler) UpdateOrderStatus(c *gin.Context) {
	id, ok := httpserver.RequireID(c)
	var req orderStatusRequest
	if !ok || !httpserver.BindJSON(c, &req) {
		return
	}
	view, err := h.uc.UpdateOrderStatus(c.Request.Context(), identity(c), id, req.Status)
	if err != nil {
		respond.Error(c, "ubah status order gagal", err)
		return
	}
	httpserver.Success(c, http.StatusOK, toOrderDetail(view))
}
