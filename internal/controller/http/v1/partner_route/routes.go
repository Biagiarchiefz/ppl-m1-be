package partner_route

import (
	"github.com/gin-gonic/gin"

	"ppl-m1-be/internal/usecase/partner_usecase"
	"ppl-m1-be/internal/usecase/utils"
	"ppl-m1-be/middleware"
)

// Mount mendaftarkan endpoint M3. Pendaftaran publik; sisanya butuh login.
func Mount(public, org *gin.RouterGroup, uc *partner_usecase.Usecase) {
	handler := &Handler{uc: uc}
	manage := middleware.RequirePermission(utils.PermPartnersManage)

	public.POST("/partner-registrations", handler.Register)
	public.GET("/partner-registrations/organizations/:code", handler.RegistrationOrganization)
	org.GET("/partner-registration-link", manage, handler.RegistrationLink)
	org.GET("/verified-partners", handler.ListVerified)
	org.GET("/partners/me", middleware.RequirePermission(utils.PermPartnerStatusRead), handler.Mine)
	org.POST("/partners/me/resubmissions", middleware.RequirePermission(utils.PermPartnerStatusRead), handler.Resubmit)
	org.GET("/partners", manage, handler.List)
	org.GET("/partners/:id", manage, handler.Get)
	org.PATCH("/partners/:id/status", manage, handler.SetStatus)
}
