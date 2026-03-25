# Backend API Specification

> **Tujuan dokumen ini**: Mendefinisikan kontrak API antara mobile app (React Native/Expo) dan backend server yang akan dibuat sebagai pengganti MSW mock.

---

## Stack & Infrastruktur

| Komponen      | Pilihan                                |
|---------------|----------------------------------------|
| Language      | Go (stdlib `net/http` + `chi` router)  |
| Database      | PostgreSQL                             |
| Autentikasi   | Tidak ada (internal/dev only)          |
| Deployment    | Local development (`localhost:8080`)   |
| Tenancy       | Single merchant                        |

### Struktur Folder Backend

```
backend/
├── cmd/
│   └── server/
│       └── main.go                  # entrypoint: load env, koneksi DB, start server
├── internal/
│   ├── db/
│   │   ├── db.go                    # setup *sql.DB dari DATABASE_URL
│   │   ├── migrate.go               # jalankan migration SQL saat startup
│   │   └── migrations/
│   │       └── 001_init.sql         # DDL semua tabel
│   ├── handlers/
│   │   ├── profile.go              # GET /api/profile/:id, PUT /api/profile/:id
│   │   └── menu.go                # GET /api/menu, GET /api/menu/:accountType
│   ├── models/
│   │   ├── profile.go              # struct Profile, ProfileResponse, EditProfileRequest
│   │   └── menu.go                # struct Menu, MenuResponse
│   ├── repository/
│   │   ├── profile.go              # GetProfileByID(), EditProfile()
│   │   └── menu.go                # GetMenu(), GetMenuByAccountType()
│   └── server/
│       ├── router.go                # daftar route + middleware (CORS, logging)
│       └── server.go                # struct Server, inject dependencies
├── seed.sql                         # data awal: 1 profile + ~9 menu
├── .env.example
└── go.mod
```

---

## Database Schema

### Tabel `profile`
Berisi data profil user.

```sql
CREATE TABLE profile (
    id            UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    bank          VARCHAR(20) NOT NULL,
    branch        VARCHAR(20) NOT NULL,
    name          VARCHAR(3) NOT NULL,
    card_number   VARCHAR(11) NOT NULL,
    card_provider VARCHAR(20) NOT NULL,
    balance       BIGINT NOT NULL DEFAULT 0 -- dalam minor unit (pence/cent)
    currency      VARCHAR(3) NOT NULL DEFAULT 'IDR',
    accountType   VARCHAR(20) NOT NULL DEFAULT 'REGULAR' -- 'REGULAR' | 'PREMIUM'
);
```

### Tabel `menu`
Menu homepage. Diurutkan berdasarkan `index ASC` untuk pagination cursor-based.

```sql
CREATE TABLE menu (
    id          VARCHAR(50) PRIMARY KEY,
    index       NUMBER UNIQUE NOT NULL,
    type        VARCHAR(20) NOT NULL,   -- 'REGULAR' | 'PREMIUM'
    title       VARCHAR(50) NOT NULL,
    icon_url    TEXT NOT NULL,
    is_active   BOOLEAN NOT NULL DEFAULT TRUE
);
```

---

## API Endpoints

### Base URL
- **Mock (MSW)**: `http://localhost:3000`
- **Real backend**: `http://localhost:8080`

### Shared Headers (semua request)

| Header         | Wajib | Keterangan                                           |
|----------------|-------|------------------------------------------------------|
| `Content-Type` | Ya    | `application/json`                                   |

---

### 1. `GET /api/profile/:id`

Mengambil data profile user berdasarkan id.

**Request**
```
GET /api/profile/:id
```

**Response 200**
```json
{
  "id":"da08ecfe-de3b-42b1-b1ce-018e144198f5",
  "bank":"Citibank",
  "branch":"Tangerang",
  "name":"Jane Doe",
  "card_number":"12355478990",
  "card_provider": "Mastercard Platinum",
  "balance": 5000000,
  "currency":"IDR",
  "accountType":"REGULAR"
}
```

| Field               | Type   | Keterangan                        |
|---------------------|--------|-----------------------------------|
| `id`                | number | UUID unik profil pengguna         |
| `bank`              | number | Nama bank pengguna                |
| `branch`            | string | Nama cabang bank                  |
| `name`              | string | Nama pengguna                     |
| `card_number`       | string | Nomor kartu pengguna              |
| `card_provider`     | string | Penyedia kartu                    |
| `balance`           | string | Jumlah dalam minor unit (pence/cent). |
| `currency`          | string | `"IDR"` atau `"USD"`              |
| `accountType`       | string | `"REGULAR"` atau `"PREMIUM"`      |

