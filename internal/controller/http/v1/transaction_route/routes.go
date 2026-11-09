package transaction_route

import (
	"github.com/gin-gonic/gin"

	"ppl-m1-be/internal/usecase/transaction_usecase"
	"ppl-m1-be/internal/usecase/utils"
	"ppl-m1-be/middleware"
)

// Mount mendaftarkan endpoint M4. Izin menentukan siapa yang boleh memanggil;
// cakupan data (milik partner/customer) diperiksa di usecase.
func Mount(rg *gin.RouterGroup, uc *transaction_usecase.Usecase) {
	handler := &Handler{uc: uc}
	mountMasterData(rg, handler)
	mountBilling(rg, handler)

	read := middleware.RequirePermission(utils.PermOrdersRead, utils.PermOrdersUpdate)
	rg.GET("/orders", read, handler.ListOrders)
	rg.GET("/orders/:id", read, handler.GetOrder)
	// Customer memakai izin baca untuk membatalkan Order miliknya; memajukan
	// status tetap ditolak di usecase.
	rg.PATCH("/orders/:id/status", read, handler.UpdateOrderStatus)
}

func mountMasterData(rg *gin.RouterGroup, handler *Handler) {
	products := rg.Group("/products", middleware.RequirePermission(utils.PermProductsManage))
	products.GET("", handler.ListProducts)
	products.POST("", handler.CreateProduct)
	products.PATCH("/:id", handler.UpdateProduct)
	products.PATCH("/:id/status", handler.SetProductStatus)

	customers := rg.Group("/customers", middleware.RequirePermission(utils.PermCustomersManage))
	customers.GET("", handler.ListCustomers)
	customers.POST("", handler.CreateCustomer)
	customers.PATCH("/:id", handler.UpdateCustomer)
}

func mountBilling(rg *gin.RouterGroup, handler *Handler) {
	readQuotations := middleware.RequirePermission(utils.PermQuotationsManage, utils.PermQuotationsRespond)
	rg.GET("/quotations", readQuotations, handler.ListQuotations)
	rg.GET("/quotations/:id", readQuotations, handler.GetQuotation)
	rg.POST("/quotations", middleware.RequirePermission(utils.PermQuotationsManage), handler.CreateQuotation)
	rg.PATCH("/quotations/:id/status", middleware.RequirePermission(utils.PermQuotationsRespond), handler.RespondQuotation)

	rg.GET("/invoices", middleware.RequirePermission(utils.PermInvoicesManage, utils.PermInvoicesPay), handler.ListInvoices)
	rg.POST("/invoices", middleware.RequirePermission(utils.PermInvoicesManage), handler.CreateInvoice)
	rg.POST("/invoices/:id/payments", middleware.RequirePermission(utils.PermInvoicesPay), handler.PayInvoice)
}
