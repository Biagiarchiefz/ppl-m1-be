# PRD MODUL 1 (M1): AUTENTIKASI, ORGANISASI, PENGGUNA, PERAN & IZIN, DAN AUDIT LOG

**Proyek:** Araya Business Hub Platform, Kelas A (Fondasi)
**Versi dokumen:** 3.0 (multi-organisasi), menggantikan v2.1. Disusun 4 Nov 2026.
**Revisi 5 Nov 2026:** hapus permanen pengguna dan organisasi, audit log dengan salinan pelaku, dan audit log platform untuk super admin (M1-EXT3).
**Revisi 7 Nov 2026:** aktivasi ulang akun, organisasi, peran, dan divisi (M1-EXT4), middleware `RequireDivision`, dan paket `pkg/m1auth` untuk modul lain.
**Pemilik modul:** Kelompok A1 (M1)

---

## 0. CARA MEMBACA DOKUMEN INI

1. **Ruang lingkup hanya M1.** Jangan membuat fitur modul lain (unit usaha, produk, order, dan sebagainya).
2. Bagian bertanda **[WAJIB SAMA]** dibaca oleh modul kelompok lain dan harus diikuti persis.
3. Bagian bertanda **[ARAHAN DOSEN]** adalah perubahan dari v2.1: platform dipakai **banyak organisasi**, dan tiap organisasi menyesuaikan pengguna, peran, izin, dan divisinya sendiri.
4. Bagian bertanda **[ASUMSI]** belum dikonfirmasi PO. Terapkan, tetapi pisahkan di konstanta/konfigurasi. Daftarnya ada di Bagian 13.
5. Implementasi acuan ada di repositori `ppl-m1-be` (Go + gin) dan `ppl-m1-fe` (React + Tailwind).

---

## 1. KONTEKS PROYEK

- Sistem ERP dibangun bersama oleh banyak kelompok (M1 sampai M14). Tiap modul punya database, backend, frontend, dan arsitektur sendiri, dan hanya terhubung lewat **kontrak API** dan **event** (Panduan API Bersama v1.0).
- **[ARAHAN DOSEN] Platform bersifat generik (multi-tenant).** Banyak organisasi memakai sistem yang sama. Tiap organisasi:
  - memiliki pengguna, peran, izin per peran, dan divisinya sendiri;
  - tidak dapat melihat data organisasi lain.
- **M1 adalah gerbang keamanan seluruh platform:**
  - satu-satunya modul yang punya login dan **menerbitkan JWT** (berlaku **8 jam**);
  - penyedia **kontrak K-01** `GET /api/m1/v1/me`;
  - pengelola **organisasi**, **peran**, dan **katalog izin** yang dipakai semua modul.
- Modul lain **membaca sendiri klaim token** (`organization_id`, `permissions`, `division`, `unit_usaha_id`) untuk memfilter data dan memeriksa hak akses, tanpa bertanya ulang ke M1 di setiap permintaan.

---

## 2. RUANG LINGKUP

### 2.1 Termasuk (in scope)

| ID | Fitur |
|---|---|
| M1-US1 | Login, penerbitan JWT, pembatasan akses per peran/izin dan divisi |
| M1-US2 | Pencatatan dan penampilan audit log (per organisasi) |
| M1-US3 | Menambah dan menonaktifkan akun pengguna |
| M1-US4 | Mengubah kata sandi |
| K-01 | `GET /api/m1/v1/me` |
| [ARAHAN DOSEN] | Pendaftaran dan pengelolaan organisasi oleh super admin platform |
| [ARAHAN DOSEN] | Peran dan izin yang diatur sendiri oleh tiap organisasi; divisi yang dikelola tiap organisasi |
| (pendukung) | Daftar/ubah pengguna, katalog izin, dokumentasi OpenAPI |

### 2.2 Tidak termasuk (non-goals)

- Registrasi mandiri pengguna atau organisasi (organisasi didaftarkan super admin).
- Refresh token, logout sisi server, reset kata sandi lewat email, 2FA, SSO.
- Satu email untuk banyak organisasi (email unik di seluruh platform).
- Data unit usaha (M2), partner (M3), customer/produk/order (M4).

---

## 3. KEPUTUSAN TEKNIS

