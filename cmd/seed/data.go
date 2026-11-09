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
	{utils.PermBusinessUnitsManage, "m2", "Mengelola unit usaha"},
	{utils.PermDashboardRead, "m2", "Melihat Dashboard Operasional Dasar"},
	{utils.PermDashboardWrite, "m2", "Mengalirkan data ke Dashboard Operasional Dasar"},
	{utils.PermPartnersManage, "m3", "Memverifikasi dan menangguhkan partner"},
	{utils.PermPartnerStatusRead, "m3", "Melihat status pendaftaran partner sendiri"},
	{utils.PermProductsManage, "m4", "Mengelola produk/jasa"},
	{utils.PermCustomersManage, "m4", "Mengelola data pelanggan"},
	{utils.PermQuotationsManage, "m4", "Membuat dan melihat quotation"},
	{utils.PermQuotationsRespond, "m4", "Menyetujui atau menolak quotation"},
	{utils.PermInvoicesManage, "m4", "Membuat dan melihat invoice"},
	{utils.PermInvoicesPay, "m4", "Membayar invoice sendiri"},
	{utils.PermOrdersRead, "m4", "Melihat order"},
	{utils.PermOrdersUpdate, "m4", "Memperbarui status order"},
	{utils.PermCartManage, "m5", "Mengelola keranjang dan checkout"},
	{utils.PermServiceRequestsSend, "m5", "Mengirim permintaan layanan"},
	{utils.PermServiceRequestsRead, "m5", "Melihat permintaan layanan yang masuk"},
}
