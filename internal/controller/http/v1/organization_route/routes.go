package organization_route

import (
	"github.com/gin-gonic/gin"

	"ppl-m1-be/internal/usecase/organization_usecase"
	"ppl-m1-be/internal/usecase/utils"
	"ppl-m1-be/middleware"
)

// Mount mendaftarkan endpoint organisasi khusus super admin platform.
func Mount(rg *gin.RouterGroup, uc *organization_usecase.Usecase, authenticated gin.HandlerFunc) {
	handler := &Handler{uc: uc}
	platform := rg.Group("/organizations", authenticated, middleware.RequirePermission(utils.PermOrganizationsManage))

	platform.GET("", handler.List)
	platform.POST("", handler.Create)
	platform.GET("/:id", handler.Get)
	platform.PATCH("/:id", handler.Rename)
	platform.PATCH("/:id/status", handler.SetStatus)
	platform.DELETE("/:id", handler.Deactivate)
	platform.DELETE("/:id/permanent", handler.Delete)
}