| Lapisan | Pilihan |
|---|---|
| Backend | Go + **gin**, struktur `internal/controller` + `internal/usecase` |
| Akses database | SQL murni dengan `pgx` + **sqlc** |
| Migrasi | **golang-migrate** (`migrations/`, dijalankan `go run ./cmd/migrate up`) |
| Seeder | `go run ./cmd/seed`: katalog izin + akun super admin dari `.env` |
| Database | PostgreSQL (Supabase), database khusus M1 |
| JWT | HS256, `github.com/golang-jwt/jwt/v5` |
| Hash kata sandi | bcrypt cost 12 |
| Dokumentasi API | `docs/openapi.yaml` + `/api/m1/v1/docs` |
| Frontend | React + Tailwind CSS |

---

## 4. ATURAN PANDUAN API BERSAMA [WAJIB SAMA]

Tidak berubah dari v2.1:

- Alamat `/api/m1/v1/{resource}`. Resource berupa kata benda jamak, huruf kecil, dengan tanda hubung. Field JSON memakai snake_case.
- Header `Authorization: Bearer <token>`. Token tidak sah dijawab 401, izin tidak cukup dijawab 403. Endpoint publik ditandai `security: []`.
- Amplop: `{success, data}`, untuk daftar ditambah `meta {page, per_page, total, total_pages}`, dan untuk galat `{success:false, error:{code, message, fields}}`.
- `fields` hanya terisi untuk galat validasi. Untuk galat lain nilainya `null`.
- Daftar kosong dijawab 200 dengan `data: []`. Tanggal ISO 8601 WIB (`+07:00`), ID bilangan bulat, nilai kosong `null`.
- Parameter daftar: `page`, `per_page` (bawaan 20, maksimal 100), `from`/`to` (YYYY-MM-DD, keduanya ikut dihitung), `q`, filter per field, dan `sort` (minus berarti menurun).
- Data yang punya riwayat tidak dihapus fisik, hanya dinonaktifkan. `DELETE /{resource}/{id}` di M1 tetap berarti menonaktifkan. Pengecualian M1: hapus permanen pengguna dan organisasi lewat endpoint terpisah `/{id}/permanent`, hanya untuk data yang sudah nonaktif, dan audit log-nya tetap tersimpan (M1-EXT3).

---

## 5. ORGANISASI, TOKEN, PERAN, DAN IZIN

### 5.1 Jenis pengguna [ARAHAN DOSEN]

| Jenis | Keterangan |
|---|---|
| **Super admin platform** | Tidak terikat organisasi (`organization_id = null`). Hanya mengelola organisasi. Dibuat oleh seeder dari `ADMIN_EMAIL`/`ADMIN_PASSWORD`. |
| **Administrator organisasi** | Memakai peran bawaan `administrator` organisasinya, yang otomatis memiliki **semua izin non-platform**, termasuk izin modul lain yang ditambahkan kemudian. Dibuat bersamaan dengan organisasi. |
| **Pengguna organisasi** | Memakai peran yang dibuat administrator organisasinya, dengan izin yang dipilih. |

### 5.2 Isi JWT [WAJIB SAMA, disepakati semua modul]

```json
{
  "user_id": 42,
  "name": "Rani Kusumawati",
  "organization_id": 7,
  "role": "kasir-toko",
  "permissions": ["m4.invoice.create", "m4.order.update"],
  "division": "penjualan",
  "unit_usaha_id": 3,
  "exp": 1793180400
}
```

- Delapan klaim ini wajib ada. Enam klaim lama tetap ada, jadi perubahan ini bersifat tambahan.
- `exp` sama dengan waktu terbit ditambah 8 jam. Token juga memuat `iss` dan `iat`.
- `organization_id`: **wajib dipakai modul lain untuk memfilter data** agar organisasi tidak saling melihat data. Nilainya `null` hanya untuk super admin.
- `role`: kode peran di organisasi (bebas dibuat tiap organisasi), atau `superadmin`. **Modul lain tidak boleh memeriksa hak akses dari nama peran**; gunakan `permissions`.
- `permissions`: daftar kode izin. Super admin hanya memiliki `["platform.organizations.manage"]`.
- `division` dan `unit_usaha_id` boleh `null`.
- **[ASUMSI]** HS256 dengan `JWT_SECRET` yang dibagikan pengelola server bersama.

### 5.3 Katalog izin [WAJIB SAMA, dikelola M1]

- Format kode: `<modul>.<resource>.<aksi>`, huruf kecil, misalnya `m4.invoice.create`.
- Katalog global disimpan di tabel `permissions` M1 dan diisi lewat seeder (`cmd/seed/data.go`).
- **Modul lain mendaftarkan kodenya lewat pull request ke repositori kontrak bersama**, lalu M1 menambahkannya ke seeder. Kode yang sudah dirilis tidak boleh diganti nama.
- Izin modul `platform` khusus super admin dan tidak dapat dipilih organisasi.

