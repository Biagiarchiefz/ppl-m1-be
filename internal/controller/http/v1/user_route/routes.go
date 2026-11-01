package user_route

import (
	"github.com/gin-gonic/gin"

	"ppl-m1-be/internal/usecase/user_usecase"
)

func Mount(rg *gin.RouterGroup, uc *user_usecase.Usecase) {
	handler := &Handler{uc: uc}
	rg.POST("/auth/registrations", handler.Register)
}