---
applyTo: "**"
---

# Progress

## Completed

### user-profile-service — SELESAI ✅

- [x] Backend API spec (`backend-spec.md`) — kontrak lengkap semua endpoint
- [x] Analisis service referensi (`addons-issuance-lc-service`) — pattern dan arsitektur
- [x] Detail implementasi (`implementation-plan.md`) — 8 phase, file mapping, verification checklist
- [x] Go project initialization (`go.mod`, folder structure) — `user-profile-service/`
- [x] Database layer (`db.go`, `migrate.go`, `001_init.sql`) — embed.FS migration
- [x] Models (Profile, Menu, StandardResponse, MenuResponse)
- [x] Repository layer (ProfileRepository, MenuRepository)
- [x] Handler layer (ProfileHandler, MenuHandler + writeJSON/writeError helpers)
- [x] Server & Router setup (chi router, CORS middleware, Logger, Recoverer)
- [x] Entrypoint (`cmd/server/main.go`) — godotenv + GetEnv pattern
- [x] Seed data (`seed.sql`) — 1 profile + 9 menu items
- [x] Docker setup (`Dockerfile`, `docker-compose.yml`, `.dockerignore`)
- [x] Additional features: CreateProfile, GetProfileByUserID, GetMyProfile (JWT), Upload Image (Azure Blob)
- [x] All 8 endpoints verified (Profile CRUD 5 + Menu 2 + Upload 1)
- [x] gRPC layer added: proto files, hand-written protogen, gRPC handlers (9 RPC)
- [x] gRPC server melayani port 9302 bersamaan dengan REST port 8080
- [x] **Search Feature API**: 3 endpoint baru (GET /api/exchange-rates, GET /api/interest-rates, GET /api/branches?q=)
  - [x] Migrations 004-006 (exchange_rate, interest_rate, branch tables)
  - [x] Models: ExchangeRate, InterestRate, Branch (raw array response, no wrapper)
  - [x] Repositories: ExchangeRateRepository, InterestRateRepository, BranchRepository (branch: GetAll + SearchByName ILIKE)
  - [x] HTTP handler: SearchHandler
  - [x] gRPC handler: GetExchangeRates, GetInterestRates, GetBranches
  - [x] Proto + protogen updated (3 new RPCs + 6 new messages, hand-written)
  - [x] Swagger docs updated (swagger.json, swagger.yaml, docs.go) — tag "Search"
  - [x] Seed data loaded into DB (4 exchange rates, 4 interest rates, 5 branches)
  - [x] `go build ./...` — compile pass ✅
- [x] **implementation-search-feature.md** — implementation plan markdown di root project
- [x] **REFACTORED ke identity-service folder structure** (2026-03-30):
  - [x] `cmd/server/main.go` → `server/main.go` (graceful shutdown, signal handling)
  - [x] `internal/handlers/` + `internal/grpchandler/` → `server/api/` (HTTP + gRPC handlers pada `*Server` struct)
  - [x] `internal/repository/` → `server/db/` (Provider pattern: provider.go, profile_provider.go, menu_provider.go, search_provider.go)
  - [x] `internal/models/` → types di-embed ke `server/db/` + `server/api/`
  - [x] `internal/server/` → di-absorb ke `server/main.go` (routes + cors inlined)
  - [x] `internal/db/` → `server/core_db.go` + `migrations/embed.go`
  - [x] Ditambahkan: `server/core_config.go`, `server/constant/constant.go`, `server/utils/utils.go`
  - [x] `server/api/api.go` — Server struct + `var _ pb.UserProfileServiceServer = (*Server)(nil)` compile-time check
  - [x] `server/api/converter.go` — model ↔ proto conversion helpers
  - [x] Dockerfile updated: `go build -o /server ./server`
  - [x] `go build ./server/` — compile pass ✅
  - [x] BFF cross-compile pass ✅
  - [x] Old `cmd/` dan `internal/` directories dihapus

### identity-service — SELESAI ✅ (dari project terpisah)

> Dikembangkan di `mini-project-copilot-identity/`, kemudian disatukan ke monorepo `mini-project-copilot/identity-service/`
> Memory bank history tersedia di `.identity-service/instructions/`

**Tahap 1 — Foundation: SELESAI ✅**

