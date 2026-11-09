package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"time"

	"ppl-m1-be/internal/usecase/audit_usecase"
	"ppl-m1-be/internal/usecase/business_unit_usecase"
	"ppl-m1-be/internal/usecase/division_usecase"
	"ppl-m1-be/internal/usecase/organization_usecase"
	"ppl-m1-be/internal/usecase/partner_usecase"
	"ppl-m1-be/internal/usecase/repository"
	"ppl-m1-be/internal/usecase/storefront_usecase"
	"ppl-m1-be/internal/usecase/transaction_usecase"
	"ppl-m1-be/internal/usecase/user_usecase"
	"ppl-m1-be/internal/usecase/utils"
	"ppl-m1-be/pkg/httpserver"
)

var seedClient = audit_usecase.Client{IP: "127.0.0.1", UserAgent: "seed-demo"}

type identity = transaction_usecase.Identity

type seeder struct {
	ctx       context.Context
	store     *repository.Store
	password  string
	orgs      *organization_usecase.Usecase
	divisions *division_usecase.Usecase
	users     *user_usecase.Usecase
	units     *business_unit_usecase.Usecase
	partners  *partner_usecase.Usecase
	tx        *transaction_usecase.Usecase
	shop      *storefront_usecase.Usecase

	orgID        int64
	admin        identity
	divisionIDs  map[string]int64
	unitIDs      map[string]int64
	partnerIDs   map[string]int64
	partnerUsers map[string]identity
	buyers       map[string]identity
	customerIDs  map[string]int64
	productIDs   map[string]int64
}

func newSeeder(ctx context.Context, store *repository.Store, password string) *seeder {
	return &seeder{
		ctx: ctx, store: store, password: password,
		orgs: organization_usecase.New(store), divisions: division_usecase.New(store), users: user_usecase.New(store),
		units: business_unit_usecase.New(store), partners: partner_usecase.New(store),
		tx: transaction_usecase.New(store), shop: storefront_usecase.New(store),
		divisionIDs: map[string]int64{}, unitIDs: map[string]int64{}, partnerIDs: map[string]int64{},
		partnerUsers: map[string]identity{}, buyers: map[string]identity{}, customerIDs: map[string]int64{}, productIDs: map[string]int64{},
	}
}

func (s *seeder) steps() []func() error {
	return []func() error{
		s.seedOrganization, s.seedDivisions, s.seedUnits, s.seedTeam, s.seedPartners, s.seedProducts, s.seedCustomers,
		s.seedQuotations, s.seedOrders, s.seedCart, s.seedServiceRequests, s.seedDashboardHistory, s.seedFinalStatuses, s.seedApplications,
	}
}

func email(local string) string {
	return local + "@" + emailDomain
}

func (s *seeder) actor(userID int64) audit_usecase.Actor {
	return audit_usecase.Actor{UserID: userID, OrganizationID: s.orgID, Client: seedClient}
}

func text(value string) sql.NullString {
	return sql.NullString{String: value, Valid: value != ""}
}

func id(value int64) sql.NullInt64 {
	return sql.NullInt64{Int64: value, Valid: value != 0}
}

func (s *seeder) seedOrganization() error {
	res, err := s.orgs.Create(s.ctx, organization_usecase.CreateInput{
		Actor: audit_usecase.Actor{Client: seedClient}, Code: demoCode, Name: "Koperasi Nusantara",
		Admin: organization_usecase.AdminInput{Name: "Dewi Lestari", Email: email("admin"), Password: s.password},
	})
	if err != nil {
		return fmt.Errorf("organisasi demo: %w", err)
	}
	s.orgID = res.Organization.ID
	s.admin = identity{Actor: s.actor(res.AdminID), Role: utils.SystemRoleCode, Permissions: []string{
		utils.PermProductsManage, utils.PermCustomersManage, utils.PermQuotationsManage, utils.PermInvoicesManage, utils.PermOrdersUpdate,
	}}
	return nil
}

