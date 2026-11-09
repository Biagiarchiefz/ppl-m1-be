package audit_route

import (
	"github.com/gin-gonic/gin"

	"ppl-m1-be/internal/usecase/audit_usecase"
	"ppl-m1-be/internal/usecase/utils"
	"ppl-m1-be/middleware"
)

// Mount mendaftarkan GET /audit-logs pada grup milik organisasi.
func Mount(org *gin.RouterGroup, uc *audit_usecase.Usecase) {
	handler := &Handler{uc: uc}
	org.GET("/audit-logs", middleware.RequirePermission(utils.PermAuditRead), handler.List)
}

// MountPlatform mendaftarkan GET /platform/audit-logs khusus super admin:
// log semua organisasi, termasuk organisasi yang sudah dihapus permanen.
func MountPlatform(rg *gin.RouterGroup, uc *audit_usecase.Usecase, authenticated gin.HandlerFunc) {
	handler := &Handler{uc: uc}
	rg.GET("/platform/audit-logs", authenticated,
		middleware.RequirePermission(utils.PermOrganizationsManage), handler.PlatformList)
}
