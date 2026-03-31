---
applyTo: "bff-service/**"
---

# bff-service

Backend for Frontend — single REST entry point for mobile app. Orchestrates calls to 3 downstream services via gRPC.

## Ports

- HTTP: 3000 (manual REST→gRPC bridge)
- gRPC: 9090

## Dependencies

- `google.golang.org/grpc`: gRPC server + clients
- `grpc-gateway/v2`: REST → gRPC gateway
- `dgrijalva/jwt-go`: JWT verify (local, same secret as identity-service)
- `urfave/cli`: CLI commands
- `swaggo/swag@v1.16.6` + `swaggo/http-swagger@v1.3.4`: Swagger
- Zap logger + FluentBit

## Downstream Services

- identity-service: gRPC port 9301 (config: `IdentityServiceAddr`)
- user-profile-service: gRPC port 9302 (config: `UserProfileServiceAddr`)
- saving-service: gRPC port 9303 (config: `SavingServiceAddr`)

## Folder Structure

```
bff-service/
├── server/
│   ├── main.go                     # Entry + CLI (grpc-server, gw-server, grpc-gw-server) + Swagger UI
│   ├── core_config.go              # Config loader (incl. 3 service addrs)
│   ├── gateway_http_handler.go     # HTTP gateway + upload handler + CORS
│   ├── http_routes.go              # REST→gRPC bridge handlers + swagger annotations
│   ├── swagger_docs.go             # Doc stubs for multi-method endpoints
│   ├── api/
│   │   ├── api.go                  # Server struct + DI
│   │   ├── bff_auth_api.go         # SignUp (orchestrated), SignIn, GetMe
│   │   ├── bff_profile_api.go      # GetMyProfile, GetProfileByID, GetProfileByUserID, CreateProfile, UpdateProfile
│   │   ├── bff_menu_api.go         # GetAllMenus, GetMenusByAccountType
│   │   ├── bff_saving_api.go       # GetExchangeRates, GetInterestRates, GetBranches
│   │   ├── bff_interceptor.go      # Chain: ProcessId → Logging → Auth
│   │   ├── bff_authInterceptor.go  # JWT verify for protected endpoints
│   │   └── error.go
│   ├── services/service.go         # ServiceConnection (3 gRPC clients)
│   ├── jwt/manager.go              # JWT Verify only (local HS256)
│   ├── lib/logger/                 # Zap logger
│   └── utils/, constant/
├── proto/ + protogen/ (BFF server + identity/profile/saving clients)
├── docs/                           # Swagger generated (docs.go, swagger.json, swagger.yaml)
├── Dockerfile                      # golang:1.24-alpine → alpine:3.20
└── docker-compose.yml              # Full stack (5 containers)
```

## All 14 REST Endpoints

```
POST /api/auth/signup        → orchestrated: identity.SignUp + profile.CreateProfile (best-effort)
POST /api/auth/signin        → proxy to identity.SignIn
GET  /api/auth/me            → proxy to identity.GetMe (JWT required)
GET  /api/profile            → JWT → profile.GetMyProfile
POST /api/profile            → profile.CreateProfile
GET  /api/profile/{id}       → profile.GetProfileByID
PUT  /api/profile/{id}       → profile.UpdateProfile
GET  /api/profile/user/{uid} → profile.GetProfileByUserID
GET  /api/menu               → profile.GetAllMenus
GET  /api/menu/{accountType} → profile.GetMenusByAccountType
POST /api/upload/image       → BFF direct to Azure Blob (multipart/form-data)
GET  /api/exchange-rates     → saving.GetExchangeRates (public)
GET  /api/interest-rates     → saving.GetInterestRates (public)
GET  /api/branches?q=        → saving.GetBranches (public)
```

## Key Patterns

- **Manual REST→gRPC bridge** (NOT grpc-gateway codegen) — routes manually defined in `http_routes.go`
- **`contextFromHTTPRequest` MUST verify JWT** and inject `user_claims` into context — gRPC interceptor chain does NOT run for direct function calls from HTTP gateway
- **`jwtMgr`** stored as package-level var, initialized in `startHTTPServer`
- **SignUp orchestration**: identity.SignUp → profile.CreateProfile (best-effort, profile failure doesn't fail signup)
- **Upload**: multipart/form-data → Azure Blob Storage direct (not via gRPC)
- **Swagger**: swaggo annotations on http_routes.go; doc stubs in swagger_docs.go for multi-method handlers. BearerAuth security definition. UI at `/swagger/bff/`
- **Interceptor chain**: ProcessId → Logging → Auth (JWT local verify)
- **CORS + security headers** middleware in gateway_http_handler.go
