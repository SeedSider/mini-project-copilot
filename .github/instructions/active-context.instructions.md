---
applyTo: "**"
---

# Active Context

## Current Focus

- `identity-service` — **SELESAI** (Tahap 1-3 + unit tests + gRPC server + codec fix). Dikembangkan di project terpisah (`mini-project-copilot-identity/`), kemudian disatukan ke monorepo ini.
- `user-profile-service` — **SELESAI + REFACTORED**. Search endpoints (exchange-rates, interest-rates, branches) telah diekstrak ke `saving-service`. Sekarang hanya menangani Profile CRUD (5 endpoint) + Menu (2 endpoint) + Upload (1 endpoint) = 8 endpoint REST + 6 gRPC.
- `saving-service` — **SELESAI + RUNNING**. Service baru hasil ekstraksi search feature dari user-profile-service. 3 endpoint REST + 3 gRPC (exchange-rates, interest-rates, branches). Compile pass, Docker Compose running, seed data loaded.
- `bff-service` — **SELESAI + RUNNING + SWAGGER**. Compile pass, semua 14 endpoint diimplementasi (11 original + 3 saving proxy). JWT auth fix applied. Swagger UI di `/swagger/bff/`. BFF sekarang terhubung ke 3 downstream services: identity (9301), user-profile (9302), saving (9303).

## Recent Changes

- **saving-service DIBUAT** — Service baru untuk financial data (exchange rates, interest rates, branches):
  - Diekstrak dari user-profile-service search feature menjadi service mandiri
  - Folder structure mengikuti BRICaMS pattern: `server/` (main.go, core_config.go, core_db.go, api/, db/, constant/, utils/, lib/)
  - 3 REST endpoint: GET /api/exchange-rates, GET /api/interest-rates, GET /api/branches?q=
  - 3 gRPC RPC: GetExchangeRates, GetInterestRates, GetBranches
  - PostgreSQL database `saving` (port 5434) dengan 3 tabel: exchange_rate, interest_rate, branch
  - Proto definitions: `saving_api.proto`, `saving_payload.proto`
  - Hand-written protogen + codec.go (JSONCodec)
  - Swagger docs + UI di `/swagger/`
  - Docker Compose: PostgreSQL 17 + saving-service (HTTP 8081, gRPC 9303)
  - Seed data loaded: 4 exchange rates, 4 interest rates, 5 branches
  - `go build ./server/` + `go vet ./server/` — pass ✅
  - CLI: `urfave/cli` dengan commands grpc-server, gw-server, grpc-gw-server
- **BFF saving-service INTEGRATION** — BFF terhubung ke saving-service:
  - `protogen/saving-service/saving_api_grpc.pb.go` — gRPC client stub
  - `server/services/service.go` — SavingService gRPC client (port 9303)
  - `server/api/bff_saving_api.go` — GetExchangeRates, GetInterestRates, GetBranches handlers
  - `server/http_routes.go` — REST routes: GET /api/exchange-rates, /api/interest-rates, /api/branches
  - `server/core_config.go` — `SavingServiceAddr` config (default: localhost:9303)
- **bff-service SWAGGER** — Swagger UI + OpenAPI spec:
  - Swagger UI route: `http://localhost:3000/swagger/bff/`
  - BearerAuth security definition untuk protected endpoints
  - `go build ./server/` + `go vet ./server/` — pass ✅

## What's Working

### bff-service

- `go build ./server/` — compile pass ✅
- **Docker container RUNNING** di port 3000 (HTTP) + 9090 (gRPC) ✅
- 14 gRPC handlers: SignUp (orchestrated), SignIn, GetMe, GetMyProfile, GetProfileByID, GetProfileByUserID, CreateProfile, UpdateProfile, GetAllMenus, GetMenusByAccountType, GetExchangeRates, GetInterestRates, GetBranches, Upload (HTTP direct)
- HTTP REST gateway on port 3000 (manual routes, no protoc codegen needed)
- gRPC server on port 9090
- Interceptor chain: ProcessId → Logging → Auth (JWT local verify)
- ServiceConnection: gRPC clients to identity-service (9301) + user-profile-service (9302) + saving-service (9303)
- SignUp orchestration: identity.SignUp → profile.CreateProfile (best-effort) ✅ VERIFIED
- GET /api/profile with JWT token ✅ VERIFIED
- Upload image: multipart/form-data → Azure Blob Storage direct
- **Swagger UI** di `/swagger/bff/` — BearerAuth security definition
- CORS + security headers middleware

