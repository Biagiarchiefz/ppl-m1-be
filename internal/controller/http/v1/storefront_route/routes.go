package storefront_route

import (
	"github.com/gin-gonic/gin"

	"ppl-m1-be/internal/usecase/storefront_usecase"
	"ppl-m1-be/internal/usecase/utils"
	"ppl-m1-be/middleware"
)

// Mount mendaftarkan endpoint M5. Toko dan pendaftaran customer publik;
// katalog di portal terbuka untuk anggota organisasi.
func Mount(public, rg *gin.RouterGroup, uc *storefront_usecase.Usecase) {
	handler := &Handler{uc: uc}
	mountPublic(public, handler)

	rg.GET("/catalog/business-units", handler.BusinessUnits)
	rg.GET("/catalog/categories", handler.Categories)
	rg.GET("/catalog/products", handler.Products)

	cart := rg.Group("", middleware.RequirePermission(utils.PermCartManage))
	cart.GET("/cart", handler.GetCart)
	cart.POST("/cart/items", handler.AddToCart)
	cart.PATCH("/cart/items/:id", handler.SetCartQty)
	cart.DELETE("/cart/items/:id", handler.RemoveFromCart)
	cart.POST("/checkouts", handler.Checkout)

	mountServiceRequests(rg, handler)
}

func mountPublic(public *gin.RouterGroup, handler *Handler) {
	public.GET("/stores/:code", handler.Store)
	public.GET("/stores/:code/products", handler.StoreProducts)
	public.POST("/customer-registrations", handler.RegisterCustomer)
	public.GET("/customer-registrations/organizations/:code", handler.RegistrationOrganization)
}

func mountServiceRequests(rg *gin.RouterGroup, handler *Handler) {
	read := middleware.RequirePermission(utils.PermServiceRequestsRead)
	send := middleware.RequirePermission(utils.PermServiceRequestsSend)
	rg.POST("/service-requests", send, handler.SendServiceRequest)
	rg.GET("/service-requests", read, handler.ServiceRequests)
	rg.GET("/service-requests/mine", send, handler.MyServiceRequests)
	rg.PATCH("/service-requests/:id/status", read, handler.SetServiceRequestStatus)
}
