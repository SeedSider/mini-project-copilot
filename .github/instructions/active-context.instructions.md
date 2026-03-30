---
applyTo: "**"
---

# Active Context

## Current Focus

- `identity-service` — **SELESAI** (Tahap 1-3 + unit tests + gRPC server + codec fix). Dikembangkan di project terpisah (`mini-project-copilot-identity/`), kemudian disatukan ke monorepo ini.
- `user-profile-service` — **SELESAI + REFACTORED**. Folder structure di-refactor agar sama dengan identity-service: `cmd/server/` + `internal/` → `server/` (main.go, core_config.go, core_db.go, api/, db/, constant/, utils/). `migrations/` dipindah ke root. Semua 11 endpoint (REST + gRPC) tetap aktif, compile pass.
- `bff-service` — **SELESAI + RUNNING + SWAGGER**. Compile pass, semua 11 endpoint diimplementasi. JWT auth fix applied. Swagger UI di `/swagger/bff/`. Docker Compose full stack UP dan terverifikasi end-to-end.

## Recent Changes

- **bff-service SWAGGER DITAMBAHKAN** — Swagger UI + OpenAPI spec untuk semua 11 endpoint:
  - Dependencies: `swaggo/swag@v1.16.6` + `swaggo/http-swagger@v1.3.4`
  - Generated docs: `docs/docs.go`, `docs/swagger.json`, `docs/swagger.yaml`
  - Swagger UI route: `http://localhost:3000/swagger/bff/`
  - Annotations pada `server/http_routes.go` (Auth, Menu, Profile/user/{id})
  - Doc stubs pada `server/swagger_docs.go` (Profile GET/POST, Profile/{id} GET/PUT, Upload)
  - Type reference import via `bff_service` alias di swagger_docs.go, `pb` alias di http_routes.go
  - `@securityDefinitions.apikey BearerAuth` untuk protected endpoints
  - 11 operasi di 5 tags: Auth (3), Profile (5), Menu (2), Upload (1)
  - `go build ./server/` + `go vet ./server/` — pass ✅
- **user-profile-service REFACTORED** — folder structure dirombak agar konsisten dengan identity-service:
  - `cmd/server/main.go` → `server/main.go` (dengan graceful shutdown, signal handling)
  - `internal/handlers/` + `internal/grpchandler/` → `server/api/` (semua HTTP + gRPC handlers di `*Server` struct)
  - `internal/repository/` → `server/db/` (Provider pattern; profile_provider.go, menu_provider.go, search_provider.go)
  - `internal/models/` → types di-embed ke `server/db/` (domain) dan `server/api/` (response wrappers)
  - `internal/server/` (router.go, server.go) → di-absorb ke `server/main.go`
  - `internal/db/` (db.go, migrate.go, migrations/) → `server/core_db.go` + `migrations/embed.go`
  - Ditambahkan: `server/core_config.go`, `server/constant/`, `server/utils/`
  - Dockerfile diperbarui: build `./server` (bukan `./cmd/server`)
  - `go build ./server/` dan BFF cross-compile **pass** ✅
  - Direktori lama `cmd/` dan `internal/` sudah dihapus
- **Docker Compose full stack UP & VERIFIED** — semua 5 containers running, end-to-end flow terverifikasi: SignUp → SignIn → GET /api/profile
- **BFF JWT fix** — `contextFromHTTPRequest` sekarang verify JWT dan inject `user_claims` ke context langsung
- **identity-service codec.go DIBUAT** — `protogen/identity-service/codec.go` yang mendaftarkan JSONCodec via `init()`

## What's Working

### bff-service

- `go build ./server/` — compile pass ✅
- **Docker container RUNNING** di port 3000 (HTTP) + 9090 (gRPC) ✅
- 11 gRPC handlers: SignUp (orchestrated), SignIn, GetMe, GetMyProfile, GetProfileByID, GetProfileByUserID, CreateProfile, UpdateProfile, GetAllMenus, GetMenusByAccountType, Upload (HTTP direct)
- HTTP REST gateway on port 3000 (manual routes, no protoc codegen needed)
- gRPC server on port 9090
- Interceptor chain: ProcessId → Logging → Auth (JWT local verify)
- ServiceConnection: gRPC clients to identity-service (9301) + user-profile-service (9302)
- SignUp orchestration: identity.SignUp → profile.CreateProfile (best-effort) ✅ VERIFIED
- GET /api/profile with JWT token ✅ VERIFIED
- Upload image: multipart/form-data → Azure Blob Storage direct
- **Swagger UI** di `/swagger/bff/` — 11 operasi, 5 tags (Auth, Profile, Menu, Upload), BearerAuth security definition
- CORS + security headers middleware
- Docker Compose full stack (`docker-compose.yml` di root): 5 containers

### identity-service