Izin M1:

| Kode | Fungsi |
|---|---|
| `platform.organizations.manage` | Mengelola organisasi (super admin) |
| `m1.users.read` | Melihat daftar dan detail pengguna |
| `m1.users.manage` | Menambah, mengubah, dan menonaktifkan pengguna |
| `m1.roles.manage` | Mengelola peran dan izinnya |
| `m1.divisions.manage` | Mengelola divisi |
| `m1.audit.read` | Melihat audit log |

### 5.4 Peran dan divisi

- Peran dan divisi **milik satu organisasi**. Kode unik per organisasi, formatnya huruf kecil, angka, `-`/`_`, maksimal 50 karakter, dan **tidak dapat diubah** setelah dibuat karena masuk ke klaim token.
- Peran bawaan `administrator` (`is_system = true`) tidak dapat diubah atau dinonaktifkan.
- Peran/divisi yang masih dipakai pengguna aktif tidak dapat dinonaktifkan.
- Perubahan peran, izin, atau divisi berlaku pada **login berikutnya** (token stateless).

### 5.5 Matriks akses endpoint M1

| Endpoint | Akses |
|---|---|
| `POST /auth/login` | Publik |
| `GET /me`, `POST /auth/change-password` | Semua pengguna login, termasuk super admin |
| `/organizations` (semua, termasuk `/{id}/status` dan `/{id}/permanent`), `GET /platform/audit-logs` | `platform.organizations.manage` |
| `GET /users`, `GET /users/{id}` | `m1.users.read` atau `m1.users.manage` |
| `POST/PATCH/DELETE /users`, `PATCH /users/{id}/status`, `DELETE /users/{id}/permanent` | `m1.users.manage` |
| `GET /roles`, `GET /divisions` | izin kelolanya, atau `m1.users.read`/`m1.users.manage` (untuk dropdown) |
| `POST/PATCH/DELETE /roles`, `PATCH /roles/{id}/status`, `GET /permissions` | `m1.roles.manage` |
| `POST/PATCH/DELETE /divisions`, `PATCH /divisions/{id}/status` | `m1.divisions.manage` |
| `GET /audit-logs` | `m1.audit.read` |

Endpoint data organisasi (users, roles, divisions, audit-logs) selalu dibatasi ke `organization_id` di token. Data organisasi lain dijawab **404** (tidak dibedakan dari data yang memang tidak ada).

---

## 6. USER STORY DAN ACCEPTANCE CRITERIA

### M1-US1: Login dan hak akses

> Sebagai Administrator, saya ingin login dan mengatur hak akses per divisi, supaya setiap pengguna hanya bisa mengakses data sesuai kewenangannya.

- **AC contoh:** *Given* kredensial valid, *When* login, *Then* akses dibatasi sesuai divisi.
- **AC tambahan:**
  1. *Given* kredensial valid, *When* login, *Then* 200 berisi JWT dengan delapan klaim Bagian 5.2 (8 jam).
  2. *Given* kredensial tidak cocok, *When* login, *Then* 401 `INVALID_CREDENTIALS` "Kredensial invalid".
  3. *Given* administrator organisasi membuat peran dan memilih izinnya, *When* pengguna dengan peran itu login, *Then* `permissions` di token berisi tepat izin tersebut, dan endpoint yang butuh izin lain dijawab 403 `FORBIDDEN`.
  4. *Given* administrator mengubah peran/divisi pengguna, *When* disimpan, *Then* login berikutnya membawa nilai baru dan perubahan tercatat di audit log.
  5. **[ARAHAN DOSEN]** *Given* pengguna organisasi A, *When* membaca data pengguna organisasi B, *Then* 404 dan data B tidak bocor.
  6. **[ARAHAN DOSEN]** *Given* organisasi dinonaktifkan super admin, *When* penggunanya login dengan kredensial benar, *Then* 401 `ORGANIZATION_INACTIVE` "Organisasi dinonaktifkan".
  7. *Given* endpoint modul memakai `RequireDivision("kode-a", "kode-b")` dari `pkg/m1auth`, *When* pengguna dari divisi lain atau tanpa divisi memanggilnya, *Then* 403 `FORBIDDEN`. Endpoint M1 sendiri memakai izin, karena kode divisi dibuat bebas oleh tiap organisasi.

