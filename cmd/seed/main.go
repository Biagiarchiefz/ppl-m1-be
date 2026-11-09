// Seeder data awal M1: katalog izin dan satu akun super admin platform dari
// ADMIN_EMAIL, ADMIN_NAME, ADMIN_PASSWORD. Aman dijalankan berulang:
// katalog izin diperbarui, akun yang sudah ada tidak ditimpa.
//
//	go run ./cmd/seed
package main

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log"
	"os"
	"strings"

	"ppl-m1-be/config"
	"ppl-m1-be/internal/app"
	"ppl-m1-be/internal/usecase/repository"
	"ppl-m1-be/internal/usecase/repository/sqlc"
	"ppl-m1-be/internal/usecase/role_usecase"
	"ppl-m1-be/internal/usecase/utils"
)

type admin struct {
	Email, Name, Password string
}

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatal(err)
	}
	adm, err := loadAdmin()
	if err != nil {
		log.Fatal(err)
	}
	db, err := app.ConnectDB(cfg)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	if err := seed(context.Background(), repository.NewStore(db), adm); err != nil {
		log.Fatal(err)
	}
	log.Println("seeder selesai: katalog izin dan akun super admin siap")
}

func loadAdmin() (admin, error) {
	adm := admin{
		Email:    strings.ToLower(strings.TrimSpace(os.Getenv("ADMIN_EMAIL"))),
		Name:     strings.TrimSpace(os.Getenv("ADMIN_NAME")),
		Password: os.Getenv("ADMIN_PASSWORD"),
	}
	if adm.Name == "" {
		adm.Name = "Super Admin"
	}
	if adm.Email == "" || adm.Password == "" {
		return admin{}, errors.New("ADMIN_EMAIL dan ADMIN_PASSWORD wajib diisi di .env")
	}
	if !utils.ValidPassword(adm.Password) || strings.EqualFold(adm.Password, adm.Email) {
		return admin{}, errors.New("ADMIN_PASSWORD minimal 8 karakter, berisi huruf dan angka, dan tidak sama dengan email")
	}
	return adm, nil
}

func seed(ctx context.Context, store *repository.Store, adm admin) error {
	hash, err := utils.HashPassword(adm.Password)
	if err != nil {
		return err
	}
	return store.InTx(ctx, func(q *sqlc.Queries) error {
		for _, p := range permissionCatalog {
			err := q.UpsertPermission(ctx, sqlc.UpsertPermissionParams{Code: p.Code, Module: p.Module, Description: p.Description})
			if err != nil {
				return fmt.Errorf("seed izin %s: %w", p.Code, err)
			}
		}
		if err := seedDefaultRoles(ctx, q); err != nil {
			return err
		}
		return seedSuperAdmin(ctx, q, adm, hash)
	})
}

// seedDefaultRoles melengkapi organisasi yang dibuat sebelum modul M2-M5 ada.
func seedDefaultRoles(ctx context.Context, q *sqlc.Queries) error {
	ids, err := q.ListAllOrganizationIDs(ctx)
	if err != nil {
		return err
	}
	for _, id := range ids {
		if err := role_usecase.ProvisionDefaults(ctx, q, id); err != nil {
			return err
		}
	}
	return nil
}

func seedSuperAdmin(ctx context.Context, q *sqlc.Queries, adm admin, hash string) error {
	_, err := q.GetUserDetailByEmail(ctx, adm.Email)
	if err == nil {
		log.Println("akun super admin sudah ada, dilewati")
		return nil
	}
	if !errors.Is(err, sql.ErrNoRows) {
		return err
	}
	_, err = q.CreateSuperAdmin(ctx, sqlc.CreateSuperAdminParams{Name: adm.Name, Email: adm.Email, PasswordHash: hash})
	if err != nil {
		return fmt.Errorf("buat akun super admin: %w", err)
	}
	return nil
}
