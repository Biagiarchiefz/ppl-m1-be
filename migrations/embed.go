// Package migrations menyimpan file migrasi golang-migrate dan
// menyematkannya ke binary agar `go run ./cmd/migrate` bisa dipakai
// tanpa memasang CLI migrate.
package migrations

import "embed"

//go:embed *.sql
var Files embed.FS