- Docker Compose running (PostgreSQL 17 + identity-service)
- 4 HTTP endpoint aktif: POST /api/auth/signup, POST /api/auth/signin, GET /api/identity/me, GET /health
- **gRPC server AKTIF** di port 9301: SignUp, SignIn, GetMe (`identity_grpc_api.go`) ✅
- **`protogen/identity-service/codec.go`** — JSONCodec registered via `init()`, gRPC server bisa handle JSON-encoded requests ✅
- JWT HS256 auth, bcrypt password hashing
- Interceptor chain: ProcessId → Logging → Errors → Auth
- Functional test 12/12 lulus (signup, signin, getme, error cases)
- Unit tests coverage meningkat: 7 test file ditambahkan

### user-profile-service

- **REFACTORED** ke identity-service folder structure pattern ✅
- Folder structure: `server/` (main.go, core_config.go, core_db.go, api/, db/, constant/, utils/), `migrations/`, `proto/`, `protogen/`
- Docker Compose running (PostgreSQL 17 + app di localhost:8080, gRPC 9302)
- **11 REST endpoint aktif**: Profile CRUD (5), Menu (2), Upload (1), Exchange Rates (1), Interest Rates (1), Branches (1)
- **9 gRPC endpoint aktif**: CreateProfile, GetProfileByID, GetProfileByUserID, UpdateProfile, GetAllMenus, GetMenusByAccountType, GetExchangeRates, GetInterestRates, GetBranches
- `server/api/` — semua HTTP + gRPC handlers pada `*Server` struct (`var _ pb.UserProfileServiceServer = (*Server)(nil)` compile-time check)
- `server/db/` — Provider pattern (provider.go, profile_provider.go, menu_provider.go, search_provider.go)
- `migrations/` — di root, dengan `embed.go` untuk embed.FS
- Seed data loaded: 1 profile + 9 menu + 4 exchange rates + 4 interest rates + 5 branches
- Business logic menu filter (PREMIUM → semua, REGULAR → hanya REGULAR) terverifikasi
- Swagger docs (docs/) tetap ada
- **Catatan seed data**: `docker-entrypoint-initdb.d` hanya jalan sekali saat volume fresh; untuk re-seed gunakan `docker cp seed.sql container:/tmp/ && docker exec psql -f /tmp/seed.sql`

## Next Steps

1. **Functional testing BFF (lanjutan)** — test 30+ test cases dari testing checklist di `backend-spec-bff-service.md`
2. **Unit tests bff-service** — target ≥ 90% coverage
3. **Unit tests user-profile-service** — belum ada test file
4. **Unit tests identity-service coverage** — verifikasi coverage ≥ 90% (go test -coverprofile)
5. **SonarQube** — static analysis pass untuk semua service

## Active Decisions

- identity-service: HTTP handler aktif (port 3031), gRPC aktif (port 9301) ✅
- user-profile-service: REST (chi, port 8080) + gRPC (port 9302) berjalan bersamaan ✅ — **refactored to `server/` pattern**
- BFF menggunakan manual REST→gRPC bridge (bukan grpc-gateway codegen)
- JWT secret key harus sama di identity-service dan BFF (lokal verification)
- Upload image di BFF langsung ke Azure Blob (tidak lewat user-profile-service)
- Docker Compose full stack: 5 containers (BFF + identity + profile + 2x PostgreSQL) ✅ RUNNING
- identity-service module path tetap `bitbucket.bri.co.id/scm/addons/addons-identity-service`
- identity-service gRPC SignUp TIDAK melakukan best-effort HTTP ke profile — BFF yang meng-orchestrate
- BFF `contextFromHTTPRequest` HARUS verify JWT dan inject `user_claims` karena HTTP gateway tidak lewat gRPC transport layer

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
- **WAJIB: setiap gRPC service yg menggunakan hand-written protogen HARUS punya `codec.go`** di package protogennya yang mendaftarkan JSONCodec via `encoding.RegisterCodec` dalam `init()`. Tanpa ini, server gRPC akan fallback ke proto codec dan gagal unmarshal.
- **BFF HTTP gateway pattern**: `contextFromHTTPRequest` HARUS verify JWT dan inject `user_claims` ke context. gRPC interceptor chain TIDAK berjalan untuk direct function calls dari HTTP gateway. Simpan `jwtMgr` sebagai package-level var, inisialisasi di `startHTTPServer`.
- **Verifikasi file.go setelah `create_file`**: selalu cek `Get-Item <path>` setelah membuat file baru untuk memastikan file benar-benar ada di disk.
- **Swagger pattern BFF**: swaggo annotations di HTTP handler files. Untuk multi-method handlers (handleProfile, handleProfileByID), gunakan doc stub functions di `swagger_docs.go`. Type references harus sesuai import alias (`pb.*` di http_routes.go, `bff_service.*` di swagger_docs.go). Swagger UI route: `http.StripPrefix` + `httpSwagger.Handler(httpSwagger.URL(...))`.
