# PRD MODUL 1 (M1): AUTENTIKASI, MANAJEMEN PENGGUNA, RBAC, DAN AUDIT LOG

**Proyek:** Araya Business Hub Platform (BPU Universitas Pasundan), Kelas A (Fondasi)
**Versi dokumen:** 2.1 (diselaraskan dengan Panduan API Bersama v1.0 draf, disusun 28 Okt 2026)
**Pemilik modul:** Kelompok A1 (M1)

---

## 0. CARA MEMBACA DOKUMEN INI (UNTUK CLAUDE CODE)

1. Baca seluruh dokumen sebelum menulis kode.
2. **Ruang lingkup hanya M1.** Jangan membuat modul lain (M2 dst.), tabel unit usaha, produk, order, dan sebagainya.
3. Bagian bertanda **[WAJIB SAMA]** berasal dari Panduan API Bersama dan harus diikuti persis, karena dibaca oleh modul-modul kelompok lain. Bagian bertanda **[BEBAS]** adalah keputusan internal M1.
4. Bagian bertanda **[ASUMSI]** adalah keputusan sementara yang belum dikonfirmasi dosen/Product Owner (PO). Terapkan dulu, tetapi pisahkan dalam kode (konstanta/konfigurasi) supaya mudah diubah. Daftar lengkapnya ada di Bagian 13.
5. Bila ada instruksi pengguna di chat yang bertentangan dengan dokumen ini, tanyakan dulu sebelum mengubah apa pun pada bagian **[WAJIB SAMA]**.

---

## 1. KONTEKS PROYEK

- Sistem ERP dibangun bersama oleh banyak kelompok. Tiap kelompok memegang satu modul (M1 sampai M14) dengan **database, kode backend, kode frontend, dan arsitektur masing-masing**.
- Antarmodul hanya terhubung lewat **kontrak API** (HTTP + JSON) dan **event**, sesuai Panduan API Bersama v1.0. Tidak ada akses database lintas modul dan tidak ada foreign key lintas modul.
- **M1 adalah gerbang keamanan seluruh platform:**
  - Satu-satunya modul yang punya fitur login dan **menerbitkan JWT** (berlaku **8 jam**).
  - Penyedia **kontrak K-01**: `GET /api/m1/v1/me` (profil pengguna dari token).
  - Pengelola **daftar peran** yang dipakai semua modul (Bagian 5.2).
- Modul lain **membaca sendiri klaim token** (`role`, `division`, `unit_usaha_id`) untuk memeriksa hak akses. Modul lain **tidak menanyakan ulang ke M1 di setiap permintaan**. Endpoint `/me` dipakai untuk mengambil profil lengkap, bukan sebagai pemeriksa token di setiap request.
- Integrasi lintas kelompok terjadi di Sprint 2. Sebelum itu, modul lain bekerja dengan data dari `fixtures/` atau mock server, jadi M1 harus menyediakan contoh data uji K-01 (Bagian 12).

---

## 2. RUANG LINGKUP

### 2.1 Termasuk (in scope)

| ID | Fitur |
|---|---|
| M1-US1 | Login, penerbitan JWT, pembatasan akses per peran dan divisi |
| M1-US2 | Pencatatan dan penampilan audit log |
| M1-US3 | Menambah dan menonaktifkan akun pengguna |
| M1-US4 | Mengubah kata sandi |
| K-01 | Endpoint `GET /api/m1/v1/me` |
| (pendukung) | Daftar peran dan divisi, daftar pengguna, ubah peran/divisi pengguna, dokumentasi OpenAPI |

### 2.2 Tidak termasuk (non-goals)

- Registrasi mandiri Customer/Partner (belum ada di user story dan belum ada kontraknya; lihat Pertanyaan Terbuka no. 3).
- Refresh token, logout sisi server, reset kata sandi lewat email, 2FA, SSO.
- Data unit usaha (milik M2), data partner (milik M3), data customer/produk/order (milik M4).
- Menerima event dari modul lain. Di daftar kontrak v1.0, M1 hanya menjadi penyedia K-01 dan tidak mengirim atau menerima event.

---

## 3. KEPUTUSAN TEKNIS [BEBAS]

Panduan membebaskan bahasa, framework, dan database di dalam modul. Berikut keputusan M1 (bukan lagi bahan diskusi):

| Lapisan | Pilihan |
|---|---|
| Backend | Go + **gin** |
| Akses database | **SQL murni** dengan `pgx` (driver) dan `sqlx` atau `sqlc`. Tanpa ORM berat. |
| Migrasi | `golang-migrate` (file `.sql` berurutan di `migrations/`) |
| Database | PostgreSQL, database terisolasi khusus M1 |
| JWT | `github.com/golang-jwt/jwt/v5` |
| Hash kata sandi | `bcrypt` (cost 12) atau `argon2id` |
| Konfigurasi | File `.env` (jangan di-commit; sediakan `.env.example`) |
| Dokumentasi API | `docs/openapi.yaml` + halaman `/api/m1/v1/docs` (Swagger UI atau Redoc) |
| Frontend | React + Tailwind CSS |

