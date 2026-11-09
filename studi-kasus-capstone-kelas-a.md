# Studi Kasus Capstone Project — Araya Business Hub Platform (Kelas A: Fondasi)

Seluruh 21 User Story Kelas A (Fondasi) Araya Business Hub Platform, lengkap dengan 2 Acceptance Criteria (AC) per User Story.

> Catatan dari dokumen asli: dokumen ini untuk pegangan penilaian dosen, bukan untuk dibagikan ke mahasiswa.

## Ringkasan Modul

| Modul | Kelompok | Cakupan | Jumlah US |
|-------|----------|---------|-----------|
| M1 | Kel. A1 | Autentikasi, hak akses per divisi, audit log, manajemen akun | 4 |
| M2 | Kel. A2 | Unit usaha & Dashboard Operasional Dasar | 4 |
| M3 | Kel. A3 | Registrasi & verifikasi Partner | 4 |
| M4 | Kel. A4 | Produk/pelanggan, quotation, invoice, pembayaran, order | 4 |
| M5 | Kel. A5 | Katalog, form layanan, keranjang, checkout, pencarian | 5 |

Format AC: **Given** (kondisi awal) – **When** (aksi) – **Then** (hasil yang diharapkan).

---

# BAGIAN 1 — User Story & Acceptance Criteria Lengkap

## Modul 1 (M1) — Kel. A1

### M1-US1
**User Story:** Sebagai Administrator, saya ingin login dan mengatur hak akses per divisi, supaya setiap pengguna hanya bisa mengakses data sesuai kewenangannya.

**Acceptance Criteria:**
1. Given kredensial valid, When login, Then akses dibatasi sesuai divisi
2. Given kredensial invalid, When login, Then akses ditolak

### M1-US2
**User Story:** Sebagai Administrator, saya ingin sistem mencatat audit log aktivitas pengguna, supaya setiap perubahan data dapat ditelusuri.

**Acceptance Criteria:**
1. Given ada perubahan data, When disimpan, Then log mencatat siapa/apa/kapan
2. Given log ingin dilihat, When admin buka halaman audit log, Then riwayat tampil berurutan waktu

### M1-US3
**User Story:** Sebagai Administrator, saya ingin menambah dan menonaktifkan akun pengguna, supaya hanya orang yang berwenang yang dapat masuk.

**Acceptance Criteria:**
1. Given admin isi data akun valid, When disimpan, Then akun aktif & dapat login sesuai divisinya
2. Given akun dinonaktifkan, When pemilik akun coba login, Then akses ditolak & data historis akun tetap tersimpan

### M1-US4
**User Story:** Sebagai Pengguna, saya ingin mengubah kata sandi saya, supaya akun tetap aman.

**Acceptance Criteria:**
1. Given pengguna sudah login, When isi kata sandi lama yang benar & kata sandi baru yang memenuhi aturan, Then kata sandi berubah & login berikutnya memakai yang baru
2. Given kata sandi lama salah, When disimpan, Then sistem tolak & kata sandi tidak berubah

---

## Modul 2 (M2) — Kel. A2

### M2-US1
**User Story:** Sebagai Administrator, saya ingin mengelola data unit usaha (profil, kategori, divisi), supaya data yang sama dipakai konsisten oleh modul lain.

**Acceptance Criteria:**
1. Given admin input profil unit usaha, When disimpan, Then data tersedia dipakai modul lain
2. Given data unit usaha diubah, When disimpan, Then modul lain otomatis melihat versi terbaru

### M2-US2
**User Story:** Sebagai Administrator, saya ingin kerangka Dashboard Operasional Dasar tersedia, supaya modul lain tinggal mengalirkan data tanpa membangun dashboard terpisah.

**Acceptance Criteria:**
1. Given modul lain kirim data, When ditampilkan, Then shell dashboard menampilkan tanpa dashboard terpisah
2. Given belum ada data terkirim, When dashboard dibuka, Then tampil kondisi kosong yang wajar (bukan error)

### M2-US3
**User Story:** Sebagai Administrator, saya ingin menonaktifkan unit usaha yang tidak beroperasi, supaya tidak tampil di modul lain.

**Acceptance Criteria:**
1. Given unit usaha aktif, When admin nonaktifkan, Then unit tidak tampil di katalog & pilihan modul lain
2. Given unit usaha nonaktif, When admin aktifkan kembali, Then unit tampil lagi dengan data semula

### M2-US4
**User Story:** Sebagai Manajer, saya ingin menyaring dashboard per unit usaha dan periode, supaya dapat memantau kinerja yang dibutuhkan.

**Acceptance Criteria:**
1. Given dashboard berisi data beberapa unit usaha, When manajer pilih unit usaha & periode, Then hanya data sesuai filter yang tampil
2. Given filter tidak menghasilkan data, When diterapkan, Then tampil kondisi kosong yang wajar (bukan error)

---

## Modul 3 (M3) — Kel. A3

### M3-US1
**User Story:** Sebagai Partner, saya ingin mendaftar ke platform, supaya bisa memasarkan produk/jasa saya.

**Acceptance Criteria:**
1. Given partner isi form pendaftaran lengkap, When submit, Then status "pending"
2. Given partner isi form tidak lengkap, When submit, Then sistem tolak & tampilkan field yang kurang

