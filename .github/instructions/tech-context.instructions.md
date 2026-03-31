---
applyTo: "**"
---

# Tech Context

## Tech Stack

| Komponen     | identity-service               | user-profile-service          | saving-service              | bff-service                 |
| ------------ | ------------------------------ | ----------------------------- | --------------------------- | --------------------------- |
| Language     | Go 1.24                        | Go 1.24                       | Go 1.24                     | Go 1.24                     |
| Transport    | HTTP (`net/http`) + gRPC handler | REST (`chi` router) + gRPC    | HTTP (`net/http`) + gRPC    | gRPC + grpc-gateway         |
| Database     | PostgreSQL (custom wrapper)    | PostgreSQL (`database/sql`)   | PostgreSQL (`database/sql`) | Tidak ada (stateless)       |
| Auth         | JWT HS256 (`dgrijalva/jwt-go`) | JWT parse di handler          | Tidak ada (public)          | JWT verify lokal            |
| Config       | `godotenv` + `os.LookupEnv`    | `godotenv` + `os.LookupEnv`   | `godotenv` + `os.LookupEnv` | `godotenv` + `os.LookupEnv` |
| Logger       | Zap + FluentBit                | stdlib `log`                  | Zap + FluentBit             | Zap + FluentBit             |
| CLI          | `urfave/cli`                   | Tidak ada                     | `urfave/cli`                | `urfave/cli`                |
| Container    | Docker + Docker Compose        | Docker + Docker Compose       | Docker + Docker Compose     | Docker + Docker Compose     |
| Code Quality | SonarQube                      | —                             | SonarQube                   | SonarQube                   |

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
│   │   ├── identity_auth_api_test.go   # Unit tests: SignUp, SignIn, GetMe (HTTP + gRPC)
│   │   ├── identity_grpc_api.go        # gRPC handler: SignUp, SignIn, GetMe via gRPC
│   │   ├── identity_authInterceptor.go # JWT auth interceptor
│   │   ├── identity_interceptor.go     # Chain: ProcessId → Logging → Errors → Auth
│   │   ├── identity_interceptor_test.go# Unit tests: interceptor chain
│   │   └── error.go                    # Error helpers (badRequest, unauthorized, conflict)
│   ├── db/
│   │   ├── provider.go                 # Provider struct + constructor
│   │   ├── identity_provider.go        # Queries: CreateUser, GetUserByUsername, CheckUsernameExists
│   │   ├── identity_provider_test.go   # Unit tests: DB queries (sqlmock)
│   │   └── error.go                    # NotFoundErr type
│   ├── jwt/
│   │   ├── manager.go                  # JWT Generate + Verify (HS256)
│   │   └── manager_test.go             # Unit tests: JWT generate + verify
│   ├── lib/
│   │   ├── database/                   # DB wrapper (connect, retry, interface, mock)
│   │   │   └── database_test.go        # Unit tests: DB wrapper
│   │   └── logger/                     # Zap structured logger + FluentBit
│   ├── utils/
│   │   ├── utils.go                    # GetProcessIdFromCtx, GetEnv, GenerateProcessId
│   │   └── utils_test.go               # Unit tests: utility functions
│   └── constant/
│       ├── constant.go                 # Response codes, date format, process_id key
│       └── constant_test.go            # Unit tests: constants
├── migrations/
│   ├── 001_init.sql                    # DDL: users, profiles
│   └── 002_rename_email_to_username.sql
├── proto/
│   ├── identity_api.proto              # Service definition (SignUp, SignIn, GetMe)
│   └── identity_payload.proto          # Request/response messages
├── www/swagger.json                    # Swagger API docs
├── Dockerfile                          # Multi-stage (golang:1.24-alpine → alpine:3.19)
├── docker-compose.yml                  # PostgreSQL 15 + identity-service
├── docker-compose.local.yml            # Dev local compose
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
- gRPC: 9301 (handler implemented in `identity_grpc_api.go`, perlu verifikasi listener exposed)

## user-profile-service

### Dependencies

- **`github.com/go-chi/chi/v5`**: Lightweight HTTP router + middleware
- **`database/sql` + `github.com/lib/pq`**: PostgreSQL stdlib
- **`github.com/joho/godotenv`**: .env file loading
- **`github.com/dgrijalva/jwt-go`**: JWT parsing (untuk GetMyProfile)
- **`github.com/swaggo/http-swagger`**: Swagger UI
- **`google.golang.org/grpc`**: gRPC server (port 9302)

### Struktur Folder

