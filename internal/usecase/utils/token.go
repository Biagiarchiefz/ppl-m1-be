package utils

import (
	"time"

	"github.com/golang-jwt/jwt/v5"

	"ppl-m1-be/pkg/m1auth"
)

type Claims = m1auth.Claims

type TokenMaker struct {
	secret   []byte
	issuer   string
	ttl      time.Duration
	verifier *m1auth.Verifier
}

func NewTokenMaker(secret, issuer string, ttl time.Duration) *TokenMaker {
	return &TokenMaker{
		secret: []byte(secret), issuer: issuer, ttl: ttl,
		verifier: m1auth.NewVerifier(secret, issuer),
	}
}

func (m *TokenMaker) TTL() time.Duration {
	return m.ttl
}

func (m *TokenMaker) Verifier() *m1auth.Verifier {
	return m.verifier
}

func (m *TokenMaker) Generate(c Claims) (string, error) {
	now := time.Now()
	c.Issuer = m.issuer
	c.IssuedAt = jwt.NewNumericDate(now)
	c.ExpiresAt = jwt.NewNumericDate(now.Add(m.ttl))
	return jwt.NewWithClaims(jwt.SigningMethodHS256, c).SignedString(m.secret)
}
