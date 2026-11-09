// Package apperr berisi galat domain M1 yang dipakai bersama oleh usecase.
// Pemetaan ke kode galat API ada di controller (paket respond).
package apperr

import (
	"database/sql"
	"errors"
)

var (
	InvalidCredentials   = errors.New("kredensial invalid")
	AccountInactive      = errors.New("akun dinonaktifkan")
	OrganizationInactive = errors.New("organisasi dinonaktifkan")
	Unauthenticated      = errors.New("pengguna token tidak ditemukan")
	OldPasswordInvalid   = errors.New("kata sandi lama salah")

	UserNotFound         = errors.New("pengguna tidak ditemukan")
	EmailTaken           = errors.New("email sudah terdaftar")
	UserAlreadyInactive  = errors.New("akun sudah nonaktif")
	CannotDeactivateSelf = errors.New("tidak dapat menonaktifkan akun sendiri")
	UserStillActive      = errors.New("akun masih aktif")
	UserAlreadyActive    = errors.New("akun sudah aktif")

	UserReferenceInactive = errors.New("peran atau divisi akun sudah nonaktif")

	RoleNotFound        = errors.New("peran tidak ditemukan")
	RoleCodeTaken       = errors.New("kode peran sudah dipakai")
	RoleAlreadyInactive = errors.New("peran sudah nonaktif")
	RoleAlreadyActive   = errors.New("peran sudah aktif")
	RoleInUse           = errors.New("peran masih dipakai")
	RoleSystemLocked    = errors.New("peran bawaan tidak dapat diubah")

	DivisionNotFound        = errors.New("divisi tidak ditemukan")
	DivisionCodeTaken       = errors.New("kode divisi sudah dipakai")
	DivisionAlreadyInactive = errors.New("divisi sudah nonaktif")
	DivisionAlreadyActive   = errors.New("divisi sudah aktif")
	DivisionInUse           = errors.New("divisi masih dipakai")

	OrganizationNotFound        = errors.New("organisasi tidak ditemukan")
	OrganizationCodeTaken       = errors.New("kode organisasi sudah dipakai")
	OrganizationAlreadyInactive = errors.New("organisasi sudah nonaktif")
	OrganizationAlreadyActive   = errors.New("organisasi sudah aktif")
	OrganizationStillActive     = errors.New("organisasi masih aktif")
)

// FieldErrors adalah galat validasi yang butuh data dari database, misalnya
// peran tidak terdaftar. Kunci = nama field JSON.
type FieldErrors map[string]string

func (f FieldErrors) Error() string {
	return "isian tidak valid"
}

// NotFound mengganti sql.ErrNoRows dengan galat domain `target`.
func NotFound(err, target error) error {
	if errors.Is(err, sql.ErrNoRows) {
		return target
	}
	return err
}
