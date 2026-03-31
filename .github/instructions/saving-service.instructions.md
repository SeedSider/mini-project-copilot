---
applyTo: "saving-service/**"
---

# saving-service

Financial data service — exchange rates, interest rates, branch locations. Extracted from user-profile-service search feature.

## Ports

- HTTP: 8081 (`net/http`)
- gRPC: 9303

## Dependencies

- Go stdlib `net/http`: HTTP server + router
- `database/sql` + `github.com/lib/pq`: PostgreSQL
- `github.com/joho/godotenv`: .env loading
- `google.golang.org/grpc`: gRPC server
- `urfave/cli`: CLI commands (grpc-server, gw-server, grpc-gw-server)
- Zap logger + FluentBit: `server/lib/logger/`
- `swaggo/http-swagger`: Swagger UI
- `google/uuid`: UUID generation

## Folder Structure

```
saving-service/
├── server/
│   ├── main.go                      # Entry + CLI + HTTP/gRPC servers
│   ├── core_config.go               # Config loader
│   ├── core_db.go                   # DB connection + migrations (embed.FS)
│   ├── api/
│   │   ├── api.go                   # Server struct + constructor
│   │   ├── saving_api.go            # HTTP: GetExchangeRates, GetInterestRates, GetBranches
│   │   ├── saving_grpc_api.go       # gRPC handlers
│   │   └── error.go                 # writeJSON, writeError
│   ├── db/
│   │   ├── provider.go              # Provider struct
│   │   ├── exchange_rate_provider.go # GetAllExchangeRates
│   │   ├── interest_rate_provider.go # GetAllInterestRates
│   │   └── branch_provider.go       # GetAllBranches, SearchBranches (ILIKE)
│   ├── constant/ + utils/ + lib/logger/
├── migrations/
│   ├── embed.go
│   ├── 001_add_exchange_rates.sql
│   ├── 002_add_interest_rates.sql
│   └── 003_add_branches.sql
├── proto/ + protogen/saving-service/ (hand-written + codec.go)
├── docs/                            # Swagger generated
├── Dockerfile                       # golang:1.24-alpine → alpine:3.20
├── docker-compose.yml / docker-compose.local.yml
├── seed.sql                         # 4 exchange rates + 4 interest rates + 5 branches
├── Makefile, sonar-project.properties
```

## Database Schema (saving) — PostgreSQL port 5434

```sql
CREATE TABLE IF NOT EXISTS exchange_rate (
    id           UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    country      VARCHAR(100) NOT NULL,
    currency     VARCHAR(10)  NOT NULL,
    country_code VARCHAR(10)  NOT NULL,
    buy          NUMERIC(10,3) NOT NULL,
    sell         NUMERIC(10,3) NOT NULL
);

CREATE TABLE IF NOT EXISTS interest_rate (
    id      UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    kind    VARCHAR(50)  NOT NULL,
    deposit VARCHAR(10)  NOT NULL,
    rate    NUMERIC(5,2) NOT NULL
);

CREATE TABLE IF NOT EXISTS branch (
    id        UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name      VARCHAR(200) NOT NULL,
    distance  VARCHAR(50)  NOT NULL,
    latitude  NUMERIC(10,6) NOT NULL,
    longitude NUMERIC(10,6) NOT NULL
);
```

## API Endpoints (3 REST + 3 gRPC) — all public, no auth

```
GET /api/exchange-rates    → GetExchangeRates (all)
GET /api/interest-rates    → GetInterestRates (all)
GET /api/branches?q=       → GetBranches (ILIKE '%query%')
GET /health                → Health check
```

## Key Patterns

- Branch search: `ILIKE '%' || $1 || '%'` (case-insensitive partial match)
- Response format: raw JSON arrays (no wrapper envelope)
- All endpoints are public — no authentication required
- Swagger UI at `/swagger/`
- CLI: `urfave/cli` with commands grpc-server, gw-server, grpc-gw-server
