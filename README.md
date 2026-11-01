# Araya Business Hub - Backend Modul M1

Backend REST API untuk **Modul M1: Manajemen Pengguna dan Hak Akses** pada proyek Araya Business Hub. Modul ini menyediakan login terpusat dan token JWT yang dipakai modul lain.

## Status

| User Story | Cakupan | Status |
|---|---|---|
| M1-US1 | Login dan hak akses per divisi | Sprint 1: login, token, `/me`, logout, middleware sudah ada. Matriks hak akses (sprint 2) belum |
| M1-US2 | Audit log aktivitas pengguna | Belum. Saat ini pencatatan hanya ke log aplikasi |
| M1-US3 | Tambah dan nonaktifkan akun | Belum |
| M1-US4 | Ubah kata sandi | Belum |

## Teknologi

- Go (versi mengikuti `go.mod`)
- Gin sebagai HTTP framework
- PostgreSQL di Supabase, diakses lewat `pgx` dan `database/sql`
- sqlc untuk membuat kode akses database dari query SQL
- JWT HS256 (`golang-jwt/jwt/v5`) dan bcrypt (`golang.org/x/crypto`)
- Deploy ke Vercel (Go runtime)

## Struktur Folder

```
.
├── main.go                       entry point (lokal dan Vercel)
├── cmd/app/main.go               entry point lokal alternatif
├── config/                       pembacaan environment variable
├── internal/
│   ├── app/                      koneksi database
│   ├── controller/http/v1/       router dan route per modul (auth_route)
│   └── usecase/
│       ├── auth_usecase/         logika login
│       ├── audit_usecase/        pencatat audit sementara (log aplikasi)
│       ├── repository/queries/   query SQL untuk sqlc
│       ├── repository/sqlc/      kode hasil sqlc generate (jangan diedit)
│       └── utils/                bcrypt dan JWT
├── middleware/                   Authenticate dan RequireRole
├── pkg/httpserver/               amplop respons dan kode galat
├── prisma/migrations/            file migrasi SQL, sumber skema untuk sqlc
├── sqlc.yaml                     konfigurasi sqlc
└── .md/                          aturan pengembangan dan PRD modul M1
```

Folder kosong seperti `audit_route`, `user_route`, dan `user_usecase` disiapkan untuk user story berikutnya.

## Menjalankan Secara Lokal

Prasyarat: Go, akses ke database PostgreSQL (Supabase), dan sqlc bila ingin mengubah query.

```bash
git clone https://github.com/nadaaputri/ppl-m1-be.git
cd ppl-m1-be
git checkout dev
cp .env.example .env      # Windows: copy .env.example .env
go mod download
go run .
```

Server berjalan di `http://localhost:8080` (sesuai `APP_PORT`). Cek dengan `GET /health` dan `GET /ready`.

## Konfigurasi

Semua nilai dibaca dari environment variable. Untuk lokal boleh memakai file `.env` (tidak masuk repo). Di Vercel isi lewat dashboard.

| Variabel | Keterangan |
|---|---|
| `APP_NAME` | Nama layanan, dipakai sebagai `iss` pada token |
| `APP_PORT` | Port lokal, bawaan 8080. Di Vercel dipakai `PORT` |
| `DB_HOST`, `DB_PORT`, `DB_USER`, `DB_PASSWORD`, `DB_NAME` | Koneksi PostgreSQL. Untuk Supabase pakai host pooler |
| `DB_SSLMODE` | `require` untuk Supabase. Bawaan `disable` |
| `JWT_SECRET` | Rahasia penandatangan token, minimal 32 byte |
| `JWT_EXPIRES_IN` | Masa berlaku token, `8h` sesuai Panduan API Bersama |
| `GIN_MODE` | Opsional, isi `release` untuk deploy |

Aplikasi berhenti saat start bila `JWT_SECRET` kurang dari 32 byte, `JWT_EXPIRES_IN` tidak valid, atau konfigurasi database tidak lengkap. Variabel `APP_ENV`, `LOG_LEVEL`, dan `DATABASE_URL` pada `.env.example` belum dibaca kode Go.

## Database dan sqlc

Skema dibuat dari file migrasi SQL di `prisma/migrations`. Tabel yang ada: `divisions`, `roles`, `users`, `permissions`, `role_permissions`, dan `audit_logs`. Semua ID bertipe integer.