### M1-US2: Audit log

> Sebagai Administrator, saya ingin sistem mencatat audit log aktivitas pengguna, supaya setiap perubahan data dapat ditelusuri.

- **AC contoh:** *Given* ada perubahan data, *When* disimpan, *Then* log mencatat siapa/apa/kapan.
- **AC tambahan:**
  1. Setiap pembuatan, perubahan, dan penonaktifan pengguna, peran, divisi, dan organisasi, serta penggantian kata sandi, mencatat satu baris (pelaku, aksi, resource, `resource_id`, waktu, organisasi) **dalam transaksi yang sama**.
  2. Login berhasil/gagal tercatat (`auth.login_succeeded` / `auth.login_failed`) beserta IP. **Kata sandi tidak pernah tercatat.**
  3. `GET /audit-logs` hanya menampilkan catatan organisasi pemanggil, terbaru di atas, dengan pagination dan filter.
  4. Tidak ada data dijawab 200 `data: []`. `to` < `from` dijawab 400 `INVALID_PARAMETER`.
  5. Tanpa izin `m1.audit.read` dijawab 403 `FORBIDDEN`.

### M1-US3: Menambah dan menonaktifkan akun

- **AC contoh:** *Given* admin isi data akun valid, *When* disimpan, *Then* akun aktif & dapat login sesuai divisinya.
- **AC tambahan:**
  1. Data tidak lengkap, atau peran/divisi bukan milik organisasi, dijawab 422 `VALIDATION_ERROR` "Isian tidak valid" dengan `fields`.
  2. Email sudah terdaftar di platform dijawab 409 `EMAIL_DUPLICATE`.
  3. Akun nonaktif login dengan kredensial benar dijawab 401 `ACCOUNT_INACTIVE`, dan riwayatnya tetap ada.
  4. Menonaktifkan akun yang sudah nonaktif dijawab 409 `USER_ALREADY_INACTIVE`. Menonaktifkan akun sendiri dijawab 409 `CANNOT_DEACTIVATE_SELF`.

### M1-US4: Mengubah kata sandi

- **AC contoh:** kata sandi lama benar dan yang baru memenuhi aturan, maka kata sandi berubah dan login berikutnya memakai yang baru.
- **AC tambahan:** kata sandi lama salah dijawab 422 `OLD_PASSWORD_INVALID` (`fields.old_password`). Kata sandi baru tidak memenuhi aturan dijawab 422 `VALIDATION_ERROR` (`fields.new_password` "Kata sandi baru tidak memenuhi aturan"). Login dengan kata sandi lama ditolak.

**Aturan kata sandi [ASUMSI]:** minimal 8 karakter, mengandung huruf dan angka, maksimal 72 byte, tidak sama dengan kata sandi lama.

### [ARAHAN DOSEN] M1-EXT1: Pengelolaan organisasi oleh super admin

1. *Given* super admin mengisi kode, nama, dan data administrator pertama, *When* disimpan, *Then* organisasi, peran `administrator` bawaan, dan akun administrator terbentuk dalam satu transaksi, dan administrator itu bisa langsung login.
2. *Given* kode organisasi sudah dipakai, *Then* 409 `ORGANIZATION_CODE_DUPLICATE`.
3. *Given* super admin, *When* memanggil endpoint data organisasi (mis. `/users`), *Then* 403.

### [ARAHAN DOSEN] M1-EXT2: Peran, izin, dan divisi per organisasi

1. Administrator organisasi dapat menambah, mengubah, dan menonaktifkan peran beserta izinnya, serta divisi.
2. Izin yang tidak ada di katalog dijawab 422 dengan `fields.permissions`.
3. Kode peran/divisi ganda di organisasi yang sama dijawab 409 `ROLE_CODE_DUPLICATE` / `DIVISION_CODE_DUPLICATE`. Kode yang sama di organisasi lain boleh.
4. Peran bawaan tidak dapat diubah (409 `ROLE_SYSTEM_LOCKED`). Peran/divisi yang dipakai pengguna aktif tidak dapat dinonaktifkan (409 `ROLE_IN_USE` / `DIVISION_IN_USE`).

### M1-EXT3: Hapus permanen dengan audit log tetap utuh

