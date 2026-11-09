package transaction_route

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"ppl-m1-be/internal/controller/http/v1/respond"
	"ppl-m1-be/internal/usecase/repository/sqlc"
	"ppl-m1-be/internal/usecase/transaction_usecase"
	"ppl-m1-be/pkg/httpserver"
)

type itemRequest struct {
	ProductID int64 `json:"product_id" binding:"required,gt=0"`
	Qty       int32 `json:"qty" binding:"required,gt=0,max=100000"`
}

type quotationRequest struct {
	CustomerID int64         `json:"customer_id" binding:"required,gt=0"`
	Notes      *string       `json:"notes"`
	Items      []itemRequest `json:"items" binding:"required,min=1,dive"`
}

type quotationStatusRequest struct {
	Status string `json:"status" binding:"required,oneof=disetujui ditolak"`
	Note   string `json:"note"`
}

type invoiceRequest struct {
	QuotationID int64 `json:"quotation_id" binding:"required,gt=0"`
}

type paymentRequest struct {
	Simulate string `json:"simulate" binding:"omitempty,oneof=berhasil gagal"`
}

type lineResponse struct {
	ProductID   *int64 `json:"product_id"`
	ProductName string `json:"product_name"`
	Price       int64  `json:"price"`
	Qty         int32  `json:"qty"`
	Subtotal    int64  `json:"subtotal"`
}

type quotationResponse struct {
	ID           int64   `json:"id"`
	CustomerID   int64   `json:"customer_id"`
	CustomerName string  `json:"customer_name"`
	PartnerID    *int64  `json:"partner_id"`
	PartnerName  *string `json:"partner_name"`
	Status       string  `json:"status"`
	Notes        *string `json:"notes"`
	Total        int64   `json:"total"`
	InvoiceID    *int64  `json:"invoice_id"`
	RespondedAt  *string `json:"responded_at"`
	CreatedAt    string  `json:"created_at"`
	// ResponseSource: customer atau offline (dicatat penjual).
	ResponseSource *string `json:"response_source"`
	ResponseNote   *string `json:"response_note"`
}

type quotationDetailResponse struct {
	quotationResponse
	Items []lineResponse `json:"items"`
}

type invoiceResponse struct {
	ID           int64   `json:"id"`
	QuotationID  *int64  `json:"quotation_id"`
	OrderID      *int64  `json:"order_id"`
	CustomerID   int64   `json:"customer_id"`
	CustomerName string  `json:"customer_name"`
	PartnerID    *int64  `json:"partner_id"`
	PartnerName  *string `json:"partner_name"`
	Total        int64   `json:"total"`
	Status       string  `json:"status"`
	PaidAt       *string `json:"paid_at"`
	CreatedAt    string  `json:"created_at"`
}

type paymentResponse struct {
	ID        int64  `json:"id"`
	Provider  string `json:"provider"`
	Reference string `json:"reference"`
	Amount    int64  `json:"amount"`
	Status    string `json:"status"`
	Message   string `json:"message"`
	CreatedAt string `json:"created_at"`
}

func toQuotationResponse(q sqlc.QuotationDetail) quotationResponse {
	return quotationResponse{
		ID: q.ID, CustomerID: q.CustomerID, CustomerName: q.CustomerName, PartnerID: httpserver.NullInt64Ptr(q.PartnerID),
		PartnerName: httpserver.NullStringPtr(q.PartnerName), Status: q.Status, Notes: httpserver.NullStringPtr(q.Notes),
		Total: q.Total, InvoiceID: httpserver.NullInt64Ptr(q.InvoiceID), RespondedAt: httpserver.NullTimeText(q.RespondedAt),
		CreatedAt: httpserver.FormatTime(q.CreatedAt), ResponseSource: httpserver.NullStringPtr(q.ResponseSource), ResponseNote: httpserver.NullStringPtr(q.ResponseNote),
	}
}

func toQuotationItem(i sqlc.QuotationItem) lineResponse {
	return lineResponse{httpserver.NullInt64Ptr(i.ProductID), i.ProductName, i.Price, i.Qty, i.Subtotal}
}

func toQuotationDetail(v transaction_usecase.QuotationView) quotationDetailResponse {
	return quotationDetailResponse{toQuotationResponse(v.Quotation), httpserver.Map(v.Items, toQuotationItem)}
}

func toInvoiceResponse(i sqlc.InvoiceDetail) invoiceResponse {
	return invoiceResponse{
		ID: i.ID, QuotationID: httpserver.NullInt64Ptr(i.QuotationID), OrderID: httpserver.NullInt64Ptr(i.OrderID), CustomerID: i.CustomerID, CustomerName: i.CustomerName,
		PartnerID: httpserver.NullInt64Ptr(i.PartnerID), PartnerName: httpserver.NullStringPtr(i.PartnerName),
		Total: i.Total, Status: i.Status, PaidAt: httpserver.NullTimeText(i.PaidAt), CreatedAt: httpserver.FormatTime(i.CreatedAt),
	}
}

