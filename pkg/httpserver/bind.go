package httpserver

import (
	"database/sql"
	"slices"
	"strings"

	"github.com/gin-gonic/gin"
)

// BindJSON membaca body ke `dst`. Bila gagal, respons galat sudah dikirim.
func BindJSON(c *gin.Context, dst any) bool {
	if err := c.ShouldBindJSON(dst); err != nil {
		BindError(c, err)
		return false
	}
	return true
}

// RequireID membaca `{id}`. Bila tidak sah, respons 400 sudah dikirim.
func RequireID(c *gin.Context) (int64, bool) {
	id, ok := PathID(c)
	if !ok {
		Fail(c, ErrInvalidParameter)
	}
	return id, ok
}

// RequirePage membaca pagination. Bila tidak sah, respons 400 sudah dikirim.
func RequirePage(c *gin.Context) (Page, bool) {
	page, ok := ParsePage(c)
	if !ok {
		Fail(c, ErrInvalidParameter)
	}
	return page, ok
}

// QueryEnum membaca query yang nilainya harus salah satu dari `allowed`.
func QueryEnum(c *gin.Context, key string, allowed ...string) (sql.NullString, bool) {
	value := QueryString(c, key)
	return value, !value.Valid || slices.Contains(allowed, value.String)
}

// Map mengubah slice untuk respons; hasilnya tidak pernah nil supaya daftar
// kosong dikirim sebagai `[]`.
func Map[T, R any](items []T, convert func(T) R) []R {
	out := make([]R, 0, len(items))
	for _, item := range items {
		out = append(out, convert(item))
	}
	return out
}

// OptionalText mengubah isian opsional menjadi nullable; teks kosong = null.
func OptionalText(s *string) sql.NullString {
	if s == nil || strings.TrimSpace(*s) == "" {
		return sql.NullString{}
	}
	return sql.NullString{String: strings.TrimSpace(*s), Valid: true}
}

// OptionalID mengubah id opsional menjadi nullable.
func OptionalID(id *int64) sql.NullInt64 {
	if id == nil {
		return sql.NullInt64{}
	}
	return sql.NullInt64{Int64: *id, Valid: true}
}

// NullInt32Ptr mengubah bilangan nullable menjadi nilai JSON atau null.
func NullInt32Ptr(n sql.NullInt32) *int32 {
	if !n.Valid {
		return nil
	}
	return &n.Int32
}

// OptionalInt32 mengubah bilangan opsional menjadi nullable.
func OptionalInt32(n *int32) sql.NullInt32 {
	if n == nil {
		return sql.NullInt32{}
	}
	return sql.NullInt32{Int32: *n, Valid: true}
}

// NullTimeText memformat waktu nullable menjadi ISO 8601 WIB atau null.
func NullTimeText(t sql.NullTime) *string {
	if !t.Valid {
		return nil
	}
	text := FormatTime(t.Time)
	return &text
}
