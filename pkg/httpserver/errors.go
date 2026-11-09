package httpserver

import "net/http"

// APIError adalah satu baris tabel kode galat M1 (PRD Bagian 8).
// `Code` tidak boleh diganti setelah dirilis. `Message` harus sama persis
// dengan teks di Acceptance Criteria.
type APIError struct {
	Status  int
	Code    string
	Message string
}

var (
	ErrInvalidCredentials = APIError{http.StatusUnauthorized, "INVALID_CREDENTIALS", "Kredensial invalid"}
	ErrUnauthenticated    = APIError{http.StatusUnauthorized, "UNAUTHENTICATED", "Token tidak ada atau tidak sah"}
	ErrAccountInactive    = APIError{http.StatusUnauthorized, "ACCOUNT_INACTIVE", "Akun dinonaktifkan"}
	ErrForbidden          = APIError{http.StatusForbidden, "FORBIDDEN", "Anda tidak berhak mengakses sumber daya ini"}
	ErrInvalidParameter   = APIError{http.StatusBadRequest, "INVALID_PARAMETER", "Parameter tidak valid"}
	ErrUserNotFound       = APIError{http.StatusNotFound, "USER_NOT_FOUND", "Pengguna tidak ditemukan"}
	ErrEmailDuplicate     = APIError{http.StatusConflict, "EMAIL_DUPLICATE", "Email sudah terdaftar"}
	ErrUserAlreadyInact   = APIError{http.StatusConflict, "USER_ALREADY_INACTIVE", "Akun sudah nonaktif"}
	ErrUserStillActive    = APIError{http.StatusConflict, "USER_STILL_ACTIVE", "Nonaktifkan akun terlebih dahulu sebelum menghapus permanen"}
	ErrUserAlreadyActive  = APIError{http.StatusConflict, "USER_ALREADY_ACTIVE", "Akun sudah aktif"}
	ErrUserRefInactive    = APIError{http.StatusConflict, "USER_REFERENCE_INACTIVE", "Peran atau divisi akun ini sudah nonaktif. Ganti terlebih dahulu sebelum mengaktifkan akun"}
	ErrCannotDeactSelf    = APIError{http.StatusConflict, "CANNOT_DEACTIVATE_SELF", "Anda tidak dapat menonaktifkan akun sendiri"}
	ErrValidation         = APIError{http.StatusUnprocessableEntity, "VALIDATION_ERROR", "Isian tidak valid"}
	ErrOldPasswordInvalid = APIError{http.StatusUnprocessableEntity, "OLD_PASSWORD_INVALID", "Kata sandi lama salah"}
	ErrInternal           = APIError{http.StatusInternalServerError, "INTERNAL_ERROR", "Terjadi kesalahan pada server"}

	// Multi-organisasi (PRD v3.0).
	ErrOrganizationInactive = APIError{http.StatusUnauthorized, "ORGANIZATION_INACTIVE", "Organisasi dinonaktifkan"}
	ErrRoleNotFound         = APIError{http.StatusNotFound, "ROLE_NOT_FOUND", "Peran tidak ditemukan"}
	ErrRoleCodeDuplicate    = APIError{http.StatusConflict, "ROLE_CODE_DUPLICATE", "Kode peran sudah dipakai"}
	ErrRoleAlreadyInactive  = APIError{http.StatusConflict, "ROLE_ALREADY_INACTIVE", "Peran sudah nonaktif"}
	ErrRoleInUse            = APIError{http.StatusConflict, "ROLE_IN_USE", "Peran masih dipakai oleh pengguna aktif"}
	ErrRoleSystemLocked     = APIError{http.StatusConflict, "ROLE_SYSTEM_LOCKED", "Peran bawaan sistem tidak dapat diubah"}
	ErrDivisionNotFound     = APIError{http.StatusNotFound, "DIVISION_NOT_FOUND", "Divisi tidak ditemukan"}
	ErrDivisionCodeDup      = APIError{http.StatusConflict, "DIVISION_CODE_DUPLICATE", "Kode divisi sudah dipakai"}
	ErrDivisionAlreadyInact = APIError{http.StatusConflict, "DIVISION_ALREADY_INACTIVE", "Divisi sudah nonaktif"}
	ErrDivisionInUse        = APIError{http.StatusConflict, "DIVISION_IN_USE", "Divisi masih dipakai oleh pengguna aktif"}
	ErrOrgNotFound          = APIError{http.StatusNotFound, "ORGANIZATION_NOT_FOUND", "Organisasi tidak ditemukan"}
	ErrOrgCodeDuplicate     = APIError{http.StatusConflict, "ORGANIZATION_CODE_DUPLICATE", "Kode organisasi sudah dipakai"}
	ErrOrgAlreadyInactive   = APIError{http.StatusConflict, "ORGANIZATION_ALREADY_INACTIVE", "Organisasi sudah nonaktif"}
	ErrRoleAlreadyActive    = APIError{http.StatusConflict, "ROLE_ALREADY_ACTIVE", "Peran sudah aktif"}
	ErrDivisionAlreadyActv  = APIError{http.StatusConflict, "DIVISION_ALREADY_ACTIVE", "Divisi sudah aktif"}
	ErrOrgAlreadyActive     = APIError{http.StatusConflict, "ORGANIZATION_ALREADY_ACTIVE", "Organisasi sudah aktif"}
	ErrOrgStillActive       = APIError{http.StatusConflict, "ORGANIZATION_STILL_ACTIVE", "Nonaktifkan organisasi terlebih dahulu sebelum menghapus permanen"}

	// Pengajuan organisasi.
	ErrApplicationNotFound = APIError{http.StatusNotFound, "ORGANIZATION_APPLICATION_NOT_FOUND", "Pengajuan organisasi tidak ditemukan"}
	ErrApplicationPending  = APIError{http.StatusConflict, "ORGANIZATION_APPLICATION_PENDING", "Masih ada pengajuan dengan kode atau email ini yang menunggu persetujuan"}
	ErrApplicationStatus   = APIError{http.StatusConflict, "ORGANIZATION_APPLICATION_STATUS_INVALID", "Pengajuan ini sudah ditinjau"}

	// M2.
	ErrBusinessUnitNotFound  = APIError{http.StatusNotFound, "BUSINESS_UNIT_NOT_FOUND", "Unit usaha tidak ditemukan"}
	ErrBusinessUnitCodeDup   = APIError{http.StatusConflict, "BUSINESS_UNIT_CODE_DUPLICATE", "Kode unit usaha sudah dipakai"}
	ErrBusinessUnitUnchanged = APIError{http.StatusConflict, "BUSINESS_UNIT_STATUS_UNCHANGED", "Status unit usaha sudah sesuai"}

	// M3.
	ErrPartnerNotFound      = APIError{http.StatusNotFound, "PARTNER_NOT_FOUND", "Partner tidak ditemukan"}
	ErrPartnerRoleMissing   = APIError{http.StatusConflict, "PARTNER_ROLE_UNAVAILABLE", "Pendaftaran partner belum dibuka untuk organisasi ini"}
	ErrPartnerStatusInvalid = APIError{http.StatusConflict, "PARTNER_STATUS_INVALID", "Perubahan status partner tidak diizinkan"}
	ErrPartnerNotVerified   = APIError{http.StatusForbidden, "PARTNER_NOT_VERIFIED", "Partner belum terverifikasi atau sedang ditangguhkan"}

	// M4.
	ErrProductNotFound        = APIError{http.StatusNotFound, "PRODUCT_NOT_FOUND", "Produk tidak ditemukan"}
	ErrProductUnchanged       = APIError{http.StatusConflict, "PRODUCT_STATUS_UNCHANGED", "Status produk sudah sesuai"}
	ErrCustomerNotFound       = APIError{http.StatusNotFound, "CUSTOMER_NOT_FOUND", "Pelanggan tidak ditemukan"}
	ErrCustomerEmailDuplicate = APIError{http.StatusConflict, "CUSTOMER_EMAIL_DUPLICATE", "Email pelanggan sudah terdaftar"}
	ErrQuotationNotFound      = APIError{http.StatusNotFound, "QUOTATION_NOT_FOUND", "Quotation tidak ditemukan"}
	ErrQuotationStatusInvalid = APIError{http.StatusConflict, "QUOTATION_STATUS_INVALID", "Quotation sudah ditanggapi"}
	ErrQuotationNotApproved   = APIError{http.StatusConflict, "QUOTATION_NOT_APPROVED", "Invoice hanya dapat dibuat dari quotation yang sudah disetujui"}
	ErrInvoiceNotFound        = APIError{http.StatusNotFound, "INVOICE_NOT_FOUND", "Invoice tidak ditemukan"}
	ErrInvoiceAlreadyExists   = APIError{http.StatusConflict, "INVOICE_ALREADY_EXISTS", "Invoice untuk quotation ini sudah dibuat"}
	ErrInvoiceAlreadyPaid     = APIError{http.StatusConflict, "INVOICE_ALREADY_PAID", "Invoice sudah lunas"}
	ErrOrderNotFound          = APIError{http.StatusNotFound, "ORDER_NOT_FOUND", "Order tidak ditemukan"}
	ErrOrderCompleted         = APIError{http.StatusConflict, "ORDER_ALREADY_COMPLETED", "Order yang sudah selesai tidak dapat diubah"}
	ErrOrderStatusInvalid     = APIError{http.StatusConflict, "ORDER_STATUS_INVALID", "Urutan status order tidak valid"}

	ErrOutOfStock        = APIError{http.StatusConflict, "OUT_OF_STOCK", "Stok produk tidak mencukupi"}
	ErrOrderCancelDenied = APIError{http.StatusConflict, "ORDER_CANCEL_NOT_ALLOWED", "Order hanya dapat dibatalkan saat berstatus baru"}
	ErrOrderAlreadyPaid  = APIError{http.StatusConflict, "ORDER_ALREADY_PAID", "Order yang sudah dibayar tidak dapat dibatalkan"}
	ErrInvoiceCancelled  = APIError{http.StatusConflict, "INVOICE_CANCELLED", "Invoice sudah dibatalkan"}

	// M5.
	ErrServiceReqNotFound  = APIError{http.StatusNotFound, "SERVICE_REQUEST_NOT_FOUND", "Permintaan layanan tidak ditemukan"}
	ErrServiceReqStatus    = APIError{http.StatusConflict, "SERVICE_REQUEST_STATUS_INVALID", "Urutan status permintaan tidak valid"}
	ErrCustomerRoleMissing = APIError{http.StatusConflict, "CUSTOMER_ROLE_UNAVAILABLE", "Pendaftaran customer belum dibuka untuk organisasi ini"}
	ErrCartEmpty           = APIError{http.StatusConflict, "CART_EMPTY", "Keranjang masih kosong"}
	ErrCartItemNotFound    = APIError{http.StatusNotFound, "CART_ITEM_NOT_FOUND", "Item tidak ada di keranjang"}
	ErrCartItemUnavailable = APIError{http.StatusConflict, "CART_ITEM_UNAVAILABLE", "Ada produk di keranjang yang sudah tidak tersedia"}
)

// Pesan per field untuk `error.fields`.
const (
	MsgFieldRequired    = "Isian ini wajib diisi"
	MsgFieldEmail       = "Format email tidak valid"
	MsgFieldType        = "Tipe data tidak valid"
	MsgPasswordTooLong  = "Kata sandi terlalu panjang"
	MsgPasswordRule     = "Kata sandi tidak memenuhi aturan"
	MsgNewPasswordRule  = "Kata sandi baru tidak memenuhi aturan"
	MsgOldPasswordWrong = "Kata sandi lama salah"
	MsgCodeFormat       = "Kode hanya boleh huruf kecil, angka, - atau _ (maks. 50 karakter)"
)