Alasan memilih SQL murni: entitas M1 sedikit (`users`, `roles`, `divisions`, `audit_logs`), relasinya sederhana, dan untuk autentikasi serta audit log kejelasan kueri (timestamp, IP, perubahan data) lebih penting daripada abstraksi ORM.

Struktur backend yang disarankan: `cmd/server`, `internal/{config,handler,middleware,service,repository,model}`, `migrations/`, `docs/openapi.yaml`.

---

## 4. ATURAN WAJIB DARI PANDUAN API BERSAMA V1.0 [WAJIB SAMA]

### 4.1 Alamat dan penamaan

- Pola: `https://[server-bersama]/api/{modul}/v1/{resource}`. Untuk M1: **`/api/m1/v1/...`**
- Resource: kata benda jamak, bahasa Inggris, huruf kecil, tanda hubung (`users`, `audit-logs`).
- Satu data: resource + ID (`/users/42`). Aksi perubahan status: `POST` ke resource + nama aksi.
- Nama field JSON: **snake_case, bahasa Inggris** (`unit_usaha_id` boleh tetap).
- Method: `GET` tidak boleh mengubah data; `POST` membuat data/menjalankan aksi (sukses membuat data = **201**); `PATCH` mengubah sebagian field; `DELETE` menonaktifkan atau menghapus. **Data yang punya riwayat tidak dihapus fisik, hanya dinonaktifkan.**
- Hindari kata kerja di alamat untuk operasi biasa (mis. `/getUsers`).

### 4.2 Autentikasi dan hak akses

| Situasi | Header |
|---|---|
| Pengguna memanggil API | `Authorization: Bearer <token>` (JWT dari M1, berlaku 8 jam) |
| Modul memanggil modul lain tanpa pengguna | `X-Service-Key: <kunci modul>` (dari `.env`, dibagikan pengelola server bersama) |
| Endpoint publik | Tanpa header; **harus ditandai "publik"** di OpenAPI (`security: []`) |

- Token tidak ada, salah, atau kedaluwarsa: **401**. Token benar tetapi peran tidak berhak: **403**.

### 4.3 Amplop respons

Sukses, satu data:

```json
{ "success": true, "data": { "id": 42 } }
```

Sukses, daftar (wajib ada `meta`):

```json
{
  "success": true,
  "data": [ { "id": 1 }, { "id": 2 } ],
  "meta": { "page": 1, "per_page": 20, "total": 57, "total_pages": 3 }
}
```

Galat:

```json
{
  "success": false,
  "error": { "code": "EMAIL_DUPLICATE", "message": "Email sudah terdaftar", "fields": { "email": "Email sudah terdaftar" } }
}
```

- `code`: HURUF_BESAR_DENGAN_GARIS_BAWAH, tidak pernah berubah. Modul pemakai membaca `code`, bukan `message`.
- `message`: bahasa Indonesia dan **sama persis dengan teks pesan di Acceptance Criteria** (supaya Test Case bisa memeriksa langsung). Karena itu, setiap AC galat di Bagian 6 menuliskan teks pesannya.
- `fields`: hanya terisi untuk galat validasi (422), berisi nama field ke pesan. Untuk galat lain nilainya `null`.

### 4.4 Kode status

| Kode | Arti | Contoh di M1 |
|---|---|---|
| 200 | Berhasil | Login berhasil, daftar audit log tampil |
| 201 | Data baru dibuat | Akun pengguna dibuat |
| 400 | Parameter tidak masuk akal | `to` lebih awal dari `from` di audit log |
| 401 | Belum login / token tidak sah | Kredensial invalid |
| 403 | Peran tidak berhak | Non-admin membuka audit log |
| 404 | Data tidak ditemukan | ID pengguna tidak ada |
| 409 | Bertentangan dengan data/status | Email ganda; menonaktifkan akun yang sudah nonaktif |
| 422 | Isian tidak lolos validasi | Field wajib kosong; kata sandi baru tidak memenuhi aturan |
| 500 | Galat server | **Tidak boleh** dipakai untuk kesalahan input pengguna |

**"Tidak ada data" bukan galat.** Daftar kosong dijawab `200` dengan `"data": []` dan `meta` yang tetap ada.

### 4.5 Format data

| Data | Format | Contoh |
|---|---|---|
| Tanggal | ISO 8601 | `"2026-10-28"` |
| Tanggal dan jam | ISO 8601 zona **WIB** | `"2026-10-28T09:15:00+07:00"` |
| ID | Bilangan bulat (bukan UUID) | `"user_id": 42` |
| Ya/tidak | `true` / `false` | `"is_active": true` |
| Kosong | `null`, bukan teks kosong | `"unit_usaha_id": null` |
| Status | Teks huruf kecil snake_case, bahasa Indonesia | `"lunas"` |