**Error Responses**

| Status | Kondisi              |
|--------|----------------------|
| 500    | Database error       |
| 404    | Profile not found    |

---

### 2. `PUT /api/profile/:id`

Mengubah data profile pengguna berdasarkan id.

**Request**
```
PUT /api/profile/da08ecfe-de3b-42b1-b1ce-018e144198f5
```

**Request Body**
```json
{
  "bank":"Citibank",
  "branch":"Tangerang",
  "name":"Jane Doe",
  "card_number":"12355478990",
  "card_provider": "Mastercard Platinum",
  "currency":"IDR"
}
```
| Field               | Type   | Keterangan                        |
|---------------------|--------|-----------------------------------|
| `bank`              | number | Nama bank pengguna                |
| `branch`            | string | Nama cabang bank                  |
| `name`              | string | Nama pengguna                     |
| `card_number`       | string | Nomor kartu pengguna              |
| `card_provider`     | string | Penyedia kartu                    |
| `currency`          | string | `"IDR"` atau `"USD"`              |

**Response 200**
```json
{
  "code": 200,
  "description": "Data pengguna berhasil diubah."
}
```

**Error Responses**

| Status | Error Code            | `recoverable` | Kondisi                                                   |
|--------|-----------------------|---------------|-----------------------------------------------------------|
| 400    | `validation_error`    | `false`       | currency tidak didukung                                   |
| 503    | `service_unavailable` | `true`        | Backend sedang down / maintenance                         |

```json
{
  "code": 400,
  "description": "Currency tidak didukung"
}
```

### 3. `GET /api/menu`

Mengambil daftar menu yang tersedia.

**Request**
```
GET /api/merchant/menu
```

**Response 200**
```json
{
  "menus": [
    {
      "id": "menu_001",
      "index": 1,
      "type": "REGULAR",
      "title": "Account and Card",
      "icon_url": "https://fonts.googleapis.com/css2?family=Material+Symbols+Outlined:opsz,wght,FILL,GRAD@20..48,100..700,0..1,-50..200&icon_names=id_card",
      "is_active": true
    },
    {
      "id": "menu_002",
      "index": 2,
      "type": "PREMIUM",
      "title": "Transfer",
      "icon_url": "https://fonts.googleapis.com/css2?family=Material+Symbols+Outlined:opsz,wght,FILL,GRAD@20..48,100..700,0..1,-50..200&icon_names=send_money",
      "is_active": true
    }
  ]
}
```

| Field       | Type   | Keterangan                                    |
|-------------|--------|-----------------------------------------------|
| `id`        | string | ID unik menu                                  |
| `index`     | number | Urutan menu                                   |
| `type`      | string | `"REGULAR"` atau `"PREMIUM"`                  |
| `title`     | string | Judul menu                                    |
| `icon_url`  | string | URL untuk icon menu                           |
| `is_active` | boolean| apakah menu aktif atau tidak                  |

**Error Responses**

| Status  Kondisi                                                    |
|--------|-----------------------------------------------------------|
| 500    | Database error                                            |

---

### 4. `GET /api/menu/:accountType`

Mengambil daftar menu yang tersedia berdasarkan accountType. Jika accountType-nya `"PREMIUM"`, semua daftar menu diambil.

**Request**
```
GET /api/menu/REGULAR
```

**Response 200**
```json
{
  "menus": [
    {
      "id": "menu_001",
      "index": 1,
      "type": "REGULAR",
      "title": "Account and Card",
      "icon_url": "https://fonts.googleapis.com/css2?family=Material+Symbols+Outlined:opsz,wght,FILL,GRAD@20..48,100..700,0..1,-50..200&icon_names=id_card",
      "is_active": true
    }
  ]
}
```

**Error Responses**

| Status | Kondisi                    |
|--------|----------------------------|
| 500    | Database error             |

---

## Business Logic

### Idempotency
- Client men-generate UUID v4 via `expo-crypto` satu kali saat user confirm payout.
- UUID dikirim sebagai header `Idempotency-Key` di setiap attempt (termasuk retry).
- Backend menyimpan key di kolom `idempotency_key UNIQUE`.
- Jika INSERT gagal karena duplicate key → query SELECT untuk kembalikan data lama.
- Jika response 200 vs 201: **frontend tidak perlu membedakan** — keduanya dianggap sukses.

