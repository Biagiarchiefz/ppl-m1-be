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