Query ditulis di `internal/usecase/repository/queries`. Setelah mengubah query atau menambah migrasi, jalankan:

```bash
sqlc generate
```

Setiap file migrasi baru harus ditambahkan ke daftar `schema` di `sqlc.yaml`.

## Endpoint

Semua endpoint M1 memakai awalan `/api/m1/v1` sesuai Panduan API Bersama.

| Method | Alamat | Akses | Fungsi |
|---|---|---|---|
| GET | `/health` | Publik | Cek server hidup |
| GET | `/ready` | Publik | Cek koneksi database |
| POST | `/api/m1/v1/auth/sessions` | Publik | Login |
| POST | `/api/m1/v1/auth/sessions/logout` | Bearer token | Logout |
| GET | `/api/m1/v1/me` | Bearer token | Profil pengguna dari token (kontrak K-01) |

### Contoh login

```http
POST /api/m1/v1/auth/sessions
Content-Type: application/json

{ "email": "kasir@contoh.id", "password": "sandi-pengguna" }
```

Respons 200:

```json
{
  "success": true,
  "data": {
    "access_token": "<jwt>",
    "token_type": "Bearer",
    "expires_in": 28800
  }
}
```

Token dikirim pada request berikutnya dengan header `Authorization: Bearer <access_token>`.

### Contoh profil

```json
{
  "success": true,
  "data": {
    "user_id": 42,
    "name": "Rani Kusumawati",
    "role": "kasir",
    "division": "penjualan",
    "unit_usaha_id": 3
  }
}
```

`division` dan `unit_usaha_id` bernilai `null` bila akun tidak memilikinya.

## Format Respons

```json
{ "success": true, "data": { } }
```

```json
{ "success": false, "error": { "code": "KODE_GALAT", "message": "Pesan" } }
```

Galat validasi (422) menambah `error.fields` berisi nama field dan pesannya.

| Status | `error.code` | Kapan |
|---|---|---|
| 400 | `BAD_REQUEST` | Body kosong, JSON rusak, atau tipe data salah |
| 401 | `INVALID_CREDENTIALS` | Login gagal, apa pun alasannya |
| 401 | `UNAUTHENTICATED` | Token tidak ada, salah, atau kedaluwarsa |
| 403 | `FORBIDDEN` | Role tidak berhak |
| 422 | `VALIDATION_ERROR` | Isian tidak lolos validasi |
| 500 | `INTERNAL_ERROR` | Kesalahan server, detail hanya di log |

## Token JWT

Token berlaku 8 jam dan berisi `user_id`, `name`, `role`, `division`, `unit_usaha_id`, dan `exp`, ditambah `iss` dan `iat`. Nama role dan divisi ditulis huruf kecil. Modul lain membaca `role` dan `unit_usaha_id` dari token tanpa bertanya ulang ke M1.

## Deploy ke Vercel

1. Hubungkan repositori ke project Vercel.
2. Di Settings, Build and Deployment, pilih **Framework Preset: Go**.
3. Isi environment variable di atas, untuk Production dan Preview.
4. Pastikan branch yang berisi kode adalah branch produksi (merge `dev` ke `main` bila Production memakai `main`).
5. Setelah deploy, cek `GET /health` dan `GET /ready`.

Vercel mendeteksi `main.go` di root dan menjalankan server pada port dari `PORT`.

## Keterbatasan yang Diketahui

- Token tidak disimpan di server. Logout hanya berlaku di sisi klien, dan akun yang dinonaktifkan tetap diterima sampai token kedaluwarsa (maksimal 8 jam).
- Pembatasan akses memakai `role` pada token. Skema database menyimpan izin per role, bukan per divisi.
- Belum ada pembatasan percobaan login berulang.
- Login dan kejadian audit baru tercatat di log aplikasi, belum di tabel `audit_logs`.
- Belum ada tes otomatis dan dokumentasi OpenAPI.

## Acuan

- Panduan API Bersama v1.0 (draf), 28 Oktober 2026
- `.md/prd-m1.md` dan `.md/aturan.md`
- Studi Kasus Capstone Project Kelas A, M1-US1 sampai M1-US4