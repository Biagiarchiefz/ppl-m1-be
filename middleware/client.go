package middleware

import (
	"github.com/gin-gonic/gin"

	"ppl-m1-be/internal/usecase/audit_usecase"
)

// Client mengambil IP dan User-Agent permintaan untuk dicatat di audit log.
func Client(c *gin.Context) audit_usecase.Client {
	return audit_usecase.Client{IP: c.ClientIP(), UserAgent: c.Request.UserAgent()}
}

// Actor membangun pelaku aksi dari token. Super admin: OrganizationID = 0.
func Actor(c *gin.Context) audit_usecase.Actor {
	actor := audit_usecase.Actor{Client: Client(c)}
	if claims := GetClaims(c); claims != nil {
		actor.UserID = claims.UserID
		if claims.OrganizationID != nil {
			actor.OrganizationID = *claims.OrganizationID
		}
	}
	return actor
}