- [x] Go module (`go.mod`) — `bitbucket.bri.co.id/scm/addons/addons-identity-service`
- [x] Proto source files (`proto/identity_api.proto`, `proto/identity_payload.proto`)
- [x] Entry point + CLI (`server/main.go`) — command `grpc-gw-server`
- [x] Config loading (`server/core_config.go`) — godotenv + os.LookupEnv
- [x] DB connection lifecycle (`server/core_db.go`) — dengan retry logic
- [x] API struct + constructor (`server/api/api.go`)
- [x] Auth interceptor JWT (`server/api/identity_authInterceptor.go`)
- [x] Interceptor chain (`server/api/identity_interceptor.go`) — ProcessId → Logging → Errors → Auth
- [x] Error helpers (`server/api/error.go`)
- [x] DB provider (`server/db/provider.go`)
- [x] DB queries (`server/db/identity_provider.go`) — CreateUser, GetUserByUsername, CheckUsernameExists
- [x] JWT manager (`server/jwt/manager.go`) — HS256 Generate + Verify
- [x] DB wrapper (`server/lib/database/`) — interface, wrapper, mock
- [x] Logger (`server/lib/logger/`) — Zap + FluentBit hook
- [x] Utils, Constants
- [x] Migration SQL (`migrations/001_init.sql`, `002_rename_email_to_username.sql`)

**Tahap 2 — API Implementation: SELESAI ✅**

- [x] `POST /api/auth/signup` — SignUp (validasi, bcrypt, INSERT + best-effort create profile)
- [x] `POST /api/auth/signin` — SignIn (bcrypt compare, JWT generate)
- [x] `GET /api/identity/me` — GetMe (JWT verify, get user)
- [x] `GET /health` — Health check
- [x] CORS middleware + security headers
- [x] gRPC → HTTP error code mapping

**Tahap 3 — Deployment: SELESAI ✅**

- [x] Dockerfile (multi-stage: golang:1.24-alpine → alpine:3.19)
- [x] Docker Compose (PostgreSQL 15 + identity-service)
- [x] Swagger JSON, Makefile, sonar-project.properties
- [x] `go build ./server/` — compile pass ✅
- [x] Docker Compose running — service UP ✅

**Functional Testing: LULUS 12/12 ✅**

- [x] Health check, SignUp (success + 3 error cases), SignIn (success + 2 error cases), GetMe (success + 2 error cases)

**Unit Tests: DITAMBAHKAN ✅**

- [x] `server/api/identity_auth_api_test.go` — SignUp, SignIn, GetMe (HTTP + gRPC)
- [x] `server/api/identity_interceptor_test.go` — interceptor chain
- [x] `server/db/identity_provider_test.go` — DB queries (sqlmock)
- [x] `server/constant/constant_test.go` — constants
- [x] `server/jwt/manager_test.go` — JWT generate + verify
- [x] `server/utils/utils_test.go` — utility functions
- [x] `server/lib/database/database_test.go` — DB wrapper
- [x] Pattern: `newTestServer(t)` → `sqlmock.New()` + `testify/assert`

**gRPC Server Handler: DITAMBAHKAN ✅**

- [x] `server/api/identity_grpc_api.go` — SignUp, SignIn, GetMe via gRPC
- [x] `var _ pb.IdentityServiceServer = (*Server)(nil)` compile-time check
- [x] SignUp gRPC TIDAK melakukan best-effort HTTP ke profile (BFF orchestrates)

### saving-service — SELESAI ✅

- [x] Diekstrak dari user-profile-service search feature menjadi service mandiri
- [x] Folder structure: `server/` (main.go, core_config.go, core_db.go, api/, db/, constant/, utils/, lib/)
- [x] 3 REST endpoint: GET /api/exchange-rates, GET /api/interest-rates, GET /api/branches?q=
- [x] 3 gRPC RPC: GetExchangeRates, GetInterestRates, GetBranches
- [x] Proto definitions: `saving_api.proto`, `saving_payload.proto`
- [x] Hand-written protogen + codec.go (JSONCodec via `encoding.RegisterCodec` in `init()`)
- [x] Migrations: 001_add_exchange_rates.sql, 002_add_interest_rates.sql, 003_add_branches.sql + embed.go
- [x] DB Provider pattern: provider.go, exchange_rate_provider.go, interest_rate_provider.go, branch_provider.go
- [x] Docker Compose: PostgreSQL 17 (port 5434) + saving-service (HTTP 8081, gRPC 9303)
- [x] Seed data: 4 exchange rates + 4 interest rates + 5 branches
- [x] Swagger docs + UI di `/swagger/`
- [x] CLI: `urfave/cli` (grpc-server, gw-server, grpc-gw-server)
- [x] Health check: GET /health
- [x] `go build ./server/` + `go vet ./server/` — compile pass ✅
- [x] Docker Compose RUNNING ✅
- [x] Seed data migrated to database ✅

