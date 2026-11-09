package v1

import (
	"database/sql"

	"github.com/gin-gonic/gin"

	"ppl-m1-be/config"
	"ppl-m1-be/internal/controller/http/v1/business_unit_route"
	"ppl-m1-be/internal/controller/http/v1/partner_route"
	"ppl-m1-be/internal/controller/http/v1/storefront_route"
	"ppl-m1-be/internal/controller/http/v1/transaction_route"
	"ppl-m1-be/internal/usecase/business_unit_usecase"
	"ppl-m1-be/internal/usecase/partner_usecase"
	"ppl-m1-be/internal/usecase/repository"
	"ppl-m1-be/internal/usecase/storefront_usecase"
	"ppl-m1-be/internal/usecase/transaction_usecase"
	"ppl-m1-be/internal/usecase/utils"
	"ppl-m1-be/middleware"
)

// mountModules mendaftarkan modul M2-M5, masing-masing di /api/{modul}/v1.
// Semuanya memakai token dari login terpusat M1.
func mountModules(router *gin.Engine, db *sql.DB, cfg config.Config) {
	store := repository.NewStore(db)
	tokens := utils.NewTokenMaker(cfg.JWTSecret, cfg.AppName, cfg.JWTExpiresIn)
	guards := []gin.HandlerFunc{middleware.Authenticate(tokens), middleware.RequireOrganization()}

	business_unit_route.Mount(router.Group("/api/m2/v1", guards...), business_unit_usecase.New(store))
	partner_route.Mount(router.Group("/api/m3/v1"), router.Group("/api/m3/v1", guards...), partner_usecase.New(store))
	transaction_route.Mount(router.Group("/api/m4/v1", guards...), transaction_usecase.New(store))
	storefront_route.Mount(router.Group("/api/m5/v1"), router.Group("/api/m5/v1", guards...), storefront_usecase.New(store))
}
