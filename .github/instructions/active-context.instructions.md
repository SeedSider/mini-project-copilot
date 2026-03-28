---
applyTo: "**"
---

# Active Context

## Current Focus

- `identity-service` — **SELESAI** (Tahap 1-3 + unit tests + gRPC server). Dikembangkan di project terpisah (`mini-project-copilot-identity/`), kemudian disatukan ke monorepo ini.
- `user-profile-service` — **SELESAI** termasuk gRPC layer (port 9302). Docker Compose running, semua endpoint (REST + gRPC) siap dipanggil BFF.
- `bff-service` — **SPEC SELESAI** (`backend-spec-bff-service.md`), belum implementasi.

## Recent Changes

- **user-profile-service gRPC layer ditambahkan** — proto files, hand-written protogen, `internal/grpchandler/` (6 RPC), `StartGRPC()` di server, port 9302
- **identity-service gRPC server diimplementasi** — `identity_grpc_api.go` (SignUp, SignIn, GetMe via gRPC; tanpa best-effort HTTP ke profile — BFF jadi orchestrator)
- **identity-service unit tests ditambahkan** — 7 file test (api, interceptor, provider, constant, jwt, utils, database); menggunakan `go-sqlmock` + `testify`
- **identity-service dan user-profile-service** masing-masing punya `docker-compose.local.yml` untuk dev lokal
- Backend spec BFF service ditulis lengkap (`backend-spec-bff-service.md`) — 11 endpoint, proto definitions, orchestration flows, Docker Compose full stack
- Memory bank files `.identity-service/instructions/` tersedia sebagai referensi history identity-service

## What's Working

### identity-service

- Docker Compose running (PostgreSQL 15 + identity-service)
- 4 HTTP endpoint aktif: POST /api/auth/signup, POST /api/auth/signin, GET /api/identity/me, GET /health
- gRPC server handler tersedia: SignUp, SignIn, GetMe (`identity_grpc_api.go`)
- JWT HS256 auth, bcrypt password hashing
- Interceptor chain: ProcessId → Logging → Errors → Auth
- Functional test 12/12 lulus (signup, signin, getme, error cases)
- Unit tests coverage meningkat: 7 test file ditambahkan

### user-profile-service

- Docker Compose running (PostgreSQL 17 + app di localhost:8080, gRPC 9302)
- 8 REST endpoint aktif: Profile CRUD (5), Menu (2), Upload (1)
- **6 gRPC endpoint aktif**: CreateProfile, GetProfileByID, GetProfileByUserID, UpdateProfile, GetAllMenus, GetMenusByAccountType
- Seed data auto-loaded (1 profile + 9 menu items)
- Business logic menu filter (PREMIUM → semua, REGULAR → hanya REGULAR) terverifikasi
- `StartGRPC()` melayani port 9302 bersamaan dengan REST port 8080

## Next Steps

1. **Implementasi bff-service** — berdasarkan `backend-spec-bff-service.md`; downstream services sudah siap (identity gRPC 9301, profile gRPC 9302)
2. **Aktifkan gRPC listener identity-service** — `identity_grpc_api.go` sudah ada, perlu pastikan server.go meng-expose port 9301
3. **Unit tests coverage** — verifikasi coverage ≥ 90% untuk identity-service (go test -coverprofile)
4. **Unit tests user-profile-service** — belum ada test file
5. **SonarQube** — static analysis pass untuk semua service

## Active Decisions

- identity-service: HTTP handler tetap aktif (port 3031), gRPC handler `identity_grpc_api.go` sudah ada
- user-profile-service: REST (chi, port 8080) + gRPC (port 9302) berjalan bersamaan
- BFF akan menggunakan grpc-gateway pattern (sama dengan addons-issuance-lc-service)
- JWT secret key harus sama di identity-service dan BFF (lokal verification)
- Upload image di BFF langsung ke Azure Blob (tidak lewat user-profile-service)
- Docker Compose full stack: 5 containers (BFF + identity + profile + 2x PostgreSQL)
- identity-service module path tetap `bitbucket.bri.co.id/scm/addons/addons-identity-service`
- identity-service gRPC SignUp TIDAK melakukan best-effort HTTP ke profile — BFF yang meng-orchestrate

## Important Patterns

- Semua user-profile-service response: `{code, description}`
- Semua identity-service error response: `{error, code, message}`
- Menu filter: PREMIUM → semua menu, REGULAR → hanya menu REGULAR
- Balance dalam minor unit (cents/pence)
- SignUp orchestration (BFF): identity.SignUp → profile.CreateProfile (best-effort)
- Interceptor chain: ProcessId → Logging → Errors → Auth
- gRPC handler pattern: `var _ pb.XxxServiceServer = (*Server)(nil)` untuk compile-time check
- Unit test pattern identity-service: `newTestServer(t)` → `sqlmock.New()` + `testify/assert`
