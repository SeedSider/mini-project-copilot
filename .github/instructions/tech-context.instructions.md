---
applyTo: '**'
---

# Tech Context

## Tech Stack

| Komponen      | Pilihan                                |
|---------------|----------------------------------------|
| Language      | Go (stdlib `net/http` + `chi` router)  |
| Database      | PostgreSQL                             |
| Autentikasi   | Tidak ada (internal/dev only)          |
| Deployment    | Docker Compose (`localhost:8080`) |
| Tenancy       | Single merchant                        |

## Dependencies Utama
- **Go stdlib `net/http`**: HTTP server foundation
- **`github.com/go-chi/chi/v5`**: Lightweight HTTP router dengan middleware support
- **`database/sql`**: Stdlib database interface
- **`github.com/lib/pq`**: PostgreSQL driver untuk Go
- **`github.com/joho/godotenv`**: Loading .env file

## Referensi dari addons-issuance-lc-service
Service referensi menggunakan tech stack yang lebih complex:
- Go 1.24.0, gRPC + gRPC-Gateway v2, GORM, Viper, Logrus, MinIO, Elastic APM, JWT
- BankEase menyederhanakan ini menjadi REST-only dengan stdlib + chi

## Struktur Folder
```
user-profile-service/
├── cmd/server/main.go              # Entrypoint: load env, DB connection, start server
├── internal/
│   ├── db/
│   │   ├── db.go                   # Setup *sql.DB dari DATABASE_URL
│   │   ├── migrate.go              # Auto-run migration saat startup (embed.FS)
│   │   └── migrations/001_init.sql # DDL create tables
│   ├── handlers/
│   │   ├── profile.go              # GET/PUT /api/profile/{id}
│   │   └── menu.go                 # GET /api/menu, GET /api/menu/{accountType}
│   ├── models/
│   │   ├── profile.go              # Profile, EditProfileRequest, StandardResponse
│   │   └── menu.go                 # Menu, MenuResponse
│   ├── repository/
│   │   ├── profile.go              # DB queries untuk profile
│   │   └── menu.go                 # DB queries untuk menu
│   └── server/
│       ├── router.go               # Route definitions + middleware (CORS, logging)
│       └── server.go               # Server struct, dependency injection
├── Dockerfile                      # Multi-stage build (golang:1.24 → alpine:3.20)
├── docker-compose.yml              # PostgreSQL 17 + app service
├── .dockerignore                   # Exclude non-build files
├── seed.sql                        # Data awal: 1 profile + 9 menu items
├── .env.example                    # Template environment variables
├── .gitignore
├── go.mod                          # Go module definition
└── go.sum                          # Dependency checksums
```

## Database

### Tabel `profile`
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
```

### Tabel `menu`
```sql
CREATE TABLE IF NOT EXISTS menu (
    id        VARCHAR(50)  PRIMARY KEY,
    "index"   INTEGER      UNIQUE NOT NULL,
    type      VARCHAR(20)  NOT NULL,
    title     VARCHAR(50)  NOT NULL,
    icon_url  TEXT         NOT NULL,
    is_active BOOLEAN      NOT NULL DEFAULT TRUE
);
```

## Environment Variables
- `DATABASE_URL`: PostgreSQL connection string
- `PORT`: Server port (default: 8080)

## Docker Setup (Recommended)
```bash
cd user-profile-service
docker compose up --build     # Start PostgreSQL + app
docker compose down           # Stop
docker compose down -v        # Stop + reset DB
```
- PostgreSQL 17 Alpine, user `bankease`, password `bankease123`, database `bankease_db`
- App builds via multi-stage Dockerfile (golang:1.24 → alpine:3.20)
- `seed.sql` otomatis dijalankan saat DB pertama kali start
- Health check memastikan app baru start setelah DB ready

## Local Development (tanpa Docker)
1. PostgreSQL harus running di local
2. Copy `.env.example` → `.env`, ubah `DATABASE_URL` ke localhost
3. `go run cmd/server/main.go`
4. Migration otomatis saat startup
5. Jalankan `psql -d bankease_db -f seed.sql` untuk data awal

## Perbedaan Kunci dengan Service Referensi
| Fitur                  | addons-issuance-lc-service | BankEase         |
|------------------------|---------------------------|------------------|
| Protocol               | gRPC + HTTP Gateway       | REST only        |
| ORM                    | GORM + protoc-gen-gorm    | database/sql     |
| Config                 | Viper + godotenv          | godotenv simple  |
| Deployment             | Kubernetes/Docker         | Docker Compose   |
| Auth                   | JWT interceptor           | None             |
| Observability          | Elastic APM + Logrus      | stdlib log       |
| CLI framework          | urfave/cli                | Tidak ada        |
| Proto codegen          | protoc-gen-go/gorm/gw     | Tidak ada        |
| CLI framework          | urfave/cli                | Tidak ada        |
| Proto codegen          | protoc-gen-go/gorm/gw     | Tidak ada        |