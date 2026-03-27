---
applyTo: "**"
---

# Active Context

## Current Focus

- `identity-service` — **SELESAI** (Tahap 1-3 + functional testing 12/12 lulus). Dikembangkan di project terpisah (`mini-project-copilot-identity/`), kemudian disatukan ke monorepo ini.
- `user-profile-service` — **SELESAI** dan terverifikasi. Docker Compose running, semua endpoint tested.
- `bff-service` — **SPEC SELESAI** (`backend-spec-bff-service.md`), belum implementasi.

## Recent Changes

- identity-service disatukan dari project terpisah ke monorepo `mini-project-copilot/identity-service/`
- Backend spec BFF service ditulis lengkap (`backend-spec-bff-service.md`) — 11 endpoint, proto definitions, orchestration flows, Docker Compose full stack
- user-profile-service sudah ditambah: CreateProfile, GetProfileByUserID, GetMyProfile (JWT), Upload Image (Azure Blob)
- Memory bank files `.identity-service/instructions/` tersedia sebagai referensi history identity-service

## What's Working

### identity-service

- Docker Compose running (PostgreSQL 15 + identity-service)
- 4 endpoint aktif: POST /api/auth/signup, POST /api/auth/signin, GET /api/identity/me, GET /health
- JWT HS256 auth, bcrypt password hashing
- Interceptor chain: ProcessId → Logging → Errors → Auth
- Functional test 12/12 lulus (signup, signin, getme, error cases)
- Cross-service: SignUp → best-effort HTTP call ke user-profile-service/CreateProfile

### user-profile-service

- Docker Compose running (PostgreSQL 17 + app di localhost:8080)
- 8 endpoint aktif: Profile CRUD (5), Menu (2), Upload (1)
- Seed data auto-loaded (1 profile + 9 menu items)
- Business logic menu filter (PREMIUM → semua, REGULAR → hanya REGULAR) terverifikasi

## Next Steps

1. **Implementasi bff-service** — berdasarkan `backend-spec-bff-service.md`
2. **Tambah gRPC ke user-profile-service** — saat ini hanya REST, perlu gRPC layer agar bisa dipanggil BFF
3. **Tambah gRPC server ke identity-service** — saat ini hanya HTTP, perlu gRPC server aktif
4. **Unit tests** — target coverage ≥ 90% untuk semua service
5. **SonarQube** — static analysis pass untuk semua service

## Active Decisions

- identity-service menggunakan `net/http` (bukan grpc-gateway aktif) — CLI ready tapi HTTP handler manual
- user-profile-service menggunakan `chi` router — REST only, perlu ditambah gRPC
- BFF akan menggunakan grpc-gateway pattern (sama dengan addons-issuance-lc-service)
- JWT secret key harus sama di identity-service dan BFF (lokal verification)
- Upload image di BFF langsung ke Azure Blob (tidak lewat user-profile-service)
- Docker Compose full stack: 5 containers (BFF + identity + profile + 2x PostgreSQL)
- identity-service module path tetap `bitbucket.bri.co.id/scm/addons/addons-identity-service`

## Important Patterns

- Semua user-profile-service response: `{code, description}`
- Semua identity-service error response: `{error, code, message}`
- Menu filter: PREMIUM → semua menu, REGULAR → hanya menu REGULAR
- Balance dalam minor unit (cents/pence)
- SignUp orchestration (BFF): identity.SignUp → profile.CreateProfile (best-effort)
- Interceptor chain: ProcessId → Logging → Errors → Auth