func (s *seeder) seedDivisions() error {
	for _, d := range [][2]string{{"operasional", "Operasional"}, {"layanan", "Layanan Pelanggan"}, {"keuangan", "Keuangan"}} {
		division, err := s.divisions.Create(s.ctx, s.admin.Actor, d[0], d[1])
		if err != nil {
			return fmt.Errorf("divisi %s: %w", d[0], err)
		}
		s.divisionIDs[d[0]] = division.ID
	}
	return nil
}

func (s *seeder) seedUnits() error {
	rows := []struct{ key, name, category, division, desc string }{
		{"kopi", "Kedai Kopi Nusantara", "Makanan & Minuman", "operasional", "Kopi pilihan dari petani lokal."},
		{"cetak", "Percetakan Nusantara", "Jasa Cetak", "layanan", "Cetak banner, kartu nama, dan undangan."},
		{"wisma", "Wisma Nusantara", "Akomodasi", "operasional", "Penginapan dan sewa aula. Sedang renovasi."},
	}
	for _, r := range rows {
		unit, err := s.units.Create(s.ctx, s.admin.Actor, business_unit_usecase.Input{
			DivisionID: id(s.divisionIDs[r.division]), Code: r.key, Name: r.name, Category: r.category,
			Description: text(r.desc), Address: text("Jl. Merdeka No. 10"), Phone: text("021-555-0100"), Email: text(email(r.key)),
		})
		if err != nil {
			return fmt.Errorf("unit %s: %w", r.key, err)
		}
		s.unitIDs[r.key] = unit.ID
	}
	return nil
}

func (s *seeder) seedTeam() error {
	rows := []struct{ local, name, role, unit string }{
		{"manajer", "Budi Santoso", utils.RoleManager, ""},
		{"manajer.kopi", "Rani Wulandari", utils.RoleManager, "kopi"},
		{"staf.kopi", "Agus Prasetyo", utils.RoleUnitStaff, "kopi"},
		{"staf.cetak", "Lina Marlina", utils.RoleUnitStaff, "cetak"},
	}
	for _, r := range rows {
		_, err := s.users.Create(s.ctx, user_usecase.CreateInput{
			Actor: s.admin.Actor, Name: r.name, Email: email(r.local), Password: s.password, Role: r.role, UnitUsahaID: id(s.unitIDs[r.unit]),
		})
		if err != nil {
			return fmt.Errorf("pengguna %s: %w", r.local, err)
		}
	}
	return nil
}

func (s *seeder) seedPartners() error {
	rows := []struct{ key, business, kind, contact, desc string }{
		{"maju", "CV Maju Bersama", "Perdagangan hasil bumi", "Hendra Wijaya", "Pemasok kopi dan madu."},
		{"rempah", "UD Rempah Asri", "Herbal", "Sari Handayani", "Sabun dan minyak herbal."},
		{"berkah", "PT Berkah Jaya", "Konveksi", "Yusuf Maulana", "Seragam dan kaos custom."},
		{"tolak", "Toko Serba Ada", "Ritel", "Dimas Anggara", "Menjual aneka barang."},
	}
	for _, r := range rows {
		if err := s.registerPartner(r.key, r.business, r.kind, r.contact, r.desc); err != nil {
			return fmt.Errorf("partner %s: %w", r.key, err)
		}
	}
	for _, key := range []string{"maju", "rempah"} {
		if _, err := s.partners.SetStatus(s.ctx, s.admin.Actor, s.partnerIDs[key], partner_usecase.StatusVerified, ""); err != nil {
			return err
		}
	}
	_, err := s.partners.SetStatus(s.ctx, s.admin.Actor, s.partnerIDs["tolak"], partner_usecase.StatusRejected, "Dokumen izin usaha belum dilampirkan.")
	return err
}

func (s *seeder) registerPartner(key, business, kind, contact, desc string) error {
	partner, err := s.partners.Register(s.ctx, partner_usecase.RegisterInput{
		OrganizationCode: demoCode, BusinessName: business, BusinessType: kind, ContactName: contact, Email: email(key),
		Password: s.password, Phone: "0812-0000-0000", Address: "Jl. Pasar Baru No. 5", Description: text(desc), Client: seedClient,
	})
	if err != nil {
		return err
	}
	s.partnerIDs[key] = partner.ID
	s.partnerUsers[key] = identity{Actor: s.actor(partner.UserID.Int64), Role: utils.RolePartner}
	return nil
}