1. *Given* akun sudah nonaktif, *When* administrator memanggil `DELETE /users/{id}/permanent`, *Then* akun terhapus, tercatat `user.deleted` dengan data terakhir akun di `old_value`, dan log lama yang pelakunya akun itu tetap menampilkan nama/email pelaku (`user_id` menjadi null).
2. *Given* akun masih aktif, *Then* 409 `USER_STILL_ACTIVE`.
3. *Given* organisasi sudah nonaktif, *When* super admin memanggil `DELETE /organizations/{id}/permanent`, *Then* organisasi beserta seluruh pengguna, peran, dan divisinya terhapus dalam satu transaksi, dan tercatat `organization.deleted`.
4. *Given* organisasi masih aktif, *Then* 409 `ORGANIZATION_STILL_ACTIVE`.
5. Audit log organisasi yang dihapus tetap tersimpan (`organization_id` null, `organization_name` tetap terisi) dan dapat dilihat super admin lewat `GET /platform/audit-logs`.

### M1-EXT4: Aktivasi ulang

**Akun**

1. *Given* akun nonaktif dengan peran dan divisi yang masih aktif, *When* administrator memanggil `PATCH /users/{id}/status` dengan `is_active: true`, *Then* akun bisa login lagi dan tercatat `user.activated`.
2. *Given* peran atau divisi akun sudah nonaktif, *Then* 409 `USER_REFERENCE_INACTIVE`. Administrator mengganti peran/divisi lewat `PATCH /users/{id}` lalu mengaktifkan akun.
3. *Given* akun sudah aktif, *When* diaktifkan, *Then* 409 `USER_ALREADY_ACTIVE`.
4. `is_active: false` berperilaku sama dengan `DELETE /users/{id}`, termasuk `USER_ALREADY_INACTIVE` dan `CANNOT_DEACTIVATE_SELF`.

**Organisasi, peran, dan divisi**

5. *Given* organisasi nonaktif, *When* super admin memanggil `PATCH /organizations/{id}/status` dengan `is_active: true`, *Then* penggunanya yang aktif bisa login lagi dan tercatat `organization.activated`. Akun yang sebelumnya sudah nonaktif tetap nonaktif.
6. *Given* peran atau divisi nonaktif, *When* administrator memanggil `PATCH /roles/{id}/status` atau `PATCH /divisions/{id}/status`, *Then* datanya bisa dipilih lagi untuk pengguna dan tercatat `role.activated` atau `division.activated`. Peran kembali dengan izin terakhirnya.
7. *Given* datanya sudah aktif, *Then* 409 `ORGANIZATION_ALREADY_ACTIVE`, `ROLE_ALREADY_ACTIVE`, atau `DIVISION_ALREADY_ACTIVE`.
8. `is_active: false` berperilaku sama dengan `DELETE` pada resource itu, termasuk `ROLE_IN_USE`, `ROLE_SYSTEM_LOCKED`, dan `DIVISION_IN_USE`.

---

## 7. SPESIFIKASI API M1

Base URL `/api/m1/v1`. Detail lengkap (request, response, contoh) ada di `docs/openapi.yaml`.

| Method | Endpoint | Fungsi |
|---|---|---|
| POST | `/auth/login` | Login (publik) |
| GET | `/me` | Profil K-01: klaim token + `email`, `organization_name`, `is_active` |
| POST | `/auth/change-password` | Ubah kata sandi sendiri |
| GET, POST | `/organizations` | Daftar / daftarkan organisasi + admin pertama (super admin) |
| GET, PATCH, DELETE | `/organizations/{id}` | Detail / ubah nama / nonaktifkan |
| DELETE | `/organizations/{id}/permanent` | Hapus permanen organisasi nonaktif beserta isinya |
| GET, POST | `/users` | Daftar (filter `q`, `role`, `division`, `unit_usaha_id`, `is_active`, `sort`) / tambah |
| GET, PATCH, DELETE | `/users/{id}` | Detail / ubah `name`, `role`, `division`, `unit_usaha_id` / nonaktifkan |
| PATCH | `/users/{id}/status`, `/organizations/{id}/status`, `/roles/{id}/status`, `/divisions/{id}/status` | Aktifkan atau nonaktifkan (`is_active`) |
| DELETE | `/users/{id}/permanent` | Hapus permanen akun nonaktif |
| GET, POST | `/roles` | Daftar peran beserta izinnya / tambah |
| GET, PATCH, DELETE | `/roles/{id}` | Detail / ubah nama, deskripsi, izin / nonaktifkan |
| GET | `/permissions` | Katalog izin yang dapat dipilih |
| GET, POST | `/divisions` | Daftar / tambah |
| GET, PATCH, DELETE | `/divisions/{id}` | Detail / ubah nama / nonaktifkan |
| GET | `/audit-logs` | Riwayat (filter `from`, `to`, `user_id`, `action`, `resource`, `sort`) |
| GET | `/platform/audit-logs` | Riwayat semua organisasi untuk super admin (filter sama + `organization_id`) |

