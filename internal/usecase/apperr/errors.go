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

	ApplicationNotFound      = errors.New("pengajuan organisasi tidak ditemukan")
	ApplicationPending       = errors.New("pengajuan organisasi masih menunggu")
	ApplicationStatusInvalid = errors.New("pengajuan organisasi sudah ditinjau")

	Forbidden = errors.New("tidak berwenang")

	BusinessUnitNotFound        = errors.New("unit usaha tidak ditemukan")
	BusinessUnitCodeTaken       = errors.New("kode unit usaha sudah dipakai")
	BusinessUnitStatusUnchanged = errors.New("status unit usaha tidak berubah")

	PartnerNotFound        = errors.New("partner tidak ditemukan")
	PartnerRoleUnavailable = errors.New("peran partner belum tersedia")
	PartnerStatusInvalid   = errors.New("perubahan status partner tidak diizinkan")
	PartnerNotVerified     = errors.New("partner belum terverifikasi")

	ProductNotFound        = errors.New("produk tidak ditemukan")
	ProductStatusUnchanged = errors.New("status produk tidak berubah")
	CustomerNotFound       = errors.New("pelanggan tidak ditemukan")
	CustomerEmailTaken     = errors.New("email pelanggan sudah terdaftar")

	QuotationNotFound      = errors.New("quotation tidak ditemukan")
	QuotationStatusInvalid = errors.New("quotation sudah ditanggapi")
	QuotationNotApproved   = errors.New("quotation belum disetujui")

	InvoiceNotFound      = errors.New("invoice tidak ditemukan")
	InvoiceAlreadyExists = errors.New("invoice sudah dibuat")
	InvoiceAlreadyPaid   = errors.New("invoice sudah lunas")

	OrderNotFound      = errors.New("order tidak ditemukan")
	OrderCompleted     = errors.New("order sudah selesai")
	OrderStatusInvalid = errors.New("urutan status order tidak valid")

	OutOfStock            = errors.New("stok tidak mencukupi")
	OrderCancelNotAllowed = errors.New("order tidak dapat dibatalkan")
	OrderAlreadyPaid      = errors.New("order sudah dibayar")
	InvoiceCancelled      = errors.New("invoice sudah dibatalkan")

	ServiceRequestNotFound      = errors.New("permintaan layanan tidak ditemukan")
	ServiceRequestStatusInvalid = errors.New("urutan status permintaan tidak valid")
	CustomerRoleUnavailable     = errors.New("peran customer belum tersedia")

	CartEmpty           = errors.New("keranjang kosong")
	CartItemNotFound    = errors.New("item keranjang tidak ditemukan")
	CartItemUnavailable = errors.New("ada produk keranjang yang tidak tersedia")
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