Field rujukan ke modul lain memakai `{nama_data}_id`. Yang relevan untuk M1: `user_id` (milik M1) dan `unit_usaha_id` (milik M2).

### 4.6 Parameter daftar

`page`, `per_page` (bawaan 20, maksimal 100), `from` dan `to` (`YYYY-MM-DD`, keduanya ikut dihitung), `q`, filter berdasarkan nama field (`?role=kasir`), `sort` (tanda minus = menurun, mis. `-created_at`).

---

## 5. TOKEN, PERAN, DAN DIVISI

### 5.1 Isi JWT [WAJIB SAMA]

```json
{
  "user_id": 42,
  "name": "Rani Kusumawati",
  "role": "kasir",
  "division": "penjualan",
  "unit_usaha_id": 3,
  "exp": 1793180400
}
```

- Enam klaim ini wajib ada. `exp` berupa Unix timestamp (detik), sama dengan waktu terbit + 8 jam.
- `unit_usaha_id` dan `division` boleh `null` untuk akun yang tidak terikat unit usaha/divisi (mis. administrator pusat).
- **[ASUMSI]** Algoritma HS256 dengan `JWT_SECRET` dari `.env`; modul lain memverifikasi dengan secret yang sama yang dibagikan pengelola server bersama. Panduan v1.0 belum menyebut cara modul lain memverifikasi tanda tangan token (lihat Pertanyaan Terbuka no. 1).

### 5.2 Daftar peran [WAJIB SAMA, dikelola M1]

Panduan menyatakan nama peran memakai daftar yang sama di semua modul, dikelola M1, dan ditulis di repositori kontrak. Nilai peran: huruf kecil. **[ASUMSI]** daftar awal, disusun dari user story dan contoh di panduan:

| Kode peran | Dipakai oleh |
|---|---|
| `administrator` | M1 dan banyak modul (kelola data, verifikasi, dsb.) |
| `manajer` | Dashboard (M2-US4) |
| `partner` | M3, M4 |
| `customer` | M4, M5 (token Customer muncul di contoh K-04) |
| `kasir` | POS (contoh di panduan) |

Peran baru dari kelompok lain diajukan lewat pull request ke repositori kontrak bersama, lalu M1 menambahkannya ke seed. File daftar peran harus dimasukkan ke repositori kontrak (Bagian 12).

### 5.3 Divisi

- `division` berupa teks kode huruf kecil (mis. `penjualan`), bukan `division_id`, karena panduan tidak mendefinisikan `division_id` sebagai field rujukan.
- M1 menyimpan daftar divisi sederhana (`divisions`) agar admin bisa memilih nilai yang valid. Data divisi di sisi unit usaha dikelola M2 (M2-US1); penyelarasannya ada di Pertanyaan Terbuka no. 4.

### 5.4 Matriks akses endpoint M1

| Endpoint | Publik | Semua pengguna login | Administrator |
|---|---|---|---|
| `POST /auth/login` | ✔ | | |
| `GET /me`, `POST /auth/change-password` | | ✔ | ✔ |
| `/users` (semua), `/audit-logs`, `/roles`, `/divisions` | | | ✔ |

---

## 6. USER STORY DAN ACCEPTANCE CRITERIA

Keterangan: **AC contoh** berasal dari lembar kerja studi kasus. **AC tambahan** adalah tambahan kelompok yang bisa disesuaikan saat Sprint Planning. Teks pesan galat di AC tambahan menjadi `message` persis di API. Pesan yang bukan berasal dari panduan ditandai **[ASUMSI]**.

### M1-US1: Login dan hak akses per divisi

> Sebagai Administrator, saya ingin login dan mengatur hak akses per divisi, supaya setiap pengguna hanya bisa mengakses data sesuai kewenangannya.

- **AC contoh:** *Given* kredensial valid, *When* login, *Then* akses dibatasi sesuai divisi.
- **AC tambahan:**
  1. *Given* kredensial valid, *When* login, *Then* respons 200 berisi token JWT yang memuat `user_id`, `name`, `role`, `division`, `unit_usaha_id`, dan `exp` (8 jam).
  2. *Given* kredensial tidak cocok, *When* login, *Then* respons 401 dengan `code` `INVALID_CREDENTIALS` dan pesan "Kredensial invalid", dan tidak ada token yang terbit.
  3. *Given* administrator mengubah `role`/`division` pengguna, *When* disimpan, *Then* login berikutnya menerbitkan token dengan nilai baru dan perubahan tercatat di audit log.
  4. *Given* pengguna non-administrator, *When* memanggil endpoint khusus administrator, *Then* respons 403 dengan `code` `FORBIDDEN` dan pesan "Anda tidak berhak mengakses sumber daya ini" **[ASUMSI]**.

### M1-US2: Audit log aktivitas pengguna

