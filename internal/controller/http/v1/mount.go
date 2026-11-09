package v1

import (
	"database/sql"

	"github.com/gin-gonic/gin"

	"ppl-m1-be/config"
	"ppl-m1-be/internal/controller/http/v1/audit_route"
	"ppl-m1-be/internal/controller/http/v1/auth_route"
	"ppl-m1-be/internal/controller/http/v1/division_route"
	"ppl-m1-be/internal/controller/http/v1/organization_route"
	"ppl-m1-be/internal/controller/http/v1/role_route"
	"ppl-m1-be/internal/controller/http/v1/user_route"
	"ppl-m1-be/internal/usecase/audit_usecase"
	"ppl-m1-be/internal/usecase/auth_usecase"
	"ppl-m1-be/internal/usecase/division_usecase"
	"ppl-m1-be/internal/usecase/organization_usecase"
	"ppl-m1-be/internal/usecase/repository"
	"ppl-m1-be/internal/usecase/role_usecase"
	"ppl-m1-be/internal/usecase/user_usecase"
	"ppl-m1-be/internal/usecase/utils"
	"ppl-m1-be/middleware"
)

// mountAPI: login publik; /me dan change-password untuk semua pengguna
// login; /organizations dan /platform/audit-logs khusus super admin;
// sisanya data milik organisasi pemanggil, dijaga per izin di tiap route.
func mountAPI(api *gin.RouterGroup, db *sql.DB, cfg config.Config) {
	store := repository.NewStore(db)
	tokens := utils.NewTokenMaker(cfg.JWTSecret, cfg.AppName, cfg.JWTExpiresIn)
	authenticated := middleware.Authenticate(tokens)
	org := api.Group("", authenticated, middleware.RequireOrganization())

	auth_route.Mount(api, auth_usecase.New(store, tokens), authenticated)
	organization_route.Mount(api, organization_usecase.New(store), authenticated)
	user_route.Mount(org, user_usecase.New(store))
	role_route.Mount(org, role_usecase.New(store))
	division_route.Mount(org, division_usecase.New(store))
	audit := audit_usecase.New(store)
	audit_route.Mount(org, audit)
	audit_route.MountPlatform(api, audit, authenticated)
}
