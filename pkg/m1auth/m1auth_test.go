package m1auth

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

const (
	testSecret = "rahasia-uji-minimal-32-byte-panjang"
	testIssuer = "ppl-m1-be"
)

func signed(t *testing.T, secret, issuer string, ttl time.Duration, c Claims) string {
	t.Helper()
	c.Issuer = issuer
	c.ExpiresAt = jwt.NewNumericDate(time.Now().Add(ttl))
	token, err := jwt.NewWithClaims(jwt.SigningMethodHS256, c).SignedString([]byte(secret))
	if err != nil {
		t.Fatal(err)
	}
	return token
}

func request(t *testing.T, guard gin.HandlerFunc, token string) (int, string) {
	t.Helper()
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.GET("/", Authenticate(NewVerifier(testSecret, testIssuer)), guard, func(c *gin.Context) {
		c.Status(http.StatusNoContent)
	})
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	var body struct {
		Error struct {
			Code string `json:"code"`
		} `json:"error"`
	}
	_ = json.Unmarshal(w.Body.Bytes(), &body)
	return w.Code, body.Error.Code
}

func allowAll(c *gin.Context) { c.Next() }

func TestAuthenticateRejectsBadTokens(t *testing.T) {
	cases := map[string]string{
		"tanpa token":      "",
		"bukan jwt":        "abc",
		"rahasia berbeda":  signed(t, "rahasia-lain-minimal-32-byte-panjang", testIssuer, time.Hour, Claims{}),
		"penerbit berbeda": signed(t, testSecret, "modul-lain", time.Hour, Claims{}),
		"kedaluwarsa":      signed(t, testSecret, testIssuer, -time.Hour, Claims{}),
	}
	for name, token := range cases {
		t.Run(name, func(t *testing.T) {
			status, code := request(t, allowAll, token)
			if status != http.StatusUnauthorized || code != "UNAUTHENTICATED" {
				t.Fatalf("dapat %d %q", status, code)
			}
		})
	}
}

type guardCase struct {
	name   string
	guard  gin.HandlerFunc
	claims Claims
	status int
}

func guardCases() []guardCase {
	orgID := int64(7)
	finance := "finance"
	user := Claims{OrganizationID: &orgID, Division: &finance, Permissions: []string{"m4.invoice.create"}}
	superAdmin := Claims{Permissions: []string{"platform.organizations.manage"}}

	return []guardCase{
		{"izin cocok", RequirePermission("m4.invoice.read", "m4.invoice.create"), user, http.StatusNoContent},
		{"izin tidak cocok", RequirePermission("m4.invoice.read"), user, http.StatusForbidden},
		{"divisi cocok", RequireDivision("hr", "finance"), user, http.StatusNoContent},
		{"divisi tidak cocok", RequireDivision("hr"), user, http.StatusForbidden},
		{"tanpa divisi", RequireDivision("finance"), superAdmin, http.StatusForbidden},
		{"punya organisasi", RequireOrganization(), user, http.StatusNoContent},
		{"tanpa organisasi", RequireOrganization(), superAdmin, http.StatusForbidden},
	}
}

func TestGuards(t *testing.T) {
	for _, tc := range guardCases() {
		t.Run(tc.name, func(t *testing.T) {
			status, _ := request(t, tc.guard, signed(t, testSecret, testIssuer, time.Hour, tc.claims))
			if status != tc.status {
				t.Fatalf("dapat %d, seharusnya %d", status, tc.status)
			}
		})
	}
}
