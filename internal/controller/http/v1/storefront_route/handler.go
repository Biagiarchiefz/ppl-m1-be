package storefront_route

import (
	"database/sql"
	"net/http"

	"github.com/gin-gonic/gin"

	"ppl-m1-be/internal/controller/http/v1/business_unit_route"
	"ppl-m1-be/internal/controller/http/v1/respond"
	"ppl-m1-be/internal/controller/http/v1/transaction_route"
	"ppl-m1-be/internal/usecase/repository/sqlc"
	"ppl-m1-be/internal/usecase/storefront_usecase"
	"ppl-m1-be/middleware"
	"ppl-m1-be/pkg/httpserver"
)

type Handler struct {
	uc *storefront_usecase.Usecase
}

type addCartRequest struct {
	ProductID int64 `json:"product_id" binding:"required,gt=0"`
	Qty       int32 `json:"qty" binding:"omitempty,gt=0,max=100000"`
}

type qtyRequest struct {
	Qty int32 `json:"qty" binding:"required,gt=0,max=100000"`
}

type serviceRequestBody struct {
	BusinessUnitID int64  `json:"business_unit_id" binding:"required,gt=0"`
	ContactName    string `json:"contact_name" binding:"required,notblank,max=150"`
	ContactEmail   string `json:"contact_email" binding:"required,email,max=255"`
	ContactPhone   string `json:"contact_phone" binding:"required,notblank,max=30"`
	Subject        string `json:"subject" binding:"required,notblank,max=150"`
	Message        string `json:"message" binding:"required,notblank"`
}

type cartItemResponse struct {
	ProductID        int64   `json:"product_id"`
	Name             string  `json:"name"`
	Price            int64   `json:"price"`
	Qty              int32   `json:"qty"`
	Subtotal         int64   `json:"subtotal"`
	BusinessUnitName *string `json:"business_unit_name"`
	PartnerName      *string `json:"partner_name"`
	Stock            *int32  `json:"stock"`
	Available        bool    `json:"available"`
}

type cartResponse struct {
	Items []cartItemResponse `json:"items"`
	Count int64              `json:"count"`
	Total int64              `json:"total"`
}

type serviceRequestResponse struct {
	ID               int64  `json:"id"`
	BusinessUnitID   int64  `json:"business_unit_id"`
	BusinessUnitName string `json:"business_unit_name"`
	ContactName      string `json:"contact_name"`
	ContactEmail     string `json:"contact_email"`
	ContactPhone     string `json:"contact_phone"`
	Subject          string `json:"subject"`
	Message          string `json:"message"`
	Status           string `json:"status"`
	CreatedAt        string `json:"created_at"`
	StatusChangedAt  string `json:"status_changed_at"`
}

func toCartItem(i sqlc.ListCartItemsRow) cartItemResponse {
	return cartItemResponse{
		ProductID: i.ProductID, Name: i.Name, Price: i.Price, Qty: i.Qty, Subtotal: i.Price * int64(i.Qty),
		BusinessUnitName: httpserver.NullStringPtr(i.BusinessUnitName), PartnerName: httpserver.NullStringPtr(i.PartnerName),
		Stock: httpserver.NullInt32Ptr(i.Stock), Available: i.Available,
	}
}

func toCart(cart storefront_usecase.Cart) cartResponse {
	return cartResponse{httpserver.Map(cart.Items, toCartItem), cart.Count, cart.Total}
}

func toServiceRequest(s sqlc.ServiceRequestDetail) serviceRequestResponse {
	return serviceRequestResponse{
		ID: s.ID, BusinessUnitID: s.BusinessUnitID, BusinessUnitName: s.BusinessUnitName, ContactName: s.ContactName,
		ContactEmail: s.ContactEmail, ContactPhone: s.ContactPhone, Subject: s.Subject, Message: s.Message,
		Status: s.Status, CreatedAt: httpserver.FormatTime(s.CreatedAt), StatusChangedAt: httpserver.FormatTime(s.StatusChangedAt),
	}
}

func (h *Handler) BusinessUnits(c *gin.Context) {
	units, err := h.uc.BusinessUnits(c.Request.Context(), middleware.Actor(c).OrganizationID)
	if err != nil {
		httpserver.Internal(c, "katalog unit usaha gagal", err)
		return
	}
	httpserver.Success(c, http.StatusOK, httpserver.Map(units, business_unit_route.ToResponse))
}

func (h *Handler) Categories(c *gin.Context) {
	categories, err := h.uc.Categories(c.Request.Context(), middleware.Actor(c).OrganizationID)
	if err != nil {
		httpserver.Internal(c, "kategori katalog gagal", err)
		return
	}
	httpserver.Success(c, http.StatusOK, httpserver.Map(categories, func(s string) string { return s }))
}

