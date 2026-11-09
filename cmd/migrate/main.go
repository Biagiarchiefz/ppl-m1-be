// Perintah migrasi database M1 (golang-migrate).
//
//	go run ./cmd/migrate up       jalankan semua migrasi yang belum ada
//	go run ./cmd/migrate down     batalkan satu migrasi terakhir
//	go run ./cmd/migrate version  tampilkan versi migrasi saat ini
package main

import (
	"database/sql"
	"errors"
	"fmt"
	"log"
	"os"

	"github.com/golang-migrate/migrate/v4"
	pgxmigrate "github.com/golang-migrate/migrate/v4/database/pgx/v5"
	"github.com/golang-migrate/migrate/v4/source/iofs"

	"ppl-m1-be/config"
	"ppl-m1-be/internal/app"
	"ppl-m1-be/migrations"
)

func main() {
	if len(os.Args) != 2 {
		log.Fatal("pemakaian: go run ./cmd/migrate [up|down|version]")
	}
	cfg, err := config.Load()
	if err != nil {
		log.Fatal(err)
	}
	db, err := app.ConnectDB(cfg)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	if err := run(db, os.Args[1]); err != nil {
		log.Fatal(err)
	}
}

func run(db *sql.DB, command string) error {
	m, err := newMigrator(db)
	if err != nil {
		return err
	}
	switch command {
	case "up":
		return ignoreNoChange(m.Up())
	case "down":
		return ignoreNoChange(m.Steps(-1))
	case "version":
		return printVersion(m)
	}
	return fmt.Errorf("perintah tidak dikenal: %s", command)
}

func newMigrator(db *sql.DB) (*migrate.Migrate, error) {
	source, err := iofs.New(migrations.Files, ".")
	if err != nil {
		return nil, err
	}
	driver, err := pgxmigrate.WithInstance(db, &pgxmigrate.Config{})
	if err != nil {
		return nil, err
	}
	return migrate.NewWithInstance("iofs", source, "pgx5", driver)
}

func ignoreNoChange(err error) error {
	if errors.Is(err, migrate.ErrNoChange) {
		log.Println("tidak ada migrasi baru")
		return nil
	}
	if err == nil {
		log.Println("migrasi selesai")
	}
	return err
}

func printVersion(m *migrate.Migrate) error {
	version, dirty, err := m.Version()
	if errors.Is(err, migrate.ErrNilVersion) {
		log.Println("belum ada migrasi yang dijalankan")
		return nil
	}
	if err == nil {
		log.Printf("versi %d (dirty=%v)", version, dirty)
	}
	return err
}
