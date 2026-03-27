---
applyTo: "**"
---

# Tech Context

## Tech Stack

| Komponen     | identity-service               | user-profile-service        | bff-service (planned)       |
| ------------ | ------------------------------ | --------------------------- | --------------------------- |
| Language     | Go 1.24                        | Go 1.24                     | Go 1.24                     |
| Transport    | HTTP (`net/http`) + gRPC ready | REST (`chi` router)         | gRPC + grpc-gateway         |
| Database     | PostgreSQL (custom wrapper)    | PostgreSQL (`database/sql`) | Tidak ada (stateless)       |
| Auth         | JWT HS256 (`dgrijalva/jwt-go`) | JWT parse di handler        | JWT verify lokal            |
| Config       | `godotenv` + `os.LookupEnv`    | `godotenv` + `os.Getenv`    | `godotenv` + `os.LookupEnv` |
| Logger       | Zap + FluentBit                | stdlib `log`                | Zap + FluentBit             |
| CLI          | `urfave/cli`                   | Tidak ada                   | `urfave/cli`                |
| Container    | Docker + Docker Compose        | Docker + Docker Compose     | Docker + Docker Compose     |
| Code Quality | SonarQube                      | —                           | SonarQube                   |

## identity-service

### Dependencies

- **Go stdlib `net/http`** + CORS middleware manual
- **`dgrijalva/jwt-go`**: JWT token parsing & validation (HS256)
- **`golang.org/x/crypto/bcrypt`**: Password hashing
- **Custom database wrapper**: `server/lib/database/` (connect, retry, transaction)
- **Zap logger + FluentBit**: `server/lib/logger/`
- **`urfave/cli`**: CLI command framework (grpc-server, gw-server, grpc-gw-server)
- **`google/uuid`**: UUID generation
- Module path: `bitbucket.bri.co.id/scm/addons/addons-identity-service`
- Asal: dikembangkan di project terpisah (`mini-project-copilot-identity/`), kemudian disatukan ke monorepo ini

### Struktur Folder

```
identity-service/
├── server/
│   ├── main.go                         # Entry point + CLI + HTTP server + auto-migration
│   ├── core_config.go                  # Config loader (env vars + .env)
│   ├── core_db.go                      # DB connection lifecycle (retry logic)
│   ├── api/
│   │   ├── api.go                      # Server struct + constructor
│   │   ├── identity_auth_api.go        # Handler: SignUp, SignIn, GetMe + HTTP handlers
│   │   ├── identity_authInterceptor.go # JWT auth interceptor
│   │   ├── identity_interceptor.go     # Chain: ProcessId → Logging → Errors → Auth
│   │   └── error.go                    # Error helpers (badRequest, unauthorized, conflict)
│   ├── db/
│   │   ├── provider.go                 # Provider struct + constructor
│   │   ├── identity_provider.go        # Queries: CreateUser, GetUserByUsername, CheckUsernameExists
│   │   └── error.go                    # NotFoundErr type
│   ├── jwt/manager.go                  # JWT Generate + Verify (HS256)
│   ├── lib/
│   │   ├── database/                   # DB wrapper (connect, retry, interface, mock)
│   │   └── logger/                     # Zap structured logger + FluentBit
│   ├── utils/utils.go                  # GetProcessIdFromCtx, GetEnv, GenerateProcessId
│   └── constant/                       # Response codes, date format, process_id key
├── migrations/
│   ├── 001_init.sql                    # DDL: users, profiles
│   └── 002_rename_email_to_username.sql
├── proto/
│   ├── identity_api.proto              # Service definition (SignUp, SignIn, GetMe)
│   └── identity_payload.proto          # Request/response messages
├── www/swagger.json                    # Swagger API docs
├── Dockerfile                          # Multi-stage (golang:1.24-alpine → alpine:3.19)
├── docker-compose.yml                  # PostgreSQL 15 + identity-service
├── Makefile                            # build, run, unit-test, docker-build
├── sonar-project.properties            # SonarQube: bricams-addons-identity-service:project
└── .env.example
```

### Database Schema (identity_db)

```sql
-- users
CREATE TABLE users (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    username VARCHAR(255) UNIQUE NOT NULL,
    password_hash TEXT NOT NULL,
    phone VARCHAR(50),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
```

### Port

- HTTP: 3031
- gRPC: 9301 (CLI ready, HTTP handler aktif)

## user-profile-service

### Dependencies

- **`github.com/go-chi/chi/v5`**: Lightweight HTTP router + middleware
- **`database/sql` + `github.com/lib/pq`**: PostgreSQL stdlib
- **`github.com/joho/godotenv`**: .env file loading
- **`github.com/dgrijalva/jwt-go`**: JWT parsing (untuk GetMyProfile)
- **`github.com/swaggo/http-swagger`**: Swagger UI

### Struktur Folder