### Menu
- Jika accountType yang digunakan di Endpoint `GET /api/menu/:accountType` adalah `"PREMIUM"`, semua daftar menu diambil datanya.

---

## Response Format

Semua response (termasuk error) menggunakan format:
```json
{
  "code": "<code error>",
  "description": "<deskripsi status code>"
}
```

Mapping ke client error hierarchy (`core/api/errors.ts`):

| HTTP Status | Client Error Class        |
|-------------|---------------------------|
| 400         | `InsufficientFundsError`  |
| 422         | `ApiError` (non-recoverable) |
| 503         | `ServiceUnavailableError` |
| Network     | `NetworkError`            |

---

## Middleware

### CORS
Wajib karena frontend (Expo web) berjalan di domain/port berbeda dari backend.

```
Access-Control-Allow-Origin: *
Access-Control-Allow-Methods: GET, POST, OPTIONS
Access-Control-Allow-Headers: Content-Type, Idempotency-Key
```

### Request Logging
Log setiap request: method, path, status, latency. Format:
```
[2026-03-16T10:00:00Z] GET /api/merchant 200 (12ms)
```

---

## Environment Variables

File: `backend/.env`

```env
DATABASE_URL=postgres://postgres:password@localhost:5432/merchant_db?sslmode=disable
PORT=8080
```

File: `backend/.env.example` (di-commit ke repo):
```env
DATABASE_URL=postgres://USER:PASSWORD@localhost:5432/merchant_db?sslmode=disable
PORT=8080
```

---

## Frontend Changes yang Diperlukan

### 1. `constants/index.ts`
Ganti `API_BASE_URL` agar bisa dikonfigurasi via env:
```ts
export const API_BASE_URL = process.env.EXPO_PUBLIC_API_URL ?? "http://localhost:8080";
```

### 2. `mocks/useMSW.ts`
Tambah flag env untuk disable MSW saat pakai real backend:
```ts
const USE_MOCK = process.env.EXPO_PUBLIC_USE_MOCK !== "false";

if (__DEV__ && USE_MOCK) { ... }
```

### 3. `.env.local` (tidak di-commit)
```env
# Pakai MSW (default dev)
EXPO_PUBLIC_USE_MOCK=true
EXPO_PUBLIC_API_URL=http://localhost:3000

# Pakai backend Go
# EXPO_PUBLIC_USE_MOCK=false
# EXPO_PUBLIC_API_URL=http://localhost:8080
```

### 4. `core/api/client.ts`
Handle HTTP 409 (idempotency conflict) — return response body tanpa throw error:
```ts
// Idempotency hit: treat as success
if (response.status === 409 || response.status === 200) {
    // handled normally
}
```
> Tidak diperlukan jika backend menggunakan pola 200/201 (bukan 409).

---

## Implementation Checklist

### Backend
- [ ] `go.mod` + install dependensi (`chi`, `lib/pq`, `godotenv`)
- [ ] `internal/db/db.go` — koneksi PostgreSQL
- [ ] `internal/db/migrations/001_init.sql` — DDL 3 tabel
- [ ] `internal/db/migrate.go` — auto-migrate saat startup
- [ ] `seed.sql` — 1 profile + 9 menu
- [ ] `internal/models/profile.go` + `models/menu.go`
- [ ] `internal/repository/profile.go` — `GetProfileByID()`, `EditProfile()`
- [ ] `internal/repository/menu.go` — `GetMenu()`, `GetMenuByAccountType()`
- [ ] `internal/handlers/profile.go` — 2 endpoints
- [ ] `internal/handlers/menu.go` — 2 endpoints
- [ ] `internal/server/router.go` — register routes + CORS + logging middleware
- [ ] `cmd/server/main.go` — entrypoint
- [ ] `.env.example`

### Frontend
- [ ] `constants/index.ts` — env-aware `API_BASE_URL`
- [ ] `mocks/useMSW.ts` — `EXPO_PUBLIC_USE_MOCK` flag
- [ ] `.env.local` — local override (gitignored)

### Verification
- [ ] `go build ./...` sukses tanpa error
- [ ] `GET /api/profile/:id` → detail payout
- [ ] `PUT /api/profile/:id` → 200
- [ ] `GET /api/menu` → semua menu
- [ ] `GET /api/menu/:REGULAR` → semua menu dengan accountType `REGULAR`
- [ ] `GET /api/menu/:PREMIUM` → semua menu
- [ ] `GET /api/profile/tidak-ada` → 404
- [ ] `npm test` → semua unit test tetap lulus (MSW tidak berubah)
