package transaction_route

import (
	"database/sql"
	"net/http"

	"github.com/gin-gonic/gin"

	"ppl-m1-be/internal/controller/http/v1/respond"
	"ppl-m1-be/internal/usecase/repository/sqlc"
	"ppl-m1-be/internal/usecase/transaction_usecase"
	"ppl-m1-be/middleware"
	"ppl-m1-be/pkg/httpserver"
)

type Handler struct {
	uc *transaction_usecase.Usecase
}

func identity(c *gin.Context) transaction_usecase.Identity {
	claims := middleware.GetClaims(c)
	return transaction_usecase.Identity{Actor: middleware.Actor(c), Role: claims.Role, Permissions: claims.Permissions, UnitUsahaID: claims.UnitUsahaID}
}

type productRequest struct {
	BusinessUnitID *int64  `json:"business_unit_id" binding:"omitempty,gt=0"`
	Stock          *int32  `json:"stock" binding:"omitempty,gte=0,max=1000000"`
	Type           string  `json:"type" binding:"required,oneof=produk jasa"`
	Name           string  `json:"name" binding:"required,notblank,max=150"`
	Category       string  `json:"category" binding:"required,notblank,max=100"`
	Description    *string `json:"description"`
	Price          *int64  `json:"price" binding:"required,gte=0,max=1000000000000"`
}

// ProductResponse dipakai juga oleh katalog M5.
type ProductResponse struct {
	ID               int64   `json:"id"`
	BusinessUnitID   *int64  `json:"business_unit_id"`
	BusinessUnitName *string `json:"business_unit_name"`
	Stock            *int32  `json:"stock"`
	PartnerID        *int64  `json:"partner_id"`
	PartnerName      *string `json:"partner_name"`
	Type             string  `json:"type"`
	Name             string  `json:"name"`
	Category         string  `json:"category"`
	Description      *string `json:"description"`
	Price            int64   `json:"price"`
	IsActive         bool    `json:"is_active"`
	CreatedAt        string  `json:"created_at"`
}

func ToProductResponse(p sqlc.ProductDetail) ProductResponse {
	return ProductResponse{
		ID: p.ID, BusinessUnitID: httpserver.NullInt64Ptr(p.BusinessUnitID), BusinessUnitName: httpserver.NullStringPtr(p.BusinessUnitName),
		Stock: httpserver.NullInt32Ptr(p.Stock), PartnerID: httpserver.NullInt64Ptr(p.PartnerID), PartnerName: httpserver.NullStringPtr(p.PartnerName),
		Type: p.Type, Name: p.Name, Category: p.Category, Description: httpserver.NullStringPtr(p.Description),
		Price: p.Price, IsActive: p.IsActive, CreatedAt: httpserver.FormatTime(p.CreatedAt),
	}
}

func (r productRequest) toInput() transaction_usecase.ProductInput {
	input := transaction_usecase.ProductInput{
		BusinessUnitID: httpserver.OptionalID(r.BusinessUnitID), Type: r.Type, Name: r.Name, Category: r.Category,
		Description: httpserver.OptionalText(r.Description), Price: *r.Price, Stock: httpserver.OptionalInt32(r.Stock),
	}
	if r.Type == "jasa" {
		input.Stock = sql.NullInt32{}
	}
	return input
}

func (h *Handler) ListProducts(c *gin.Context) {
	page, ok := httpserver.RequirePage(c)
	if !ok {
		return
	}
	result, err := h.uc.ListProducts(c.Request.Context(), identity(c), httpserver.QueryString(c, "q"), page.Limit(), page.Offset())
	if err != nil {
		respond.Error(c, "daftar produk gagal", err)
		return
	}
	httpserver.SuccessList(c, httpserver.Map(result.Items, ToProductResponse), page.Meta(result.Total))
}

