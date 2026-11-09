package middleware

import (
	"github.com/gin-gonic/gin"

	"ppl-m1-be/internal/usecase/utils"
	"ppl-m1-be/pkg/m1auth"
)

// Authenticate memverifikasi token memakai kunci yang sama dengan penerbitnya.
// Pemeriksaan selengkapnya ada di pkg/m1auth, yang juga dipakai modul lain.
func Authenticate(tokens *utils.TokenMaker) gin.HandlerFunc {
	return m1auth.Authenticate(tokens.Verifier())
}

func RequirePermission(allowed ...string) gin.HandlerFunc {
	return m1auth.RequirePermission(allowed...)
}

func RequireOrganization() gin.HandlerFunc {
	return m1auth.RequireOrganization()
}

func GetClaims(c *gin.Context) *utils.Claims {
	return m1auth.GetClaims(c)
}
