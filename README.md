# Araya Business Hub - Backend Modul M1

Backend REST API untuk **Modul M1: Autentikasi, Organisasi, Pengguna, Peran & Izin, dan Audit Log** pada proyek Araya Business Hub. Platform dipakai **banyak organisasi** (multi-tenant): super admin platform mendaftarkan organisasi, lalu tiap organisasi mengatur pengguna, peran beserta izinnya, dan divisinya sendiri. M1 adalah satu-satunya modul yang menerbitkan JWT untuk seluruh platform.

Acuan: `PRD_M1_Fondasi_v3.0.md` (v2.1 disimpan sebagai riwayat) dan `.md/aturan.md`.

## Status

| User Story | Status |
|---|---|
| M1-US1 Login & hak akses | Selesai: login, `/me` (K-01), hak akses berbasis izin, isolasi organisasi, `RequireDivision` untuk modul lain |
| M1-US2 Audit log | Selesai: tercatat dalam transaksi yang sama, per organisasi |
| M1-US3 Tambah & nonaktifkan akun | Selesai, ditambah aktivasi ulang dan hapus permanen akun nonaktif |
| M1-US4 Ubah kata sandi | Selesai |
| Organisasi (arahan dosen) | Selesai: super admin mendaftarkan, menonaktifkan, dan menghapus permanen organisasi, serta melihat audit log semua organisasi |
| Peran, izin, divisi per organisasi | Selesai |

Uji otomatis di repositori baru mencakup `pkg/m1auth` (`go test ./...`). Belum ada: uji usecase dan API di repositori, koleksi Postman, fixtures K-01.

## Teknologi

- Go (versi mengikuti `go.mod`) dan Gin
- PostgreSQL (Supabase), diakses lewat `pgx` + `database/sql`
- sqlc untuk kode akses database dari query SQL murni
- golang-migrate untuk migrasi (file di `migrations/`, disematkan ke binary)
- JWT HS256 (`golang-jwt/jwt/v5`) dan bcrypt cost 12

## Struktur Folder

```
.
├── main.go                          entry point (lokal dan Vercel)
├── cmd/
│   ├── app/                         entry point lokal alternatif
│   ├── migrate/                     go run ./cmd/migrate up|down|version
│   └── seed/                        go run ./cmd/seed (katalog izin + super admin)
├── config/                          pembacaan environment variable
├── migrations/                      file migrasi golang-migrate (sumber skema sqlc)
├── internal/
│   ├── app/                         koneksi database
│   ├── controller/http/v1/          router, respond (pemetaan galat), *_route per resource
│   └── usecase/
│       ├── apperr/                  galat domain bersama
│       ├── auth_usecase/            login, /me, ubah kata sandi, perhitungan izin token
│       ├── organization_usecase/    organisasi (super admin)
│       ├── user_usecase/            pengguna organisasi
│       ├── role_usecase/            peran + izin, katalog izin
│       ├── division_usecase/        divisi organisasi
│       ├── audit_usecase/           pencatatan dan daftar audit log
│       ├── repository/              Store (transaksi) + query sqlc
│       └── utils/                   bcrypt, aturan kata sandi, JWT, kode izin
├── middleware/                      Authenticate, RequirePermission, RequireOrganization, CORS
├── pkg/httpserver/                  amplop respons, kode galat, pagination, WIB
├── pkg/m1auth/                      verifikasi JWT + middleware, bisa disalin ke modul lain
└── docs/openapi.yaml                dokumentasi API, tampil di /api/m1/v1/docs
```

## Menjalankan Secara Lokal

Prasyarat: Go dan database PostgreSQL (Supabase, pakai **Session pooler** port 5432).

```bash
cp .env.example .env          # Windows: copy .env.example .env, lalu isi nilainya
go mod download
go run ./cmd/migrate up       # buat tabel
go run ./cmd/seed             # katalog izin + akun super admin dari ADMIN_* di .env
go run .
```

Alur pertama kali:
1. Login sebagai super admin (`ADMIN_EMAIL` / `ADMIN_PASSWORD`).
2. Daftarkan organisasi lewat `POST /api/m1/v1/organizations` (atau halaman Organisasi di frontend) beserta administrator pertamanya.
3. Login sebagai administrator organisasi itu, lalu buat divisi, peran (pilih izinnya), dan pengguna.

Server berjalan di `http://localhost:8080`. Cek `GET /health` dan `GET /ready`, dokumentasi di `http://localhost:8080/api/m1/v1/docs`.

## Konfigurasi

| Variabel | Keterangan |
|---|---|
| `APP_NAME` | Nama layanan, dipakai sebagai `iss` pada token |
| `APP_PORT` | Port lokal, bawaan 8080. Di Vercel dipakai `PORT` |
| `DB_HOST`, `DB_PORT`, `DB_USER`, `DB_PASSWORD`, `DB_NAME` | Koneksi PostgreSQL. `DB_USER` Supabase berformat `postgres.<project_ref>` |
| `DB_SSLMODE` | `require` untuk Supabase. Bawaan `disable` |
| `JWT_SECRET` | Rahasia penandatangan token, minimal 32 byte |
| `JWT_EXPIRES_IN` | **`8h`**, sesuai kontrak bersama |
| `CORS_ALLOWED_ORIGINS` | Origin frontend dipisah koma, atau `*`. Kosong = tanpa header CORS |
| `ADMIN_EMAIL`, `ADMIN_NAME`, `ADMIN_PASSWORD` | Akun **super admin platform** untuk `cmd/seed` |
| `GIN_MODE` | Opsional, isi `release` untuk deploy |

