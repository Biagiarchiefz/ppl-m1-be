package v1

import (
	"database/sql"

	"github.com/gin-gonic/gin"

	"ppl-m1-be/internal/controller/http/v1/user_route"
	"ppl-m1-be/internal/usecase/audit_usecase"
	"ppl-m1-be/internal/usecase/repository/sqlc"
	"ppl-m1-be/internal/usecase/user_usecase"
)

func mountUser(api *gin.RouterGroup, db *sql.DB) {
	uc := user_usecase.New(sqlc.New(db), audit_usecase.NewStdLogger())
	user_route.Mount(api, uc)
}