> Sebagai Administrator, saya ingin sistem mencatat audit log aktivitas pengguna, supaya setiap perubahan data dapat ditelusuri.

- **AC contoh:** *Given* ada perubahan data, *When* disimpan, *Then* log mencatat siapa/apa/kapan.
- **AC tambahan:**
  1. *Given* akun dibuat, diubah, dinonaktifkan, atau kata sandi diganti, *When* perubahan berhasil disimpan, *Then* satu baris audit log tercatat dengan pelaku (`user_id`), aksi, resource, `resource_id`, dan waktu, dalam transaksi database yang sama dengan perubahannya.
  2. *Given* login gagal atau berhasil, *When* diproses, *Then* tercatat `auth.login_failed` / `auth.login_succeeded` beserta alamat IP, dan **kata sandi tidak pernah** tercatat.
  3. *Given* administrator membuka halaman audit log, *When* data dimuat, *Then* riwayat tampil berurutan waktu (bawaan terbaru di atas) dengan pagination.
  4. *Given* belum ada catatan atau filter tidak menghasilkan data, *When* halaman dibuka, *Then* API menjawab 200 dengan `data: []` dan UI menampilkan pesan kosong yang wajar (bukan error).
  5. *Given* pengguna non-administrator, *When* memanggil `GET /audit-logs`, *Then* respons 403 `FORBIDDEN`.

### M1-US3: Menambah dan menonaktifkan akun

> Sebagai Administrator, saya ingin menambah dan menonaktifkan akun pengguna, supaya hanya orang yang berwenang yang dapat masuk.

- **AC contoh:** *Given* admin isi data akun valid, *When* disimpan, *Then* akun aktif & dapat login sesuai divisinya.
- **AC tambahan:**
  1. *Given* administrator mengirim data akun yang tidak lengkap, *When* disimpan, *Then* respons 422 `VALIDATION_ERROR`, pesan "Isian tidak valid" **[ASUMSI]**, dan `fields` menunjuk field yang bermasalah.
  2. *Given* email sudah terdaftar, *When* disimpan, *Then* respons 409 `EMAIL_DUPLICATE` dengan pesan "Email sudah terdaftar".
  3. *Given* akun dinonaktifkan, *When* pemilik akun mencoba login dengan kredensial benar, *Then* respons 401 `ACCOUNT_INACTIVE` dengan pesan "Akun dinonaktifkan" **[ASUMSI]**, dan data historis akun (termasuk audit log) tetap tersimpan.
  4. *Given* akun sudah nonaktif, *When* dinonaktifkan lagi, *Then* respons 409 `USER_ALREADY_INACTIVE`; *Given* administrator menonaktifkan akunnya sendiri, *Then* respons 409 `CANNOT_DEACTIVATE_SELF` **[ASUMSI]**.

### M1-US4: Mengubah kata sandi

> Sebagai Pengguna, saya ingin mengubah kata sandi saya, supaya akun tetap aman.

- **AC contoh:** *Given* pengguna sudah login, *When* isi kata sandi lama yang benar & kata sandi baru yang memenuhi aturan, *Then* kata sandi berubah & login berikutnya memakai yang baru.
- **AC tambahan:**
  1. *Given* kata sandi lama salah, *When* disimpan, *Then* respons 422 `OLD_PASSWORD_INVALID` dengan `fields.old_password` "Kata sandi lama salah", dan kata sandi tidak berubah.
  2. *Given* kata sandi baru tidak memenuhi aturan, *When* disimpan, *Then* respons 422 `VALIDATION_ERROR` dengan `fields.new_password` "Kata sandi baru tidak memenuhi aturan" **[ASUMSI]**.
  3. *Given* kata sandi berhasil diganti, *When* pengguna login dengan kata sandi lama, *Then* ditolak dengan "Kredensial invalid"; perubahan tercatat di audit log tanpa menyimpan nilai kata sandi.

**Aturan kata sandi [ASUMSI]:** minimal 8 karakter, mengandung huruf dan angka, tidak sama dengan kata sandi lama. Letakkan aturan ini di satu fungsi validator agar mudah diubah.

---

## 7. SPESIFIKASI API M1

Base URL: `https://[server-bersama]/api/m1/v1`. Semua respons memakai amplop Bagian 4.3.

### 7.1 Ringkasan endpoint