```
user-profile-service/
├── server/
│   ├── main.go                         # Entry point + HTTP chi router + gRPC server + graceful shutdown
│   ├── core_config.go                  # Config struct + initConfig (godotenv + os.LookupEnv)
│   ├── core_db.go                      # DB connection lifecycle + runMigration (embed.FS)
│   ├── api/
│   │   ├── api.go                      # Server struct + constructor + pb.UserProfileServiceServer check
│   │   ├── profile_auth_api.go         # HTTP: GetMyProfile (JWT), GetProfile, UpdateProfile, CreateProfile, GetProfileByUserID
│   │   ├── profile_grpc_api.go         # gRPC: CreateProfile, GetProfileByID, GetProfileByUserID, UpdateProfile
│   │   ├── menu_api.go                 # HTTP: GetAllMenus, GetMenusByAccountType
│   │   ├── menu_grpc_api.go            # gRPC: GetAllMenus, GetMenusByAccountType
│   │   ├── search_api.go               # HTTP: GetExchangeRates, GetInterestRates, GetBranches
│   │   ├── search_grpc_api.go          # gRPC: GetExchangeRates, GetInterestRates, GetBranches
│   │   ├── upload_api.go               # HTTP: UploadImage (Azure Blob)
│   │   ├── converter.go                # Model ↔ proto conversion helpers
│   │   └── error.go                    # writeJSON, writeError, StandardResponse, UploadResponse
│   ├── db/
│   │   ├── provider.go                 # Provider struct + constructor
│   │   ├── profile_provider.go         # Queries: GetByID, GetByUserID, Create, Update + domain types
│   │   ├── menu_provider.go            # Queries: GetAll, GetByAccountType + Menu/MenuResponse types
│   │   └── search_provider.go          # Queries: ExchangeRates, InterestRates, Branches + domain types
│   ├── constant/
│   │   └── constant.go                 # Response codes, date format
│   └── utils/
│       └── utils.go                    # GetEnv helper
├── migrations/
│   ├── embed.go                        # embed.FS for SQL migration files
│   ├── 001_init.sql                    # DDL: profile, menu
│   ├── 002_add_image_to_profile.sql
│   ├── 003_add_user_id_to_profile.sql
│   ├── 004_add_exchange_rates.sql
│   ├── 005_add_interest_rates.sql
│   └── 006_add_branches.sql
├── proto/
│   ├── user_profile_api.proto          # Service definition (9 RPC methods)
│   └── user_profile_payload.proto      # Request/response messages
├── protogen/user-profile-service/
│   ├── codec.go                        # JSON codec for gRPC
│   ├── user_profile_api_grpc.pb.go     # Hand-written gRPC service interface
│   └── user_profile_payload.pb.go      # Hand-written message structs
├── docs/                               # Swagger generated docs
├── Dockerfile                          # Multi-stage (golang:1.24-alpine → alpine:3.20)
├── docker-compose.yml                  # PostgreSQL 17 + app
├── docker-compose.local.yml            # Dev local compose
├── seed.sql                            # 1 profile + 9 menu items
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
- gRPC: 9302

## saving-service

### Dependencies

- **Go stdlib `net/http`**: HTTP server + router
- **`database/sql` + `github.com/lib/pq`**: PostgreSQL stdlib
- **`github.com/joho/godotenv`**: .env file loading
- **`google.golang.org/grpc`**: gRPC server (port 9303)
- **`urfave/cli`**: CLI command framework
- **Zap logger + FluentBit**: `server/lib/logger/`
- **`swaggo/http-swagger`**: Swagger UI
- **`google/uuid`**: UUID generation

### Struktur Folder

```
saving-service/
├── server/
│   ├── main.go                         # Entry point + CLI + HTTP/gRPC servers
│   ├── core_config.go                  # Config loader (env vars + .env)
│   ├── core_db.go                      # DB connection + migrations (embed.FS)
│   ├── api/
│   │   ├── api.go                      # Server struct + constructor
│   │   ├── saving_api.go               # HTTP: GetExchangeRates, GetInterestRates, GetBranches
│   │   ├── saving_grpc_api.go          # gRPC: GetExchangeRates, GetInterestRates, GetBranches
│   │   └── error.go                    # writeJSON, writeError helpers
│   ├── db/
│   │   ├── provider.go                 # Provider struct + constructor
│   │   ├── exchange_rate_provider.go    # GetAllExchangeRates
│   │   ├── interest_rate_provider.go    # GetAllInterestRates
│   │   └── branch_provider.go          # GetAllBranches, SearchBranches (ILIKE)
│   ├── constant/
│   │   └── constant.go
│   ├── utils/
│   │   └── utils.go
│   └── lib/
│       └── logger/                     # Zap structured logger + FluentBit
├── migrations/
│   ├── embed.go                        # embed.FS for SQL migration files
│   ├── 001_add_exchange_rates.sql       # DDL: exchange_rate table
│   ├── 002_add_interest_rates.sql       # DDL: interest_rate table
│   └── 003_add_branches.sql             # DDL: branch table
├── proto/
│   ├── saving_api.proto                 # Service definition (3 RPC methods)
│   └── saving_payload.proto             # Request/response messages
├── protogen/saving-service/
│   ├── codec.go                        # JSON codec for gRPC
│   ├── saving_api_grpc.pb.go            # Hand-written gRPC service interface
│   └── saving_payload.pb.go             # Hand-written message structs
├── docs/                               # Swagger generated docs
├── Dockerfile                          # Multi-stage (golang:1.24-alpine → alpine:3.20)
├── docker-compose.yml                  # PostgreSQL 17 + saving-service
├── docker-compose.local.yml            # Dev local compose
├── seed.sql                            # 4 exchange rates + 4 interest rates + 5 branches
├── Makefile
└── sonar-project.properties
```

### Database Schema (saving)

```sql
-- exchange_rate
CREATE TABLE IF NOT EXISTS exchange_rate (
    id           UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    country      VARCHAR(100) NOT NULL,
    currency     VARCHAR(10)  NOT NULL,
    country_code VARCHAR(10)  NOT NULL,
    buy          NUMERIC(10,3) NOT NULL,
    sell         NUMERIC(10,3) NOT NULL
);

