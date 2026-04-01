---
name: project-progress
description: "BankEase project progress tracker and historical audit log. Use when: reviewing project status, updating memory bank, sprint planning, checking what's completed vs remaining, reviewing architecture decisions log, or checking known issues. Contains detailed checklist of all completed work across 4 services."
argument-hint: "What to check (e.g. 'completed items', 'remaining work', 'known issues', 'architecture decisions')"
---

# BankEase Project Progress

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
- [x] Search Feature API: 3 endpoint (exchange-rates, interest-rates, branches)
- [x] REFACTORED to identity-service folder structure (2026-03-30)
- [x] `go build ./server/` — compile pass ✅

### identity-service — SELESAI ✅

> Dikembangkan di `mini-project-copilot-identity/`, disatukan ke monorepo

- [x] Tahap 1 — Foundation: Go module, proto, CLI, config, DB, API struct, interceptors, JWT, DB provider, logger, migrations
- [x] Tahap 2 — API: SignUp, SignIn, GetMe, Health check, CORS, gRPC→HTTP mapping
- [x] Tahap 3 — Deployment: Dockerfile, Docker Compose, Swagger, Makefile
- [x] Functional Testing: 12/12 passed (signup, signin, getme, error cases)
- [x] Unit Tests: 7 test files (api, interceptor, db, constant, jwt, utils, database wrapper)
- [x] gRPC Server: SignUp, SignIn, GetMe via gRPC (`identity_grpc_api.go`)
- [x] **Forgot Password** (2026-04-01): `ValidateOtp` (random 6-digit via crypto/rand) + `UpdatePassword` (JWT-protected, username from JWT claims)
  - [x] `identity_forgot_password_api.go` + `identity_forgot_password_api_test.go` (29 tests)
  - [x] `UpdatePasswordByUsername` added to `identity_provider.go` + tests
  - [x] Proto + protogen updated (ValidateOtpRequest/Response, UpdatePasswordRequest/Response)
  - [x] HTTP routes: POST /api/auth/validate-otp, PUT /api/auth/update-password
  - [x] Coverage: api 91.2%, db 100%, jwt 93.3% ✅
  - [x] Docker Compose running & end-to-end verified ✅

### saving-service — SELESAI ✅

- [x] Extracted from user-profile-service search feature
- [x] 3 REST endpoints + 3 gRPC RPCs (exchange-rates, interest-rates, branches)
- [x] Proto + hand-written protogen + codec.go
- [x] Migrations + DB Provider pattern
- [x] Docker Compose: PostgreSQL 17 (port 5434) + service (HTTP 8081, gRPC 9303)
- [x] Seed data: 4 exchange rates + 4 interest rates + 5 branches
- [x] Swagger docs + UI, CLI (urfave/cli), Health check
- [x] `go build ./server/` + `go vet ./server/` — compile pass ✅

### BFF Service — SELESAI ✅

- [x] 17 endpoints implemented (14 original + 3 payment proxy)
- [x] Proto + protogen (BFF server + 4 downstream clients)
- [x] Server: CLI, config, JWT (verify only), ServiceConnection (4 gRPC clients)
- [x] Interceptor chain: ProcessId → Logging → Auth
- [x] Auth, Profile, Menu, Saving, Payment handlers
- [x] HTTP gateway + routes (manual REST→gRPC bridge)
- [x] Upload: multipart/form-data → Azure Blob direct
- [x] JWT fix: `contextFromHTTPRequest` verify + inject `user_claims`
- [x] Docker Compose full stack RUNNING & VERIFIED (signup → signin → profile e2e)
- [x] Swagger UI at `/swagger/bff/` with BearerAuth
- [x] saving-service integration (3 routes)
- [x] payment-service integration (3 routes: providers, internet-bill, currency-list)

### Infrastructure

- [x] Memory Bank initialized and updated
- [x] identity-service merged into monorepo
- [x] Docker Compose full stack UP (5 containers) ✅ VERIFIED 2026-03-30
- [x] End-to-end flow verified: POST /signup → POST /signin → GET /api/profile ✅

### payment-service — SELESAI ✅

- [x] Phase 1 — Foundation: go.mod, proto, protogen (codec.go + hand-written stubs), migrations (3 SQL + embed.go), lib/, constant/, utils/, seed.sql, .env.example
- [x] Phase 2 — Core Infrastructure: core_config.go (Config + initConfig + JWTSecret), core_db.go (DB wrapper + retry + embed.FS migration), jwt/manager.go (Verify only)
- [x] Phase 3 — Data Layer: db/provider.go, db/bill_provider.go (GetAllProviders, GetInternetBillByUserID), db/currency_provider.go (GetAllCurrencies), db/error.go (NotFoundErr)
- [x] Phase 4 — API Layer: api.go (Server struct + compile-time check), payment_api.go (3 HTTP handlers), payment_grpc_api.go (3 gRPC RPCs), payment_authInterceptor.go (JWT, skip public), payment_interceptor.go (chain), converter.go, error.go
- [x] Phase 5 — Server Entry Point: main.go (CLI + HTTP/gRPC + jwtMiddleware + cors + routes)
- [x] Phase 6 — Docker & Config: Dockerfile (multi-stage), docker-compose.yml, docker-compose.local.yml, Makefile, sonar-project.properties
- [x] Phase 7 — Swagger Docs: docs/docs.go, swagger.json, swagger.yaml
- [x] Phase 8 — BFF Integration: protogen client stubs, bff_payment_api.go, ServiceConnection (+PaymentService), http_routes (+3), bff_api_grpc.pb.go (+3 RPCs), bff_payload.pb.go (+payment types), core_config (+PaymentServiceAddr), docker-compose updated
- [x] Phase 9 — Verification: `go build ./server/` + `go vet ./server/` pass for payment-service AND bff-service

