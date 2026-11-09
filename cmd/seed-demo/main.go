// Seeder data demo: mengisi semua halaman agar alur aplikasi bisa dicoba.
// Terpisah dari cmd/seed supaya data contoh tidak ikut ke lingkungan produksi.
// Semua data dibuat lewat usecase, jadi audit log, stok, dan invoice ikut
// terbentuk seperti pemakaian sungguhan.
//
// Jalankan dulu `go run ./cmd/migrate up` dan `go run ./cmd/seed`, lalu isi
// DEMO_PASSWORD di .env.
//
//	go run ./cmd/seed-demo        buat data demo (dilewati bila sudah ada)
//	go run ./cmd/seed-demo reset  hapus data demo saja
package main

import (
	"context"
	"database/sql"
	"errors"
	"log"
	"os"

	"ppl-m1-be/config"
	"ppl-m1-be/internal/app"
	"ppl-m1-be/internal/usecase/apperr"
	"ppl-m1-be/internal/usecase/audit_usecase"
	"ppl-m1-be/internal/usecase/repository"
	"ppl-m1-be/internal/usecase/utils"
)

const (
	demoCode    = "demo-nusantara"
	secondCode  = "demo-harapan"
	emailDomain = "demo.araya.test"

	// minPermissions memastikan cmd/seed sudah mengisi katalog izin M1-M5.
	minPermissions = 20
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatal(err)
	}
	db, err := app.ConnectDB(cfg)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()
	s := newSeeder(context.Background(), repository.NewStore(db), os.Getenv("DEMO_PASSWORD"))
	if err := dispatch(s); err != nil {
		log.Fatal(err)
	}
}

func dispatch(s *seeder) error {
	if len(os.Args) == 2 && os.Args[1] == "reset" {
		return s.reset()
	}
	if len(os.Args) != 1 {
		return errors.New("pemakaian: go run ./cmd/seed-demo [reset]")
	}
	if !utils.ValidPassword(s.password) {
		return errors.New("DEMO_PASSWORD wajib diisi: minimal 8 karakter, berisi huruf dan angka")
	}
	return s.seed()
}

func (s *seeder) seed() error {
	_, err := s.store.GetOrganizationByCode(s.ctx, demoCode)
	if err == nil {
		log.Println("data demo sudah ada, dilewati. Jalankan `reset` untuk membuat ulang")
		return nil
	}
	if !errors.Is(err, sql.ErrNoRows) {
		return err
	}
	if err := s.requireCatalog(); err != nil {
		return err
	}
	for _, step := range s.steps() {
		if err := step(); err != nil {
			return err
		}
	}
	s.printAccounts()
	return nil
}

func (s *seeder) requireCatalog() error {
	count, err := s.store.CountPermissions(s.ctx)
	if err != nil {
		return err
	}
	if count < minPermissions {
		return errors.New("katalog izin belum lengkap: jalankan `go run ./cmd/seed` lebih dulu")
	}
	return nil
}

// reset hanya menyentuh organisasi dan pengajuan milik data demo.
func (s *seeder) reset() error {
	for _, code := range []string{demoCode, secondCode} {
		if err := s.removeOrganization(code); err != nil {
			return err
		}
	}
	err := s.store.DeleteOrganizationApplicationsByEmail(s.ctx, "%@"+emailDomain)
	if err == nil {
		log.Println("data demo dihapus")
	}
	return err
}

func (s *seeder) removeOrganization(code string) error {
	org, err := s.store.GetOrganizationByCode(s.ctx, code)
	if errors.Is(err, sql.ErrNoRows) {
		return nil
	}
	if err != nil {
		return err
	}
	actor := audit_usecase.Actor{Client: seedClient}
	err = s.orgs.Deactivate(s.ctx, actor, org.ID)
	if err != nil && !errors.Is(err, apperr.OrganizationAlreadyInactive) {
		return err
	}
	return s.orgs.Delete(s.ctx, actor, org.ID)
}