`role` dan `division` di body maupun respons pengguna berupa **kode** di organisasi pemanggil.

---

## 8. DAFTAR KODE GALAT M1 [WAJIB SAMA setelah dirilis]

| Status | `code` | `message` |
|---|---|---|
| 400 | `INVALID_PARAMETER` | Parameter tidak valid |
| 401 | `INVALID_CREDENTIALS` | Kredensial invalid |
| 401 | `UNAUTHENTICATED` | Token tidak ada atau tidak sah |
| 401 | `ACCOUNT_INACTIVE` | Akun dinonaktifkan |
| 401 | `ORGANIZATION_INACTIVE` | Organisasi dinonaktifkan |
| 403 | `FORBIDDEN` | Anda tidak berhak mengakses sumber daya ini |
| 404 | `USER_NOT_FOUND` | Pengguna tidak ditemukan |
| 404 | `ROLE_NOT_FOUND` | Peran tidak ditemukan |
| 404 | `DIVISION_NOT_FOUND` | Divisi tidak ditemukan |
| 404 | `ORGANIZATION_NOT_FOUND` | Organisasi tidak ditemukan |
| 409 | `EMAIL_DUPLICATE` | Email sudah terdaftar |
| 409 | `USER_ALREADY_INACTIVE` | Akun sudah nonaktif |
| 409 | `CANNOT_DEACTIVATE_SELF` | Anda tidak dapat menonaktifkan akun sendiri |
| 409 | `USER_STILL_ACTIVE` | Nonaktifkan akun terlebih dahulu sebelum menghapus permanen |
| 409 | `USER_ALREADY_ACTIVE` | Akun sudah aktif |
| 409 | `USER_REFERENCE_INACTIVE` | Peran atau divisi akun ini sudah nonaktif. Ganti terlebih dahulu sebelum mengaktifkan akun |
| 409 | `ROLE_CODE_DUPLICATE` | Kode peran sudah dipakai |
| 409 | `ROLE_ALREADY_INACTIVE` | Peran sudah nonaktif |
| 409 | `ROLE_IN_USE` | Peran masih dipakai oleh pengguna aktif |
| 409 | `ROLE_SYSTEM_LOCKED` | Peran bawaan sistem tidak dapat diubah |
| 409 | `DIVISION_CODE_DUPLICATE` | Kode divisi sudah dipakai |
| 409 | `DIVISION_ALREADY_INACTIVE` | Divisi sudah nonaktif |
| 409 | `DIVISION_IN_USE` | Divisi masih dipakai oleh pengguna aktif |
| 409 | `ORGANIZATION_CODE_DUPLICATE` | Kode organisasi sudah dipakai |
| 409 | `ORGANIZATION_ALREADY_INACTIVE` | Organisasi sudah nonaktif |
| 409 | `ORGANIZATION_ALREADY_ACTIVE` | Organisasi sudah aktif |
| 409 | `ROLE_ALREADY_ACTIVE` | Peran sudah aktif |
| 409 | `DIVISION_ALREADY_ACTIVE` | Divisi sudah aktif |
| 409 | `ORGANIZATION_STILL_ACTIVE` | Nonaktifkan organisasi terlebih dahulu sebelum menghapus permanen |
| 422 | `VALIDATION_ERROR` | Isian tidak valid (`fields` terisi) |
| 422 | `OLD_PASSWORD_INVALID` | Kata sandi lama salah |
| 500 | `INTERNAL_ERROR` | Terjadi kesalahan pada server |

Pesan selain yang berasal dari panduan berstatus **[ASUMSI]**.

---

## 9. MODEL DATA

Sumber kebenaran: folder `migrations/` (`000001_init_m1`, `000002_permanent_delete`).