-- interest_rate
CREATE TABLE IF NOT EXISTS interest_rate (
    id      UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    kind    VARCHAR(50)  NOT NULL,
    deposit VARCHAR(10)  NOT NULL,
    rate    NUMERIC(5,2) NOT NULL
);

-- branch
CREATE TABLE IF NOT EXISTS branch (
    id        UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name      VARCHAR(200) NOT NULL,
    distance  VARCHAR(50)  NOT NULL,
    latitude  NUMERIC(10,6) NOT NULL,
    longitude NUMERIC(10,6) NOT NULL
);
```

### Port

- HTTP: 8081
- gRPC: 9303

## bff-service

### Dependencies

- **`google.golang.org/grpc`**: gRPC server
- **`grpc-gateway/v2`**: REST → gRPC gateway
- **`dgrijalva/jwt-go`**: JWT verify lokal
- **`urfave/cli`**: CLI commands
- **`swaggo/swag@v1.16.6`**: Swagger annotation parser + doc generator
- **`swaggo/http-swagger@v1.3.4`**: Swagger UI HTTP handler

### Struktur Folder

```
bff-service/
├── proto/                          # BFF proto definitions
├── protogen/                       # Generated Go code (BFF + downstream clients)
├── docs/                           # Swagger generated docs (docs.go, swagger.json, swagger.yaml)
├── server/
│   ├── main.go                     # Entry + CLI (grpc-server, gw-server, grpc-gw-server) + Swagger UI route
│   ├── core_config.go              # Config loader
│   ├── gateway_http_handler.go     # HTTP gateway + custom upload handler
│   ├── http_routes.go              # REST→gRPC bridge handlers + swagger annotations
│   ├── swagger_docs.go             # Swagger doc stubs for multi-method endpoints
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
# Per service (local dev)
cd saving-service && docker compose -f docker-compose.local.yml up --build -d
cd identity-service && docker compose -f docker-compose.local.yml up --build -d
cd user-profile-service && docker compose -f docker-compose.local.yml up --build -d

# BFF full stack
cd bff-service && docker compose up --build
```

| Service              | HTTP | gRPC | Database      |
| -------------------- | ---- | ---- | ------------- |
| bff-service          | 3000 | 9090 | — (stateless) |
| identity-service     | 3031 | 9301 | identity_db   |
| user-profile-service | 8080 | 9302 | bankease_db   |
| saving-service       | 8081 | 9303 | saving        |

## Perbedaan Kunci dengan Service Referensi

| Fitur             | addons-issuance-lc-service | BankEase Services                     |
| ----------------- | -------------------------- | ------------------------------------- |
| Protocol          | gRPC + HTTP Gateway        | Mixed (REST + gRPC gateway planned)   |
| ORM               | GORM + protoc-gen-gorm     | database/sql (stdlib)                 |
| Config            | Viper + godotenv           | godotenv + os.LookupEnv/os.Getenv     |
| Deployment        | Kubernetes/Docker          | Docker Compose                        |
| External services | 15+ gRPC clients           | 3 gRPC clients (BFF)                  |
| Observability     | Elastic APM + Logrus       | Zap (identity, saving, BFF) / stdlib log (profile) |
| Code Quality      | SonarQube                  | SonarQube (identity, saving, BFF)     |
| CLI framework     | urfave/cli                 | urfave/cli (identity, saving, BFF)    |
| Proto codegen     | protoc-gen-go/gorm/gw      | Hand-written protogen                 |
