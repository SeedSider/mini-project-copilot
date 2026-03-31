---
applyTo: "**"
---

# Active Context

## Current Focus

- All 4 services **SELESAI** — compile pass, Docker Compose running, verified end-to-end.
- Next priority: unit tests (bff, user-profile, saving) → SonarQube.

## Next Steps

1. Unit tests bff-service (target ≥ 90%)
2. Unit tests user-profile-service (target ≥ 90%)
3. Unit tests saving-service (target ≥ 90%)
4. Unit tests identity-service coverage verification ≥ 90%
5. Functional testing BFF — 30+ test cases from checklist
6. SonarQube analysis pass for all services

## Active Decisions

- BFF uses manual REST→gRPC bridge (not grpc-gateway codegen)
- BFF connects to 3 downstream: identity (9301), user-profile (9302), saving (9303)
- JWT secret key same in identity-service and BFF (local verification)
- Upload image: BFF direct to Azure Blob (not via user-profile-service gRPC)
- Docker Compose per service (local dev); full stack compose in bff-service/
- Search endpoints extracted from user-profile to saving-service

## Important Patterns

- **codec.go WAJIB**: every gRPC service with hand-written protogen MUST have `codec.go` registering JSONCodec via `encoding.RegisterCodec` in `init()`
- **BFF HTTP gateway**: `contextFromHTTPRequest` MUST verify JWT and inject `user_claims`. gRPC interceptor chain does NOT run for direct HTTP gateway calls. Store `jwtMgr` as package-level var.
- **Verify files after `create_file`**: always check `Get-Item <path>` after creating new files
- gRPC handler: `var _ pb.XxxServiceServer = (*Server)(nil)` compile-time check
- Unit test: `newTestServer(t)` → `sqlmock.New()` + `testify/assert`
- Seed data: `docker-entrypoint-initdb.d` only on fresh volume; `docker cp` + `psql -f` for re-seed
- Menu filter: PREMIUM → all, REGULAR → only REGULAR
- SignUp orchestration (BFF): identity.SignUp → profile.CreateProfile (best-effort)
