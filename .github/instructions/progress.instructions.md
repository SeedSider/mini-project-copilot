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
- [x] gRPC layer added: proto files, hand-written protogen, `internal/grpchandler/` (6 RPC)
- [x] `StartGRPC()` melayani port 9302 bersamaan dengan REST port 8080

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

### BFF Service Spec — SELESAI ✅

- [x] Backend spec BFF service (`backend-spec-bff-service.md`) — 11 endpoint lengkap
- [x] Proto definitions (BFF + user-profile-service baru)
- [x] Orchestration flows (SignUp, SignIn, GetMe, Profile, Menu, Upload)
- [x] Docker Compose full stack (5 containers)
- [x] Testing checklist 30+ test cases

### Infrastructure

- [x] Memory Bank initialized dan updated (semua core files)
- [x] identity-service disatukan ke monorepo

## In Progress

- [ ] Verifikasi gRPC listener identity-service aktif di port 9301 (identity_grpc_api.go sudah ada, perlu cek server.go expose port)
- [ ] Unit tests coverage ≥ 90% identity-service (go test -coverprofile)

## Not Started

### Implementasi bff-service

- [ ] Go project scaffolding (`bff-service/`)
- [ ] Proto code generation (BFF, identity client, profile client)
- [ ] gRPC server + grpc-gateway setup
- [ ] ServiceConnection (identity + profile gRPC clients)
- [ ] Auth handlers (SignUp orchestration, SignIn proxy, GetMe)
- [ ] Profile handlers (proxy to user-profile-service)
- [ ] Menu handlers (proxy to user-profile-service)
- [ ] Upload handler (direct to Azure Blob)
- [ ] JWT verify (lokal, secret sama dgn identity)
- [ ] Interceptor chain (ProcessId → Logging → Auth)
- [ ] Docker Compose full stack integration

### Prerequisite: Aktifkan gRPC listener identity-service

- [ ] Pastikan gRPC server listener di-expose di port 9301 dari `server/main.go`

### Unit Tests & Quality

- [ ] Unit tests identity-service coverage ≥ 90% (test files sudah ada, perlu verifikasi coverage)
- [ ] Unit tests user-profile-service (target ≥ 90%) — belum ada test file
- [ ] Unit tests bff-service (target ≥ 90%)
- [ ] SonarQube analysis pass untuk semua service

### Future Enhancements

- [ ] Frontend integration — hubungkan mobile app ke backend real
- [ ] Email verification, forgot password, refresh token
- [ ] Rate limiting pada login endpoint
- [ ] Graceful shutdown

## Known Issues

- identity-service: perlu verifikasi port 9301 gRPC listener benar-benar di-expose dari `server/main.go` (handler sudah ada di `identity_grpc_api.go`)
- user-profile-service: belum ada unit tests
- Kedua services belum melalui SonarQube analysis

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
