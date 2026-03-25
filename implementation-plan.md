# BankEase Backend — Detail Implementasi

> **Tujuan**: Membangun Go REST API backend (`backend/`) sebagai pengganti MSW mock untuk mobile banking app BankEase.  
> **Arsitektur**: Mengikuti pattern `addons-issuance-lc-service` (layered DI, `GetEnv` config, `database/sql` + context) namun disederhanakan — tanpa gRPC, tanpa GORM, tanpa JWT. Hanya `chi` router + stdlib.

---

## Daftar Isi

- [Tech Stack](#tech-stack)
- [Struktur Folder](#struktur-folder)
- [Phase 1: Project Scaffolding](#phase-1-project-scaffolding)
- [Phase 2: Database Layer](#phase-2-database-layer)
- [Phase 3: Models](#phase-3-models)
- [Phase 4: Repository Layer](#phase-4-repository-layer)
- [Phase 5: Handler Layer](#phase-5-handler-layer)
- [Phase 6: Server & Router](#phase-6-server--router)
- [Phase 7: Entrypoint](#phase-7-entrypoint)
- [Phase 8: Seed Data](#phase-8-seed-data)
- [File Mapping: Referensi → BankEase](#file-mapping-referensi--bankease)
- [Verification Checklist](#verification-checklist)
- [Keputusan Arsitektur](#keputusan-arsitektur)
- [Catatan Tambahan](#catatan-tambahan)

---

## Tech Stack

| Komponen       | Pilihan                               | Referensi di `addons-issuance-lc-service`       |
|----------------|---------------------------------------|--------------------------------------------------|
| Language       | Go (stdlib `net/http` + `chi` router) | Go 1.24, gRPC + gRPC-Gateway                    |
| Database       | PostgreSQL via `database/sql`         | PostgreSQL via GORM + `database/sql`             |
| DB Driver      | `github.com/lib/pq`                  | `github.com/lib/pq`                             |
| Router         | `github.com/go-chi/chi/v5`           | `grpc-gateway/v2/runtime` + custom mux          |
| Config         | `github.com/joho/godotenv`           | `godotenv` + `viper`                            |
| Auth           | Tidak ada (dev only)                  | JWT interceptor                                  |
| Logging        | `log` stdlib                          | `logrus` + Elastic APM                           |
| DI Pattern     | Manual via Server struct              | Manual via Server struct                         |

---

## Struktur Folder

```
backend/
├── cmd/
│   └── server/
│       └── main.go                     # Entrypoint: load env, koneksi DB, start server
├── internal/
│   ├── db/
│   │   ├── db.go                       # Setup *sql.DB dari DATABASE_URL
│   │   ├── migrate.go                  # Auto-run migration saat startup (embed.FS)
│   │   └── migrations/
│   │       └── 001_init.sql            # DDL: CREATE TABLE profile, menu
│   ├── handlers/
│   │   ├── profile.go                  # GET/PUT /api/profile/{id}
│   │   └── menu.go                     # GET /api/menu, GET /api/menu/{accountType}
│   ├── models/
│   │   ├── profile.go                  # Profile, EditProfileRequest, StandardResponse
│   │   └── menu.go                     # Menu, MenuResponse
│   ├── repository/
│   │   ├── profile.go                  # GetProfileByID(), UpdateProfile()
│   │   └── menu.go                     # GetAllMenus(), GetMenusByAccountType()
│   └── server/
│       ├── router.go                   # Route definitions + middleware (CORS, logging)
│       └── server.go                   # Server struct, dependency injection
├── seed.sql                            # Data awal: 1 profile + 9 menu items
├── .env.example                        # Template environment variables
└── go.mod                              # Module definition + dependencies
```

---

## Phase 1: Project Scaffolding

> Tidak ada dependency antar langkah — semua bisa dikerjakan paralel.

### 1.1 Initialize Go Module

```bash
mkdir -p backend
cd backend
go mod init github.com/bankease/backend
go get github.com/go-chi/chi/v5
go get github.com/lib/pq
go get github.com/joho/godotenv
```

### 1.2 Buat Folder Structure

Semua folder dan file sesuai [Struktur Folder](#struktur-folder) di atas.

### 1.3 Buat `.env.example`

```env
DATABASE_URL=postgres://USER:PASSWORD@localhost:5432/bankease_db?sslmode=disable
PORT=8080
```

---

## Phase 2: Database Layer

> *Depends on*: Phase 1 selesai.

### 2.1 `internal/db/db.go` — Koneksi Database

**Pattern dari**: `server/core_db.go` + `server/lib/database/database.go`

| Aspek              | Referensi (`addons-issuance-lc-service`)         | BankEase                                   |
|--------------------|---------------------------------------------------|--------------------------------------------|
| Driver             | GORM wrapper + `database/sql`                     | `database/sql` langsung                    |
| Connection         | `database.InitConnectionDB()` + retry logic       | `sql.Open()` + `db.Ping()` (no retry)     |
| Pool               | `MaxIdleConns(0)`, `MaxOpenConns(100)`             | `MaxIdleConns(5)`, `MaxOpenConns(25)`, `ConnMaxLifetime(5min)` |

**Fungsi utama**:

```go
func NewDB(databaseURL string) (*sql.DB, error)
```

- Buka koneksi: `sql.Open("postgres", databaseURL)`
- Verifikasi: `db.Ping()`
- Set pool: `SetMaxOpenConns(25)`, `SetMaxIdleConns(5)`, `SetConnMaxLifetime(5 * time.Minute)`
- Return `*sql.DB` atau error

### 2.2 `internal/db/migrations/001_init.sql` — DDL

```sql
CREATE TABLE IF NOT EXISTS profile (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    bank            VARCHAR(50)  NOT NULL,
    branch          VARCHAR(50)  NOT NULL,
    name            VARCHAR(100) NOT NULL,
    card_number     VARCHAR(20)  NOT NULL,
    card_provider   VARCHAR(50)  NOT NULL,
    balance         BIGINT       NOT NULL DEFAULT 0,
    currency        VARCHAR(3)   NOT NULL DEFAULT 'IDR',
    account_type    VARCHAR(20)  NOT NULL DEFAULT 'REGULAR'
);

CREATE TABLE IF NOT EXISTS menu (
    id        VARCHAR(50)  PRIMARY KEY,
    "index"   INTEGER      UNIQUE NOT NULL,
    type      VARCHAR(20)  NOT NULL,
    title     VARCHAR(50)  NOT NULL,
    icon_url  TEXT         NOT NULL,
    is_active BOOLEAN      NOT NULL DEFAULT TRUE
);
```

> **Catatan koreksi dari spec**:
> - `VARCHAR(3)` untuk name → `VARCHAR(100)` (spec contoh: "Jane Doe" = 8 karakter)
> - `NUMBER` → `INTEGER` (tipe PostgreSQL yang benar)
> - `accountType` → `account_type` (snake_case convention di PostgreSQL)
> - `"index"` di-quote karena merupakan reserved word di SQL

### 2.3 `internal/db/migrate.go` — Auto-Migration

**Mekanisme**: Menggunakan Go `embed.FS` untuk embed file SQL ke dalam binary.

```go
//go:embed migrations/*.sql
var migrationFS embed.FS

func RunMigrations(db *sql.DB) error
```

- Baca `001_init.sql` dari embedded filesystem
- Eksekusi DDL via `db.Exec()`
- Idempotent karena `CREATE TABLE IF NOT EXISTS`

---

## Phase 3: Models

> Bisa dikerjakan **paralel** dengan Phase 2.

### 3.1 `internal/models/profile.go`

```go
// Profile — domain model untuk tabel profile
type Profile struct {
    ID           string `json:"id"`
    Bank         string `json:"bank"`
    Branch       string `json:"branch"`
    Name         string `json:"name"`
    CardNumber   string `json:"card_number"`
    CardProvider string `json:"card_provider"`
    Balance      int64  `json:"balance"`
    Currency     string `json:"currency"`
    AccountType  string `json:"accountType"`       // camelCase di JSON, snake_case di DB
}

// EditProfileRequest — field yang boleh diubah via PUT
type EditProfileRequest struct {
    Bank         string `json:"bank"`
    Branch       string `json:"branch"`
    Name         string `json:"name"`
    CardNumber   string `json:"card_number"`
    CardProvider string `json:"card_provider"`
    Currency     string `json:"currency"`
}

// StandardResponse — format response konsisten {code, description}
type StandardResponse struct {
    Code        int    `json:"code"`
    Description string `json:"description"`
}
```

### 3.2 `internal/models/menu.go`

```go
// Menu — domain model untuk tabel menu
type Menu struct {
    ID       string `json:"id"`
    Index    int    `json:"index"`
    Type     string `json:"type"`
    Title    string `json:"title"`
    IconURL  string `json:"icon_url"`
    IsActive bool   `json:"is_active"`
}

// MenuResponse — wrapper response untuk list menu
type MenuResponse struct {
    Menus []Menu `json:"menus"`
}
```

---

## Phase 4: Repository Layer

> *Depends on*: Phase 2 (DB) + Phase 3 (Models) selesai.

**Pattern dari**: `server/db/issued_lc_provider.go`
- Prepared statements untuk query parameterized
- `context.WithTimeout(ctx, 5*time.Second)` untuk setiap operasi DB
- Manual row scanning ke struct

### 4.1 `internal/repository/profile.go`

```go
type ProfileRepository struct {
    DB *sql.DB
}
```

| Method | Signature | Query | Error Handling |
|--------|-----------|-------|----------------|
| `GetProfileByID` | `(ctx, id string) (*models.Profile, error)` | `SELECT * FROM profile WHERE id = $1` | `sql.ErrNoRows` → not found |
| `UpdateProfile` | `(ctx, id string, req models.EditProfileRequest) error` | `UPDATE profile SET bank=$1, branch=$2, ... WHERE id=$7` | `RowsAffected() == 0` → not found |

### 4.2 `internal/repository/menu.go`

```go
type MenuRepository struct {
    DB *sql.DB
}
```

| Method | Signature | Query | Logic |
|--------|-----------|-------|-------|
| `GetAllMenus` | `(ctx) ([]models.Menu, error)` | `SELECT * FROM menu WHERE is_active = TRUE ORDER BY "index" ASC` | Return semua menu aktif |
| `GetMenusByAccountType` | `(ctx, accountType string) ([]models.Menu, error)` | Conditional | `PREMIUM` → return ALL menus; `REGULAR` → `WHERE type = 'REGULAR'` |

**Business logic menu filtering**:
```
if accountType == "PREMIUM":
    → SELECT * FROM menu WHERE is_active = TRUE ORDER BY "index" ASC  (semua menu)
if accountType == "REGULAR":
    → SELECT * FROM menu WHERE type = 'REGULAR' AND is_active = TRUE ORDER BY "index" ASC
```

---

## Phase 5: Handler Layer

> *Depends on*: Phase 4 (Repository) selesai.

**Pattern dari**: `server/api/issued_lc_data_api.go`
- Handler sebagai method di struct (bukan standalone function)
- Flow: validate → call repo → format response
- Dependency injection via struct field

### 5.1 `internal/handlers/profile.go`

```go
type ProfileHandler struct {
    Repo *repository.ProfileRepository
}
```

| Endpoint | Method | Flow |
|----------|--------|------|
| `GET /api/profile/{id}` | `GetProfile(w, r)` | Extract `id` dari `chi.URLParam` → `Repo.GetProfileByID()` → 200 profile JSON / 404 / 500 |
| `PUT /api/profile/{id}` | `UpdateProfile(w, r)` | Extract `id` + decode body → validate currency (IDR/USD) → `Repo.UpdateProfile()` → 200 `{code, description}` / 400 / 404 |

**Helper functions** (reusable di semua handler):
```go
func writeJSON(w http.ResponseWriter, status int, data interface{})
func writeError(w http.ResponseWriter, status int, description string)
```

**Validasi di `UpdateProfile`**:
- Currency harus `"IDR"` atau `"USD"` → jika tidak, return 400 `"Currency tidak didukung"`

### 5.2 `internal/handlers/menu.go`

```go
type MenuHandler struct {
    Repo *repository.MenuRepository
}
```

| Endpoint | Method | Flow |
|----------|--------|------|
| `GET /api/menu` | `GetAllMenus(w, r)` | `Repo.GetAllMenus()` → 200 `MenuResponse` / 500 |
| `GET /api/menu/{accountType}` | `GetMenusByAccountType(w, r)` | Extract `accountType` → `Repo.GetMenusByAccountType()` → 200 `MenuResponse` / 500 |

---

## Phase 6: Server & Router

> *Depends on*: Phase 5 (Handlers) selesai.

### 6.1 `internal/server/server.go` — DI & Server Struct

**Pattern dari**: `server/main.go` (dependency wiring via struct)

```go
type Server struct {
    DB     *sql.DB
    Router chi.Router
    Port   string
}

func NewServer(db *sql.DB, port string) *Server
func (s *Server) Start() error   // → http.ListenAndServe(":"+s.Port, s.Router)
```

**Dependency wiring di `NewServer()`**:
```
db → ProfileRepository{DB: db}  → ProfileHandler{Repo: profileRepo}
db → MenuRepository{DB: db}     → MenuHandler{Repo: menuRepo}
                                 → setupRoutes(profileHandler, menuHandler)
```

### 6.2 `internal/server/router.go` — Routes & Middleware

**Pattern dari**: `server/gateway_http_handler.go` (CORS middleware, route mounting)

**Middleware stack**:
1. `middleware.Logger` — chi built-in request logger
2. `middleware.Recoverer` — panic recovery
3. Custom CORS middleware

**Route table**:

| Method | Path | Handler |
|--------|------|---------|
| `GET` | `/api/profile/{id}` | `profileHandler.GetProfile` |
| `PUT` | `/api/profile/{id}` | `profileHandler.UpdateProfile` |
| `GET` | `/api/menu` | `menuHandler.GetAllMenus` |
| `GET` | `/api/menu/{accountType}` | `menuHandler.GetMenusByAccountType` |

**CORS middleware** (dari pattern `cors()` di `gateway_http_handler.go`):
```
Access-Control-Allow-Origin: *
Access-Control-Allow-Methods: GET, PUT, OPTIONS
Access-Control-Allow-Headers: Content-Type, Idempotency-Key
```

---

## Phase 7: Entrypoint

> *Depends on*: Phase 6 (Server) selesai.

### 7.1 `cmd/server/main.go`

**Pattern dari**: `server/main.go` (init → config → DB → server) + `server/core_config.go` (`GetEnv` helper)

**Startup flow**:
```
1. godotenv.Load()                          ← load .env file
2. GetEnv("DATABASE_URL", "")               ← baca config (pattern dari core_config.go)
3. GetEnv("PORT", "8080")
4. db.NewDB(databaseURL)                    ← establish DB connection
5. db.RunMigrations(database)               ← auto-migrate tables
6. server.NewServer(database, port)         ← wire dependencies
7. log.Printf("Server started on :%s", port)
8. server.Start()                           ← listen & serve
```

**`GetEnv` helper** (disalin dari `core_config.go`):
```go
func GetEnv(key, fallback string) string {
    if value, ok := os.LookupEnv(key); ok {
        return value
    }
    return fallback
}
```

---

## Phase 8: Seed Data

> Bisa dikerjakan **paralel** dengan Phase 7.

### 8.1 `seed.sql`

**Data**:
- 1 record profile (Jane Doe, Citibank, Tangerang, balance 5.000.000, IDR, REGULAR)
- 9 record menu (campuran REGULAR dan PREMIUM)

**Idempotency**: Menggunakan `INSERT ... ON CONFLICT DO NOTHING`

**Contoh menu items**:

| ID | Index | Type | Title |
|----|-------|------|-------|
| menu_001 | 1 | REGULAR | Account and Card |
| menu_002 | 2 | PREMIUM | Transfer |
| menu_003 | 3 | REGULAR | Payment |
| menu_004 | 4 | REGULAR | Top Up |
| menu_005 | 5 | PREMIUM | Investment |
| menu_006 | 6 | REGULAR | History |
| menu_007 | 7 | PREMIUM | Wealth Management |
| menu_008 | 8 | REGULAR | Settings |
| menu_009 | 9 | PREMIUM | Priority Services |

---

## File Mapping: Referensi → BankEase

Tabel berikut menunjukkan file mana di `addons-issuance-lc-service` yang menjadi referensi untuk setiap file BankEase:

| File BankEase | Referensi | Pattern yang Diambil |
|---------------|-----------|----------------------|
| `cmd/server/main.go` | `server/main.go` | Startup flow: init → config → DB → server |
| `cmd/server/main.go` | `server/core_config.go` | `GetEnv(key, fallback)` helper function |
| `internal/db/db.go` | `server/core_db.go` | DB connection init, pool configuration |
| `internal/db/db.go` | `server/lib/database/database.go` | Connection pooling settings |
| `internal/db/migrate.go` | *(baru — pakai embed.FS)* | Lebih bersih dari pattern referensi |
| `internal/repository/*.go` | `server/db/issued_lc_provider.go` | Prepared statements, `context.WithTimeout`, manual `rows.Scan()` |
| `internal/handlers/*.go` | `server/api/issued_lc_data_api.go` | Handler struct method, validate → call → response flow |
| `internal/server/server.go` | `server/main.go` | DI via struct fields |
| `internal/server/router.go` | `server/gateway_http_handler.go` | CORS middleware, route registration |

---

## Verification Checklist

### Build & Setup
- [ ] `cd backend && go build ./...` — kompilasi tanpa error
- [ ] PostgreSQL running, database `bankease_db` dibuat
- [ ] `go run cmd/server/main.go` — log "Server started on :8080", tabel auto-migrate

### Seed & Test Data
- [ ] `psql -d bankease_db -f seed.sql` — insert data awal

### API Endpoint Testing

| # | Test | Command | Expected |
|---|------|---------|----------|
| 1 | GET profile | `curl localhost:8080/api/profile/<uuid>` | 200 + profile JSON |
| 2 | GET profile not found | `curl localhost:8080/api/profile/00000000-0000-0000-0000-000000000000` | 404 + `{code, description}` |
| 3 | PUT profile | `curl -X PUT -H "Content-Type: application/json" -d '{"bank":"BRI","branch":"Jakarta","name":"John","card_number":"123","card_provider":"Visa","currency":"IDR"}' localhost:8080/api/profile/<uuid>` | 200 + `{code, description}` |
| 4 | PUT invalid currency | Body dengan `"currency":"EUR"` | 400 + `{code: 400, description: "Currency tidak didukung"}` |
| 5 | GET all menus | `curl localhost:8080/api/menu` | 200 + 9 menu items |
| 6 | GET menu REGULAR | `curl localhost:8080/api/menu/REGULAR` | 200 + hanya menu type REGULAR |
| 7 | GET menu PREMIUM | `curl localhost:8080/api/menu/PREMIUM` | 200 + SEMUA menu (REGULAR + PREMIUM) |
| 8 | CORS preflight | `curl -X OPTIONS -H "Origin: http://localhost:3000" localhost:8080/api/menu` | CORS headers present |

---

## Keputusan Arsitektur

| Keputusan | Alasan | Perbedaan dari Referensi |
|-----------|--------|--------------------------|
| Module name `github.com/bankease/backend` | Simplified path | Referensi: `bitbucket.bri.co.id/scm/addons/...` |
| `database/sql` (tanpa GORM) | Query simple, lebih ringan | Referensi: GORM + `protoc-gen-gorm` |
| `chi` router (tanpa gRPC) | REST-only scope, lightweight | Referensi: gRPC + gRPC-Gateway |
| `godotenv` (tanpa Viper) | Config sederhana, cukup `.env` + `os.Getenv` | Referensi: Viper + godotenv |
| `embed.FS` untuk migrations | SQL embedded di binary, no external file dependency | Referensi: GORM auto-migrate dari proto |
| snake_case di DB, camelCase di JSON | PostgreSQL convention + spec response format | — |
| `"index"` di-quote dalam SQL | Reserved word di PostgreSQL | — |
| Tanpa auth layer | Dev/internal service only | Referensi: JWT interceptor |
| Tanpa idempotency di v1 | Belum ada endpoint transaksional | Spec mendokumentasikan pattern untuk future use |
| Chi `middleware.Logger` | Built-in, cukup untuk dev logging | Referensi: Logrus + Elastic APM |

---

## Catatan Tambahan

### 1. Graceful Shutdown (Opsional)
Referensi (`server/main.go`) mengimplementasikan `os.Signal` listening untuk menutup DB connection dengan bersih saat `Ctrl+C`. Direkomendasikan untuk ditambahkan di Phase 7 karena hanya beberapa baris kode dan mencegah DB connection leak.

### 2. Koreksi dari Spec DDL
| Spec Original | Koreksi | Alasan |
|---------------|---------|--------|
| `name VARCHAR(3)` | `name VARCHAR(100)` | Contoh di spec: "Jane Doe" (8 chars) |
| `NUMBER` | `INTEGER` | PostgreSQL type yang benar |
| `accountType` | `account_type` | snake_case convention di PostgreSQL |
| `index` tanpa quote | `"index"` | Reserved word di SQL |

### 3. Dependency Graph

```
Phase 1 (Scaffolding)
    ├── Phase 2 (Database Layer) ─┐
    │                             ├── Phase 4 (Repository) ── Phase 5 (Handler) ── Phase 6 (Server) ── Phase 7 (Entrypoint)
    └── Phase 3 (Models) ────────┘
                                                                                     Phase 8 (Seed) ── (paralel)
```