## Database, Migrasi, dan sqlc

Tabel: `organizations`, `permissions`, `roles`, `role_permissions`, `divisions`, `users`, `audit_logs`, dan view `user_details`. Rinciannya ada di PRD v3.0 Bagian 9.

`audit_logs` menyimpan salinan nama/email pelaku dan nama organisasi saat kejadian, dengan FK `ON DELETE SET NULL`. Jadi pengguna dan organisasi bisa dihapus permanen tanpa menghilangkan riwayatnya.

Untuk dibagikan ke orang lain, gabungan akhir semua migrasi tersedia di `docs/schema.sql`. File ini bisa dijalankan sekali di SQL Editor Supabase dan sudah berisi katalog izin M1. Sumber kebenaran tetap `migrations/`.

Mengubah skema:

1. Buat pasangan file baru di `migrations/`, misalnya `000002_nama.up.sql` dan `.down.sql`. Jangan mengubah migrasi yang sudah dijalankan di database bersama.
2. `go run ./cmd/migrate up`
3. Sesuaikan query di `internal/usecase/repository/queries/`, lalu `sqlc generate`.

Menambah izin modul lain: tambahkan kodenya ke `cmd/seed/data.go`, lalu jalankan `go run ./cmd/seed` lagi. Seeder aman diulang. Peran bawaan `administrator` tiap organisasi otomatis mendapat izin baru itu.

## Endpoint

Semua endpoint memakai awalan `/api/m1/v1`. Detail di `docs/openapi.yaml`.

| Method | Alamat | Izin |
|---|---|---|
| POST | `/auth/login` | Publik |
| GET | `/me` | Login |
| POST | `/auth/change-password` | Login |
| GET, POST, PATCH, DELETE | `/organizations[/{id}]` | `platform.organizations.manage` (super admin) |
| PATCH | `/organizations/{id}/status` | `platform.organizations.manage` (aktifkan atau nonaktifkan) |
| DELETE | `/organizations/{id}/permanent` | `platform.organizations.manage` (organisasi harus sudah nonaktif) |
| GET | `/platform/audit-logs` | `platform.organizations.manage` |
| GET | `/users[/{id}]` | `m1.users.read` atau `m1.users.manage` |
| POST, PATCH, DELETE | `/users[/{id}]` | `m1.users.manage` |
| PATCH | `/users/{id}/status` | `m1.users.manage` (aktifkan atau nonaktifkan akun) |
| DELETE | `/users/{id}/permanent` | `m1.users.manage` (akun harus sudah nonaktif) |
| GET | `/roles[/{id}]` | `m1.roles.manage`, `m1.users.manage`, atau `m1.users.read` |
| POST, PATCH, DELETE | `/roles[/{id}]` | `m1.roles.manage` |
| PATCH | `/roles/{id}/status` | `m1.roles.manage` (aktifkan atau nonaktifkan) |
| GET | `/permissions` | `m1.roles.manage` |
| GET | `/divisions[/{id}]` | `m1.divisions.manage`, `m1.users.manage`, atau `m1.users.read` |
| POST, PATCH, DELETE | `/divisions[/{id}]` | `m1.divisions.manage` |
| PATCH | `/divisions/{id}/status` | `m1.divisions.manage` (aktifkan atau nonaktifkan) |
| GET | `/audit-logs` | `m1.audit.read` |

Endpoint daftar menerima `page` dan `per_page` (bawaan 20, maksimal 100). Data selalu dibatasi ke organisasi di token; data organisasi lain dijawab 404.

`DELETE /{id}` hanya menonaktifkan, dan data nonaktif bisa diaktifkan lagi lewat `PATCH /{id}/status`. Hapus permanen memakai `/{id}/permanent` dan hanya berlaku untuk akun dan organisasi yang sudah nonaktif.

## Token JWT

Berlaku 8 jam:

```json
{ "user_id": 42, "name": "Rani", "organization_id": 7, "role": "kasir-toko",
  "permissions": ["m1.users.read"], "division": "penjualan", "unit_usaha_id": 3, "exp": 1793180400 }
```

Modul lain **memfilter data dengan `organization_id`** dan **memeriksa hak akses dengan `permissions`**, bukan dengan nama `role`, karena nama peran dibuat bebas oleh tiap organisasi. Super admin memiliki `organization_id: null` dan izin `platform.organizations.manage`. Perubahan peran atau izin berlaku pada login berikutnya.

## Integrasi Modul Lain (M2–M5)

Tiap modul punya database dan backend sendiri. M1 menerbitkan token, lalu modul lain memverifikasinya sendiri tanpa memanggil M1 di setiap permintaan.

### 1. Alur

