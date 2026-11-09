package main

import "ppl-m1-be/internal/usecase/utils"

type permission struct {
	Code        string
	Module      string
	Description string
}

// permissionCatalog adalah katalog izin global. Peran dan divisi kini milik
// tiap organisasi, jadi tidak di-seed lagi. Izin modul lain (M2-M5)
// ditambahkan di sini setelah kelompoknya mengirim daftar kodenya lewat
// repositori kontrak; seeder bisa dijalankan ulang kapan saja.
var permissionCatalog = []permission{
	{utils.PermOrganizationsManage, "platform", "Mengelola organisasi (khusus super admin)"},
	{utils.PermUsersRead, "m1", "Melihat daftar dan detail pengguna"},
	{utils.PermUsersManage, "m1", "Menambah, mengubah, dan menonaktifkan pengguna"},
	{utils.PermRolesManage, "m1", "Mengelola peran dan izinnya"},
	{utils.PermDivisionsManage, "m1", "Mengelola divisi"},
	{utils.PermAuditRead, "m1", "Melihat audit log"},
}