### BFF Service — SELESAI ✅

- [x] Backend spec BFF service (`backend-spec-bff-service.md`) — 14 endpoint lengkap (11 original + 3 saving proxy)
- [x] Go project scaffolding (`bff-service/`, `go.mod`, folder structure)
- [x] Proto definitions (`bff_api.proto`, `bff_payload.proto`)
- [x] Protogen — hand-written gRPC stubs (BFF server, identity client, profile client, saving client)
- [x] Server entry point + CLI (`server/main.go`) — `grpc-gw-server` command
- [x] Config loading (`server/core_config.go`) — godotenv + os.LookupEnv
- [x] JWT Manager (`server/jwt/manager.go`) — Verify only (HS256, local verification)
- [x] ServiceConnection (`server/services/service.go`) — gRPC clients to identity + profile + saving
- [x] Interceptor chain (`server/api/bff_interceptor.go`) — ProcessId → Logging → Auth
- [x] Auth interceptor (`server/api/bff_authInterceptor.go`) — JWT verify for GetMe, GetMyProfile
- [x] Auth handlers (`server/api/bff_auth_api.go`) — SignUp (orchestrated), SignIn (proxy), GetMe
- [x] Profile handlers (`server/api/bff_profile_api.go`) — GetMyProfile, GetProfileByID, GetProfileByUserID, CreateProfile, UpdateProfile
- [x] Menu handlers (`server/api/bff_menu_api.go`) — GetAllMenus, GetMenusByAccountType
- [x] **Saving handlers** (`server/api/bff_saving_api.go`) — GetExchangeRates, GetInterestRates, GetBranches
- [x] HTTP gateway + routes (`server/http_routes.go`) — manual REST→gRPC bridge (14 routes)
- [x] Upload handler (`server/gateway_http_handler.go`) — multipart/form-data → Azure Blob direct
- [x] CORS + security headers middleware
- [x] Error helpers + gRPC→HTTP status code mapping
- [x] Logger (`server/lib/logger/`) — Zap structured logger
- [x] Utils, Constants
- [x] Dockerfile (multi-stage: golang:1.24-alpine → alpine:3.20)
- [x] Docker Compose full stack (`docker-compose.yml` di root) — containers for BFF + downstream services
- [x] `.env.example`, `Makefile`, `sonar-project.properties`, `.dockerignore`
- [x] `go build ./server/` — compile pass ✅
- [x] **JWT fix** — `contextFromHTTPRequest` verify JWT + inject `user_claims`; `jwtMgr` package-level var ✅
- [x] **`protogen/identity-service/codec.go`** — JSONCodec registered, gRPC server handle JSON requests ✅
- [x] **Docker Compose full stack RUNNING & VERIFIED** — SignUp → SignIn → GET /api/profile end-to-end ✅
- [x] **Swagger UI DITAMBAHKAN** — BearerAuth security definition
  - [x] Dependencies: `swaggo/swag@v1.16.6` + `swaggo/http-swagger@v1.3.4`
  - [x] Generated docs: `docs/docs.go`, `docs/swagger.json`, `docs/swagger.yaml`
  - [x] Swagger UI route: `http://localhost:3000/swagger/bff/`
  - [x] Annotations pada `server/http_routes.go` (Auth 3, Menu 2, Profile/user/{id} 1)
  - [x] Doc stubs pada `server/swagger_docs.go` (Profile GET/POST, Profile/{id} GET/PUT, Upload)
  - [x] `@securityDefinitions.apikey BearerAuth` untuk protected endpoints
  - [x] `go build ./server/` + `go vet ./server/` — pass ✅
- [x] **saving-service integration** — 3 new routes: GET /api/exchange-rates, /api/interest-rates, /api/branches
  - [x] `protogen/saving-service/saving_api_grpc.pb.go` — gRPC client stub
  - [x] `server/services/service.go` — SavingService gRPC client (port 9303)
  - [x] `server/api/bff_saving_api.go` — GetExchangeRates, GetInterestRates, GetBranches handlers
  - [x] `server/core_config.go` — `SavingServiceAddr` config (default: localhost:9303)

### BFF Service Spec — SELESAI ✅

- [x] Backend spec BFF service (`backend-spec-bff-service.md`) — 11 endpoint lengkap
- [x] Proto definitions (BFF + user-profile-service baru)
- [x] Orchestration flows (SignUp, SignIn, GetMe, Profile, Menu, Upload)
- [x] Docker Compose full stack (5 containers)
- [x] Testing checklist 30+ test cases

### Infrastructure