func catalogFilter(c *gin.Context) (storefront_usecase.CatalogFilter, bool) {
	unit, okUnit := httpserver.QueryInt64(c, "business_unit_id")
	partner, okPartner := httpserver.QueryInt64(c, "partner_id")
	return storefront_usecase.CatalogFilter{
		Q: httpserver.QueryString(c, "q"), Category: httpserver.QueryString(c, "category"), BusinessUnitID: unit, PartnerID: partner,
	}, okUnit && okPartner
}

// Products mengembalikan `data: []` bila tidak ada hasil; bukan galat.
func (h *Handler) Products(c *gin.Context) {
	page, ok := httpserver.RequirePage(c)
	if !ok {
		return
	}
	filter, ok := catalogFilter(c)
	if !ok {
		httpserver.Fail(c, httpserver.ErrInvalidParameter)
		return
	}
	result, err := h.uc.Products(c.Request.Context(), middleware.Actor(c).OrganizationID, filter, page.Limit(), page.Offset())
	if err != nil {
		httpserver.Internal(c, "katalog produk gagal", err)
		return
	}
	httpserver.SuccessList(c, httpserver.Map(result.Items, transaction_route.ToProductResponse), page.Meta(result.Total))
}

func (h *Handler) GetCart(c *gin.Context) {
	cart, err := h.uc.GetCart(c.Request.Context(), middleware.Actor(c).UserID)
	if err != nil {
		respond.Error(c, "ambil keranjang gagal", err)
		return
	}
	httpserver.Success(c, http.StatusOK, toCart(cart))
}

func (h *Handler) AddToCart(c *gin.Context) {
	var req addCartRequest
	if !httpserver.BindJSON(c, &req) {
		return
	}
	if req.Qty == 0 {
		req.Qty = 1
	}
	cart, err := h.uc.AddToCart(c.Request.Context(), middleware.Actor(c), req.ProductID, req.Qty)
	if err != nil {
		respond.Error(c, "tambah keranjang gagal", err)
		return
	}
	httpserver.Success(c, http.StatusOK, toCart(cart))
}

func (h *Handler) SetCartQty(c *gin.Context) {
	id, ok := httpserver.RequireID(c)
	var req qtyRequest
	if !ok || !httpserver.BindJSON(c, &req) {
		return
	}
	cart, err := h.uc.SetCartQty(c.Request.Context(), middleware.Actor(c).UserID, id, req.Qty)
	if err != nil {
		respond.Error(c, "ubah keranjang gagal", err)
		return
	}
	httpserver.Success(c, http.StatusOK, toCart(cart))
}

func (h *Handler) RemoveFromCart(c *gin.Context) {
	id, ok := httpserver.RequireID(c)
	if !ok {
		return
	}
	cart, err := h.uc.RemoveFromCart(c.Request.Context(), middleware.Actor(c).UserID, id)
	if err != nil {
		respond.Error(c, "hapus item keranjang gagal", err)
		return
	}
	httpserver.Success(c, http.StatusOK, toCart(cart))
}

func (h *Handler) Checkout(c *gin.Context) {
	orderIDs, err := h.uc.Checkout(c.Request.Context(), middleware.Actor(c))
	if err != nil {
		respond.Error(c, "checkout gagal", err)
		return
	}
	httpserver.Success(c, http.StatusCreated, gin.H{"order_ids": orderIDs})
}

func (h *Handler) SendServiceRequest(c *gin.Context) {
	var req serviceRequestBody
	if !httpserver.BindJSON(c, &req) {
		return
	}
	created, err := h.uc.SendServiceRequest(c.Request.Context(), middleware.Actor(c), storefront_usecase.ServiceRequestInput(req))
	if err != nil {
		respond.Error(c, "kirim permintaan layanan gagal", err)
		return
	}
	httpserver.Success(c, http.StatusCreated, gin.H{"id": created.ID, "status": created.Status, "created_at": httpserver.FormatTime(created.CreatedAt)})
}

// serviceRequestUnit membatasi staf unit usaha ke unitnya sendiri; pengelola
// tanpa unit boleh memilih lewat filter.
func serviceRequestUnit(c *gin.Context) (sql.NullInt64, bool) {
	if unit := middleware.GetClaims(c).UnitUsahaID; unit != nil {
		return sql.NullInt64{Int64: *unit, Valid: true}, true
	}
	return httpserver.QueryInt64(c, "business_unit_id")
}

func (h *Handler) ServiceRequests(c *gin.Context) {
	page, ok := httpserver.RequirePage(c)
	if !ok {
		return
	}
	unit, ok := serviceRequestUnit(c)
	if !ok {
		httpserver.Fail(c, httpserver.ErrInvalidParameter)
		return
	}
	result, err := h.uc.ServiceRequests(c.Request.Context(), middleware.Actor(c).OrganizationID, unit, page.Limit(), page.Offset())
	if err != nil {
		httpserver.Internal(c, "daftar permintaan layanan gagal", err)
		return
	}
	httpserver.SuccessList(c, httpserver.Map(result.Items, toServiceRequest), page.Meta(result.Total))
}