| Tabel | Kolom utama | Catatan |
|---|---|---|
| `organizations` | `id`, `code` (unik), `name`, `is_active`, `deactivated_at` | Tenant |
| `permissions` | `code` (PK), `module`, `description` | Katalog global |
| `roles` | `id`, `organization_id`, `code`, `name`, `description`, `is_system`, `is_active` | Unik (`organization_id`, `code`) |
| `role_permissions` | `role_id`, `permission_code` | Izin per peran |
| `divisions` | `id`, `organization_id`, `code`, `name`, `is_active` | Unik (`organization_id`, `code`) |
| `users` | `id`, `organization_id`, `name`, `email` (unik, tanpa beda huruf besar/kecil), `password_hash`, `role_id`, `division_id`, `unit_usaha_id`, `is_super_admin`, `is_active`, `created_by`, `deactivated_at` | FK komposit menjamin peran/divisi dari organisasi yang sama |
| `audit_logs` | `id`, `organization_id`, `organization_name`, `user_id`, `actor_name`, `actor_email`, `action`, `resource`, `resource_id`, `old_value`, `new_value`, `ip_address`, `user_agent`, `created_at` | Append-only. FK `user_id` → `users` dan `organization_id` → `organizations` dengan `ON DELETE SET NULL` |
| view `user_details` | pengguna + kode peran/divisi + status organisasi | Untuk login dan daftar |

Ketentuan:
- `unit_usaha_id` milik M2, disimpan tanpa foreign key.
- `actor_name`, `actor_email`, dan `organization_name` disalin saat log ditulis, sehingga log tetap terbaca setelah pengguna/organisasinya dihapus permanen. `resource_id` tanpa foreign key karena menunjuk ke tabel yang berbeda-beda sesuai `resource`.
- `users.created_by` memakai `ON DELETE SET NULL` agar akun yang pernah membuat akun lain tetap bisa dihapus.
- `old_value`/`new_value` tidak boleh memuat `password`, `password_hash`, atau token.
- Aksi audit: `user.created/updated/activated/deactivated/deleted/password_changed`, `auth.login_succeeded/failed`, `role.created/updated/activated/deactivated`, `division.created/updated/activated/deactivated`, `organization.created/updated/activated/deactivated/deleted`.

---

## 10. ATURAN BISNIS DAN KEAMANAN

1. Kata sandi disimpan sebagai hash bcrypt dan tidak pernah dikembalikan.
2. Login gagal tidak membedakan "email tidak ada" dan "kata sandi salah". `ACCOUNT_INACTIVE` / `ORGANIZATION_INACTIVE` hanya dikembalikan setelah kata sandi benar.
3. Hak akses diperiksa di middleware dari `permissions` di token, bukan hanya di UI.
4. Semua query data organisasi selalu dibatasi `organization_id` pemanggil.
5. Audit log ditulis dalam transaksi yang sama dengan perubahannya. Kalau gagal, perubahan dibatalkan.
6. JWT stateless: akun/organisasi yang dinonaktifkan masih diterima modul lain sampai `exp` (maksimal 8 jam). `/me` di M1 menolaknya.
7. Konfigurasi sensitif hanya dari `.env` (`JWT_SECRET`, kredensial database, `ADMIN_*`). `.env` tidak di-commit.
8. CORS dikonfigurasi lewat `CORS_ALLOWED_ORIGINS`.

---

## 11. FRONTEND M1

| Halaman | Akses | Isi |
|---|---|---|
| `/login` | Publik | Email + kata sandi, "Ingat saya" |
| `/platform/organizations` | Super admin | Daftar, daftarkan (dengan admin pertama), ubah nama, nonaktifkan, hapus permanen (bila nonaktif) |
| `/platform/audit-logs` | Super admin | Audit log semua organisasi, termasuk yang sudah dihapus |
| `/admin` | Pengelola organisasi | Ringkasan |
| `/admin/users` | `m1.users.read` / `manage` | Tabel, cari/filter, tambah, ubah peran/divisi, nonaktifkan, hapus permanen (bila nonaktif) |
| `/admin/roles` | `m1.roles.manage` | Peran + centang izin per modul |
| `/admin/divisions` | `m1.divisions.manage` | Tambah, ubah nama, nonaktifkan |
| `/admin/audit-logs` | `m1.audit.read` | Riwayat dengan nama/email pelaku, filter, dan pesan kosong |
| `/profile` | Semua, lewat ikon profil di navbar | Profil dari `/me`, dengan modal ubah kata sandi. Halaman awal pengguna tanpa izin admin |
| 403 / 404 | Semua | Halaman galat |

Menu disembunyikan berdasarkan `permissions`, tetapi tetap dijaga backend. Kalau API menjawab 401, token dihapus dan pengguna diarahkan ke `/login`.

---

## 12. KONTRAK, DOKUMENTASI, DAN PENGUJIAN