### saving-service

- `go build ./server/` + `go vet ./server/` — compile pass ✅
- **Docker Compose RUNNING** — PostgreSQL 17 (port 5434) + saving-service (HTTP 8081, gRPC 9303)
- 3 REST endpoint aktif: GET /api/exchange-rates, GET /api/interest-rates, GET /api/branches?q=
- 3 gRPC RPC aktif: GetExchangeRates, GetInterestRates, GetBranches
- Health check: GET /health
- Swagger UI di `/swagger/`
- Seed data loaded: 4 exchange rates + 4 interest rates + 5 branches
- Branch search: ILIKE case-insensitive partial match
- Database `saving`: 3 tabel (exchange_rate, interest_rate, branch)
- CLI: `urfave/cli` (grpc-server, gw-server, grpc-gw-server)

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
- **8 REST endpoint aktif**: Profile CRUD (5), Menu (2), Upload (1)
- **6 gRPC endpoint aktif**: CreateProfile, GetProfileByID, GetProfileByUserID, UpdateProfile, GetAllMenus, GetMenusByAccountType
- **Search endpoints diekstrak ke saving-service** — exchange-rates, interest-rates, branches sudah dipindah
- `server/api/` — semua HTTP + gRPC handlers pada `*Server` struct (`var _ pb.UserProfileServiceServer = (*Server)(nil)` compile-time check)
- `server/db/` — Provider pattern (provider.go, profile_provider.go, menu_provider.go)
- `migrations/` — di root, dengan `embed.go` untuk embed.FS
- Seed data loaded: 1 profile + 9 menu
- Business logic menu filter (PREMIUM → semua, REGULAR → hanya REGULAR) terverifikasi
- Swagger docs (docs/) tetap ada

## Next Steps

1. **Functional testing BFF (lanjutan)** — test 30+ test cases dari testing checklist di `backend-spec-bff-service.md`
2. **Unit tests bff-service** — target ≥ 90% coverage
3. **Unit tests user-profile-service** — belum ada test file
4. **Unit tests saving-service** — belum ada test file
5. **Unit tests identity-service coverage** — verifikasi coverage ≥ 90% (go test -coverprofile)
6. **SonarQube** — static analysis pass untuk semua service

## Active Decisions

- identity-service: HTTP handler aktif (port 3031), gRPC aktif (port 9301) ✅
- user-profile-service: REST (chi, port 8080) + gRPC (port 9302) berjalan bersamaan ✅ — **refactored to `server/` pattern**
- saving-service: REST (net/http, port 8081) + gRPC (port 9303) — **service baru, standalone**
- BFF menggunakan manual REST→gRPC bridge (bukan grpc-gateway codegen)
- BFF terhubung ke 3 downstream services: identity (9301), user-profile (9302), saving (9303)
- JWT secret key harus sama di identity-service dan BFF (lokal verification)
- Upload image di BFF langsung ke Azure Blob (tidak lewat user-profile-service)
- Docker Compose per service (local dev); full stack compose di bff-service/docker-compose.yml
- identity-service module path tetap `bitbucket.bri.co.id/scm/addons/addons-identity-service`
- identity-service gRPC SignUp TIDAK melakukan best-effort HTTP ke profile — BFF yang meng-orchestrate
- BFF `contextFromHTTPRequest` HARUS verify JWT dan inject `user_claims` karena HTTP gateway tidak lewat gRPC transport layer
- Search endpoints (exchange-rates, interest-rates, branches) diekstrak dari user-profile ke saving-service

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