| Method | Endpoint | Fungsi | Autentikasi | Asal |
|---|---|---|---|---|
| `POST` | `/auth/login` | Login, terbitkan JWT | Publik | M1-US1 |
| `GET` | `/me` | Profil dari token (**K-01**) | Bearer | M1-US1, K-01 |
| `POST` | `/auth/change-password` | Ubah kata sandi sendiri | Bearer | M1-US4 |
| `POST` | `/users` | Tambah akun | Bearer (admin) | M1-US3 |
| `DELETE` | `/users/{id}` | Nonaktifkan akun (soft delete) | Bearer (admin) | M1-US3 |
| `GET` | `/audit-logs` | Riwayat audit log | Bearer (admin) | M1-US2 |
| `GET` | `/users` | Daftar akun (filter, pagination) | Bearer (admin) | pendukung |
| `GET` | `/users/{id}` | Detail satu akun | Bearer (admin) | pendukung |
| `PATCH` | `/users/{id}` | Ubah `name`, `role`, `division`, `unit_usaha_id` | Bearer (admin) | pendukung (M1-US1) |
| `GET` | `/roles` | Daftar peran | Bearer (admin) | pendukung |
| `GET` | `/divisions` | Daftar divisi | Bearer (admin) | pendukung |

Hanya `GET /me` (K-01) yang disebut langsung di Panduan sebagai kontrak antarmodul. Endpoint lain adalah rancangan M1 yang mengikuti aturan penamaan Panduan, dan semuanya wajib masuk `docs/openapi.yaml`. Endpoint pendukung bersifat internal M1 (boleh berbeda selama bentuk respons tetap mengikuti amplop).

### 7.2 `POST /auth/login` (publik)

Request:

```json
{ "email": "rani@example.com", "password": "rahasia123" }
```

Respons 200:

```json
{
  "success": true,
  "data": {
    "access_token": "<jwt>",
    "token_type": "Bearer",
    "expires_in": 28800,
    "user": { "user_id": 42, "name": "Rani Kusumawati", "role": "kasir", "division": "penjualan", "unit_usaha_id": 3 }
  }
}
```

Galat: `401 INVALID_CREDENTIALS` ("Kredensial invalid", untuk email tidak ada atau kata sandi salah, tanpa membedakan keduanya); `401 ACCOUNT_INACTIVE` (kredensial benar tetapi akun nonaktif); `422 VALIDATION_ERROR` (email/kata sandi kosong).

> Identitas login memakai `email` **[ASUMSI]**; bila tim memilih `username`, ubah di satu tempat saja.

### 7.3 `GET /me` (K-01)

Respons 200:

```json
{
  "success": true,
  "data": {
    "user_id": 42,
    "name": "Rani Kusumawati",
    "email": "rani@example.com",
    "role": "kasir",
    "division": "penjualan",
    "unit_usaha_id": 3,
    "is_active": true
  }
}
```

Galat: `401 UNAUTHENTICATED` ("Token tidak ada atau tidak sah"); `401 ACCOUNT_INACTIVE` bila akun dinonaktifkan setelah token terbit (M1 memeriksa database di endpoint ini).

### 7.4 `POST /auth/change-password` (Bearer)

Request: `{ "old_password": "...", "new_password": "..." }`

Respons 200: `{ "success": true, "data": { "changed_at": "2026-11-04T10:00:00+07:00" } }`

Galat: `422 OLD_PASSWORD_INVALID`, `422 VALIDATION_ERROR`, `401 UNAUTHENTICATED`.

### 7.5 `POST /users` (admin)

Request:

```json
{
  "name": "Budi Santoso",
  "email": "budi@example.com",
  "password": "Sandi12345",
  "role": "partner",
  "division": null,
  "unit_usaha_id": null
}
```

Respons **201**: objek pengguna `{ id, name, email, role, division, unit_usaha_id, is_active: true, created_at }` (tanpa `password`/`password_hash`).

Galat: `403 FORBIDDEN`, `409 EMAIL_DUPLICATE`, `422 VALIDATION_ERROR` (termasuk `role`/`division` yang tidak ada di daftar).

> `unit_usaha_id` disimpan sebagai bilangan bulat tanpa foreign key (data milik M2). Validasi ke K-02 (`GET /api/m2/v1/businessunits`) opsional dan baru dilakukan setelah M2 siap di Sprint 2.

### 7.6 `DELETE /users/{id}` (admin)

Menonaktifkan akun: `is_active = false`, `deactivated_at` diisi, **tidak ada penghapusan fisik**. Respons 200: `{ "success": true, "data": { "id": 42, "is_active": false } }`.

Galat: `404 USER_NOT_FOUND`, `409 USER_ALREADY_INACTIVE`, `409 CANNOT_DEACTIVATE_SELF`, `403 FORBIDDEN`.

### 7.7 `GET /audit-logs` (admin)

Parameter: `page`, `per_page`, `from`, `to`, `user_id` (pelaku), `action`, `resource`, `sort` (bawaan `-created_at`).

Respons 200:

```json
{
  "success": true,
  "data": [
    {
      "id": 301,
      "user_id": 1,
      "action": "user.deactivated",
      "resource": "users",
      "resource_id": 42,
      "old_value": { "is_active": true },
      "new_value": { "is_active": false },
      "ip_address": "10.0.0.5",
      "created_at": "2026-11-04T09:15:00+07:00"
    }
  ],
  "meta": { "page": 1, "per_page": 20, "total": 1, "total_pages": 1 }
}
```

