package auth_usecase

import (
	"ppl-m1-be/internal/usecase/audit_usecase"
	"ppl-m1-be/internal/usecase/repository/sqlc"
	"ppl-m1-be/internal/usecase/utils"
)

// Alasan login gagal yang dicatat di audit log (tidak pernah dikirim ke klien).
const (
	reasonUnknownEmail  = "email_tidak_terdaftar"
	reasonWrongPassword = "kata_sandi_salah"
	reasonInactive      = "akun_nonaktif"
	reasonOrgInactive   = "organisasi_nonaktif"
)

type LoginInput struct {
	Email    string
	Password string
	Client   audit_usecase.Client
}

type LoginResult struct {
	AccessToken string
	ExpiresIn   int
	Claims      utils.Claims
}

// Profile adalah isi K-01 (/me): data pengguna terbaru plus izinnya.
type Profile struct {
	User        sqlc.UserDetail
	Permissions []string
}

type ChangePasswordInput struct {
	Actor       audit_usecase.Actor
	OldPassword string
	NewPassword string
}