type productRow struct {
	key, owner, unit, kind, name, category, desc string
	price                                        int64
	stock                                        int32
}

func stockOf(kind string, stock int32) sql.NullInt32 {
	return sql.NullInt32{Int32: stock, Valid: kind == "produk"}
}

func (s *seeder) seedProducts() error {
	rows := []productRow{
		{"gayo", "", "kopi", "produk", "Kopi Gayo 250 g", "Kopi", "Arabika sangrai sedang.", 65000, 40},
		{"susu", "", "kopi", "produk", "Kopi Susu Botol", "Minuman", "Siap minum, 250 ml.", 22000, 60},
		{"tumbler", "", "kopi", "produk", "Tumbler Edisi Lama", "Merchandise", "Produk lama yang sudah tidak dijual.", 90000, 5},
		{"banner", "", "cetak", "jasa", "Cetak Banner per Meter", "Cetak", "Bahan flexi 280 gsm.", 35000, 0},
		{"kartu", "", "cetak", "jasa", "Kartu Nama 1 Box", "Cetak", "Isi 100 lembar, art carton.", 55000, 0},
		{"aula", "", "wisma", "jasa", "Sewa Aula Harian", "Akomodasi", "Kapasitas 100 orang.", 2500000, 0},
		{"toraja", "maju", "", "produk", "Biji Kopi Toraja 1 kg", "Kopi", "Dipanen dari dataran tinggi Toraja.", 180000, 12},
		{"madu", "maju", "", "produk", "Madu Hutan 500 ml", "Herbal", "Stok sedang kosong.", 95000, 0},
		{"menu", "maju", "kopi", "jasa", "Konsultasi Menu Kopi", "Konsultasi", "Sesi 2 jam untuk pemilik kedai.", 400000, 0},
		{"sabun", "rempah", "", "produk", "Sabun Sereh Wangi", "Herbal", "Sabun batang alami.", 28000, 30},
	}
	for _, r := range rows {
		if err := s.createProduct(r); err != nil {
			return fmt.Errorf("produk %s: %w", r.key, err)
		}
	}
	return nil
}

func (s *seeder) createProduct(r productRow) error {
	seller := s.admin
	if r.owner != "" {
		seller = s.partnerUsers[r.owner]
	}
	product, err := s.tx.CreateProduct(s.ctx, seller, transaction_usecase.ProductInput{
		BusinessUnitID: id(s.unitIDs[r.unit]), Type: r.kind, Name: r.name, Category: r.category,
		Description: text(r.desc), Price: r.price, Stock: stockOf(r.kind, r.stock),
	})
	if err != nil {
		return err
	}
	s.productIDs[r.key] = product.ID
	return nil
}

func (s *seeder) seedCustomers() error {
	for _, c := range [][2]string{{"siti", "Siti Rahma"}, {"andi", "Andi Saputra"}} {
		if err := s.registerBuyer(c[0], c[1]); err != nil {
			return fmt.Errorf("customer %s: %w", c[0], err)
		}
	}
	_, err := s.tx.CreateCustomer(s.ctx, s.admin, transaction_usecase.CustomerInput{
		Name: "PT Sejahtera Abadi", Email: email("sejahtera"), Phone: text("021-555-0199"), Address: text("Jl. Sudirman Kav. 21"),
	})
	return err
}

func (s *seeder) registerBuyer(key, name string) error {
	err := s.shop.RegisterCustomer(s.ctx, storefront_usecase.CustomerRegistration{
		OrganizationCode: demoCode, Name: name, Email: email(key), Password: s.password, Phone: "0857-0000-0000", Client: seedClient,
	})
	if err != nil {
		return err
	}
	user, err := s.store.GetUserDetailByEmail(s.ctx, email(key))
	if err != nil {
		return err
	}
	s.buyers[key] = identity{Actor: s.actor(user.ID), Role: utils.RoleCustomer}
	customer, err := s.store.GetCustomerForUser(s.ctx, user.ID)
	s.customerIDs[key] = customer.ID
	return err
}

