---
applyTo: "**"
---

# System Patterns

## Arsitektur Overview

BankEase menggunakan arsitektur microservices dengan BFF pattern. Mobile app berkomunikasi hanya dengan BFF via REST, BFF meneruskan ke downstream services via gRPC.

```
┌──────────────┐         REST (HTTP)          ┌─────────────────────────┐
│  Mobile App  │ ──────────────────────────▶  │   BFF Service           │
│  (React      │  POST /api/auth/signup       │   (grpc-gateway)        │
│   Native)    │  POST /api/auth/signin       │                         │
│              │  GET  /api/profile            │   Port 3000 (HTTP)      │
│              │  GET  /api/menu/...           │   Port 9090 (gRPC)      │
└──────────────┘  POST /api/upload/image      └──────┬───────┬──────────┘
                                                     │       │
                                          gRPC       │       │  gRPC
                                                     ▼       ▼
                                    ┌────────────────┐ ┌──────────────────┐
                                    │ identity-      │ │ user-profile-    │
                                    │ service        │ │ service          │
                                    │ Port 9301 gRPC │ │ Port 9302 gRPC   │
                                    │ Port 3031 HTTP │ │ Port 8080 HTTP   │
                                    └───────┬────────┘ └────────┬─────────┘
                                            ▼                   ▼
                                    ┌────────────────┐ ┌──────────────────┐
                                    │ PostgreSQL     │ │ PostgreSQL       │
                                    │ identity_db    │ │ bankease_db      │
                                    └────────────────┘ └──────────────────┘
```

## Per-Service Architecture

### user-profile-service (BRICaMS Pattern — same as identity-service)

```
┌─────────────────────────────────────────┐
│  server/main.go (entry + HTTP + gRPC)  │
│    graceful shutdown, signal handling  │
├─────────────────────────────────────────┤
│  server/core_config.go → Config struct │
│  server/core_db.go → DB + migrations   │
├─────────────────────────────────────────┤
│  server/api/ (HTTP + gRPC handlers)    │
│    - api.go           → Server struct  │
│    - profile_auth_api.go → HTTP CRUD   │
│    - profile_grpc_api.go → gRPC CRUD   │
│    - menu_api.go      → HTTP menus     │
│    - menu_grpc_api.go → gRPC menus     │
│    - search_api.go    → HTTP search    │
│    - search_grpc_api.go → gRPC search  │
│    - upload_api.go    → HTTP upload    │
│    - converter.go     → model↔proto    │
│    - error.go         → error helpers  │
├─────────────────────────────────────────┤
│  server/db/ (Provider pattern)         │
│    - provider.go      → Provider struct│
│    - profile_provider.go → CRUD queries│
│    - menu_provider.go → menu queries   │
│    - search_provider.go → rates/branch │
├─────────────────────────────────────────┤
│  server/constant/, server/utils/       │
├─────────────────────────────────────────┤
│  migrations/ (SQL DDL + embed.go)      │
│  proto/ (gRPC definitions)             │
│  protogen/ (hand-written gRPC code)    │
└─────────────────────────────────────────┘
```

### identity-service (BRICaMS Pattern)

```
┌─────────────────────────────────────────┐
│  server/main.go (entry + CLI + HTTP)    │
├─────────────────────────────────────────┤
│  server/api/ (handlers + interceptors)  │
│    - api.go           → Server struct   │
│    - identity_auth_api.go → SignUp/In/Me│
│    - identity_interceptor.go → chain    │
│    - identity_authInterceptor.go → JWT  │
│    - error.go         → error helpers   │
├─────────────────────────────────────────┤
│  server/db/ (Provider pattern)          │
│    - provider.go      → Provider struct │
│    - identity_provider.go → DB queries  │
├─────────────────────────────────────────┤
│  server/jwt/ → JWTManager (HS256)       │
├─────────────────────────────────────────┤
│  server/lib/ (database wrapper, logger) │
├─────────────────────────────────────────┤
│  migrations/ (SQL DDL)                  │
│  proto/ (gRPC definitions)              │
└─────────────────────────────────────────┘
```

### bff-service (Gateway Pattern — addons-issuance-lc-service style)

```
┌─────────────────────────────────────────┐
│  server/main.go (entry + CLI)           │
│    grpc-server / gw-server / grpc-gw    │
├─────────────────────────────────────────┤
│  server/gateway_http_handler.go         │
│    grpc-gateway mux + custom handlers   │
│    (upload, CORS, error mapping)        │
├─────────────────────────────────────────┤
│  server/api/ (gRPC handlers)            │
│    - api.go → Server struct + DI        │
│    - bff_auth_api.go → orchestration    │
│    - bff_profile_api.go → proxy         │
│    - bff_menu_api.go → proxy            │
│    - bff_interceptor.go → chain         │
├─────────────────────────────────────────┤
│  server/services/ (ServiceConnection)   │
│    - gRPC clients ke identity + profile │
├─────────────────────────────────────────┤
│  server/jwt/ → JWT Verify (lokal)       │
├─────────────────────────────────────────┤
│  proto/ → BFF proto definitions         │
│  protogen/ → generated Go code          │
└─────────────────────────────────────────┘
```

