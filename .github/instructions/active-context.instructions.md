---
applyTo: "**"
---

# Active Context

## Current Focus

- All **5 services SELESAI** — compile pass (`go build` + `go vet`), all verified.
- **payment-service IMPLEMENTED** ✅ — 30 new files, 3 endpoints (providers, internet-bill, currency-list)
- **BFF integration with payment-service DONE** ✅ — 4th downstream (gRPC 9304), 3 proxy routes added
- Unit tests user-profile-service ✅ (5 test files)
- Unit tests saving-service ✅ (4 test files)
- Next priority: unit tests payment-service → unit tests bff-service → SonarQube.

## Next Steps

1. **Unit tests payment-service** (target ≥ 90%) — payment_api, payment_grpc_api, bill_provider, currency_provider, interceptor
2. **Unit tests bff-service** (target ≥ 90%) — all proxy handlers including payment
3. Unit tests identity-service coverage verification ≥ 90%
4. Functional testing BFF — 30+ test cases from checklist
5. SonarQube analysis pass for all services
6. Docker Compose full stack test (all 5 services + 4 DBs)

## Active Decisions

- BFF uses manual REST→gRPC bridge (not grpc-gateway codegen)
- BFF connects to **4 downstream**: identity (9301), user-profile (9302), saving (9303), payment (9304)
- JWT secret key same in identity-service, payment-service, and BFF (local verification)
- Upload image: BFF direct to Azure Blob (not via user-profile-service gRPC)
- Docker Compose per service (local dev); full stack compose in bff-service/
- Search endpoints extracted from user-profile to saving-service
- **payment-service**: HTTP 8082, gRPC 9304, PostgreSQL 5435 (DB: payment)
- payment-service auth: hybrid — internet-bill protected (JWT), providers + currency-list public
- payment-service HTTP JWT middleware in main.go (jwtMiddleware wraps HandleGetInternetBill)

## Important Patterns

- **codec.go WAJIB**: every gRPC service with hand-written protogen MUST have `codec.go` registering JSONCodec via `encoding.RegisterCodec` in `init()`
- **BFF HTTP gateway**: `contextFromHTTPRequest` MUST verify JWT and inject `user_claims`. gRPC interceptor chain does NOT run for direct HTTP gateway calls. Store `jwtMgr` as package-level var.
- **Verify files after `create_file`**: always check `Get-Item <path>` after creating new files
- gRPC handler: `var _ pb.XxxServiceServer = (*Server)(nil)` compile-time check
- Unit test: `newTestServer(t)` → `sqlmock.New()` + `testify/assert`
- Seed data: `docker-entrypoint-initdb.d` only on fresh volume; `docker cp` + `psql -f` for re-seed
- Menu filter: PREMIUM → all, REGULAR → only REGULAR
- SignUp orchestration (BFF): identity.SignUp → profile.CreateProfile (best-effort)
- **payment-service pattern**: HTTP jwtMiddleware in main.go extracts claims → context.WithValue("user_claims") → handler reads from ctx
- **BFF ServiceConnection**: InitServicesConn now takes 4 addrs (identity, profile, saving, payment)
