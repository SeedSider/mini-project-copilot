---
applyTo: "**"
---

# Active Context

## Current Focus

- `identity-service` — **SELESAI** (Tahap 1-3 + unit tests + gRPC server). Dikembangkan di project terpisah (`mini-project-copilot-identity/`), kemudian disatukan ke monorepo ini.
- `user-profile-service` — **SELESAI** termasuk gRPC layer (port 9302) + Search Feature API (exchange-rates, interest-rates, branches). Docker Compose running, semua 11 endpoint (REST + gRPC) siap dipanggil BFF.
- `bff-service` — **IMPLEMENTASI SELESAI**. Compile pass, semua 11 endpoint diimplementasi (auth, profile, menu, upload). Docker Compose full stack ready.

## Recent Changes

- **user-profile-service Search Feature API DIIMPLEMENTASI** — 3 endpoint baru (GET /api/exchange-rates, GET /api/interest-rates, GET /api/branches?q=), REST + gRPC, migration 004-006, seed data, swagger docs updated
- **implementation-search-feature.md** — implementation plan markdown dibuat di root project
- **Seed data berhasil dimuat** ke database via `docker cp` + `psql -f` (bukan via docker-entrypoint-initdb.d karena volume sudah ada)
- **bff-service DIIMPLEMENTASI** — 11 endpoint (3 auth, 5 profile, 2 menu, 1 upload), gRPC + HTTP gateway, hand-written protogen, interceptor chain, ServiceConnection, JWT local verify, Azure Blob upload
- **Docker Compose full stack** dibuat di root project (`docker-compose.yml`) — 5 containers (BFF + identity + profile + 2x PostgreSQL)
- **user-profile-service gRPC layer ditambahkan** — proto files, hand-written protogen, `internal/grpchandler/` (6 RPC), `StartGRPC()` di server, port 9302
- **identity-service gRPC server diimplementasi** — `identity_grpc_api.go` (SignUp, SignIn, GetMe via gRPC; tanpa best-effort HTTP ke profile — BFF jadi orchestrator)
- **identity-service unit tests ditambahkan** — 7 file test (api, interceptor, provider, constant, jwt, utils, database); menggunakan `go-sqlmock` + `testify`

## What's Working

### bff-service

- `go build ./server/` — compile pass ✅
- 11 gRPC handlers: SignUp (orchestrated), SignIn, GetMe, GetMyProfile, GetProfileByID, GetProfileByUserID, CreateProfile, UpdateProfile, GetAllMenus, GetMenusByAccountType, Upload (HTTP direct)
- HTTP REST gateway on port 3000 (manual routes, no protoc codegen needed)
- gRPC server on port 9090
- Interceptor chain: ProcessId → Logging → Auth (JWT local verify)
- ServiceConnection: gRPC clients to identity-service (9301) + user-profile-service (9302)
- SignUp orchestration: identity.SignUp → profile.CreateProfile (best-effort)
- Upload image: multipart/form-data → Azure Blob Storage direct
- CORS + security headers middleware
- Docker Compose full stack (`docker-compose.yml` di root): 5 containers

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
- **11 REST endpoint aktif**: Profile CRUD (5), Menu (2), Upload (1), Exchange Rates (1), Interest Rates (1), Branches (1)
- **9 gRPC endpoint aktif**: CreateProfile, GetProfileByID, GetProfileByUserID, UpdateProfile, GetAllMenus, GetMenusByAccountType, GetExchangeRates, GetInterestRates, GetBranches
- Seed data loaded: 1 profile + 9 menu + 4 exchange rates + 4 interest rates + 5 branches
- Business logic menu filter (PREMIUM → semua, REGULAR → hanya REGULAR) terverifikasi
- `StartGRPC()` melayani port 9302 bersamaan dengan REST port 8080
- Swagger docs updated (swagger.json, swagger.yaml, docs.go) dengan 3 endpoint baru tag "Search"
- **Catatan seed data**: `docker-entrypoint-initdb.d` hanya jalan sekali saat volume fresh; untuk re-seed gunakan `docker cp seed.sql container:/tmp/ && docker exec psql -f /tmp/seed.sql`

## Next Steps

1. **Aktifkan gRPC listener identity-service** — `identity_grpc_api.go` sudah ada, perlu pastikan server.go meng-expose port 9301
2. **Docker Compose full stack test** — jalankan `docker compose up --build` dari root, verifikasi semua 5 containers UP
3. **Functional testing BFF** — test 30+ test cases dari testing checklist di `backend-spec-bff-service.md`
4. **Unit tests bff-service** — target ≥ 90% coverage
5. **Unit tests user-profile-service** — belum ada test file
6. **Unit tests identity-service coverage** — verifikasi coverage ≥ 90% (go test -coverprofile)
7. **SonarQube** — static analysis pass untuk semua service

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

- Semua user-profile-service response: raw JSON arrays untuk Search endpoints; `{code, description}` untuk Profile/Menu
- Semua identity-service error response: `{error, code, message}`
- Search endpoints (exchange-rates, interest-rates, branches) — no auth required, public
- Branch search: `ILIKE '%' || $1 || '%'` (case-insensitive partial match)
- Seed data pattern: `docker-entrypoint-initdb.d` hanya jalan saat volume fresh; gunakan `docker cp` + `psql -f` untuk re-seed
- Menu filter: PREMIUM → semua menu, REGULAR → hanya menu REGULAR
- Balance dalam minor unit (cents/pence)
- SignUp orchestration (BFF): identity.SignUp → profile.CreateProfile (best-effort)
- Interceptor chain: ProcessId → Logging → Errors → Auth
- gRPC handler pattern: `var _ pb.XxxServiceServer = (*Server)(nil)` untuk compile-time check
- Unit test pattern identity-service: `newTestServer(t)` → `sqlmock.New()` + `testify/assert`
