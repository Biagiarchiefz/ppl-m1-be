package business_unit_route

import (
	"github.com/gin-gonic/gin"

	"ppl-m1-be/internal/usecase/business_unit_usecase"
	"ppl-m1-be/internal/usecase/utils"
	"ppl-m1-be/middleware"
)

// Mount mendaftarkan endpoint M2. Daftar unit usaha terbuka untuk semua
// anggota organisasi karena dipakai sebagai pilihan di modul lain.
func Mount(rg *gin.RouterGroup, uc *business_unit_usecase.Usecase) {
	handler := &Handler{uc: uc}
	manage := middleware.RequirePermission(utils.PermBusinessUnitsManage)

	rg.GET("/business-units", handler.List)
	rg.GET("/business-units/:id", handler.Get)
	rg.POST("/business-units", manage, handler.Create)
	rg.PATCH("/business-units/:id", manage, handler.Update)
	rg.PATCH("/business-units/:id/status", manage, handler.SetStatus)

	rg.GET("/dashboard", middleware.RequirePermission(utils.PermDashboardRead), handler.Report)
	rg.POST("/dashboard/metrics", middleware.RequirePermission(utils.PermDashboardWrite), handler.SubmitMetric)
}
