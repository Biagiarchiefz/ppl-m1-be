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

	rg.POST("/organization-applications", handler.Apply)
	rg.POST("/organization-application-lookups", handler.LookupApplication)
	review := rg.Group("/organization-applications", authenticated, middleware.RequirePermission(utils.PermOrganizationsManage))
	review.GET("", handler.ListApplications)
	review.PATCH("/:id/status", handler.ReviewApplication)
}
