package user_route

import (
	"github.com/gin-gonic/gin"

	"ppl-m1-be/internal/usecase/user_usecase"
	"ppl-m1-be/internal/usecase/utils"
	"ppl-m1-be/middleware"
)

// Mount mendaftarkan endpoint pengguna pada grup milik organisasi.
func Mount(org *gin.RouterGroup, uc *user_usecase.Usecase) {
	handler := &Handler{uc: uc}
	read := middleware.RequirePermission(utils.PermUsersRead, utils.PermUsersManage)
	manage := middleware.RequirePermission(utils.PermUsersManage)

	org.GET("/users", read, handler.List)
	org.GET("/users/:id", read, handler.Get)
	org.POST("/users", manage, handler.Create)
	org.PATCH("/users/:id", manage, handler.Update)
	org.PATCH("/users/:id/status", manage, handler.SetStatus)
	org.DELETE("/users/:id", manage, handler.Deactivate)
	org.DELETE("/users/:id/permanent", manage, handler.Delete)
}
