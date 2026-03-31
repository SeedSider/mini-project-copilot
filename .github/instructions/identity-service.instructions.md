---
applyTo: "identity-service/**"
---

# identity-service

Module path: `bitbucket.bri.co.id/scm/addons/addons-identity-service`
Origin: developed in separate project (`mini-project-copilot-identity/`), merged into monorepo.

## Ports

- HTTP: 3031
- gRPC: 9301

## Dependencies

- Go stdlib `net/http` + CORS middleware manual
- `dgrijalva/jwt-go`: JWT HS256 parsing & validation
- `golang.org/x/crypto/bcrypt`: Password hashing
- Custom database wrapper: `server/lib/database/` (connect, retry, transaction)
- Zap logger + FluentBit: `server/lib/logger/`
- `urfave/cli`: CLI commands (grpc-server, gw-server, grpc-gw-server)
- `google/uuid`: UUID generation

## Folder Structure

```
identity-service/
├── server/
│   ├── main.go                         # Entry point + CLI + HTTP server + auto-migration
│   ├── core_config.go                  # Config loader (env vars + .env)
│   ├── core_db.go                      # DB connection lifecycle (retry logic)
│   ├── api/
│   │   ├── api.go                      # Server struct + constructor
│   │   ├── identity_auth_api.go        # Handler: SignUp, SignIn, GetMe + HTTP handlers
│   │   ├── identity_auth_api_test.go   # Unit tests
│   │   ├── identity_grpc_api.go        # gRPC handler: SignUp, SignIn, GetMe
│   │   ├── identity_authInterceptor.go # JWT auth interceptor
│   │   ├── identity_interceptor.go     # Chain: ProcessId → Logging → Errors → Auth
│   │   ├── identity_interceptor_test.go
│   │   └── error.go                    # Error helpers
│   ├── db/
│   │   ├── provider.go                 # Provider struct + constructor
│   │   ├── identity_provider.go        # CreateUser, GetUserByUsername, CheckUsernameExists
│   │   ├── identity_provider_test.go
│   │   └── error.go                    # NotFoundErr type
│   ├── jwt/
│   │   ├── manager.go                  # JWT Generate + Verify (HS256)
│   │   └── manager_test.go
│   ├── lib/
│   │   ├── database/                   # DB wrapper (connect, retry, interface, mock)
│   │   └── logger/                     # Zap structured logger + FluentBit
│   ├── utils/ + constant/
├── migrations/
│   ├── 001_init.sql                    # DDL: users table
│   └── 002_rename_email_to_username.sql
├── proto/
│   ├── identity_api.proto
│   └── identity_payload.proto
├── protogen/identity-service/          # Hand-written gRPC stubs + codec.go
├── Dockerfile                          # golang:1.24-alpine → alpine:3.19
├── docker-compose.yml / docker-compose.local.yml
├── Makefile, sonar-project.properties, .env.example
└── www/swagger.json
```

## Database Schema (identity_db)

```sql
CREATE TABLE users (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    username VARCHAR(255) UNIQUE NOT NULL,
    password_hash TEXT NOT NULL,
    phone VARCHAR(50),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
```

## API Endpoints

```
POST /api/auth/signup    → SignUp (validate, bcrypt, INSERT)
POST /api/auth/signin    → SignIn (bcrypt compare, JWT generate)
GET  /api/identity/me    → GetMe (JWT verify, get user)
GET  /health             → Health check
```

## Key Patterns

- Interceptor chain: ProcessId → Logging → Errors → Auth
- gRPC handler: `var _ pb.IdentityServiceServer = (*Server)(nil)` compile-time check
- gRPC SignUp does NOT do best-effort HTTP to profile — BFF orchestrates
- Error response: `{ "error": true, "code": 401, "message": "Unauthorized" }`
- Unit test pattern: `newTestServer(t)` → `sqlmock.New()` + `testify/assert`
