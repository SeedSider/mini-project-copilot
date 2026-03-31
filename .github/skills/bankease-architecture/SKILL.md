---
name: bankease-architecture
description: "Cross-service architecture reference for BankEase microservices. Use when: creating new services, doing cross-service work, architectural reviews, understanding inter-service communication, Docker Compose setup, routing patterns, interceptor chains, response formats, or design decisions. Covers all 4 services: identity, user-profile, saving, bff."
argument-hint: "Topic to look up (e.g. 'Docker Compose', 'routing', 'interceptor chain', 'response format')"
---

# BankEase Architecture Reference

## Architecture Overview

BankEase uses microservices with BFF pattern. Mobile app communicates only with BFF via REST, BFF forwards to downstream services via gRPC.

```
┌──────────────┐         REST (HTTP)          ┌─────────────────────────┐
│  Mobile App  │ ──────────────────────────▶  │   BFF Service           │
│  (React      │  POST /api/auth/signup       │   (grpc-gateway)        │
│   Native)    │  POST /api/auth/signin       │                         │
│              │  GET  /api/profile            │   Port 3000 (HTTP)      │
│              │  GET  /api/menu/...           │   Port 9090 (gRPC)      │
│              │  GET  /api/exchange-rates     │                         │
│              │  GET  /api/interest-rates     │                         │
│              │  GET  /api/branches           │                         │
└──────────────┘  POST /api/upload/image      └──────┬───────┬──────────┘
                                                     │       │
                                          gRPC       │       │  gRPC
                                                     ▼       ▼
                                    ┌────────────────┐ ┌──────────────────┐ ┌──────────────────┐
                                    │ identity-      │ │ user-profile-    │ │ saving-          │
                                    │ service        │ │ service          │ │ service          │
                                    │ Port 9301 gRPC │ │ Port 9302 gRPC   │ │ Port 9303 gRPC   │
                                    │ Port 3031 HTTP │ │ Port 8080 HTTP   │ │ Port 8081 HTTP   │
                                    └───────┬────────┘ └────────┬─────────┘ └────────┬─────────┘
                                            ▼                   ▼                    ▼
                                    ┌────────────────┐ ┌──────────────────┐ ┌──────────────────┐
                                    │ PostgreSQL     │ │ PostgreSQL       │ │ PostgreSQL       │
                                    │ identity_db    │ │ bankease_db      │ │ saving           │
                                    └────────────────┘ └──────────────────┘ └──────────────────┘
```

## Service Comparison Matrix

| Komponen     | identity-service               | user-profile-service          | saving-service              | bff-service                 |
| ------------ | ------------------------------ | ----------------------------- | --------------------------- | --------------------------- |
| Language     | Go 1.24                        | Go 1.24                       | Go 1.24                     | Go 1.24                     |
| Transport    | HTTP (`net/http`) + gRPC       | REST (`chi`) + gRPC           | HTTP (`net/http`) + gRPC    | gRPC + HTTP gateway         |
| Database     | PostgreSQL (custom wrapper)    | PostgreSQL (`database/sql`)   | PostgreSQL (`database/sql`) | Stateless                   |
| Auth         | JWT HS256 (`dgrijalva/jwt-go`) | JWT parse di handler          | None (public)               | JWT verify lokal            |
| Config       | `godotenv` + `os.LookupEnv`   | `godotenv` + `os.LookupEnv`  | `godotenv` + `os.LookupEnv` | `godotenv` + `os.LookupEnv` |
| Logger       | Zap + FluentBit                | stdlib `log`                  | Zap + FluentBit             | Zap + FluentBit             |
| CLI          | `urfave/cli`                   | None                          | `urfave/cli`                | `urfave/cli`                |
| Code Quality | SonarQube                      | —                             | SonarQube                   | SonarQube                   |
| HTTP Port    | 3031                           | 8080                          | 8081                        | 3000                        |
| gRPC Port    | 9301                           | 9302                          | 9303                        | 9090                        |
| Database     | identity_db                    | bankease_db                   | saving                      | —                           |

## Reference Pattern (addons-issuance-lc-service)

| Aspek             | addons-issuance-lc-service | BankEase Services                     |
| ----------------- | -------------------------- | ------------------------------------- |
| Protocol          | gRPC + HTTP Gateway        | Mixed (REST + gRPC gateway)           |
| ORM               | GORM + protoc-gen-gorm     | database/sql (stdlib)                 |
| Config            | Viper + godotenv           | godotenv + os.LookupEnv               |
| Deployment        | Kubernetes/Docker          | Docker Compose                        |
| External services | 15+ gRPC clients           | 3 gRPC clients (BFF)                  |
| Observability     | Elastic APM + Logrus       | Zap / stdlib log                      |
| Proto codegen     | protoc-gen-go/gorm/gw      | Hand-written protogen                 |

## Key Design Decisions

### 1. BFF Pattern (Backend for Frontend)
- Mobile app only communicates with BFF (single entry point)
- BFF orchestrates calls to downstream services
- Example: SignUp = identity.SignUp → profile.CreateProfile (best-effort)

### 2. gRPC Inter-Service Communication
- BFF → identity-service: gRPC (port 9301)
- BFF → user-profile-service: gRPC (port 9302)
- BFF → saving-service: gRPC (port 9303)
- Mobile app → BFF: REST (via manual HTTP gateway)

### 3. JWT Local Verification di BFF
- BFF verifies JWT locally (same secret key as identity-service)
- No need to call identity-service for every token verification
- Protected endpoints: GET /api/auth/me, GET /api/profile

### 4. Triple Database
- identity-service: PostgreSQL `identity_db` (table `users`)
- user-profile-service: PostgreSQL `bankease_db` (tables `profile`, `menu`)
- saving-service: PostgreSQL `saving` (tables `exchange_rate`, `interest_rate`, `branch`)
- BFF: stateless, no database

