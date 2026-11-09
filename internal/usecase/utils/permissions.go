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

	PermBusinessUnitsManage = "m2.business_units.manage"
	PermDashboardRead       = "m2.dashboard.read"
	PermDashboardWrite      = "m2.dashboard.write"

	PermPartnersManage    = "m3.partners.manage"
	PermPartnerStatusRead = "m3.partner_status.read"

	PermProductsManage    = "m4.products.manage"
	PermCustomersManage   = "m4.customers.manage"
	PermQuotationsManage  = "m4.quotations.manage"
	PermQuotationsRespond = "m4.quotations.respond"
	PermInvoicesManage    = "m4.invoices.manage"
	PermInvoicesPay       = "m4.invoices.pay"
	PermOrdersRead        = "m4.orders.read"
	PermOrdersUpdate      = "m4.orders.update"

	PermCartManage          = "m5.cart.manage"
	PermServiceRequestsSend = "m5.service_requests.send"
	PermServiceRequestsRead = "m5.service_requests.read"
)

const (
	// RoleSuperAdmin adalah nilai klaim `role` untuk super admin platform.
	RoleSuperAdmin = "superadmin"
	// SystemRoleCode adalah kode peran administrator bawaan tiap organisasi.
	SystemRoleCode = "administrator"
	SystemRoleName = "Administrator"

	RolePartner  = "partner"
	RoleCustomer = "customer"
	RoleManager  = "manajer"
	// RoleUnitStaff memproses Order dan permintaan layanan milik unit usahanya.
	RoleUnitStaff = "staf_unit"
)

// DefaultRole adalah peran siap pakai untuk aktor di studi kasus. Izinnya
// tetap bisa diubah administrator organisasi.
type DefaultRole struct {
	Code        string
	Name        string
	Permissions []string
}

var DefaultRoles = []DefaultRole{
	{RolePartner, "Partner", []string{
		PermPartnerStatusRead, PermProductsManage, PermCustomersManage, PermQuotationsManage,
		PermQuotationsRespond, PermInvoicesManage, PermOrdersRead, PermOrdersUpdate,
	}},
	{RoleCustomer, "Customer", []string{
		PermQuotationsRespond, PermInvoicesPay, PermOrdersRead, PermCartManage, PermServiceRequestsSend,
	}},
	{RoleManager, "Manajer", []string{PermDashboardRead}},
	{RoleUnitStaff, "Staf Unit Usaha", []string{PermOrdersRead, PermOrdersUpdate, PermServiceRequestsRead}},
}

var codePattern = regexp.MustCompile(`^[a-z0-9]+([_-][a-z0-9]+)*$`)

// ValidCode memeriksa kode organisasi/peran/divisi: huruf kecil, angka,
// dipisah - atau _, maksimal 50 karakter.
func ValidCode(code string) bool {
	return len(code) <= 50 && codePattern.MatchString(code)
}