func toPaymentResponse(p sqlc.Payment) paymentResponse {
	return paymentResponse{p.ID, p.Provider, p.Reference, p.Amount, p.Status, p.Message, httpserver.FormatTime(p.CreatedAt)}
}

func toItemInputs(items []itemRequest) []transaction_usecase.ItemInput {
	return httpserver.Map(items, func(i itemRequest) transaction_usecase.ItemInput {
		return transaction_usecase.ItemInput{ProductID: i.ProductID, Qty: i.Qty}
	})
}

func (h *Handler) ListQuotations(c *gin.Context) {
	page, ok := httpserver.RequirePage(c)
	status, okStatus := httpserver.QueryEnum(c, "status", transaction_usecase.QuotationWaiting, transaction_usecase.QuotationApproved, transaction_usecase.QuotationRejected)
	if !ok {
		return
	}
	if !okStatus {
		httpserver.Fail(c, httpserver.ErrInvalidParameter)
		return
	}
	result, err := h.uc.ListQuotations(c.Request.Context(), identity(c), status, page.Limit(), page.Offset())
	if err != nil {
		respond.Error(c, "daftar quotation gagal", err)
		return
	}
	httpserver.SuccessList(c, httpserver.Map(result.Items, toQuotationResponse), page.Meta(result.Total))
}

func (h *Handler) GetQuotation(c *gin.Context) {
	id, ok := httpserver.RequireID(c)
	if !ok {
		return
	}
	view, err := h.uc.GetQuotation(c.Request.Context(), identity(c), id)
	if err != nil {
		respond.Error(c, "ambil quotation gagal", err)
		return
	}
	httpserver.Success(c, http.StatusOK, toQuotationDetail(view))
}

func (h *Handler) CreateQuotation(c *gin.Context) {
	var req quotationRequest
	if !httpserver.BindJSON(c, &req) {
		return
	}
	view, err := h.uc.CreateQuotation(c.Request.Context(), identity(c), transaction_usecase.QuotationInput{
		CustomerID: req.CustomerID, Notes: httpserver.OptionalText(req.Notes), Items: toItemInputs(req.Items),
	})
	if err != nil {
		respond.Error(c, "buat quotation gagal", err)
		return
	}
	httpserver.Success(c, http.StatusCreated, toQuotationDetail(view))
}

func (h *Handler) RespondQuotation(c *gin.Context) {
	id, ok := httpserver.RequireID(c)
	var req quotationStatusRequest
	if !ok || !httpserver.BindJSON(c, &req) {
		return
	}
	view, err := h.uc.RespondQuotation(c.Request.Context(), identity(c), id, req.Status, req.Note)
	if err != nil {
		respond.Error(c, "ubah status quotation gagal", err)
		return
	}
	httpserver.Success(c, http.StatusOK, toQuotationDetail(view))
}

func (h *Handler) ListInvoices(c *gin.Context) {
	page, ok := httpserver.RequirePage(c)
	status, okStatus := httpserver.QueryEnum(c, "status", transaction_usecase.InvoiceUnpaid, transaction_usecase.InvoicePaid, transaction_usecase.InvoiceVoid)
	if !ok {
		return
	}
	if !okStatus {
		httpserver.Fail(c, httpserver.ErrInvalidParameter)
		return
	}
	result, err := h.uc.ListInvoices(c.Request.Context(), identity(c), status, page.Limit(), page.Offset())
	if err != nil {
		respond.Error(c, "daftar invoice gagal", err)
		return
	}
	httpserver.SuccessList(c, httpserver.Map(result.Items, toInvoiceResponse), page.Meta(result.Total))
}

func (h *Handler) CreateInvoice(c *gin.Context) {
	var req invoiceRequest
	if !httpserver.BindJSON(c, &req) {
		return
	}
	invoice, err := h.uc.CreateInvoice(c.Request.Context(), identity(c), req.QuotationID)
	if err != nil {
		respond.Error(c, "buat invoice gagal", err)
		return
	}
	httpserver.Success(c, http.StatusCreated, toInvoiceResponse(invoice))
}

// PayInvoice menjawab 201 walau gateway gagal: percobaannya tetap tercatat.
// Hasilnya dibaca dari `payment.status` dan `invoice.status`.
func (h *Handler) PayInvoice(c *gin.Context) {
	id, ok := httpserver.RequireID(c)
	var req paymentRequest
	if !ok || !httpserver.BindJSON(c, &req) {
		return
	}
	res, err := h.uc.PayInvoice(c.Request.Context(), identity(c), id, req.Simulate)
	if err != nil {
		respond.Error(c, "bayar invoice gagal", err)
		return
	}
	httpserver.Success(c, http.StatusCreated, gin.H{"payment": toPaymentResponse(res.Payment), "invoice": toInvoiceResponse(res.Invoice)})
}