Galat: `400 INVALID_PARAMETER` (mis. `to` < `from`), `403 FORBIDDEN`, `401 UNAUTHENTICATED`. Tidak ada endpoint untuk mengubah atau menghapus audit log.

### 7.8 Endpoint pendukung

- `GET /users`: filter `q` (nama/email), `role`, `division`, `unit_usaha_id`, `is_active`; pagination dan `sort`.
- `PATCH /users/{id}`: hanya field yang dikirim yang berubah. Perubahan `role`/`division` dicatat sebagai `user.updated` dengan `old_value`/`new_value`. Token yang sudah terbit tetap berlaku sampai habis; nilai baru berlaku pada login berikutnya.
- `GET /roles`, `GET /divisions`: daftar `{ code, name }` untuk dropdown di UI admin.

---

## 8. DAFTAR KODE GALAT M1 [WAJIB SAMA setelah dirilis, karena `code` tidak boleh berubah]

| Status | `code` | `message` | Keterangan |
|---|---|---|---|
| 401 | `INVALID_CREDENTIALS` | Kredensial invalid | Teks dari panduan |
| 401 | `UNAUTHENTICATED` | Token tidak ada atau tidak sah | Kode ini dipakai juga di contoh K-04 panduan |
| 401 | `ACCOUNT_INACTIVE` | Akun dinonaktifkan | [ASUMSI] |
| 403 | `FORBIDDEN` | Anda tidak berhak mengakses sumber daya ini | [ASUMSI] |
| 400 | `INVALID_PARAMETER` | Parameter tidak valid | [ASUMSI] |
| 404 | `USER_NOT_FOUND` | Pengguna tidak ditemukan | [ASUMSI] |
| 409 | `EMAIL_DUPLICATE` | Email sudah terdaftar | Pola sama dengan `NIK_DUPLICATE` di panduan |
| 409 | `USER_ALREADY_INACTIVE` | Akun sudah nonaktif | [ASUMSI] |
| 409 | `CANNOT_DEACTIVATE_SELF` | Anda tidak dapat menonaktifkan akun sendiri | [ASUMSI] |
| 422 | `VALIDATION_ERROR` | Isian tidak valid | `fields` terisi |
| 422 | `OLD_PASSWORD_INVALID` | Kata sandi lama salah | `fields.old_password` |
| 500 | `INTERNAL_ERROR` | Terjadi kesalahan pada server | Jangan bocorkan detail internal |

Menambah `code` baru termasuk perubahan yang boleh langsung. Mengganti nama atau menghapus `code` adalah perubahan yang merusak (Bagian 12.3).

---

## 9. MODEL DATA [BEBAS]

Semua ID memakai `BIGSERIAL`/`BIGINT` (bilangan bulat). Waktu memakai `TIMESTAMPTZ`; keluaran API dikonversi ke WIB (`+07:00`).

```sql
CREATE TABLE roles (
  code        VARCHAR(50) PRIMARY KEY,         -- administrator, manajer, partner, customer, kasir
  name        VARCHAR(100) NOT NULL,
  description TEXT
);

CREATE TABLE divisions (
  code      VARCHAR(50) PRIMARY KEY,           -- penjualan, ...
  name      VARCHAR(100) NOT NULL,
  is_active BOOLEAN NOT NULL DEFAULT TRUE
);

CREATE TABLE users (
  id             BIGSERIAL PRIMARY KEY,
  name           VARCHAR(150) NOT NULL,
  email          VARCHAR(255) NOT NULL,
  password_hash  TEXT NOT NULL,
  role           VARCHAR(50) NOT NULL REFERENCES roles(code),
  division       VARCHAR(50) REFERENCES divisions(code),   -- boleh NULL
  unit_usaha_id  BIGINT,                                   -- milik M2, TANPA foreign key
  is_active      BOOLEAN NOT NULL DEFAULT TRUE,
  created_by     BIGINT REFERENCES users(id),
  created_at     TIMESTAMPTZ NOT NULL DEFAULT now(),
  updated_at     TIMESTAMPTZ NOT NULL DEFAULT now(),
  deactivated_at TIMESTAMPTZ
);
CREATE UNIQUE INDEX users_email_lower_uidx ON users (lower(email));

CREATE TABLE audit_logs (
  id          BIGSERIAL PRIMARY KEY,
  user_id     BIGINT,                 -- pelaku; NULL bila login gagal dan email tidak dikenal
  action      VARCHAR(80) NOT NULL,   -- user.created, user.updated, user.deactivated, user.password_changed, auth.login_succeeded, auth.login_failed
  resource    VARCHAR(50) NOT NULL,   -- users, auth
  resource_id BIGINT,
  old_value   JSONB,
  new_value   JSONB,
  ip_address  VARCHAR(45),
  user_agent  TEXT,
  created_at  TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE INDEX audit_logs_created_at_idx ON audit_logs (created_at DESC);
CREATE INDEX audit_logs_user_id_idx ON audit_logs (user_id);
```

