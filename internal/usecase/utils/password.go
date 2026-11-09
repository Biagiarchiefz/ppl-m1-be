package utils

import (
	"unicode"

	"golang.org/x/crypto/bcrypt"
)

const (
	// MaxPasswordBytes adalah batas bcrypt; lebih dari ini ditolak.
	MaxPasswordBytes = 72
	minPasswordChars = 8
	bcryptCost       = 12
)

func HashPassword(plain string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(plain), bcryptCost)
	return string(hash), err
}

func CheckPassword(hash, plain string) error {
	return bcrypt.CompareHashAndPassword([]byte(hash), []byte(plain))
}

// ValidPassword adalah satu-satunya tempat aturan kata sandi [ASUMSI PRD 6]:
// minimal 8 karakter, mengandung huruf dan angka, maksimal 72 byte.
// Aturan "tidak sama dengan kata sandi lama" diperiksa oleh pemanggil.
func ValidPassword(plain string) bool {
	if len([]rune(plain)) < minPasswordChars || len(plain) > MaxPasswordBytes {
		return false
	}
	var hasLetter, hasDigit bool
	for _, r := range plain {
		hasLetter = hasLetter || unicode.IsLetter(r)
		hasDigit = hasDigit || unicode.IsDigit(r)
	}
	return hasLetter && hasDigit
}
