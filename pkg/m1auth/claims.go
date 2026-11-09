// Package m1auth memverifikasi JWT yang diterbitkan M1 dan menyediakan
// middleware gin untuk modul lain.
//
// Paket ini mandiri: hanya bergantung pada gin dan golang-jwt, tanpa database
// dan tanpa paket lain di repositori M1. Modul Go lain cukup menyalin folder
// ini ke repositorinya. Panduannya ada di README bagian "Integrasi Modul Lain".
package m1auth

import "github.com/golang-jwt/jwt/v5"

// Claims adalah isi JWT M1 (PRD Bagian 5.2).
type Claims struct {
	UserID         int64    `json:"user_id"`
	Name           string   `json:"name"`
	OrganizationID *int64   `json:"organization_id"`
	Role           string   `json:"role"`
	Permissions    []string `json:"permissions"`
	Division       *string  `json:"division"`
	UnitUsahaID    *int64   `json:"unit_usaha_id"`
	jwt.RegisteredClaims
}

// Verifier memeriksa tanda tangan, penerbit, dan masa berlaku token.
type Verifier struct {
	secret []byte
	issuer string
}

// NewVerifier menerima JWT_SECRET dan nama penerbit (APP_NAME di M1).
func NewVerifier(secret, issuer string) *Verifier {
	return &Verifier{secret: []byte(secret), issuer: issuer}
}

// Parse mengembalikan klaim token yang sah. Hanya HS256 yang diterima.
func (v *Verifier) Parse(raw string) (*Claims, error) {
	claims := &Claims{}
	_, err := jwt.ParseWithClaims(raw, claims, v.keyFunc,
		jwt.WithValidMethods([]string{"HS256"}),
		jwt.WithIssuer(v.issuer),
		jwt.WithExpirationRequired(),
	)
	if err != nil {
		return nil, err
	}
	return claims, nil
}

func (v *Verifier) keyFunc(_ *jwt.Token) (any, error) {
	return v.secret, nil
}