### 5. Layered Architecture (per service)
- **Handlers/API** (`server/api/`): Request handling, validation, response formatting
- **Repository/Provider** (`server/db/`): Database interaction
- **Models**: Domain objects embedded in `server/db/` + `server/api/`

### 6. Hand-written Protogen
- No `protoc` toolchain — all gRPC stubs hand-written in `protogen/` directories
- **WAJIB**: every gRPC service with hand-written protogen MUST have `codec.go` in its protogen package that registers JSONCodec via `encoding.RegisterCodec` in `init()`. Without this, gRPC server falls back to proto codec and fails to unmarshal.

## Interceptor Chain (identity + BFF)

```
ProcessIdInterceptor → LoggingInterceptor → ErrorsInterceptor → AuthInterceptor
```

- ProcessId: assigns unique ID to each request
- Logging: structured logging with Zap
- Errors: panic recovery + error formatting
- Auth: JWT verification for protected endpoints

## Response Formats

**user-profile-service:**
```json
{ "code": 200, "description": "Success" }
```

**identity-service (error):**
```json
{ "error": true, "code": 401, "message": "Unauthorized" }
```

**saving-service:** raw JSON arrays (no wrapper envelope)

## Routing Patterns

### BFF (all endpoints — single entry point for mobile)
```
POST /api/auth/signup       → orchestrated: identity.SignUp + profile.CreateProfile
POST /api/auth/signin       → proxy to identity.SignIn
GET  /api/auth/me           → proxy to identity.GetMe (JWT required)
GET  /api/profile           → JWT → profile.GetMyProfile
POST /api/profile           → profile.CreateProfile
GET  /api/profile/{id}      → profile.GetProfileByID
PUT  /api/profile/{id}      → profile.UpdateProfile
GET  /api/profile/user/{uid}→ profile.GetProfileByUserID
GET  /api/menu              → profile.GetAllMenus
GET  /api/menu/{accountType}→ profile.GetMenusByAccountType
POST /api/upload/image      → BFF direct to Azure Blob
GET  /api/exchange-rates    → saving.GetExchangeRates (public)
GET  /api/interest-rates    → saving.GetInterestRates (public)
GET  /api/branches?q=       → saving.GetBranches (public)
```

### identity-service
```
POST /api/auth/signup       → SignUp (validate, bcrypt, INSERT)
POST /api/auth/signin       → SignIn (bcrypt compare, JWT generate)
GET  /api/identity/me       → GetMe (JWT verify, get user)
GET  /health                → Health check
```

### user-profile-service
```
GET  /api/profile           → GetMyProfile (JWT)
POST /api/profile           → CreateProfile
GET  /api/profile/{id}      → GetProfileByID
PUT  /api/profile/{id}      → UpdateProfile
GET  /api/profile/user/{uid}→ GetProfileByUserID
GET  /api/menu              → GetAllMenus
GET  /api/menu/{accountType}→ GetMenusByAccountType
POST /api/upload/image      → UploadImage (Azure Blob)
```

### saving-service
```
GET  /api/exchange-rates    → GetExchangeRates (all)
GET  /api/interest-rates    → GetInterestRates (all)
GET  /api/branches?q=       → GetBranches (ILIKE search)
GET  /health                → Health check
```

## Docker Compose Setup

```bash
# Per service (local dev)
cd saving-service && docker compose -f docker-compose.local.yml up --build -d
cd identity-service && docker compose -f docker-compose.local.yml up --build -d
cd user-profile-service && docker compose -f docker-compose.local.yml up --build -d

# BFF full stack (all 5 containers)
cd bff-service && docker compose up --build
```

| Service              | HTTP | gRPC | Database      | Docker Compose                    |
| -------------------- | ---- | ---- | ------------- | --------------------------------- |
| bff-service          | 3000 | 9090 | — (stateless) | `bff-service/docker-compose.yml`  |
| identity-service     | 3031 | 9301 | identity_db   | `identity-service/docker-compose.local.yml` |
| user-profile-service | 8080 | 9302 | bankease_db   | `user-profile-service/docker-compose.local.yml` |
| saving-service       | 8081 | 9303 | saving        | `saving-service/docker-compose.local.yml` |

## BFF HTTP Gateway Pattern

- BFF uses manual REST→gRPC bridge (NOT grpc-gateway codegen)
- `contextFromHTTPRequest` MUST verify JWT and inject `user_claims` into context
- gRPC interceptor chain does NOT run for direct function calls from HTTP gateway
- Store `jwtMgr` as package-level var, initialize in `startHTTPServer`
- Upload handler: multipart/form-data → Azure Blob Storage direct (not via gRPC)

## Swagger Pattern (BFF)

- swaggo annotations on HTTP handler files
- For multi-method handlers (handleProfile, handleProfileByID), use doc stub functions in `swagger_docs.go`
- Type references: `pb.*` in http_routes.go, `bff_service.*` in swagger_docs.go
- Swagger UI route: `http.StripPrefix` + `httpSwagger.Handler(httpSwagger.URL(...))`
- URL: `http://localhost:3000/swagger/bff/`

## Important Patterns

- Seed data: `docker-entrypoint-initdb.d` only runs on fresh volume; use `docker cp` + `psql -f` for re-seed
- Menu filter: PREMIUM → all menus, REGULAR → only REGULAR menus
- Balance in minor unit (cents/pence)
- gRPC handler pattern: `var _ pb.XxxServiceServer = (*Server)(nil)` for compile-time check
- Unit test pattern: `newTestServer(t)` → `sqlmock.New()` + `testify/assert`
