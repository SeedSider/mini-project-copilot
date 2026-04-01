---
applyTo: "**"
---

# Active Context

## Current Focus

- **Forgot Password feature SELESAI** ✅ — ValidateOtp + UpdatePassword di identity-service + BFF, end-to-end verified.
- identity-service unit tests: 91.2% coverage (api), 100% (db), 93.3% (jwt) ✅
- identity-service + bff-service RUNNING via Docker Compose ✅ (April 1, 2026)
- Next priority: unit tests bff-service → SonarQube analysis.

## Next Steps

1. Unit tests bff-service (target ≥ 90%)
2. Functional testing BFF — 30+ test cases from checklist
3. SonarQube analysis pass for all services

## Active Decisions

- BFF uses manual REST→gRPC bridge (not grpc-gateway codegen)
- BFF connects to 3 downstream: identity (9301), user-profile (9302), saving (9303)
- JWT secret key same in identity-service and BFF (local verification)
- Upload image: BFF direct to Azure Blob (not via user-profile-service gRPC)
- Docker Compose per service (local dev); full stack compose in bff-service/
- Search endpoints extracted from user-profile to saving-service
- **Forgot Password**: ValidateOtp is public; UpdatePassword is JWT-protected (BFF extracts username from JWT claims, forwards to identity gRPC)
- **OTP**: `crypto/rand` random 6-digit (100000–999999); NOT hardcoded
- **UpdatePassword flow**: Client → BFF (JWT verify, extract username) → identity-service gRPC (UpdatePassword with username + newPassword)

## Important Patterns

- **codec.go WAJIB**: every gRPC service with hand-written protogen MUST have `codec.go` registering JSONCodec via `encoding.RegisterCodec` in `init()`
- **BFF HTTP gateway**: `contextFromHTTPRequest` MUST verify JWT and inject `user_claims`. gRPC interceptor chain does NOT run for direct HTTP gateway calls. Store `jwtMgr` as package-level var.
- **Verify files after `create_file`**: always check `Get-Item <path>` after creating new files
- gRPC handler: `var _ pb.XxxServiceServer = (*Server)(nil)` compile-time check
- Unit test: `newTestServer(t)` → `sqlmock.New()` + `testify/assert`
- Seed data: `docker-entrypoint-initdb.d` only on fresh volume; `docker cp` + `psql -f` for re-seed
- Menu filter: PREMIUM → all, REGULAR → only REGULAR
- SignUp orchestration (BFF): identity.SignUp → profile.CreateProfile (best-effort)
- **Forgot Password unit test**: inject `otpGenerator func() (int, error)` into handler via closure for testability; mock `crypto/rand` errors via generator injection
- **Protected route pattern**: add path to `protectedPaths` map in `bff_authInterceptor.go` AND in identity-service `identity_authInterceptor.go` skip-list for JWT-protected endpoints