func (s *seeder) quote(partner, buyer, product string, qty int32, notes string) (int64, error) {
	seller := s.partnerUsers[partner]
	linked, err := s.tx.CreateCustomer(s.ctx, seller, transaction_usecase.CustomerInput{Name: buyer, Email: email(buyer), Phone: text("0857-0000-0000")})
	if err != nil {
		return 0, err
	}
	view, err := s.tx.CreateQuotation(s.ctx, seller, transaction_usecase.QuotationInput{
		CustomerID: linked.ID, Notes: text(notes), Items: []transaction_usecase.ItemInput{{ProductID: s.productIDs[product], Qty: qty}},
	})
	return view.Quotation.ID, err
}

func (s *seeder) seedQuotations() error {
	steps := []func() error{s.quoteWaiting, s.quoteBilled, s.quotePaid, s.quoteOffline, s.quoteRejected}
	for _, step := range steps {
		if err := step(); err != nil {
			return fmt.Errorf("quotation: %w", err)
		}
	}
	return nil
}

func (s *seeder) quoteWaiting() error {
	_, err := s.quote("maju", "siti", "menu", 1, "Penawaran sesi konsultasi bulan depan.")
	return err
}

func (s *seeder) invoiceApproved(buyer, product string, qty int32) (int64, error) {
	quotation, err := s.quote("maju", buyer, product, qty, "Pesanan partai besar.")
	if err != nil {
		return 0, err
	}
	if _, err := s.tx.RespondQuotation(s.ctx, s.buyers[buyer], quotation, transaction_usecase.QuotationApproved, ""); err != nil {
		return 0, err
	}
	invoice, err := s.tx.CreateInvoice(s.ctx, s.partnerUsers["maju"], quotation)
	return invoice.ID, err
}

// quoteBilled menyisakan invoice belum lunas dengan satu percobaan bayar gagal.
func (s *seeder) quoteBilled() error {
	invoice, err := s.invoiceApproved("siti", "toraja", 2)
	if err != nil {
		return err
	}
	_, err = s.tx.PayInvoice(s.ctx, s.buyers["siti"], invoice, transaction_usecase.PaymentFailed)
	return err
}

func (s *seeder) quotePaid() error {
	invoice, err := s.invoiceApproved("andi", "menu", 1)
	if err != nil {
		return err
	}
	_, err = s.tx.PayInvoice(s.ctx, s.buyers["andi"], invoice, transaction_usecase.PaymentSucceeded)
	return err
}

func (s *seeder) quoteOffline() error {
	quotation, err := s.quote("maju", "andi", "toraja", 1, "Dikonfirmasi lewat telepon.")
	if err != nil {
		return err
	}
	_, err = s.tx.RespondQuotation(s.ctx, s.partnerUsers["maju"], quotation, transaction_usecase.QuotationApproved, "Disetujui lewat WhatsApp oleh Andi.")
	return err
}

func (s *seeder) quoteRejected() error {
	quotation, err := s.quote("maju", "andi", "madu", 3, "Penawaran madu kemasan besar.")
	if err != nil {
		return err
	}
	_, err = s.tx.RespondQuotation(s.ctx, s.buyers["andi"], quotation, transaction_usecase.QuotationRejected, "")
	return err
}

func (s *seeder) buy(buyer, product string, qty int32) (int64, error) {
	actor := s.buyers[buyer].Actor
	if _, err := s.shop.AddToCart(s.ctx, actor, s.productIDs[product], qty); err != nil {
		return 0, err
	}
	orders, err := s.shop.Checkout(s.ctx, actor)
	if err != nil {
		return 0, err
	}
	return orders[0], nil
}

func (s *seeder) payOrder(buyer string, orderID int64) error {
	unpaid := sql.NullString{String: transaction_usecase.InvoiceUnpaid, Valid: true}
	page, err := s.tx.ListInvoices(s.ctx, s.buyers[buyer], unpaid, 100, 0)
	if err != nil {
		return err
	}
	for _, invoice := range page.Items {
		if invoice.OrderID.Int64 == orderID {
			_, err := s.tx.PayInvoice(s.ctx, s.buyers[buyer], invoice.ID, transaction_usecase.PaymentSucceeded)
			return err
		}
	}
	return fmt.Errorf("invoice order %d tidak ditemukan", orderID)
}

