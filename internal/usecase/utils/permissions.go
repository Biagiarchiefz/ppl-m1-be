package utils

import "regexp"

// Kode izin M1. Modul lain mendaftarkan kodenya sendiri lewat repositori
// kontrak (format: <modul>.<resource>.<aksi>), lalu ditambahkan ke seeder.
const (
	PermOrganizationsManage = "platform.organizations.manage"
	PermUsersRead           = "m1.users.read"
	PermUsersManage         = "m1.users.manage"
	PermRolesManage         = "m1.roles.manage"
	PermDivisionsManage     = "m1.divisions.manage"
	PermAuditRead           = "m1.audit.read"
)

const (
	// RoleSuperAdmin adalah nilai klaim `role` untuk super admin platform.
	RoleSuperAdmin = "superadmin"
	// SystemRoleCode adalah kode peran administrator bawaan tiap organisasi.
	SystemRoleCode = "administrator"
	SystemRoleName = "Administrator"
)

var codePattern = regexp.MustCompile(`^[a-z0-9]+([_-][a-z0-9]+)*$`)

// ValidCode memeriksa kode organisasi/peran/divisi: huruf kecil, angka,
// dipisah - atau _, maksimal 50 karakter.
func ValidCode(code string) bool {
	return len(code) <= 50 && codePattern.MatchString(code)
}