Catatan:
- `users` tidak pernah dihapus fisik, sehingga riwayat dan audit log tetap utuh.
- `audit_logs` bersifat append-only. Disarankan mencabut hak `UPDATE`/`DELETE` pada tabel ini untuk role database aplikasi.
- `old_value`/`new_value` tidak boleh memuat `password`, `password_hash`, atau token.
- **Seed** lewat migrasi/perintah terpisah: peran awal (Bagian 5.2), minimal satu divisi contoh, dan satu akun `administrator` awal yang kredensialnya diambil dari variabel `.env` (jangan di-hardcode di repositori).

---

## 10. ATURAN BISNIS DAN KEAMANAN

1. Kata sandi disimpan sebagai hash (bcrypt/argon2id), tidak pernah disimpan atau dikembalikan dalam bentuk asli.
2. Pesan login gagal tidak membedakan "email tidak ada" dan "kata sandi salah". `ACCOUNT_INACTIVE` hanya dikembalikan setelah kata sandi benar.
3. Pemeriksaan peran dilakukan di middleware, bukan hanya di UI. Endpoint admin selalu memeriksa `role == administrator` dari token.
4. Pencatatan audit log dilakukan dalam transaksi yang sama dengan perubahan datanya. Bila log gagal tersimpan, perubahan dibatalkan.
5. JWT bersifat stateless: tidak ada logout sisi server. Token pengguna yang dinonaktifkan masih berlaku di modul lain sampai `exp` (maksimal 8 jam); hal ini diterima sebagai konsekuensi aturan "modul tidak menanyakan ulang ke M1" (Pertanyaan Terbuka no. 2).
6. Konfigurasi sensitif (`JWT_SECRET`, `X-Service-Key` M1, kredensial database) hanya dari `.env`. Sediakan `.env.example`; jangan commit `.env`.
7. CORS dikonfigurasi lewat `.env` agar frontend modul lain dan frontend M1 bisa memanggil API.
8. Jangan menaruh data pribadi berlebih di token (tetap enam klaim di Bagian 5.1).

---

## 11. FRONTEND M1 (React + Tailwind) [BEBAS, ikuti Panduan UI Bersama v1.0 bila sudah tersedia]

| Halaman | Isi |
|---|---|
| `/login` | Form email dan kata sandi; menampilkan pesan galat dari `error.message` |
| `/change-password` | Kata sandi lama, baru, konfirmasi; galat per field dari `error.fields` |
| `/users` | Tabel pengguna (cari, filter, pagination), tambah akun, ubah peran/divisi, nonaktifkan dengan dialog konfirmasi (khusus admin) |
| `/audit-logs` | Tabel berurutan waktu, filter tanggal/pelaku/aksi, pagination, **kondisi kosong yang wajar** (khusus admin) |
| 403 / 404 | Halaman akses ditolak dan tidak ditemukan |

- Format tampilan (mis. "5 Okt 2026", "Rp 1.250.000") adalah urusan UI. API selalu mengirim data mentah.
- Bila API menjawab 401, hapus token lalu arahkan ke `/login`. Menu dan rute admin disembunyikan untuk non-admin (tetapi tetap dijaga di backend).
- Token disimpan di sisi klien dan dikirim lewat header `Authorization: Bearer` **[ASUMSI: `localStorage`/`sessionStorage`]**.

---

## 12. KONTRAK, DOKUMENTASI, DAN PENGUJIAN

### 12.1 Yang harus ada

| Artefak | Lokasi |
|---|---|
| Semua endpoint M1 (termasuk yang internal) | `docs/openapi.yaml` di repositori M1; `/auth/login` ditandai publik (`security: []`) |
| Halaman dokumentasi | `/api/m1/v1/docs` (otomatis dari OpenAPI) |
| Kontrak K-01 | `contracts/K-01-me.yaml` di repositori kontrak bersama (lewat pull request, disetujui semua pemakai, di-merge PO) |
| Contoh data uji | `fixtures/K-01-*.json` (sukses dan galat 401) agar modul lain bisa bekerja sebelum M1 selesai |
| Daftar peran | File daftar peran di repositori kontrak (Bagian 5.2) |
| Koleksi Postman | Minimal 1 uji sukses dan 1 uji galat per kontrak, dijalankan setiap endpoint berubah |

Kerangka OpenAPI mengikuti lampiran Panduan: `openapi: 3.0.3`, `securitySchemes` (`bearerAuth`, `serviceKey`), skema `Error` standar.

### 12.2 Pengujian

- Uji unit/integrasi untuk tiap AC di Bagian 6, termasuk pemeriksaan teks `message` persis.
- Uji kontrak K-01 (sukses dan galat) di Postman.
- Sprint 2: integration test bersama modul pemakai di server bersama; hasilnya dicatat di Dokumen Teknis kedua pihak.