func (s *seeder) advance(by identity, orderID int64, statuses ...string) error {
	for _, status := range statuses {
		if _, err := s.tx.UpdateOrderStatus(s.ctx, by, orderID, status); err != nil {
			return err
		}
	}
	return nil
}

func (s *seeder) seedOrders() error {
	steps := []func() error{s.orderNew, s.orderProcessing, s.orderDone, s.orderCancelled, s.orderPartnerNew, s.orderBeforeSuspend}
	for _, step := range steps {
		if err := step(); err != nil {
			return fmt.Errorf("order: %w", err)
		}
	}
	return nil
}

func (s *seeder) orderNew() error {
	_, err := s.buy("siti", "gayo", 2)
	return err
}

func (s *seeder) orderProcessing() error {
	order, err := s.buy("andi", "banner", 4)
	if err != nil {
		return err
	}
	if err := s.payOrder("andi", order); err != nil {
		return err
	}
	return s.advance(s.admin, order, transaction_usecase.OrderProcessing)
}

func (s *seeder) orderDone() error {
	order, err := s.buy("siti", "toraja", 1)
	if err != nil {
		return err
	}
	if err := s.payOrder("siti", order); err != nil {
		return err
	}
	return s.advance(s.partnerUsers["maju"], order, transaction_usecase.OrderProcessing, transaction_usecase.OrderDone)
}

func (s *seeder) orderCancelled() error {
	order, err := s.buy("andi", "susu", 3)
	if err != nil {
		return err
	}
	return s.advance(s.buyers["andi"], order, transaction_usecase.OrderCancelled)
}

func (s *seeder) orderPartnerNew() error {
	_, err := s.buy("andi", "toraja", 2)
	return err
}

func (s *seeder) orderBeforeSuspend() error {
	_, err := s.buy("siti", "sabun", 2)
	return err
}

func (s *seeder) seedCart() error {
	actor := s.buyers["siti"].Actor
	for _, product := range []string{"susu", "kartu"} {
		if _, err := s.shop.AddToCart(s.ctx, actor, s.productIDs[product], 1); err != nil {
			return fmt.Errorf("keranjang: %w", err)
		}
	}
	return nil
}

func (s *seeder) seedServiceRequests() error {
	rows := []struct {
		buyer, unit, subject, message string
		steps                         []string
	}{
		{"siti", "kopi", "Pesanan kopi untuk rapat", "Mohon info harga 50 cup untuk Jumat.", nil},
		{"andi", "cetak", "Cetak undangan pernikahan", "Butuh 300 lembar, bisa antar?", []string{"diproses"}},
		{"siti", "cetak", "Revisi desain banner", "Terima kasih, hasil revisi sudah sesuai.", []string{"diproses", "selesai"}},
	}
	for _, r := range rows {
		if err := s.sendRequest(r.buyer, r.unit, r.subject, r.message, r.steps); err != nil {
			return fmt.Errorf("permintaan layanan: %w", err)
		}
	}
	return nil
}

func (s *seeder) sendRequest(buyer, unit, subject, message string, steps []string) error {
	created, err := s.shop.SendServiceRequest(s.ctx, s.buyers[buyer].Actor, storefront_usecase.ServiceRequestInput{
		BusinessUnitID: s.unitIDs[unit], ContactName: buyer, ContactEmail: email(buyer), ContactPhone: "0857-0000-0000", Subject: subject, Message: message,
	})
	if err != nil {
		return err
	}
	for _, status := range steps {
		if err := s.shop.SetServiceRequestStatus(s.ctx, s.admin.Actor, sql.NullInt64{}, created.ID, status); err != nil {
			return err
		}
	}
	return nil
}

// seedDashboardHistory mengisi 14 hari ke belakang agar filter periode terlihat.
func (s *seeder) seedDashboardHistory() error {
	sources := []sql.NullInt64{id(s.unitIDs["kopi"]), id(s.unitIDs["cetak"]), {}}
	now := time.Now().In(httpserver.WIB)
	for day := int64(1); day <= 14; day++ {
		date := sql.NullTime{Time: now.AddDate(0, 0, int(-day)), Valid: true}
		for i, unit := range sources {
			if err := s.pushDay(unit, date, day+int64(i)); err != nil {
				return fmt.Errorf("riwayat dashboard: %w", err)
			}
		}
	}
	return nil
}