1. Frontend login ke `POST /api/m1/v1/auth/login` dan mendapat `data.access_token`.
2. Frontend mengirim `Authorization: Bearer <token>` ke modul mana pun.
3. Backend modul memverifikasi token, lalu memakai klaimnya untuk memfilter data dan memeriksa izin.

### 2. Yang perlu disamakan

| Di modul Anda | Nilai |
|---|---|
| Rahasia token | Sama dengan `JWT_SECRET` M1. Minta ke pengelola M1 lewat jalur privat, jangan di-commit |
| Penerbit token | Sama dengan `APP_NAME` M1 |
| Algoritma | HS256. Token kedaluwarsa wajib ditolak |

### 3. Backend Go + gin

Salin folder `pkg/m1auth` ke repositori Anda, misalnya menjadi `internal/m1auth`. Paket ini hanya butuh gin dan `github.com/golang-jwt/jwt/v5`. `go get` belum bisa dipakai karena nama modul di `go.mod` M1 adalah `ppl-m1-be`, bukan alamat repositori.

```go
verifier := m1auth.NewVerifier(os.Getenv("JWT_SECRET"), os.Getenv("M1_ISSUER"))

api := router.Group("/api/m4/v1", m1auth.Authenticate(verifier), m1auth.RequireOrganization())
api.POST("/invoices", m1auth.RequirePermission("m4.invoice.create"), createInvoice)
api.GET("/laporan", m1auth.RequireDivision("keuangan"), getLaporan)

func createInvoice(c *gin.Context) {
    claims := m1auth.GetClaims(c)
    // Simpan dan filter data dengan *claims.OrganizationID.
}
```

| Middleware | Hasil bila gagal |
|---|---|
| `Authenticate(verifier)` | 401 `UNAUTHENTICATED` |
| `RequirePermission(kode...)` | 403 `FORBIDDEN` bila tidak punya satu pun izin |
| `RequireDivision(kode...)` | 403 `FORBIDDEN` bila divisi tidak cocok atau kosong |
| `RequireOrganization()` | 403 `FORBIDDEN` untuk super admin |

### 4. Backend bukan Go

Pakai pustaka JWT di bahasa Anda dan terapkan hal yang sama: terima HS256 saja, cek `iss` dan `exp`, baca klaim seperti pada bagian Token JWT, dan balas galat dengan amplop Panduan API Bersama.

### 5. Aturan yang wajib diikuti

- **Selalu filter data dengan `organization_id`** dari token, jangan dari body atau query.
- **Periksa hak akses dengan `permissions`**, bukan dengan `role`. Nama peran dibuat bebas oleh tiap organisasi.
- **Pakai `RequireDivision` hanya untuk kode yang sudah disepakati.** Kode divisi juga dibuat bebas oleh tiap organisasi, jadi tiap organisasi harus membuat divisi dengan kode itu.
- **Daftarkan kode izin baru** dengan format `<modul>.<resource>.<aksi>` ke M1, lalu M1 menambahkannya ke `cmd/seed/data.go`.
- **Simpan `user_id` sebagai angka biasa tanpa foreign key**, karena database modul terpisah. Simpan juga salinan nama bila perlu ditampilkan, karena akun bisa dihapus permanen dan `GET /users/{id}` lalu menjawab 404.
- **Untuk aksi sensitif, cek status terkini lewat `GET /api/m1/v1/me`.** Token akun yang baru dinonaktifkan tetap sah sampai kedaluwarsa, maksimal 8 jam, sedangkan `/me` langsung menolaknya.

### 6. Contoh login

```bash
curl -X POST http://localhost:8080/api/m1/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"email":"admin@contoh.id","password":"KataSandi123"}'
```

## Format Respons dan Kode Galat

```json
{ "success": true, "data": { } }
{ "success": true, "data": [ ], "meta": { "page": 1, "per_page": 20, "total": 0, "total_pages": 0 } }
{ "success": false, "error": { "code": "KODE", "message": "Pesan", "fields": null } }
```

Daftar lengkap kode galat ada di PRD v3.0 Bagian 8 dan `pkg/httpserver/errors.go`. Pemetaan galat domain ke kode API ada di `internal/controller/http/v1/respond`.

## Deploy ke Vercel

1. Hubungkan repositori ke project Vercel, pilih **Framework Preset: Go**.
2. Isi environment variable di atas untuk Production dan Preview.
3. Jalankan migrasi dan seeder dari lokal ke database produksi.
4. Setelah deploy, cek `GET /health` dan `GET /ready`.

## Keterbatasan yang Diketahui

- JWT stateless: akun/organisasi yang dinonaktifkan tetap diterima modul lain sampai token kedaluwarsa (maksimal 8 jam). `/me` di M1 sudah menolaknya.
- Partner dan Customer belum bisa mendaftar sendiri. Ini menunggu keputusan tentang organisasi tujuan pendaftarnya.
- Modul lain belum punya jalur untuk menulis ke audit log M1.
- Katalog izin baru berisi izin M1. Izin modul lain menunggu daftar dari kelompoknya.
- Belum ada pembatasan percobaan login berulang.