- `docs/openapi.yaml` memuat semua endpoint M1 dan ditampilkan di `/api/m1/v1/docs`.
- Repositori kontrak bersama perlu diperbarui: `contracts/K-01-me.yaml` (klaim baru), **katalog izin** semua modul, dan fixtures K-01.
- Uji end-to-end API (110 skenario: login, isolasi organisasi, izin, peran bawaan, nonaktif, aktivasi ulang, hapus permanen, audit organisasi dan platform) sudah dijalankan terhadap database Supabase, tetapi skripnya belum ada di repositori. Uji otomatis di repositori baru mencakup `pkg/m1auth`. Koleksi Postman **belum dibuat**.
- Perubahan dari v2.1 ke v3.0 menambah klaim token. Mengganti makna `role` (dari daftar peran tetap menjadi kode peran per organisasi) adalah **perubahan yang merusak** bagi modul yang memeriksa `role`, jadi semua modul harus beralih ke `permissions` (sudah disepakati).

---

## 13. PERTANYAAN TERBUKA

1. Kode izin modul M2–M5 belum dikirim kelompok masing-masing. Saat ini katalog hanya berisi izin M1.
2. Relasi divisi M1 dengan data divisi unit usaha di M2 (M2-US1).
3. Teks pesan galat dan aturan kata sandi yang bertanda [ASUMSI].
4. Cara modul lain memverifikasi tanda tangan JWT (HS256 dengan secret bersama, atau RS256).
5. Modul lain yang menyimpan `user_id` atau `organization_id` dari M1 perlu menyepakati perilaku saat data itu dihapus permanen di M1 (misalnya menyimpan salinan nama, seperti audit log M1).

---

## 14. DAFTAR PERIKSA SEBELUM INTEGRASI

- [x] Semua alamat memakai `/api/m1/v1/...`
- [x] JWT memuat delapan klaim, berlaku 8 jam
- [x] Amplop `success`/`data`/`error`, daftar memakai `meta`
- [x] Data organisasi terisolasi per `organization_id`
- [x] Hak akses berbasis `permissions`
- [x] Akun/organisasi nonaktif ditolak login, data historis tetap ada
- [x] `docs/openapi.yaml` sesuai kode
- [ ] Katalog izin semua modul di repositori kontrak
- [ ] K-01 versi baru dan fixtures di repositori kontrak
- [ ] Koleksi Postman dan uji otomatis

---

## LAMPIRAN: PERUBAHAN DARI v2.1

- Platform menjadi multi-organisasi: tabel `organizations`, super admin platform, organisasi didaftarkan beserta admin pertamanya.
- Peran tidak lagi daftar tetap global. Tiap organisasi membuat perannya sendiri dan memilih izin dari katalog global. Peran bawaan `administrator` otomatis memiliki semua izin organisasi.
- Divisi dapat dikelola tiap organisasi (sebelumnya hanya dari seeder).
- Token menambah `organization_id` dan `permissions`. Hak akses dibaca dari `permissions`, bukan `role`.
- `/me` menambah `organization_id`, `organization_name`, dan `permissions`.
- Endpoint baru: `/organizations`, `POST/PATCH/DELETE /roles`, `/permissions`, `POST/PATCH/DELETE /divisions`, `GET /roles/{id}`, `GET /divisions/{id}`.
- Kode galat baru: `ORGANIZATION_INACTIVE`, `*_NOT_FOUND`, `*_CODE_DUPLICATE`, `*_ALREADY_INACTIVE`, `ROLE_IN_USE`, `DIVISION_IN_USE`, `ROLE_SYSTEM_LOCKED`.
- Audit log menyimpan `organization_id` dan ditampilkan per organisasi.
- Revisi 5 Nov 2026: hapus permanen pengguna/organisasi (`/{id}/permanent`), kolom salinan `actor_name`/`actor_email`/`organization_name` dan FK `ON DELETE SET NULL` di `audit_logs`, endpoint `GET /platform/audit-logs`, kode galat `USER_STILL_ACTIVE` dan `ORGANIZATION_STILL_ACTIVE`.
- Revisi 7 Nov 2026: endpoint `PATCH /{users,organizations,roles,divisions}/{id}/status`, aksi `*.activated`, kode galat `*_ALREADY_ACTIVE` dan `USER_REFERENCE_INACTIVE`, paket `pkg/m1auth` dengan `RequireDivision`, dan tema gelap di frontend. Tidak ada perubahan skema database.
