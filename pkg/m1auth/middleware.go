package m1auth

import (
	"net/http"
	"slices"
	"strings"

	"github.com/gin-gonic/gin"
)

const claimsKey = "claims"

// Authenticate menolak permintaan tanpa token sah dengan 401, lalu menyimpan
// klaimnya di context. Ambil lewat GetClaims.
func Authenticate(verifier *Verifier) gin.HandlerFunc {
	return func(c *gin.Context) {
		claims, err := verifier.Parse(bearerToken(c.GetHeader("Authorization")))
		if err != nil {
			unauthenticated(c)
			return
		}
		c.Set(claimsKey, claims)
		c.Next()
	}
}

// RequirePermission meloloskan pengguna yang memiliki salah satu izin `allowed`.
// Izin dibaca dari token, jadi perubahan peran berlaku pada login berikutnya.
func RequirePermission(allowed ...string) gin.HandlerFunc {
	return require(func(claims *Claims) bool {
		return slices.ContainsFunc(allowed, func(p string) bool {
			return slices.Contains(claims.Permissions, p)
		})
	})
}

// RequireDivision meloloskan pengguna dari salah satu kode divisi `allowed`.
// Kode divisi dibuat bebas oleh tiap organisasi, jadi pakai ini hanya untuk
// kode yang sudah disepakati. Untuk hak akses umum, pakai RequirePermission.
// Token tanpa divisi, termasuk super admin, ditolak.
func RequireDivision(allowed ...string) gin.HandlerFunc {
	return require(func(claims *Claims) bool {
		return claims.Division != nil && slices.Contains(allowed, *claims.Division)
	})
}

// RequireOrganization menolak token tanpa organisasi (super admin) untuk
// endpoint yang datanya milik organisasi.
func RequireOrganization() gin.HandlerFunc {
	return require(func(claims *Claims) bool {
		return claims.OrganizationID != nil
	})
}

// GetClaims mengembalikan klaim yang disimpan Authenticate, atau nil.
func GetClaims(c *gin.Context) *Claims {
	value, ok := c.Get(claimsKey)
	if !ok {
		return nil
	}
	claims, _ := value.(*Claims)
	return claims
}

func require(allow func(*Claims) bool) gin.HandlerFunc {
	return func(c *gin.Context) {
		claims := GetClaims(c)
		if claims == nil {
			unauthenticated(c)
			return
		}
		if !allow(claims) {
			abort(c, http.StatusForbidden, "FORBIDDEN", "Anda tidak berhak mengakses sumber daya ini")
			return
		}
		c.Next()
	}
}

func bearerToken(header string) string {
	scheme, token, found := strings.Cut(header, " ")
	if !found || !strings.EqualFold(scheme, "Bearer") {
		return ""
	}
	return strings.TrimSpace(token)
}

func unauthenticated(c *gin.Context) {
	abort(c, http.StatusUnauthorized, "UNAUTHENTICATED", "Token tidak ada atau tidak sah")
}

// abort mengirim amplop galat Panduan API Bersama. Paket ini menulisnya
// sendiri agar tetap mandiri saat disalin ke modul lain.
func abort(c *gin.Context, status int, code, message string) {
	c.AbortWithStatusJSON(status, gin.H{
		"success": false,
		"error":   gin.H{"code": code, "message": message, "fields": nil},
	})
}
