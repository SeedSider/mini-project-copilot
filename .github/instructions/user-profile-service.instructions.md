---
applyTo: "user-profile-service/**"
---

# user-profile-service

Manages user profiles, homepage menus, and image upload. Refactored to `server/` pattern (same as identity-service).

## Ports

- HTTP: 8080 (chi router)
- gRPC: 9302

## Dependencies

- `github.com/go-chi/chi/v5`: HTTP router + middleware
- `database/sql` + `github.com/lib/pq`: PostgreSQL
- `github.com/joho/godotenv`: .env loading
- `github.com/dgrijalva/jwt-go`: JWT parsing (GetMyProfile)
- `github.com/swaggo/http-swagger`: Swagger UI
- `google.golang.org/grpc`: gRPC server

## Folder Structure

```
user-profile-service/
├── server/
│   ├── main.go                    # Entry + chi router + gRPC server + graceful shutdown
│   ├── core_config.go             # Config struct + initConfig
│   ├── core_db.go                 # DB connection + runMigration (embed.FS)
│   ├── api/
│   │   ├── api.go                 # Server struct + pb.UserProfileServiceServer check
│   │   ├── profile_auth_api.go    # HTTP: GetMyProfile (JWT), GetProfile, UpdateProfile, CreateProfile, GetProfileByUserID
│   │   ├── profile_grpc_api.go    # gRPC: CreateProfile, GetProfileByID, GetProfileByUserID, UpdateProfile
│   │   ├── menu_api.go            # HTTP: GetAllMenus, GetMenusByAccountType
│   │   ├── menu_grpc_api.go       # gRPC: GetAllMenus, GetMenusByAccountType
│   │   ├── upload_api.go          # HTTP: UploadImage (Azure Blob)
│   │   ├── converter.go           # Model ↔ proto conversion helpers
│   │   └── error.go               # writeJSON, writeError, StandardResponse
│   ├── db/
│   │   ├── provider.go            # Provider struct + constructor
│   │   ├── profile_provider.go    # GetByID, GetByUserID, Create, Update + domain types
│   │   └── menu_provider.go       # GetAll, GetByAccountType + Menu/MenuResponse types
│   ├── constant/ + utils/
├── migrations/
│   ├── embed.go                   # embed.FS
│   ├── 001_init.sql               # profile + menu tables
│   ├── 002_add_image_to_profile.sql
│   └── 003_add_user_id_to_profile.sql
├── proto/ + protogen/user-profile-service/ (hand-written + codec.go)
├── docs/                          # Swagger generated
├── Dockerfile                     # golang:1.24-alpine → alpine:3.20
├── docker-compose.yml / docker-compose.local.yml
└── seed.sql                       # 1 profile + 9 menu items
```

## Database Schema (bankease_db)

```sql
CREATE TABLE IF NOT EXISTS profile (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id         UUID UNIQUE,
    bank            VARCHAR(50)  NOT NULL,
    branch          VARCHAR(50)  NOT NULL,
    name            VARCHAR(100) NOT NULL,
    card_number     VARCHAR(20)  NOT NULL,
    card_provider   VARCHAR(50)  NOT NULL,
    balance         BIGINT       NOT NULL DEFAULT 0,
    currency        VARCHAR(3)   NOT NULL DEFAULT 'IDR',
    account_type    VARCHAR(20)  NOT NULL DEFAULT 'REGULAR',
    image           TEXT         DEFAULT ''
);

CREATE TABLE IF NOT EXISTS menu (
    id        VARCHAR(50)  PRIMARY KEY,
    "index"   INTEGER      UNIQUE NOT NULL,
    type      VARCHAR(20)  NOT NULL,
    title     VARCHAR(50)  NOT NULL,
    icon_url  TEXT         NOT NULL,
    is_active BOOLEAN      NOT NULL DEFAULT TRUE
);
```

## API Endpoints (8 REST + 6 gRPC)

```
GET  /api/profile            → GetMyProfile (JWT required)
POST /api/profile            → CreateProfile
GET  /api/profile/{id}       → GetProfileByID
PUT  /api/profile/{id}       → UpdateProfile
GET  /api/profile/user/{uid} → GetProfileByUserID
GET  /api/menu               → GetAllMenus
GET  /api/menu/{accountType} → GetMenusByAccountType
POST /api/upload/image       → UploadImage (Azure Blob)
```

## Key Patterns

- Provider pattern: `server/db/provider.go` + domain-specific providers
- Compile-time check: `var _ pb.UserProfileServiceServer = (*Server)(nil)`
- Converter: `server/api/converter.go` for model ↔ proto conversion
- Response format: `{ "code": 200, "description": "Success" }`
- Menu filter logic: PREMIUM → all menus, REGULAR → only REGULAR
- Search endpoints extracted to saving-service