### 12.3 Perubahan kontrak

- Boleh langsung: menambah field opsional, endpoint baru, atau nilai `error.code` baru.
- Merusak (butuh persetujuan semua pemakai; umumkan paling lambat 1 minggu sebelumnya atau buat `v2`): menghapus/mengganti nama field, mengubah tipe/format data, menjadikan field opsional menjadi wajib.
- Kontrak hanya sah bila tertulis di file OpenAPI/repositori kontrak. Kesepakatan lisan atau lewat chat tidak berlaku.

---

## 13. PERTANYAAN TERBUKA (konfirmasi ke dosen/PO)

1. **Verifikasi tanda tangan JWT oleh modul lain.** Panduan tidak menyebut caranya. Usulan: HS256 dengan secret bersama dari pengelola server, atau RS256 dengan kunci publik yang dibagikan.
2. **Akun nonaktif vs token yang sudah terbit.** Karena modul lain tidak bertanya ke M1, akun nonaktif tetap diterima sampai token habis. Apakah ini dapat diterima?
3. **Pembuatan akun Customer dan Partner.** Contoh K-04 memakai token Customer, tetapi M1-US3 hanya memungkinkan admin membuat akun, dan tidak ada kontrak registrasi di daftar K-01 s.d. K-14. Perlu diputuskan: admin membuat manual, endpoint registrasi publik baru (usulan kontrak), atau M3 memicu pembuatan akun partner setelah verifikasi.
4. **Kepemilikan data divisi.** M2-US1 mengelola divisi sebagai bagian data unit usaha, sedangkan token M1 membawa `division`. Perlu disepakati apakah M1 memiliki daftar divisi sendiri (seperti di dokumen ini) atau mengikuti M2.
5. **Audit log lintas modul.** AC M1-US2 berbunyi "ada perubahan data", tetapi tidak ada kontrak bagi modul lain untuk mengirim jejak audit ke M1. Dokumen ini membatasi audit log pada aktivitas data M1. Bila perlu lintas modul, harus diusulkan kontrak baru (mis. event `audit.recorded` lewat `X-Service-Key`).
6. **Daftar peran awal, aturan kata sandi, dan teks pesan galat** yang bertanda [ASUMSI] di dokumen ini.
7. Panduan API Bersama masih **draf v1.0**. Periksa perubahan sebelum Sprint 2.

---

## 14. DAFTAR PERIKSA SEBELUM INTEGRASI (diadaptasi dari Panduan, khusus M1)

- [ ] Semua alamat memakai `/api/m1/v1/...`
- [ ] JWT memuat enam klaim (`user_id`, `name`, `role`, `division`, `unit_usaha_id`, `exp`), berlaku 8 jam
- [ ] Semua respons memakai amplop `success` / `data` / `error` (daftar memakai `meta`)
- [ ] `message` galat sama persis dengan teks di AC; `code` huruf besar dan stabil
- [ ] Kode status sesuai tabel Bab 3 panduan; daftar kosong dijawab 200
- [ ] Tanggal ISO 8601 (WIB), ID bilangan bulat, nilai kosong `null`, nama field snake_case
- [ ] Akun nonaktif ditolak login dan data historis tidak terhapus
- [ ] `docs/openapi.yaml` sesuai kode dan terbuka di `/api/m1/v1/docs`
- [ ] K-01, daftar peran, dan fixtures sudah di-merge di repositori kontrak bersama
- [ ] Koleksi Postman (sukses + galat) berjalan tanpa galat di server bersama

---

## LAMPIRAN: PERUBAHAN DARI VERSI 2.0 (untuk manusia; boleh dihapus sebelum diberikan ke Claude Code)

- Klaim JWT dilengkapi (`name`, `exp`) dan peran `/me` diluruskan: modul lain membaca token sendiri, bukan memanggil `/me` setiap request.
- Ditambahkan daftar peran yang dikelola M1, amplop daftar dengan `meta`, tabel status 400/404/409/500, tabel kode galat, format data, dan parameter daftar.
- Label "sesuai Panduan v1.0" diperjelas: hanya `/me` yang disebut di panduan; endpoint lain adalah rancangan M1 yang mengikuti aturan penamaan.
- Ditambahkan endpoint pendukung (`GET/PATCH /users`, `GET /users/{id}`, `/roles`, `/divisions`) agar "mengatur hak akses per divisi" benar-benar bisa dilakukan.
- Bagian perbandingan ORM diganti menjadi keputusan teknis. "Prisma" dihapus karena itu ORM JavaScript/TypeScript, bukan untuk Go.
- Ditambahkan model data, aturan keamanan, halaman frontend, kewajiban OpenAPI/kontrak/Postman, serta pertanyaan terbuka.
- Acceptance Criteria dipisah menjadi AC contoh (lembar kerja) dan AC tambahan kelompok, dengan teks pesan galat yang eksplisit.