## Referensi Pattern dari addons-issuance-lc-service

| Aspek             | addons-issuance-lc-service | identity-service             | user-profile-service         | bff-service                   |
| ----------------- | -------------------------- | ---------------------------- | ---------------------------- | ----------------------------- |
| Transport         | gRPC + HTTP Gateway        | HTTP (net/http) + gRPC ready | REST (chi router) + gRPC     | gRPC + HTTP Gateway           |
| Config management | Viper + godotenv           | godotenv + os.LookupEnv      | godotenv + os.LookupEnv      | godotenv + os.LookupEnv       |
| Database          | GORM + protoc-gen-gorm     | database/sql (wrapper)       | database/sql (stdlib)        | Tidak ada (stateless)         |
| Auth              | JWT interceptor            | JWT interceptor              | JWT parse di handler         | JWT verify lokal              |
| Logging           | Logrus + Fluent            | Zap + FluentBit              | stdlib log                   | Zap + FluentBit               |
| DI Pattern        | Manual (Server struct)     | Manual (Server struct)       | Manual (Server struct)       | Manual (Server struct)        |
| Folder structure  | server/ pattern            | server/ pattern              | server/ pattern (refactored) | server/ pattern               |
| External services | 15+ gRPC clients           | 1 HTTP call (profile)        | Tidak ada                    | 2 gRPC clients                |

## Key Design Decisions

### 1. BFF Pattern (Backend for Frontend)

- Mobile app hanya berkomunikasi dengan BFF (single entry point)
- BFF mengorkestrasi calls ke downstream services
- Contoh: SignUp = identity.SignUp → profile.CreateProfile

### 2. gRPC untuk Inter-Service Communication

- BFF → identity-service: gRPC
- BFF → user-profile-service: gRPC
- Mobile app → BFF: REST (via grpc-gateway)

### 3. JWT Local Verification di BFF

- BFF memverifikasi JWT secara lokal (secret key sama dengan identity-service)
- Tidak perlu call ke identity-service untuk setiap verifikasi token
- Protected endpoints: GET /api/auth/me, GET /api/profile

### 4. Dual Database

- identity-service: PostgreSQL `identity_db` (tabel `users`)
- user-profile-service: PostgreSQL `bankease_db` (tabel `profile`, `menu`)
- BFF: stateless, tidak punya database sendiri

### 5. Layered Architecture (per service)

- **Handlers/API**: Menerima request, validasi, format response
- **Repository/Provider**: Interaksi database
- **Models**: Domain objects dan request/response payloads

### 6. Interceptor Chain (identity + BFF)

```
ProcessIdInterceptor → LoggingInterceptor → ErrorsInterceptor → AuthInterceptor
```

### 7. Response Format Konsisten

**user-profile-service:**

```json
{ "code": 200, "description": "Success" }
```

**identity-service (error):**

```json
{ "error": true, "code": 401, "message": "Unauthorized" }
```

## Routing Pattern

### user-profile-service

```
/api/profile           → GET (my profile, JWT), POST (create)
/api/profile/{id}      → GET (by ID), PUT (update)
/api/profile/user/{uid}→ GET (by user_id)
/api/menu              → GET (all menus)
/api/menu/{accountType}→ GET (filtered)
/api/upload/image      → POST (Azure Blob)
```

### identity-service

```
/api/auth/signup       → POST (register)
/api/auth/signin       → POST (login)
/api/identity/me       → GET (current user, JWT)
/health                → GET (health check)
```

### bff-service (planned — all of the above via single entry point)

```
/api/auth/signup       → POST (orchestrated: identity + profile)
/api/auth/signin       → POST (proxy to identity)
/api/auth/me           → GET (proxy to identity, JWT)
/api/profile           → GET (JWT → profile), POST
/api/profile/{id}      → GET, PUT
/api/profile/user/{uid}→ GET
/api/menu              → GET
/api/menu/{accountType}→ GET
/api/upload/image      → POST (BFF direct to Azure Blob)
```

## Error Handling Pattern

- Validasi di handler/API layer sebelum panggil repository/provider
- gRPC status codes di-map ke HTTP status codes (via grpc-gateway atau manual)
- Response body selalu berformat sesuai standard per service

## Docker Deployment Pattern

```
docker-compose.yml (full stack)
├── identity-db (postgres:17-alpine, port 5432)
├── profile-db (postgres:17-alpine, port 5433)
├── identity-service (port 3031 HTTP, 9301 gRPC)
├── user-profile-service (port 8080 HTTP, 9302 gRPC)
└── bff-service (port 3000 HTTP, 9090 gRPC)
```

- Multi-stage Dockerfiles: build di `golang:1.24-alpine`, run di `alpine`
- Health checks memastikan DB ready sebelum services start
- DDL + seed data via PostgreSQL init directory
