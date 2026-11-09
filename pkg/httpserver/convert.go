package httpserver

import (
	"database/sql"
	"time"
)

// WIB dipakai untuk semua tanggal dan jam di keluaran API (Panduan 4.5).
// Memakai zona tetap agar tidak bergantung pada tzdata di server.
var WIB = time.FixedZone("WIB", 7*60*60)

func FormatTime(t time.Time) string {
	return t.In(WIB).Format(time.RFC3339)
}

func NullStringPtr(s sql.NullString) *string {
	if !s.Valid {
		return nil
	}
	return &s.String
}

func NullInt64Ptr(n sql.NullInt64) *int64 {
	if !n.Valid {
		return nil
	}
	return &n.Int64
}