### M3-US2
**User Story:** Sebagai Administrator, saya ingin memverifikasi (menyetujui/menolak) pendaftaran Partner, supaya hanya partner terverifikasi yang tampil di platform.

**Acceptance Criteria:**
1. Given status pending, When admin setujui, Then status "terverifikasi" & tampil di platform
2. Given status pending, When admin tolak, Then status "ditolak" + alasan tercatat

### M3-US3
**User Story:** Sebagai Partner, saya ingin melihat status pendaftaran saya, supaya tahu perkembangan verifikasi.

**Acceptance Criteria:**
1. Given partner sudah submit pendaftaran, When membuka halaman status, Then status terkini (pending/terverifikasi/ditolak) tampil
2. Given pendaftaran ditolak, When partner membuka halaman status, Then alasan penolakan tampil

### M3-US4
**User Story:** Sebagai Administrator, saya ingin menangguhkan dan mengaktifkan kembali Partner, supaya partner bermasalah tidak tampil di platform.

**Acceptance Criteria:**
1. Given partner terverifikasi, When admin tangguhkan + isi alasan, Then partner tidak tampil di platform & alasan tercatat
2. Given partner ditangguhkan, When admin aktifkan kembali, Then partner kembali tampil di platform

---

## Modul 4 (M4) — Kel. A4

### M4-US1
**User Story:** Sebagai Administrator/Partner, saya ingin menginput data produk/jasa dan pelanggan, supaya transaksi tercatat rapi.

**Acceptance Criteria:**
1. Given data produk/pelanggan valid, When disimpan, Then tersedia untuk transaksi
2. Given data tidak lengkap, When disimpan, Then sistem tolak & minta lengkapi

### M4-US2
**User Story:** Sebagai Partner, saya ingin membuat penawaran (quotation) dan invoice, supaya transaksi tercatat jelas dan dapat ditagih.

**Acceptance Criteria:**
1. Given quotation disetujui, When dibuat invoice, Then invoice berstatus "belum lunas"
2. Given quotation belum disetujui, When coba buat invoice, Then sistem tidak mengizinkan

### M4-US3
**User Story:** Sebagai Customer, saya ingin membayar Invoice melalui payment gateway (simulasi sandbox), supaya status Invoice otomatis lunas tanpa verifikasi manual.

**Acceptance Criteria:**
1. Given invoice belum lunas, When bayar via gateway, Then status otomatis "lunas"
2. Given pembayaran gagal, When gateway respon gagal, Then status invoice tetap "belum lunas"

### M4-US4
**User Story:** Sebagai Partner, saya ingin memperbarui status Order (diproses, selesai), supaya Customer mengetahui progres pesanannya.

**Acceptance Criteria:**
1. Given Order berstatus baru, When partner ubah ke "diproses", Then status tersimpan & waktu perubahan tercatat
2. Given Order berstatus "selesai", When partner coba ubah status lagi, Then sistem tolak

---

## Modul 5 (M5) — Kel. A5

### M5-US1
**User Story:** Sebagai Customer, saya ingin melihat profil dan katalog seluruh unit usaha BPU dalam satu halaman, supaya mudah menemukan produk/jasa yang dicari.

**Acceptance Criteria:**
1. Given customer buka halaman, When load, Then seluruh unit usaha tampil dalam 1 halaman
2. Given tidak ada unit usaha aktif, When load, Then tampil pesan kosong yang wajar

### M5-US2
**User Story:** Sebagai Customer, saya ingin mengisi formulir permintaan layanan, supaya dapat menghubungi unit usaha secara langsung.

**Acceptance Criteria:**
1. Given form diisi lengkap, When submit, Then permintaan diteruskan ke unit usaha
2. Given form tidak lengkap, When submit, Then sistem tolak & tampilkan validasi

### M5-US3
**User Story:** Sebagai Customer, saya ingin menambahkan produk ke keranjang saat menjelajahi katalog, supaya bisa mengumpulkan beberapa produk sebelum checkout.

**Acceptance Criteria:**
1. Given produk tampil di katalog, When klik "+ Keranjang", Then item tersimpan & badge bertambah 1
2. Given produk sudah ada di keranjang, When klik "+ Keranjang" lagi, Then qty bertambah (bukan entri baru)

### M5-US4
**User Story:** Sebagai Customer, saya ingin checkout dari keranjang menjadi Order, supaya produk yang saya pilih diproses jadi transaksi. (memakai kontrak Order dari M4)

**Acceptance Criteria:**
1. Given keranjang berisi item, When checkout, Then Order terbentuk lewat kontrak M4
2. Given keranjang kosong, When coba checkout, Then sistem cegah & tampilkan pesan

### M5-US5
**User Story:** Sebagai Customer, saya ingin mencari dan menyaring katalog berdasarkan kata kunci, kategori, atau unit usaha, supaya cepat menemukan produk/jasa yang dicari.

**Acceptance Criteria:**
1. Given katalog tampil, When customer cari kata kunci atau pilih kategori/unit usaha, Then hanya produk/jasa yang sesuai tampil
2. Given tidak ada hasil, When pencarian dijalankan, Then tampil pesan "tidak ditemukan" (bukan error)

---

