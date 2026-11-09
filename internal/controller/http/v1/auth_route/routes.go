package auth_route

import (
	"github.com/gin-gonic/gin"

	"ppl-m1-be/internal/usecase/auth_usecase"
)

// Mount mendaftarkan endpoint autentikasi. `authenticated` adalah middleware
// pemeriksa Bearer token; /me dan change-password terbuka untuk semua
// pengguna yang login, termasuk super admin.
func Mount(rg *gin.RouterGroup, uc *auth_usecase.Usecase, authenticated gin.HandlerFunc) {
	handler := &Handler{uc: uc}

	rg.POST("/auth/login", handler.Login)
	rg.POST("/auth/change-password", authenticated, handler.ChangePassword)
	rg.GET("/me", authenticated, handler.Me)
}
