package role_route

import (
	"github.com/gin-gonic/gin"

	"ppl-m1-be/internal/usecase/role_usecase"
	"ppl-m1-be/internal/usecase/utils"
	"ppl-m1-be/middleware"
)

// Mount mendaftarkan endpoint peran dan katalog izin. Daftar peran juga
// boleh dibaca pengelola pengguna (untuk dropdown form akun).
func Mount(org *gin.RouterGroup, uc *role_usecase.Usecase) {
	handler := &Handler{uc: uc}
	read := middleware.RequirePermission(utils.PermRolesManage, utils.PermUsersManage, utils.PermUsersRead)
	manage := middleware.RequirePermission(utils.PermRolesManage)

	org.GET("/roles", read, handler.List)
	org.GET("/roles/:id", read, handler.Get)
	org.POST("/roles", manage, handler.Create)
	org.PATCH("/roles/:id", manage, handler.Update)
	org.PATCH("/roles/:id/status", manage, handler.SetStatus)
	org.DELETE("/roles/:id", manage, handler.Deactivate)
	org.GET("/permissions", manage, handler.Permissions)
}