func (h *Handler) CreateProduct(c *gin.Context) {
	var req productRequest
	if !httpserver.BindJSON(c, &req) {
		return
	}
	product, err := h.uc.CreateProduct(c.Request.Context(), identity(c), req.toInput())
	if err != nil {
		respond.Error(c, "tambah produk gagal", err)
		return
	}
	httpserver.Success(c, http.StatusCreated, ToProductResponse(product))
}

func (h *Handler) UpdateProduct(c *gin.Context) {
	id, ok := httpserver.RequireID(c)
	var req productRequest
	if !ok || !httpserver.BindJSON(c, &req) {
		return
	}
	product, err := h.uc.UpdateProduct(c.Request.Context(), identity(c), id, req.toInput())
	if err != nil {
		respond.Error(c, "ubah produk gagal", err)
		return
	}
	httpserver.Success(c, http.StatusOK, ToProductResponse(product))
}

func (h *Handler) SetProductStatus(c *gin.Context) {
	ctx, id, active, ok := httpserver.StatusUpdate(c)
	if !ok {
		return
	}
	if err := h.uc.SetProductStatus(ctx, identity(c), id, active); err != nil {
		respond.Error(c, "ubah status produk gagal", err)
		return
	}
	httpserver.StatusUpdated(c, id, active)
}

type customerRequest struct {
	Name    string  `json:"name" binding:"required,notblank,max=150"`
	Email   string  `json:"email" binding:"required,email,max=255"`
	Phone   string  `json:"phone" binding:"required,notblank,max=30"`
	Address *string `json:"address"`
}

type customerResponse struct {
	ID        int64   `json:"id"`
	Name      string  `json:"name"`
	Email     string  `json:"email"`
	Phone     *string `json:"phone"`
	Address   *string `json:"address"`
	HasLogin  bool    `json:"has_login"`
	CreatedAt string  `json:"created_at"`
}

func toCustomerResponse(v sqlc.Customer) customerResponse {
	return customerResponse{
		ID: v.ID, Name: v.Name, Email: v.Email, Phone: httpserver.NullStringPtr(v.Phone),
		Address: httpserver.NullStringPtr(v.Address), HasLogin: v.UserID.Valid, CreatedAt: httpserver.FormatTime(v.CreatedAt),
	}
}

func (r customerRequest) toInput() transaction_usecase.CustomerInput {
	return transaction_usecase.CustomerInput{
		Name: r.Name, Email: r.Email, Phone: httpserver.OptionalText(&r.Phone), Address: httpserver.OptionalText(r.Address),
	}
}

func (h *Handler) ListCustomers(c *gin.Context) {
	page, ok := httpserver.RequirePage(c)
	if !ok {
		return
	}
	result, err := h.uc.ListCustomers(c.Request.Context(), identity(c), httpserver.QueryString(c, "q"), page.Limit(), page.Offset())
	if err != nil {
		respond.Error(c, "daftar pelanggan gagal", err)
		return
	}
	httpserver.SuccessList(c, httpserver.Map(result.Items, toCustomerResponse), page.Meta(result.Total))
}

func (h *Handler) CreateCustomer(c *gin.Context) {
	var req customerRequest
	if !httpserver.BindJSON(c, &req) {
		return
	}
	customer, err := h.uc.CreateCustomer(c.Request.Context(), identity(c), req.toInput())
	if err != nil {
		respond.Error(c, "tambah pelanggan gagal", err)
		return
	}
	httpserver.Success(c, http.StatusCreated, toCustomerResponse(customer))
}

func (h *Handler) UpdateCustomer(c *gin.Context) {
	id, ok := httpserver.RequireID(c)
	var req customerRequest
	if !ok || !httpserver.BindJSON(c, &req) {
		return
	}
	customer, err := h.uc.UpdateCustomer(c.Request.Context(), identity(c), id, req.toInput())
	if err != nil {
		respond.Error(c, "ubah pelanggan gagal", err)
		return
	}
	httpserver.Success(c, http.StatusOK, toCustomerResponse(customer))
}
