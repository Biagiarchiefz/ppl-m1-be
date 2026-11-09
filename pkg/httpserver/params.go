package httpserver

import (
	"database/sql"
	"slices"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
)

const dateLayout = "2006-01-02"

// PathID membaca parameter `{id}` berupa bilangan bulat positif.
func PathID(c *gin.Context) (int64, bool) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	return id, err == nil && id > 0
}

func QueryString(c *gin.Context, key string) sql.NullString {
	v := c.Query(key)
	return sql.NullString{String: v, Valid: v != ""}
}

func QueryInt64(c *gin.Context, key string) (sql.NullInt64, bool) {
	raw := c.Query(key)
	if raw == "" {
		return sql.NullInt64{}, true
	}
	n, err := strconv.ParseInt(raw, 10, 64)
	return sql.NullInt64{Int64: n, Valid: err == nil}, err == nil
}

func QueryBool(c *gin.Context, key string) (sql.NullBool, bool) {
	raw := c.Query(key)
	if raw == "" {
		return sql.NullBool{}, true
	}
	b, err := strconv.ParseBool(raw)
	return sql.NullBool{Bool: b, Valid: err == nil}, err == nil
}

// QuerySort membaca `sort`; hanya nilai di `allowed` yang diterima.
func QuerySort(c *gin.Context, allowed []string, fallback string) (string, bool) {
	v := c.DefaultQuery("sort", fallback)
	return v, slices.Contains(allowed, v)
}

// QueryDate membaca tanggal `YYYY-MM-DD` sebagai awal hari dalam WIB.
func QueryDate(c *gin.Context, key string) (sql.NullTime, bool) {
	raw := c.Query(key)
	if raw == "" {
		return sql.NullTime{}, true
	}
	t, err := time.ParseInLocation(dateLayout, raw, WIB)
	return sql.NullTime{Time: t, Valid: err == nil}, err == nil
}