## Not Started

### Remaining Tasks

- [ ] Unit tests payment-service (target ≥ 90%)
- [ ] Functional testing BFF — 30+ test cases from testing checklist
- [ ] SonarQube analysis pass for all services
- [ ] Docker Compose full stack test (5 services + 4 DBs)

## In Progress / Recently Completed (2026-04-01)

- [x] payment-service IMPLEMENTED ✅ — 30 new files, 9 phases complete
  - 3 endpoints: GET /api/pay-the-bill/providers, GET /api/pay-the-bill/internet-bill (JWT), GET /api/currency-list
  - gRPC + HTTP dual server, JWT auth interceptor, PostgreSQL (port 5435)
  - Docker Compose, Makefile, Swagger docs, sonar-project.properties
- [x] BFF payment integration ✅ — 4th downstream service
  - New: `bff-service/protogen/payment-service/`, `bff-service/server/api/bff_payment_api.go`
  - Modified: core_config (+PaymentServiceAddr), services/service.go (+PaymentService client), main.go (+4th arg), http_routes.go (+3 handlers), bff_api_grpc.pb.go (+3 RPCs), bff_payload.pb.go (+payment types)
- [x] All 5 services compile verified: `go build ./server/` pass for identity, user-profile, saving, payment, bff
- [x] Unit tests user-profile-service ✅ — 5 test files
- [x] Unit tests saving-service ✅ — 4 test files

## Known Issues

- bff-service: no unit tests yet
- payment-service: no unit tests yet
- All services: no SonarQube analysis yet
- Seed data caveat: `docker-entrypoint-initdb.d` only runs on fresh volume; use `docker exec -i <container> psql -U postgres -d <db> -f /docker-entrypoint-initdb.d/seed.sql` for re-seed
- `create_file` tool caveat: tool may report success but file not created on disk. Always verify with `Get-Item <path>`.

## Architecture Decisions Log

| Keputusan                                 | Alasan                                                                           | Tanggal    |
| ----------------------------------------- | -------------------------------------------------------------------------------- | ---------- |
| REST-only user-profile-service            | Scope awal kecil, mobile app cuma perlu REST                                     | 2026-03-25 |
| database/sql (tanpa GORM)                 | Lebih ringan, query simple                                                       | 2026-03-25 |
| chi router (user-profile)                 | Lightweight, idiomatic Go, middleware support                                    | 2026-03-25 |
| net/http (identity, tanpa grpc-gateway)   | Tidak ada toolchain protoc, HTTP handler pragmatis                               | 2026-03-25 |
| Docker Compose per service                | Satu command untuk start semua                                                   | 2026-03-25 |
| identity-service module path bitbucket    | Konsisten dengan naming convention BRI                                           | 2026-03-25 |
| BFF pattern — single entry point          | Mobile app hanya perlu 1 endpoint                                                | 2026-03-27 |
| gRPC inter-service communication          | Standard BRICaMS pattern                                                         | 2026-03-27 |
| JWT local verification di BFF             | Tidak perlu call identity utk setiap request                                     | 2026-03-27 |
| Upload langsung di BFF ke Azure Blob      | Tidak perlu forward file besar via gRPC                                          | 2026-03-27 |
| identity-service dari project terpisah    | Dikembangkan paralel, disatukan ke monorepo                                      | 2026-03-27 |
| Search endpoints — raw array response     | Sesuai api.txt spec; tidak perlu wrapper envelope                                | 2026-03-30 |
| Branch search — ILIKE SQL                 | Case-insensitive partial match sesuai api.txt spec                               | 2026-03-30 |
| Re-seed via docker cp + psql              | docker-entrypoint-initdb.d tidak re-run jika volume sudah ada                    | 2026-03-30 |
| BFF JWT fix via contextFromHTTPRequest    | HTTP gateway calls gRPC handlers directly — interceptor chain tidak jalan        | 2026-03-30 |
| codec.go wajib di setiap protogen pkg     | Hand-written types tidak implement proto.Message; gRPC fallback ke proto codec   | 2026-03-30 |
| user-profile-service refactor ke server/  | Konsistensi folder structure dengan identity-service & BFF                       | 2026-03-30 |
| BFF Swagger via swaggo                    | Konsisten dengan user-profile-service; swaggo annotation-based doc gen           | 2026-03-30 |
| Search endpoints → saving-service         | Separation of concerns; search data bukan tanggung jawab user-profile            | 2026-03-31 |
| saving-service standalone                 | Service mandiri dengan DB sendiri (PostgreSQL `saving`, port 5434)               | 2026-03-31 |
| BFF terhubung ke 3 downstream services    | identity (9301) + user-profile (9302) + saving (9303)                            | 2026-03-31 |
| payment-service planned                   | 5th service: providers, internet-bill (JWT), currency-list; ports 8082/9304/5435 | 2026-04-01 |
| payment-service auth hybrid               | internet-bill protected (JWT), providers + currency-list public                  | 2026-04-01 |
| payment-service DB wrapper identity-style | Consistency: use lib/database/ wrapper across all services                       | 2026-04-01 |
| BFF terhubung ke 4 downstream             | identity (9301) + user-profile (9302) + saving (9303) + payment (9304)           | 2026-04-01 |
| payment-service IMPLEMENTED               | 30 files, 3 endpoints, HTTP+gRPC, JWT intercept, Swagger, Docker                 | 2026-04-01 |