func (s *seeder) pushDay(unit sql.NullInt64, date sql.NullTime, seed int64) error {
	metric := business_unit_usecase.Metric{OrganizationID: s.orgID, BusinessUnitID: unit, SourceModule: "m4", PeriodDate: date}
	items, revenue := metric, metric
	items.Key, items.Label, items.Unit, items.Value = "items_ordered", "Item dipesan", "item", 2+seed%6
	revenue.Key, revenue.Label, revenue.Unit, revenue.Value = "revenue", "Pendapatan lunas", "rupiah", 150000+seed*35000
	if err := business_unit_usecase.Push(s.ctx, s.store.Queries, items); err != nil {
		return err
	}
	return business_unit_usecase.Push(s.ctx, s.store.Queries, revenue)
}

// seedFinalStatuses dijalankan terakhir karena status ini menyembunyikan data
// dari katalog yang dibutuhkan langkah sebelumnya.
func (s *seeder) seedFinalStatuses() error {
	if err := s.tx.SetProductStatus(s.ctx, s.admin, s.productIDs["tumbler"], false); err != nil {
		return err
	}
	if err := s.units.SetStatus(s.ctx, s.admin.Actor, s.unitIDs["wisma"], false); err != nil {
		return err
	}
	_, err := s.partners.SetStatus(s.ctx, s.admin.Actor, s.partnerIDs["rempah"], partner_usecase.StatusSuspended, "Banyak keluhan keterlambatan pengiriman.")
	return err
}

func (s *seeder) seedApplications() error {
	rows := []struct{ code, name, local, status, reason string }{
		{"demo-bina-usaha", "Yayasan Bina Usaha", "pemohon.bina", "", ""},
		{"demo-kembang", "Kelompok Tani Kembang", "pemohon.kembang", organization_usecase.ApplicationRejected, "Profil organisasi belum lengkap."},
		{secondCode, "Koperasi Harapan", "admin.harapan", organization_usecase.ApplicationApproved, ""},
	}
	for _, r := range rows {
		if err := s.applyOrganization(r.code, r.name, r.local, r.status, r.reason); err != nil {
			return fmt.Errorf("pengajuan %s: %w", r.code, err)
		}
	}
	return nil
}

func (s *seeder) applyOrganization(code, name, local, status, reason string) error {
	application, err := s.orgs.Apply(s.ctx, organization_usecase.ApplicationInput{
		Code: code, OrganizationName: name, AdminName: "Pemohon " + name, AdminEmail: email(local), Password: s.password,
		Phone: "0811-0000-0000", Description: text("Ingin menjalankan marketplace untuk anggota."), Client: seedClient,
	})
	if err != nil || status == "" {
		return err
	}
	_, err = s.orgs.ReviewApplication(s.ctx, audit_usecase.Actor{Client: seedClient}, application.ID, status, reason)
	return err
}

func (s *seeder) printAccounts() {
	log.Println("data demo siap. Kata sandi semua akun = DEMO_PASSWORD")
	log.Printf("toko publik demo: /toko/%s", demoCode)
	rows := [][2]string{
		{"Administrator organisasi", "admin"}, {"Manajer (semua unit)", "manajer"}, {"Manajer unit kopi", "manajer.kopi"},
		{"Staf unit kopi", "staf.kopi"}, {"Staf unit cetak", "staf.cetak"}, {"Partner terverifikasi", "maju"}, {"Partner ditangguhkan", "rempah"},
		{"Partner pending", "berkah"}, {"Partner ditolak", "tolak"}, {"Customer", "siti"}, {"Customer", "andi"},
		{"Admin organisasi kedua", "admin.harapan"}, {"Pemohon organisasi (cek status)", "pemohon.bina"}, {"Pemohon ditolak (cek status)", "pemohon.kembang"},
	}
	for _, row := range rows {
		log.Printf("  %-32s %s", row[0], email(row[1]))
	}
}