- [x] Memory Bank initialized dan updated (semua core files)
- [x] identity-service disatukan ke monorepo
- [x] Docker Compose full stack UP (5 containers) ✅ VERIFIED 2026-03-30
- [x] End-to-end flow verified: POST /signup → POST /signin → GET /api/profile ✅

## In Progress

_(tidak ada item in progress saat ini)_

## Not Started

### Docker Compose Full Stack Test

- [x] Jalankan `docker compose up --build` dari root, verifikasi semua 5 containers UP ✅
- [x] Functional testing BFF (partial) — signup, signin, GET /api/profile ✅
- [ ] Functional testing BFF lanjutan — 30+ test cases dari testing checklist

### Unit Tests & Quality

- [ ] Unit tests bff-service (target ≥ 90%)
- [ ] Unit tests user-profile-service (target ≥ 90%) — belum ada test file
- [ ] Unit tests saving-service (target ≥ 90%) — belum ada test file
- [ ] Unit tests identity-service coverage ≥ 90% (test files sudah ada, perlu verifikasi coverage)
- [ ] SonarQube analysis pass untuk semua service

### Future Enhancements

- [ ] Frontend integration — hubungkan mobile app ke backend real
- [ ] Email verification, forgot password, refresh token
- [ ] Rate limiting pada login endpoint
- [ ] Graceful shutdown

## Known Issues

- user-profile-service: belum ada unit tests
- saving-service: belum ada unit tests
- Semua services belum melalui SonarQube analysis
- **Seed data caveat**: `docker-entrypoint-initdb.d` hanya jalan saat volume fresh; gunakan `docker exec -i <container> psql -U postgres -d <db> -f /docker-entrypoint-initdb.d/seed.sql` untuk re-seed
- **`create_file` tool caveat**: tool bisa melaporkan sukses tapi file tidak terbuat di disk. SELALU verifikasi dengan `Get-Item <path>` setelah membuat file baru.

## Architecture Decisions Log

| Keputusan                               | Alasan                                             | Tanggal    |
| --------------------------------------- | -------------------------------------------------- | ---------- |
| REST-only user-profile-service          | Scope awal kecil, mobile app cuma perlu REST       | 2026-03-25 |
| database/sql (tanpa GORM)               | Lebih ringan, query simple                         | 2026-03-25 |
| chi router (user-profile)               | Lightweight, idiomatic Go, middleware support      | 2026-03-25 |
| net/http (identity, tanpa grpc-gateway) | Tidak ada toolchain protoc, HTTP handler pragmatis | 2026-03-25 |
| Docker Compose per service              | Satu command untuk start semua                     | 2026-03-25 |
| identity-service module path bitbucket  | Konsisten dengan naming convention BRI             | 2026-03-25 |
| BFF pattern — single entry point        | Mobile app hanya perlu 1 endpoint                  | 2026-03-27 |
| gRPC inter-service communication        | Standard BRICaMS pattern                           | 2026-03-27 |
| JWT local verification di BFF           | Tidak perlu call identity utk setiap request       | 2026-03-27 |
| Upload langsung di BFF ke Azure Blob    | Tidak perlu forward file besar via gRPC            | 2026-03-27 |
| identity-service dari project terpisah  | Dikembangkan paralel, disatukan ke monorepo        | 2026-03-27 |
| Search endpoints — raw array response   | Sesuai api.txt spec; tidak perlu wrapper envelope  | 2026-03-30 |
| Branch search — ILIKE SQL               | Case-insensitive partial match sesuai api.txt spec | 2026-03-30 |
| Re-seed via docker cp + psql            | docker-entrypoint-initdb.d tidak re-run jika volume sudah ada | 2026-03-30 |
| BFF JWT fix via contextFromHTTPRequest  | HTTP gateway calls gRPC handlers directly — interceptor chain tidak jalan | 2026-03-30 |
| codec.go wajib di setiap protogen pkg   | Hand-written types tidak implement proto.Message; gRPC fallback ke proto codec | 2026-03-30 |
| user-profile-service refactor ke server/ | Konsistensi folder structure dengan identity-service & BFF | 2026-03-30 |
| BFF Swagger via swaggo                  | Konsisten dengan user-profile-service; swaggo annotation-based doc gen | 2026-03-30 |
| Search endpoints → saving-service        | Separation of concerns; search data bukan tanggung jawab user-profile  | 2026-03-31 |
| saving-service standalone                | Service mandiri dengan DB sendiri (PostgreSQL `saving`, port 5434)      | 2026-03-31 |
| BFF terhubung ke 3 downstream services   | identity (9301) + user-profile (9302) + saving (9303)                  | 2026-03-31 |
