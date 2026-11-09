package middleware

import (
	"net/http"
	"slices"

	"github.com/gin-gonic/gin"
)

const (
	corsMethods = "GET, POST, PATCH, DELETE, OPTIONS"
	corsHeaders = "Authorization, Content-Type, X-Service-Key"
)

// CORS mengizinkan origin dari CORS_ALLOWED_ORIGINS (PRD 10.7). Isi "*" untuk
// mengizinkan semua origin. Daftar kosong berarti header CORS tidak dikirim.
func CORS(allowed []string) gin.HandlerFunc {
	allowAll := slices.Contains(allowed, "*")
	return func(c *gin.Context) {
		origin := c.GetHeader("Origin")
		if origin != "" && (allowAll || slices.Contains(allowed, origin)) {
			h := c.Writer.Header()
			h.Set("Access-Control-Allow-Origin", origin)
			h.Add("Vary", "Origin")
			h.Set("Access-Control-Allow-Methods", corsMethods)
			h.Set("Access-Control-Allow-Headers", corsHeaders)
			h.Set("Access-Control-Max-Age", "600")
		}
		if c.Request.Method == http.MethodOptions {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}
		c.Next()
	}
}
