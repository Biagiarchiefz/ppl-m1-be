package storefront_route

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"ppl-m1-be/internal/controller/http/v1/respond"
	"ppl-m1-be/internal/controller/http/v1/transaction_route"
	"ppl-m1-be/internal/usecase/repository/sqlc"
	"ppl-m1-be/pkg/httpserver"
)

// Data toko bersifat publik, jadi hanya memuat profil yang pantas dilihat umum.
type storeUnit struct {
	ID          int64   `json:"id"`
	Name        string  `json:"name"`
	Category    string  `json:"category"`
	Description *string `json:"description"`
	Address     *string `json:"address"`
	Phone       *string `json:"phone"`
}

type storePartner struct {
	ID           int64   `json:"id"`
	BusinessName string  `json:"business_name"`
	BusinessType string  `json:"business_type"`
	Description  *string `json:"description"`
}

func toStoreUnit(b sqlc.BusinessUnitDetail) storeUnit {
	return storeUnit{b.ID, b.Name, b.Category, httpserver.NullStringPtr(b.Description), httpserver.NullStringPtr(b.Address), httpserver.NullStringPtr(b.Phone)}
}

func toStorePartner(p sqlc.Partner) storePartner {
	return storePartner{p.ID, p.BusinessName, p.BusinessType, httpserver.NullStringPtr(p.Description)}
}

// Store menjawab profil toko; `id` dipakai frontend untuk mengenali apakah
// akun yang login milik organisasi toko ini.
func (h *Handler) Store(c *gin.Context) {
	store, err := h.uc.Store(c.Request.Context(), c.Param("code"))
	if err != nil {
		respond.Error(c, "ambil toko gagal", err)
		return
	}
	httpserver.Success(c, http.StatusOK, gin.H{
		"id": store.Organization.ID, "code": store.Organization.Code, "name": store.Organization.Name,
		"business_units": httpserver.Map(store.Units, toStoreUnit), "partners": httpserver.Map(store.Partners, toStorePartner),
		"categories": httpserver.Map(store.Categories, func(s string) string { return s }),
	})
}

func (h *Handler) StoreProducts(c *gin.Context) {
	page, ok := httpserver.RequirePage(c)
	if !ok {
		return
	}
	filter, ok := catalogFilter(c)
	if !ok {
		httpserver.Fail(c, httpserver.ErrInvalidParameter)
		return
	}
	result, err := h.uc.StoreProducts(c.Request.Context(), c.Param("code"), filter, page.Limit(), page.Offset())
	if err != nil {
		respond.Error(c, "ambil produk toko gagal", err)
		return
	}
	httpserver.SuccessList(c, httpserver.Map(result.Items, transaction_route.ToProductResponse), page.Meta(result.Total))
}
