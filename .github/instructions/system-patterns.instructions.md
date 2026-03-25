---
applyTo: '**'
---

# System Patterns

## Arsitektur Overview
BankEase menggunakan arsitektur layered (Clean Architecture simplified) yang terinspirasi dari `addons-issuance-lc-service`, tetapi disederhanakan untuk REST API murni (tanpa gRPC).

```
┌─────────────────────────────────────────┐
│  cmd/server/main.go (entrypoint)        │
├─────────────────────────────────────────┤
│  internal/server/ (router + server)     │
│    - router.go  → route + middleware    │
│    - server.go  → dependency injection  │
├─────────────────────────────────────────┤
│  internal/handlers/ (HTTP handlers)     │
│    - profile.go → GET/PUT /api/profile  │
│    - menu.go    → GET /api/menu         │
├─────────────────────────────────────────┤
│  internal/repository/ (data access)     │
│    - profile.go → query profile         │
│    - menu.go    → query menu            │
├─────────────────────────────────────────┤
│  internal/models/ (domain structs)      │
│    - profile.go → Profile, Request/Resp │
│    - menu.go    → Menu, MenuResponse    │
├─────────────────────────────────────────┤
│  internal/db/ (database layer)          │
│    - db.go       → *sql.DB setup        │
│    - migrate.go  → auto-migration       │
│    - migrations/ → SQL DDL files        │
└─────────────────────────────────────────┘
```

## Referensi Pattern dari addons-issuance-lc-service

| Aspek                | addons-issuance-lc-service    | BankEase (simplified)        |
|----------------------|-------------------------------|------------------------------|
| Transport            | gRPC + HTTP Gateway           | REST (chi router) saja       |
| Config management    | Viper + godotenv              | godotenv / os.Getenv         |
| Database             | GORM + protoc-gen-gorm        | database/sql (stdlib)        |
| Routing              | gRPC-Gateway + custom mux     | chi router                   |
| Auth                 | JWT interceptor               | Tidak ada (dev only)         |
| File storage         | MinIO integration             | Tidak ada                    |
| Logging              | Logrus + Fluent               | stdlib log / slog            |
| Dependency injection | Manual (via struct fields)    | Manual (via Server struct)   |

## Key Design Decisions

### 1. Layered Architecture
- **Handlers**: Menerima HTTP request, validasi input, memanggil repository, format response
- **Repository**: Interaksi langsung dengan database, return domain model
- **Models**: Struct Go untuk domain objects dan request/response payloads

### 2. Dependency Injection via Server Struct
Mirip pattern di `addons-issuance-lc-service` dimana semua service dependencies di-inject melalui struct:
```go
type Server struct {
    DB     *sql.DB
    Router chi.Router
}
```

### 3. Database Migration on Startup
SQL migration dijalankan otomatis saat server start (mirip pattern core_db.go di referensi).

### 4. Idempotency Pattern
- Client generate UUID v4 → kirim via `Idempotency-Key` header
- Backend simpan di kolom `idempotency_key UNIQUE`
- Jika duplicate → SELECT data lama, return response yang sama

### 5. Response Format Konsisten
Semua API response (sukses maupun error) menggunakan format:
```json
{
  "code": "<status code>",
  "description": "<deskripsi>"
}
```

## Routing Pattern
```
/api/profile/:id     → GET (read profile), PUT (update profile)
/api/menu            → GET (all menus)
/api/menu/:accountType → GET (filtered menus)
```

## Error Handling Pattern
- Validasi di handler layer sebelum panggil repository
- HTTP status code standard (200, 400, 404, 500, 503)
- Response body selalu berformat `{code, description}`

## Docker Deployment Pattern
```
docker-compose.yml
├── db (postgres:17-alpine)
│   ├── healthcheck: pg_isready
│   ├── volume: pgdata (persistent)
│   └── init: 01-schema.sql + 02-seed.sql via /docker-entrypoint-initdb.d/
└── app (multi-stage build)
    ├── depends_on: db (service_healthy)
    └── env: DATABASE_URL → db:5432
```
- Multi-stage Dockerfile: build di `golang:1.24-alpine`, run di `alpine:3.20`
- App waits for DB healthy sebelum start
- DDL (`001_init.sql`) dimount sebagai `01-schema.sql` → tables dibuat dulu
- Seed data (`seed.sql`) dimount sebagai `02-seed.sql` → insert setelah tables ada
- PostgreSQL init scripts hanya dijalankan saat DB volume pertama kali dibuat