```
user-profile-service/
├── cmd/server/main.go              # Entrypoint: load env, DB connection, start server
├── internal/
│   ├── db/
│   │   ├── db.go                   # Setup *sql.DB dari DATABASE_URL
│   │   ├── migrate.go              # Auto-run migration (embed.FS)
│   │   └── migrations/
│   │       ├── 001_init.sql        # DDL: profile, menu
│   │       ├── 002_add_image_to_profile.sql
│   │       └── 003_add_user_id_to_profile.sql
│   ├── handlers/
│   │   ├── profile.go              # CRUD /api/profile + GetMyProfile (JWT)
│   │   ├── menu.go                 # GET /api/menu, /api/menu/{accountType}
│   │   └── upload.go               # POST /api/upload/image (Azure Blob)
│   ├── models/
│   │   ├── profile.go              # Profile, EditProfileRequest, CreateProfileRequest, StandardResponse
│   │   └── menu.go                 # Menu, MenuResponse
│   ├── repository/
│   │   ├── profile.go              # DB queries: GetByID, GetByUserID, Create, Update
│   │   └── menu.go                 # DB queries: GetAll, GetByAccountType
│   └── server/
│       ├── router.go               # Routes + middleware (CORS, logging)
│       └── server.go               # Server struct, dependency injection
├── docs/                           # Swagger generated docs
├── Dockerfile                      # Multi-stage (golang:1.24 → alpine:3.20)
├── docker-compose.yml              # PostgreSQL 17 + app
├── seed.sql                        # 1 profile + 9 menu items
└── .env.example
```

### Database Schema (bankease_db)

```sql
-- profile
CREATE TABLE IF NOT EXISTS profile (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id         UUID UNIQUE,
    bank            VARCHAR(50)  NOT NULL,
    branch          VARCHAR(50)  NOT NULL,
    name            VARCHAR(100) NOT NULL,
    card_number     VARCHAR(20)  NOT NULL,
    card_provider   VARCHAR(50)  NOT NULL,
    balance         BIGINT       NOT NULL DEFAULT 0,
    currency        VARCHAR(3)   NOT NULL DEFAULT 'IDR',
    account_type    VARCHAR(20)  NOT NULL DEFAULT 'REGULAR',
    image           TEXT         DEFAULT ''
);

-- menu
CREATE TABLE IF NOT EXISTS menu (
    id        VARCHAR(50)  PRIMARY KEY,
    "index"   INTEGER      UNIQUE NOT NULL,
    type      VARCHAR(20)  NOT NULL,
    title     VARCHAR(50)  NOT NULL,
    icon_url  TEXT         NOT NULL,
    is_active BOOLEAN      NOT NULL DEFAULT TRUE
);
```

### Port

- HTTP: 8080

## bff-service (Planned)

### Dependencies (planned)

- **`google.golang.org/grpc`**: gRPC server
- **`grpc-gateway/v2`**: REST → gRPC gateway
- **`dgrijalva/jwt-go`**: JWT verify lokal
- **`urfave/cli`**: CLI commands

### Struktur Folder

```
bff-service/
├── proto/                          # BFF proto definitions
├── protogen/                       # Generated Go code (BFF + downstream clients)
├── server/
│   ├── main.go                     # Entry + CLI (grpc-server, gw-server, grpc-gw-server)
│   ├── core_config.go              # Config loader
│   ├── gateway_http_handler.go     # HTTP gateway + custom upload handler
│   ├── api/                        # gRPC handlers (auth, profile, menu orchestration)
│   ├── services/service.go         # ServiceConnection (identity + profile gRPC clients)
│   ├── jwt/manager.go              # JWT Verify only (lokal)
│   ├── lib/logger/                 # Zap logger
│   └── utils/, constant/
├── Dockerfile
└── docker-compose.yml              # Full stack (5 containers)
```

### Port

- HTTP: 3000
- gRPC: 9090

## Docker Setup (Full Stack)

```bash
# Dari root project
docker compose up --build     # Start semua (BFF + identity + profile + 2x DB)
docker compose down -v        # Stop + reset
```

| Service              | HTTP | gRPC | Database      |
| -------------------- | ---- | ---- | ------------- |
| bff-service          | 3000 | 9090 | — (stateless) |
| identity-service     | 3031 | 9301 | identity_db   |
| user-profile-service | 8080 | 9302 | bankease_db   |

## Perbedaan Kunci dengan Service Referensi

| Fitur             | addons-issuance-lc-service | BankEase Services                     |
| ----------------- | -------------------------- | ------------------------------------- |
| Protocol          | gRPC + HTTP Gateway        | Mixed (REST + gRPC gateway planned)   |
| ORM               | GORM + protoc-gen-gorm     | database/sql (stdlib)                 |
| Config            | Viper + godotenv           | godotenv + os.LookupEnv/os.Getenv     |
| Deployment        | Kubernetes/Docker          | Docker Compose                        |
| External services | 15+ gRPC clients           | 2 gRPC clients (BFF)                  |
| Observability     | Elastic APM + Logrus       | Zap (identity) / stdlib log (profile) |
| Code Quality      | SonarQube                  | SonarQube (identity, BFF)             |
| CLI framework     | urfave/cli                 | Tidak ada                             |
| Proto codegen     | protoc-gen-go/gorm/gw      | Tidak ada                             |
| CLI framework     | urfave/cli                 | Tidak ada                             |
| Proto codegen     | protoc-gen-go/gorm/gw      | Tidak ada                             |
