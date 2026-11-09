package division_route

import (
	"github.com/gin-gonic/gin"

	"ppl-m1-be/internal/usecase/division_usecase"
	"ppl-m1-be/internal/usecase/utils"
	"ppl-m1-be/middleware"
)

// Mount mendaftarkan endpoint divisi. Daftar divisi juga boleh dibaca
// pengelola pengguna (untuk dropdown form akun).
func Mount(org *gin.RouterGroup, uc *division_usecase.Usecase) {
	handler := &Handler{uc: uc}
	read := middleware.RequirePermission(utils.PermDivisionsManage, utils.PermUsersManage, utils.PermUsersRead)
	manage := middleware.RequirePermission(utils.PermDivisionsManage)

	org.GET("/divisions", read, handler.List)
	org.GET("/divisions/:id", read, handler.Get)
	org.POST("/divisions", manage, handler.Create)
	org.PATCH("/divisions/:id", manage, handler.Rename)
	org.PATCH("/divisions/:id/status", manage, handler.SetStatus)
	org.DELETE("/divisions/:id", manage, handler.Deactivate)
}
